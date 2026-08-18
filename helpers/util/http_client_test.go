package util

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewHTTPClient(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		method      string
		token       []string
		expectError bool
	}{
		{"valid params", "http://test.com", "GET", []string{"token"}, false},
		{"empty url", "", "GET", nil, true},
		{"empty method", "http://test.com", "", nil, true},
		{"no token", "http://test.com", "GET", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewHTTPClient(tt.url, tt.method, tt.token...)
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if !tt.expectError && client == nil {
				t.Error("Expected client, got nil")
			}
		})
	}
}

func TestHTTPClient_MakeRequest(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, "GET", "test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.MakeRequest(context.Background(), nil)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTPClient_MakeRequest_WithBody(t *testing.T) {
	// Mock server that echoes the request body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, "POST", "test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	requestBody := map[string]string{"key": "value"}
	resp, err := client.MakeRequest(context.Background(), requestBody)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify the body was sent correctly
	var responseBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if responseBody["key"] != "value" {
		t.Errorf("Expected key=value, got %v", responseBody)
	}
}

func TestHTTPClient_MakeRequest_DifferentMethods(t *testing.T) {
	tests := []struct {
		method         string
		expectedStatus []int
	}{
		{"GET", []int{http.StatusOK}},
		{"POST", []int{http.StatusOK, http.StatusCreated, http.StatusNoContent}},
		{"PUT", []int{http.StatusOK}},
		{"DELETE", []int{http.StatusOK, http.StatusCreated, http.StatusNoContent, http.StatusAccepted}},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.method {
					t.Errorf("Expected method %s, got %s", tt.method, r.Method)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client, err := NewHTTPClient(server.URL, tt.method, "test-token")
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			resp, err := client.MakeRequest(context.Background(), nil)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
		})
	}
}

func TestHTTPClient_MakeRequest_Headers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Authorization: Bearer test-token, got %s", auth)
		}

		// Check Content-Type header
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", contentType)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, "POST", "test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.MakeRequest(context.Background(), map[string]string{"test": "data"})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
}

func TestHTTPClient_MakeRequest_NoToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that no Authorization header is present
		auth := r.Header.Get("Authorization")
		if auth != "" {
			t.Errorf("Expected no Authorization header, got %s", auth)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, "GET")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.MakeRequest(context.Background(), nil)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
}

func TestHTTPClient_MakeRequest_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, "POST", "test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Try to send something that can't be marshaled to JSON
	invalidBody := make(chan int)
	_, err = client.MakeRequest(context.Background(), invalidBody)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestHTTPClient_MakeRequest_ErrorStatusCodes(t *testing.T) {
	tests := []struct {
		method      string
		statusCode  int
		expectError bool
	}{
		{"GET", http.StatusOK, false},
		{"GET", http.StatusBadRequest, true},
		{"GET", http.StatusInternalServerError, true},
		{"POST", http.StatusCreated, false},
		{"POST", http.StatusBadRequest, true},
		{"DELETE", http.StatusAccepted, false},
		{"DELETE", http.StatusNotFound, true},
	}

	for _, tt := range tests {
		t.Run(tt.method+"_"+http.StatusText(tt.statusCode), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"error": "test error"}`))
			}))
			defer server.Close()

			client, err := NewHTTPClient(server.URL, tt.method, "test-token")
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			resp, err := client.MakeRequest(context.Background(), nil)
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if resp != nil {
				resp.Body.Close()
			}
		})
	}
}

func TestHTTPClient_MakeRequest_UnexpectedStatusError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Ref-Number", "ref-1")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":30001}}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewHTTPClient(server.URL, http.MethodPost)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.MakeRequest(context.Background(), nil)
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected error")
	}
	var statusErr *UnexpectedStatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected *UnexpectedStatusError, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("StatusCode = %d, want 400", statusErr.StatusCode)
	}
	if !strings.Contains(statusErr.Body, "30001") {
		t.Fatalf("Body = %q, want 30001", statusErr.Body)
	}
	msg := err.Error()
	if !strings.Contains(msg, "unexpected status code 400") || !strings.Contains(msg, "ref-1") {
		t.Fatalf("Error() = %q, want status code 400 and X-Ref-Number", msg)
	}
}

func TestHTTPClient_handleErrorResponse(t *testing.T) {
	client := &HTTPClient{}

	tests := []struct {
		name          string
		statusCode    int
		expectedCodes []int
		expectError   bool
	}{
		{"success", http.StatusOK, []int{http.StatusOK}, false},
		{"success multiple", http.StatusCreated, []int{http.StatusOK, http.StatusCreated}, false},
		{"error", http.StatusBadRequest, []int{http.StatusOK}, true},
		{"error not in list", http.StatusInternalServerError, []int{http.StatusOK, http.StatusCreated}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock response
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader(`{"error": "test"}`)),
			}

			err := client.handleErrorResponse(resp, tt.expectedCodes)
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestHTTPClient_MakeRequest_NetworkError(t *testing.T) {
	// Use an invalid URL to simulate network error
	client, err := NewHTTPClient("http://invalid-url", "GET", "test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.MakeRequest(context.Background(), nil)
	if err == nil {
		t.Error("Expected network error, got nil")
	}
}

func TestHTTPClient_MakeRequest_ContextCancellation(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This won't be reached due to context cancellation
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, "GET", "test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = client.MakeRequest(ctx, nil)
	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestHTTPClient_MakeRequest_CustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customHeader := r.Header.Get("X-Custom-Header")
		if customHeader != "custom-value" {
			t.Errorf("Expected X-Custom-Header: custom-value, got %s", customHeader)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, "GET", "test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Add custom header
	client.Headers["X-Custom-Header"] = "custom-value"

	resp, err := client.MakeRequest(context.Background(), nil)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
}
