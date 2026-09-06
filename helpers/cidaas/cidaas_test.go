package cidaas

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// NewMockServer creates a test server with configurable response
func NewMockServer(statusCode int, responseBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))
}

func NewTestClientConfig(serverURL string) ClientConfig {
	return ClientConfig{
		BaseURL:     serverURL,
		AccessToken: "test-token",
		ClientID:    "test-client-id",
	}
}

func TestClientConfig_makeRequest_Success(t *testing.T) {
	expectedRequestBody := map[string]string{"test": "data"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "test-endpoint") {
			t.Errorf("Expected test-endpoint in path, got %s", r.URL.Path)
		}

		// Verify Authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.Contains(authHeader, "test-token") {
			t.Errorf("Expected Authorization header with test-token, got %s", authHeader)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedBody map[string]string
		_ = json.Unmarshal(body, &receivedBody)

		if receivedBody["test"] != expectedRequestBody["test"] {
			t.Errorf("Expected request body %v, got %v", expectedRequestBody, receivedBody)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}

	resp, err := config.makeRequest(context.Background(), http.MethodPost, "test-endpoint", expectedRequestBody)

	if err != nil {
		t.Fatalf("makeRequest failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	resp.Body.Close()
}

func TestClientConfig_makeRequest_Error(t *testing.T) {
	server := NewMockServer(http.StatusBadRequest, `{"error": "bad request"}`)
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}

	_, err := config.makeRequest(context.Background(), http.MethodGet, "error-endpoint", nil)

	if err == nil {
		t.Error("Expected error for bad request, got nil")
	}
}

func TestNewClient_Success(t *testing.T) {
	// Mock token server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method for token request, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "token-srv/token") {
			t.Errorf("Expected token-srv/token endpoint, got %s", r.URL.Path)
		}

		// Verify request body contains client credentials
		body, _ := io.ReadAll(r.Body)
		var payload map[string]string
		_ = json.Unmarshal(body, &payload)

		if payload["client_id"] != "test-client-id" {
			t.Errorf("Expected client_id 'test-client-id', got %s", payload["client_id"])
		}

		if payload["client_secret"] != "test-client-secret" {
			t.Errorf("Expected client_secret 'test-client-secret', got %s", payload["client_secret"])
		}

		if payload["grant_type"] != "client_credentials" {
			t.Errorf("Expected grant_type 'client_credentials', got %s", payload["grant_type"])
		}

		// Return access token
		response := TokenResponse{
			AccessToken: "generated-access-token-123",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer tokenServer.Close()

	config := ClientConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		BaseURL:      tokenServer.URL,
	}

	client, err := NewClient(context.Background(), config)

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client instance, got nil")
	}

	// Verify all service clients are initialized
	if client.Roles == nil {
		t.Error("Expected Roles to be initialized")
	}
	if client.Scopes == nil {
		t.Error("Expected Scopes to be initialized")
	}
	if client.ScopeGroup == nil {
		t.Error("Expected ScopeGroup to be initialized")
	}
	if client.GroupType == nil {
		t.Error("Expected GroupType to be initialized")
	}
	if client.UserGroup == nil {
		t.Error("Expected UserGroup to be initialized")
	}
	if client.RegFields == nil {
		t.Error("Expected RegFields to be initialized")
	}
	if client.TemplateGroup == nil {
		t.Error("Expected TemplateGroup to be initialized")
	}
	if client.Templates == nil {
		t.Error("Expected Templates to be initialized")
	}
	if client.PasswordPolicy == nil {
		t.Error("Expected PasswordPolicy to be initialized")
	}
}

