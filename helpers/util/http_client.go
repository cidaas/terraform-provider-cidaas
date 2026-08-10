package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPClient provides a configurable HTTP client with authentication and error handling.
// It supports common HTTP methods and automatic JSON marshaling.
type HTTPClient struct {
	Token      string
	HTTPMethod string
	URL        string
	Headers    map[string]string
}

type HTTPClientInterface interface {
	MakeRequest(body interface{}) (*http.Response, error)
}

// NewHTTPClient creates a new HTTP client with the specified URL and method.
// Optional token parameter enables Bearer authentication.
func NewHTTPClient(url, method string, token ...string) (*HTTPClient, error) {
	if url == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}
	if method == "" {
		return nil, fmt.Errorf("HTTP method cannot be empty")
	}

	var tokenValue string
	if len(token) > 0 {
		tokenValue = token[0]
	}
	return &HTTPClient{
		URL:        url,
		HTTPMethod: method,
		Token:      tokenValue,
		Headers:    make(map[string]string),
	}, nil
}

// MakeRequest executes an HTTP request with the configured method, URL, and headers.
// It automatically marshals the request body to JSON and handles authentication if a token is set.
func (h *HTTPClient) MakeRequest(ctx context.Context, requestBody interface{}) (*http.Response, error) {
	var reqBodyByte io.Reader
	if requestBody == nil {
		reqBodyByte = nil
	} else {
		bodyByte, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON body, %w", err)
		}
		reqBodyByte = bytes.NewBuffer(bodyByte)
	}

	client := http.DefaultClient
	req, err := http.NewRequestWithContext(ctx, h.HTTPMethod, h.URL, reqBodyByte)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request, %w", err)
	}
	if h.Token != "" {
		req.Header.Add("Authorization", "Bearer "+h.Token)
	}
	for k, v := range h.Headers {
		req.Header.Add(k, v)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Add("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return resp, fmt.Errorf("request failed, %w", err)
	}

	var expectedCodes []int
	switch h.HTTPMethod {
	case http.MethodGet, http.MethodPut:
		expectedCodes = []int{http.StatusOK}
	case http.MethodPost:
		expectedCodes = []int{http.StatusOK, http.StatusCreated, http.StatusNoContent}
	case http.MethodPatch:
		expectedCodes = []int{http.StatusOK, http.StatusNoContent}
	case http.MethodDelete:
		expectedCodes = []int{http.StatusOK, http.StatusCreated, http.StatusNoContent, http.StatusAccepted}
	default:
		expectedCodes = []int{http.StatusOK} // Default fallback
	}

	if err := h.handleErrorResponse(resp, expectedCodes); err != nil {
		return resp, err
	}
	return resp, nil
}

// MakeRequestReadBody executes the request, reads the full response body, and returns the HTTP status code
// and response headers (for correlation, e.g. X-Ref-Number). Only transport/body-read errors are returned as err;
// non-2xx status codes are not treated as errors.
func (h *HTTPClient) MakeRequestReadBody(ctx context.Context, requestBody interface{}) (int, []byte, http.Header, error) {
	var reqBodyByte io.Reader
	if requestBody == nil {
		reqBodyByte = nil
	} else {
		bodyByte, marshalErr := json.Marshal(requestBody)
		if marshalErr != nil {
			return 0, nil, nil, fmt.Errorf("failed to marshal JSON body, %w", marshalErr)
		}
		reqBodyByte = bytes.NewBuffer(bodyByte)
	}
	req, err := http.NewRequestWithContext(ctx, h.HTTPMethod, h.URL, reqBodyByte)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to create HTTP request, %w", err)
	}
	if h.Token != "" {
		req.Header.Add("Authorization", "Bearer "+h.Token)
	}
	for k, v := range h.Headers {
		req.Header.Add(k, v)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Add("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("request failed, %w", err)
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return resp.StatusCode, nil, resp.Header, fmt.Errorf("failed to read response body, %w", readErr)
	}
	return resp.StatusCode, body, resp.Header, nil
}

// handleErrorResponse validates the HTTP response status code against expected codes.
// Returns an error if the status code is not in the expected list, including response body for debugging.
func (h *HTTPClient) handleErrorResponse(resp *http.Response, expectedCodes []int) error {
	for _, code := range expectedCodes {
		if resp.StatusCode == code {
			return nil
		}
	}

	bodyStr, bodyErr := ResponseToString(resp)
	if bodyErr != nil {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("%w: failed to read response body: %v%s", ErrResourceNotFound, bodyErr, xRefNumberSuffix(resp))
		}
		return fmt.Errorf("unexpected status code %d, failed to read response body: %w%s", resp.StatusCode, bodyErr, xRefNumberSuffix(resp))
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%w: %s%s", ErrResourceNotFound, bodyStr, xRefNumberSuffix(resp))
	}
	return fmt.Errorf("unexpected status code %d, response body: %s%s", resp.StatusCode, bodyStr, xRefNumberSuffix(resp))
}

// HTTP header name used by Cidaas for request correlation (support / tracing).
const httpHeaderXRefNumber = "X-Ref-Number"

// xRefNumberSuffix returns a fragment to append to HTTP error messages when the response includes X-Ref-Number.
func xRefNumberSuffix(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	return XRefNumberSuffixFromHeader(resp.Header)
}

// XRefNumberSuffixFromHeader returns a fragment like ", X-Ref-Number: <value>" when the header is set.
// Use when building errors from MakeRequestReadBody (which does not use handleErrorResponse).
func XRefNumberSuffixFromHeader(h http.Header) string {
	if h == nil {
		return ""
	}
	ref := h.Get(httpHeaderXRefNumber)
	if ref == "" {
		return ""
	}
	return fmt.Sprintf(", %s: %s", httpHeaderXRefNumber, ref)
}
