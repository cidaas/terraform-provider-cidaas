package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPClient is a small Bearer JSON client for cidaas APIs.
type HTTPClient struct {
	URL        string
	Method     string
	Token      string
	Headers    map[string]string
	HTTPClient *http.Client
}

func NewHTTPClient(url, method, token string) *HTTPClient {
	return &HTTPClient{
		URL:        url,
		Method:     method,
		Token:      token,
		Headers:    map[string]string{},
		HTTPClient: http.DefaultClient,
	}
}

func (h *HTTPClient) DoJSON(ctx context.Context, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, h.Method, h.URL, reader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if h.Token != "" {
		req.Header.Set("Authorization", "Bearer "+h.Token)
	}
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}
	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

// DecodeJSON unmarshals a JSON response body into dest. Caller must close resp.Body.
func DecodeJSON(resp *http.Response, dest any) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("decode body: %w", err)
	}
	return nil
}

func ExpectStatus(resp *http.Response, codes ...int) error {
	for _, c := range codes {
		if resp.StatusCode == c {
			return nil
		}
	}
	body, _ := io.ReadAll(resp.Body)
	ref := resp.Header.Get("X-Ref-Number")
	msg := fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	if ref != "" {
		msg += ", X-Ref-Number: " + ref
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%w: %s", ErrNotFound, msg)
	}
	return fmt.Errorf("%s", msg)
}

// ErrNotFound indicates a 404 from the API.
var ErrNotFound = fmt.Errorf("resource not found")
