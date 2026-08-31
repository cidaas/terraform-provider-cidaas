// helpers/cidaas/scope_test.go
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

func TestNewScope(t *testing.T) {
	config := NewTestClientConfig("http://test.com")

	scope := NewScope(config)

	if scope == nil {
		t.Fatal("Expected scope instance, got nil")
	}

	if scope.BaseURL != config.BaseURL {
		t.Errorf("Expected BaseURL %s, got %s", config.BaseURL, scope.BaseURL)
	}

	if scope.AccessToken != config.AccessToken {
		t.Errorf("Expected AccessToken %s, got %s", config.AccessToken, scope.AccessToken)
	}
}

func TestScope_Upsert_Success(t *testing.T) {
	expectedScope := ScopeModel{
		ID:                  "scope-123",
		ScopeKey:            "read:profile",
		SecurityLevel:       "PUBLIC",
		RequiredUserConsent: true,
		GroupName:           []string{"user", "admin"},
		ScopeOwner:          "system",
		LocaleWiseDescription: []ScopeLocalDescription{
			{
				Locale:      "en-US",
				Language:    "en",
				Title:       "Read Profile",
				Description: "Allows reading user profile information",
			},
			{
				Locale:      "de-DE",
				Language:    "de",
				Title:       "Profil lesen",
				Description: "Ermöglicht das Lesen von Benutzerprofilinformationen",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "scopes-srv/scope") {
			t.Errorf("Expected scopes-srv/scope endpoint, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedScope ScopeModel
		json.Unmarshal(body, &receivedScope)

		if receivedScope.ScopeKey != expectedScope.ScopeKey {
			t.Errorf("Expected ScopeKey %s, got %s", expectedScope.ScopeKey, receivedScope.ScopeKey)
		}

		if receivedScope.SecurityLevel != expectedScope.SecurityLevel {
			t.Errorf("Expected SecurityLevel %s, got %s", expectedScope.SecurityLevel, receivedScope.SecurityLevel)
		}

		if receivedScope.RequiredUserConsent != expectedScope.RequiredUserConsent {
			t.Errorf("Expected RequiredUserConsent %t, got %t", expectedScope.RequiredUserConsent, receivedScope.RequiredUserConsent)
		}

		if len(receivedScope.LocaleWiseDescription) != len(expectedScope.LocaleWiseDescription) {
			t.Errorf("Expected %d locale descriptions, got %d", len(expectedScope.LocaleWiseDescription), len(receivedScope.LocaleWiseDescription))
		}

		// Send success response
		response := ScopeResponse{
			Success: true,
			Status:  200,
			Data:    expectedScope,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	result, err := scope.Upsert(context.Background(), expectedScope)

	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.ScopeKey != expectedScope.ScopeKey {
		t.Errorf("Expected ScopeKey %s, got %s", expectedScope.ScopeKey, result.Data.ScopeKey)
	}

	if result.Data.SecurityLevel != expectedScope.SecurityLevel {
		t.Errorf("Expected SecurityLevel %s, got %s", expectedScope.SecurityLevel, result.Data.SecurityLevel)
	}

	if len(result.Data.LocaleWiseDescription) != len(expectedScope.LocaleWiseDescription) {
		t.Errorf("Expected %d locale descriptions, got %d", len(expectedScope.LocaleWiseDescription), len(result.Data.LocaleWiseDescription))
	}
}

func TestScope_Upsert_WithMinimalData(t *testing.T) {
	minimalScope := ScopeModel{
		ScopeKey:            "basic:read",
		RequiredUserConsent: false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedScope ScopeModel
		json.Unmarshal(body, &receivedScope)

		if receivedScope.ScopeKey != minimalScope.ScopeKey {
			t.Errorf("Expected ScopeKey %s, got %s", minimalScope.ScopeKey, receivedScope.ScopeKey)
		}

		if receivedScope.RequiredUserConsent != minimalScope.RequiredUserConsent {
			t.Errorf("Expected RequiredUserConsent %t, got %t", minimalScope.RequiredUserConsent, receivedScope.RequiredUserConsent)
		}

		response := ScopeResponse{
			Success: true,
			Status:  200,
			Data:    minimalScope,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	result, err := scope.Upsert(context.Background(), minimalScope)

	if err != nil {
		t.Fatalf("Upsert with minimal data failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.ScopeKey != minimalScope.ScopeKey {
		t.Errorf("Expected ScopeKey %s, got %s", minimalScope.ScopeKey, result.Data.ScopeKey)
	}
}

func TestScope_Upsert_ServerError(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "internal server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	scopeModel := ScopeModel{ScopeKey: "test:scope"}
	_, err := scope.Upsert(context.Background(), scopeModel)

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestScope_Upsert_InvalidJSON(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{invalid json}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	scopeModel := ScopeModel{ScopeKey: "test:scope"}
	_, err := scope.Upsert(context.Background(), scopeModel)

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestScope_Upsert_ContextCancellation(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{"success": true}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	scopeModel := ScopeModel{ScopeKey: "test:scope"}
	_, err := scope.Upsert(ctx, scopeModel)

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestScope_Upsert_WithMultipleGroups(t *testing.T) {
	scopeWithGroups := ScopeModel{
		ScopeKey:            "admin:manage",
		SecurityLevel:       "CONFIDENTIAL",
		RequiredUserConsent: true,
		GroupName:           []string{"admin", "super-admin", "moderator"},
		ScopeOwner:          "admin-service",
		LocaleWiseDescription: []ScopeLocalDescription{
			{
				Locale:      "en-US",
				Language:    "en",
				Title:       "Admin Management",
				Description: "Full administrative access",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedScope ScopeModel
		json.Unmarshal(body, &receivedScope)

		if len(receivedScope.GroupName) != 3 {
			t.Errorf("Expected 3 groups, got %d", len(receivedScope.GroupName))
		}

		expectedGroups := []string{"admin", "super-admin", "moderator"}
		for i, group := range receivedScope.GroupName {
			if group != expectedGroups[i] {
				t.Errorf("Expected group %s at index %d, got %s", expectedGroups[i], i, group)
			}
		}

		response := ScopeResponse{
			Success: true,
			Status:  200,
			Data:    scopeWithGroups,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	result, err := scope.Upsert(context.Background(), scopeWithGroups)

	if err != nil {
		t.Fatalf("Upsert with multiple groups failed: %v", err)
	}

	if len(result.Data.GroupName) != 3 {
		t.Errorf("Expected 3 groups in response, got %d", len(result.Data.GroupName))
	}
}

func TestScope_Upsert_WithErrorResponse(t *testing.T) {
	errorResponse := `{
        "success": false,
        "status": 400,
        "error": "Invalid scope key format"
    }`

	server := NewMockServer(http.StatusBadRequest, errorResponse)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	scopeModel := ScopeModel{ScopeKey: "invalid scope key"}
	_, err := scope.Upsert(context.Background(), scopeModel)

	if err == nil {
		t.Error("Expected error for bad request, got nil")
	}
}

func TestScope_Get_Success(t *testing.T) {
	expectedScope := ScopeModel{
		ID:                  "scope-123",
		ScopeKey:            "read:profile",
		SecurityLevel:       "PUBLIC",
		RequiredUserConsent: true,
		GroupName:           []string{"user", "admin"},
		ScopeOwner:          "system",
		LocaleWiseDescription: []ScopeLocalDescription{
			{
				Locale:      "en-US",
				Language:    "en",
				Title:       "Read Profile",
				Description: "Allows reading user profile information",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "scopes-srv/scope") {
			t.Errorf("Expected scopes-srv/scope endpoint, got %s", r.URL.Path)
		}

		// Verify scopekey query param: same string as passed to Get (casing preserved, not lowercased).
		scopeKeyParam := r.URL.Query().Get("scopekey")
		if scopeKeyParam != "READ:PROFILE" {
			t.Errorf("Expected scopekey parameter 'READ:PROFILE', got %s", scopeKeyParam)
		}

		// Send success response
		response := ScopeResponse{
			Success: true,
			Status:  200,
			Data:    expectedScope,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	result, err := scope.Get(context.Background(), "READ:PROFILE")

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.ScopeKey != expectedScope.ScopeKey {
		t.Errorf("Expected ScopeKey %s, got %s", expectedScope.ScopeKey, result.Data.ScopeKey)
	}

	if result.Data.SecurityLevel != expectedScope.SecurityLevel {
		t.Errorf("Expected SecurityLevel %s, got %s", expectedScope.SecurityLevel, result.Data.SecurityLevel)
	}
}

func TestScope_Get_PreservesScopeKeyCasing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scopeKeyParam := r.URL.Query().Get("scopekey")
		if scopeKeyParam != "ADMIN:Write" {
			t.Errorf("Expected scopekey forwarded unchanged as 'ADMIN:Write', got %s", scopeKeyParam)
		}
		response := ScopeResponse{
			Success: true,
			Status:  200,
			Data: ScopeModel{
				ScopeKey: "admin:write",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	_, err := scope.Get(context.Background(), "ADMIN:Write")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
}

func TestScope_Get_NotFound(t *testing.T) {
	server := NewMockServer(http.StatusNotFound, `{"error": "scope not found"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	_, err := scope.Get(context.Background(), "nonexistent:scope")

	if err == nil {
		t.Error("Expected error for not found scope, got nil")
	}
}

func TestScope_Get_ServerError(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "internal server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	_, err := scope.Get(context.Background(), "test:scope")

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestScope_Get_InvalidJSON(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{invalid json}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	_, err := scope.Get(context.Background(), "test:scope")

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestScope_Delete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		// Verify path includes the scope key segment as passed to Delete (casing preserved, not lowercased).
		if !strings.Contains(r.URL.Path, "READ:PROFILE") {
			t.Errorf("Expected 'READ:PROFILE' in URL path, got %s", r.URL.Path)
		}

		if !strings.Contains(r.URL.Path, "scopes-srv/scope") {
			t.Errorf("Expected scopes-srv/scope endpoint, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	err := scope.Delete(context.Background(), "READ:PROFILE")

	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestScope_Delete_PreservesScopeKeyCasing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "ADMIN:Delete") {
			t.Errorf("Expected scope key forwarded unchanged in path ('ADMIN:Delete'), got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	err := scope.Delete(context.Background(), "ADMIN:Delete")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestScope_Delete_ServerError(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "internal server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	err := scope.Delete(context.Background(), "test:scope")

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestScope_Delete_NotFound(t *testing.T) {
	server := NewMockServer(http.StatusNotFound, `{"error": "scope not found"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	err := scope.Delete(context.Background(), "nonexistent:scope")

	if err == nil {
		t.Error("Expected error for not found scope, got nil")
	}
}

func TestScope_GetAll_Success(t *testing.T) {
	expectedScopes := []ScopeModel{
		{
			ID:                  "scope-1",
			ScopeKey:            "read:profile",
			SecurityLevel:       "PUBLIC",
			RequiredUserConsent: true,
			GroupName:           []string{"user"},
			LocaleWiseDescription: []ScopeLocalDescription{
				{
					Locale:      "en-US",
					Language:    "en",
					Title:       "Read Profile",
					Description: "Read user profile",
				},
			},
		},
		{
			ID:                  "scope-2",
			ScopeKey:            "write:profile",
			SecurityLevel:       "CONFIDENTIAL",
			RequiredUserConsent: true,
			GroupName:           []string{"admin"},
			LocaleWiseDescription: []ScopeLocalDescription{
				{
					Locale:      "en-US",
					Language:    "en",
					Title:       "Write Profile",
					Description: "Modify user profile",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "scopes-srv/scope/list") {
			t.Errorf("Expected scopes-srv/scope/list endpoint, got %s", r.URL.Path)
		}

		// Send success response
		response := AllScopeResp{
			Success: true,
			Status:  200,
			Data:    expectedScopes,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	result, err := scope.GetAll(context.Background())

	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 scopes, got %d", len(result))
	}

	if result[0].ScopeKey != "read:profile" {
		t.Errorf("Expected first scope key 'read:profile', got %s", result[0].ScopeKey)
	}

	if result[1].ScopeKey != "write:profile" {
		t.Errorf("Expected second scope key 'write:profile', got %s", result[1].ScopeKey)
	}

	// Verify complex nested data
	if len(result[0].LocaleWiseDescription) != 1 {
		t.Errorf("Expected 1 locale description for first scope, got %d", len(result[0].LocaleWiseDescription))
	}

	if result[0].LocaleWiseDescription[0].Title != "Read Profile" {
		t.Errorf("Expected title 'Read Profile', got %s", result[0].LocaleWiseDescription[0].Title)
	}
}

func TestScope_GetAll_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := AllScopeResp{
			Success: true,
			Status:  200,
			Data:    []ScopeModel{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	result, err := scope.GetAll(context.Background())

	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 scopes, got %d", len(result))
	}
}

func TestScope_GetAll_ServerError(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "internal server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	_, err := scope.GetAll(context.Background())

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestScope_GetAll_InvalidJSON(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{invalid json}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	_, err := scope.GetAll(context.Background())

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestScope_ContextCancellation_Get(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{"success": true}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := scope.Get(ctx, "test:scope")

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestScope_ContextCancellation_Delete(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{"success": true}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := scope.Delete(ctx, "test:scope")

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestScope_ContextCancellation_GetAll(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{"success": true}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scope := NewScope(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := scope.GetAll(ctx)

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}
