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

func TestRole_UpsertRole(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "data": {"role": "test-role"}}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	role := NewRole(config)

	roleModel := RoleModel{
		Role:        "test-role",
		Name:        "Test Role",
		Description: "Test Description",
	}

	resp, err := role.UpsertRole(context.Background(), roleModel)
	if err != nil {
		t.Fatalf("UpsertRole failed: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestNewRole(t *testing.T) {
	config := ClientConfig{
		BaseURL:     "http://test.com",
		AccessToken: "test-token",
	}

	role := NewRole(config)

	if role == nil {
		t.Fatal("Expected role instance, got nil")
	}

	if role.BaseURL != config.BaseURL {
		t.Errorf("Expected BaseURL %s, got %s", config.BaseURL, role.BaseURL)
	}

	if role.AccessToken != config.AccessToken {
		t.Errorf("Expected AccessToken %s, got %s", config.AccessToken, role.AccessToken)
	}
}

func TestRole_UpsertRole_Success(t *testing.T) {
	expectedRole := RoleModel{
		Name:        "Test Role",
		Description: "Test Description",
		Role:        "test-role",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "roles-srv/role") {
			t.Errorf("Expected roles-srv/role endpoint, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedRole RoleModel
		json.Unmarshal(body, &receivedRole)

		if receivedRole.Role != expectedRole.Role {
			t.Errorf("Expected role %s, got %s", expectedRole.Role, receivedRole.Role)
		}

		// Send success response
		response := RoleResponse{
			Success: true,
			Status:  200,
			Data:    expectedRole,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	role := NewRole(config)

	result, err := role.UpsertRole(context.Background(), expectedRole)

	if err != nil {
		t.Fatalf("UpsertRole failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.Role != expectedRole.Role {
		t.Errorf("Expected role %s, got %s", expectedRole.Role, result.Data.Role)
	}
}

func TestRole_UpsertRole_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	role := NewRole(config)

	roleModel := RoleModel{Role: "test-role"}
	_, err := role.UpsertRole(context.Background(), roleModel)

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to upsert role") {
		t.Errorf("Expected 'failed to upsert role' in error, got %s", err.Error())
	}
}

func TestRole_GetRole_Success(t *testing.T) {
	expectedRole := RoleModel{
		Name:        "Test Role",
		Description: "Test Description",
		Role:        "test-role",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		// Verify query parameter
		roleParam := r.URL.Query().Get("role")
		if roleParam != "test-role" {
			t.Errorf("Expected role parameter 'test-role', got %s", roleParam)
		}

		// Send success response
		response := RoleResponse{
			Success: true,
			Status:  200,
			Data:    expectedRole,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	role := NewRole(config)

	result, err := role.GetRole(context.Background(), "test-role")

	if err != nil {
		t.Fatalf("GetRole failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.Role != expectedRole.Role {
		t.Errorf("Expected role %s, got %s", expectedRole.Role, result.Data.Role)
	}
}

func TestRole_GetRole_EmptyRole(t *testing.T) {
	config := ClientConfig{}
	role := NewRole(config)

	_, err := role.GetRole(context.Background(), "")

	if err == nil {
		t.Error("Expected error for empty role, got nil")
	}

	if err.Error() != "role cannot be empty" {
		t.Errorf("Expected 'role cannot be empty', got %s", err.Error())
	}
}

func TestRole_GetRole_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "role not found"}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	role := NewRole(config)

	_, err := role.GetRole(context.Background(), "nonexistent-role")

	if err == nil {
		t.Error("Expected error for not found role, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get role") {
		t.Errorf("Expected 'failed to get role' in error, got %s", err.Error())
	}
}

func TestRole_DeleteRole_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		// Verify query parameter
		roleParam := r.URL.Query().Get("role")
		if roleParam != "test-role" {
			t.Errorf("Expected role parameter 'test-role', got %s", roleParam)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	role := NewRole(config)

	err := role.DeleteRole(context.Background(), "test-role")

	if err != nil {
		t.Fatalf("DeleteRole failed: %v", err)
	}
}

func TestRole_DeleteRole_EmptyRole(t *testing.T) {
	config := ClientConfig{}
	role := NewRole(config)

	err := role.DeleteRole(context.Background(), "")

	if err == nil {
		t.Error("Expected error for empty role, got nil")
	}

	if err.Error() != "role cannot be empty" {
		t.Errorf("Expected 'role cannot be empty', got %s", err.Error())
	}
}

func TestRole_DeleteRole_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	role := NewRole(config)

	err := role.DeleteRole(context.Background(), "test-role")

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to delete role") {
		t.Errorf("Expected 'failed to delete role' in error, got %s", err.Error())
	}
}

func TestRole_GetAll_Success(t *testing.T) {
	expectedRoles := []RoleModel{
		{Name: "Role 1", Description: "Description 1", Role: "role1"},
		{Name: "Role 2", Description: "Description 2", Role: "role2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "groups-srv/graph/roles") {
			t.Errorf("Expected groups-srv/graph/roles endpoint, got %s", r.URL.Path)
		}

		// Verify empty body
		body, _ := io.ReadAll(r.Body)
		if string(body) != "{}" {
			t.Errorf("Expected empty object body, got %s", string(body))
		}

		// Send success response
		response := AllRoleResponse{
			Success: true,
			Status:  200,
			Data:    expectedRoles,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	role := NewRole(config)

	result, err := role.GetAll(context.Background())

	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(result))
	}

	if result[0].Role != "role1" {
		t.Errorf("Expected first role 'role1', got %s", result[0].Role)
	}

	if result[1].Role != "role2" {
		t.Errorf("Expected second role 'role2', got %s", result[1].Role)
	}
}

func TestRole_GetAll_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := AllRoleResponse{
			Success: true,
			Status:  200,
			Data:    []RoleModel{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	role := NewRole(config)

	result, err := role.GetAll(context.Background())

	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 roles, got %d", len(result))
	}
}

func TestRole_GetAll_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	role := NewRole(config)

	_, err := role.GetAll(context.Background())

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get all roles") {
		t.Errorf("Expected 'failed to get all roles' in error, got %s", err.Error())
	}
}

func TestRole_GetAll_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	role := NewRole(config)

	_, err := role.GetAll(context.Background())

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// Test context cancellation
func TestRole_UpsertRole_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This shouldn't be reached due to context cancellation
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	role := NewRole(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	roleModel := RoleModel{Role: "test-role"}
	_, err := role.UpsertRole(ctx, roleModel)

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}