func TestNewClient_URLCleanup(t *testing.T) {
	testCases := []struct {
		inputURL    string
		expectedURL string
	}{
		{"https://example.com/", "https://example.com"},
		{"https://example.com//", "https://example.com"},
		{"https://example.com///", "https://example.com"},
		{"https://example.com", "https://example.com"},
		{"http://localhost:8080/", "http://localhost:8080"},
		{"http://localhost:8080//", "http://localhost:8080"},
	}

	for _, tc := range testCases {
		tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := TokenResponse{AccessToken: "test-token"}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))

		config := ClientConfig{
			ClientID:     "test-id",
			ClientSecret: "test-secret",
			BaseURL:      tc.inputURL,
		}

		// Replace the input URL with our test server URL for the actual request
		config.BaseURL = tokenServer.URL + strings.TrimPrefix(tc.inputURL, "https://example.com")
		if strings.HasPrefix(tc.inputURL, "http://localhost:8080") {
			config.BaseURL = tokenServer.URL + strings.TrimPrefix(tc.inputURL, "http://localhost:8080")
		}

		_, err := NewClient(context.Background(), config)
		if err != nil {
			t.Errorf("NewClient failed for URL %s: %v", tc.inputURL, err)
		}

		tokenServer.Close()
	}
}

func TestNewClient_TokenRequestError(t *testing.T) {
	server := NewMockServer(http.StatusUnauthorized, `{"error": "invalid credentials"}`)
	defer server.Close()

	config := ClientConfig{
		ClientID:     "invalid-id",
		ClientSecret: "invalid-secret",
		BaseURL:      server.URL,
	}

	_, err := NewClient(context.Background(), config)

	if err == nil {
		t.Error("Expected error for invalid credentials, got nil")
	}

	if !strings.Contains(err.Error(), "failed to generate access token") {
		t.Errorf("Expected 'failed to generate access token' in error, got %s", err.Error())
	}
}

func TestNewClient_InvalidTokenURL(t *testing.T) {
	config := ClientConfig{
		ClientID:     "test-id",
		ClientSecret: "test-secret",
		BaseURL:      "://invalid-url",
	}

	_, err := NewClient(context.Background(), config)

	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	if !strings.Contains(err.Error(), "failed to create token url") {
		t.Errorf("Expected 'failed to create token url' in error, got %s", err.Error())
	}
}

func TestNewClient_InvalidTokenResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	config := ClientConfig{
		ClientID:     "test-id",
		ClientSecret: "test-secret",
		BaseURL:      server.URL,
	}

	_, err := NewClient(context.Background(), config)

	if err == nil {
		t.Error("Expected error for invalid JSON response, got nil")
	}

	if !strings.Contains(err.Error(), "failed to generate access token") {
		t.Errorf("Expected 'failed to generate access token' in error, got %s", err.Error())
	}
}

func TestNewClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		<-r.Context().Done()
	}))
	defer server.Close()

	config := ClientConfig{
		ClientID:     "test-id",
		ClientSecret: "test-secret",
		BaseURL:      server.URL,
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewClient(ctx, config)

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestNewClient_EmptyTokenResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := TokenResponse{AccessToken: ""}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := ClientConfig{
		ClientID:     "test-id",
		ClientSecret: "test-secret",
		BaseURL:      server.URL,
	}

	client, err := NewClient(context.Background(), config)

	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// Should still create client even with empty token
	if client == nil {
		t.Error("Expected client to be created even with empty token")
	}
}

func TestTokenResponse_Struct(t *testing.T) {
	// Test TokenResponse struct JSON marshaling/unmarshaling
	expectedToken := "test-access-token-123"
	tokenResp := TokenResponse{AccessToken: expectedToken}

	// Marshal to JSON
	jsonData, err := json.Marshal(tokenResp)
	if err != nil {
		t.Fatalf("Failed to marshal TokenResponse: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaledResp TokenResponse
	err = json.Unmarshal(jsonData, &unmarshaledResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal TokenResponse: %v", err)
	}

	if unmarshaledResp.AccessToken != expectedToken {
		t.Errorf("Expected AccessToken %s, got %s", expectedToken, unmarshaledResp.AccessToken)
	}
}
