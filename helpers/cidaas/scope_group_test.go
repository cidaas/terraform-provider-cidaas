// helpers/cidaas/scope_group_test.go
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

func TestNewScopeGroup(t *testing.T) {
	config := NewTestClientConfig("http://test.com")

	scopeGroup := NewScopeGroup(config)

	if scopeGroup == nil {
		t.Fatal("Expected scope group instance, got nil")
	}

	if scopeGroup.BaseURL != config.BaseURL {
		t.Errorf("Expected BaseURL %s, got %s", config.BaseURL, scopeGroup.BaseURL)
	}

	if scopeGroup.AccessToken != config.AccessToken {
		t.Errorf("Expected AccessToken %s, got %s", config.AccessToken, scopeGroup.AccessToken)
	}
}

func TestScopeGroup_Upsert_Success(t *testing.T) {
	expectedScopeGroup := ScopeGroupConfig{
		ID:          "scope-group-123",
		GroupName:   "admin-group",
		Description: "Administrative scope group",
		CreatedTime: "2024-01-15T10:30:00Z",
		UpdatedTime: "2024-01-15T11:45:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "scopes-srv/group") {
			t.Errorf("Expected scopes-srv/group endpoint, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedScopeGroup ScopeGroupConfig
		json.Unmarshal(body, &receivedScopeGroup)

		if receivedScopeGroup.GroupName != expectedScopeGroup.GroupName {
			t.Errorf("Expected GroupName %s, got %s", expectedScopeGroup.GroupName, receivedScopeGroup.GroupName)
		}

		if receivedScopeGroup.Description != expectedScopeGroup.Description {
			t.Errorf("Expected Description %s, got %s", expectedScopeGroup.Description, receivedScopeGroup.Description)
		}

		// Send success response
		response := ScopeGroupResponse{
			Success: true,
			Status:  200,
			Data:    expectedScopeGroup,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	result, err := scopeGroup.Upsert(context.Background(), expectedScopeGroup)

	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.GroupName != expectedScopeGroup.GroupName {
		t.Errorf("Expected GroupName %s, got %s", expectedScopeGroup.GroupName, result.Data.GroupName)
	}

	if result.Data.Description != expectedScopeGroup.Description {
		t.Errorf("Expected Description %s, got %s", expectedScopeGroup.Description, result.Data.Description)
	}

	if result.Data.ID != expectedScopeGroup.ID {
		t.Errorf("Expected ID %s, got %s", expectedScopeGroup.ID, result.Data.ID)
	}
}

func TestScopeGroup_Upsert_WithMinimalData(t *testing.T) {
	minimalScopeGroup := ScopeGroupConfig{
		GroupName:   "basic-group",
		Description: "Basic scope group",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedScopeGroup ScopeGroupConfig
		json.Unmarshal(body, &receivedScopeGroup)

		if receivedScopeGroup.GroupName != minimalScopeGroup.GroupName {
			t.Errorf("Expected GroupName %s, got %s", minimalScopeGroup.GroupName, receivedScopeGroup.GroupName)
		}

		if receivedScopeGroup.Description != minimalScopeGroup.Description {
			t.Errorf("Expected Description %s, got %s", minimalScopeGroup.Description, receivedScopeGroup.Description)
		}

		// Server might add timestamps and ID
		responseData := minimalScopeGroup
		responseData.ID = "generated-id-123"
		responseData.CreatedTime = "2024-01-15T10:30:00Z"
		responseData.UpdatedTime = "2024-01-15T10:30:00Z"

		response := ScopeGroupResponse{
			Success: true,
			Status:  200,
			Data:    responseData,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	result, err := scopeGroup.Upsert(context.Background(), minimalScopeGroup)

	if err != nil {
		t.Fatalf("Upsert with minimal data failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.GroupName != minimalScopeGroup.GroupName {
		t.Errorf("Expected GroupName %s, got %s", minimalScopeGroup.GroupName, result.Data.GroupName)
	}

	// Verify server-generated fields
	if result.Data.ID == "" {
		t.Error("Expected server to generate an ID")
	}

	if result.Data.CreatedTime == "" {
		t.Error("Expected server to set CreatedTime")
	}
}

func TestScopeGroup_Upsert_UpdateExisting(t *testing.T) {
	existingScopeGroup := ScopeGroupConfig{
		ID:          "existing-group-123",
		GroupName:   "updated-group",
		Description: "Updated description",
		CreatedTime: "2024-01-15T10:30:00Z",
		UpdatedTime: "2024-01-15T12:00:00Z", // Updated timestamp
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedScopeGroup ScopeGroupConfig
		json.Unmarshal(body, &receivedScopeGroup)

		// Verify ID is present for update
		if receivedScopeGroup.ID != existingScopeGroup.ID {
			t.Errorf("Expected ID %s, got %s", existingScopeGroup.ID, receivedScopeGroup.ID)
		}

		if receivedScopeGroup.GroupName != existingScopeGroup.GroupName {
			t.Errorf("Expected GroupName %s, got %s", existingScopeGroup.GroupName, receivedScopeGroup.GroupName)
		}

		response := ScopeGroupResponse{
			Success: true,
			Status:  200,
			Data:    existingScopeGroup,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	result, err := scopeGroup.Upsert(context.Background(), existingScopeGroup)

	if err != nil {
		t.Fatalf("Upsert update failed: %v", err)
	}

	if result.Data.ID != existingScopeGroup.ID {
		t.Errorf("Expected ID %s, got %s", existingScopeGroup.ID, result.Data.ID)
	}

	if result.Data.UpdatedTime != existingScopeGroup.UpdatedTime {
		t.Errorf("Expected UpdatedTime %s, got %s", existingScopeGroup.UpdatedTime, result.Data.UpdatedTime)
	}
}

func TestScopeGroup_Upsert_WithTimestamps(t *testing.T) {
	scopeGroupWithTimestamps := ScopeGroupConfig{
		GroupName:   "timestamp-group",
		Description: "Group with timestamps",
		CreatedTime: "2024-01-15T10:30:00Z",
		UpdatedTime: "2024-01-15T11:45:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedScopeGroup ScopeGroupConfig
		json.Unmarshal(body, &receivedScopeGroup)

		if receivedScopeGroup.CreatedTime != scopeGroupWithTimestamps.CreatedTime {
			t.Errorf("Expected CreatedTime %s, got %s", scopeGroupWithTimestamps.CreatedTime, receivedScopeGroup.CreatedTime)
		}

		if receivedScopeGroup.UpdatedTime != scopeGroupWithTimestamps.UpdatedTime {
			t.Errorf("Expected UpdatedTime %s, got %s", scopeGroupWithTimestamps.UpdatedTime, receivedScopeGroup.UpdatedTime)
		}

		response := ScopeGroupResponse{
			Success: true,
			Status:  200,
			Data:    scopeGroupWithTimestamps,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	result, err := scopeGroup.Upsert(context.Background(), scopeGroupWithTimestamps)

	if err != nil {
		t.Fatalf("Upsert with timestamps failed: %v", err)
	}

	if result.Data.CreatedTime != scopeGroupWithTimestamps.CreatedTime {
		t.Errorf("Expected CreatedTime %s, got %s", scopeGroupWithTimestamps.CreatedTime, result.Data.CreatedTime)
	}

	if result.Data.UpdatedTime != scopeGroupWithTimestamps.UpdatedTime {
		t.Errorf("Expected UpdatedTime %s, got %s", scopeGroupWithTimestamps.UpdatedTime, result.Data.UpdatedTime)
	}
}

func TestScopeGroup_Upsert_ServerError(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "internal server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	scopeGroupConfig := ScopeGroupConfig{GroupName: "test-group"}
	_, err := scopeGroup.Upsert(context.Background(), scopeGroupConfig)

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to upsert scope group") {
		t.Errorf("Expected 'failed to upsert scope group' in error, got %s", err.Error())
	}
}

func TestScopeGroup_Upsert_InvalidJSON(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{invalid json}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	scopeGroupConfig := ScopeGroupConfig{GroupName: "test-group"}
	_, err := scopeGroup.Upsert(context.Background(), scopeGroupConfig)

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestScopeGroup_Upsert_ContextCancellation(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{"success": true}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	scopeGroupConfig := ScopeGroupConfig{GroupName: "test-group"}
	_, err := scopeGroup.Upsert(ctx, scopeGroupConfig)

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestScopeGroup_Upsert_EmptyGroupName(t *testing.T) {
	emptyScopeGroup := ScopeGroupConfig{
		GroupName:   "", // Empty group name
		Description: "Group with empty name",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedScopeGroup ScopeGroupConfig
		json.Unmarshal(body, &receivedScopeGroup)

		if receivedScopeGroup.GroupName != "" {
			t.Errorf("Expected empty GroupName, got %s", receivedScopeGroup.GroupName)
		}

		response := ScopeGroupResponse{
			Success: true,
			Status:  200,
			Data:    emptyScopeGroup,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	result, err := scopeGroup.Upsert(context.Background(), emptyScopeGroup)

	if err != nil {
		t.Fatalf("Upsert with empty group name failed: %v", err)
	}

	if result.Data.GroupName != "" {
		t.Errorf("Expected empty GroupName in response, got %s", result.Data.GroupName)
	}
}

func TestScopeGroup_Upsert_LongDescription(t *testing.T) {
	longDescription := strings.Repeat("This is a very long description for the scope group. ", 20)
	scopeGroupWithLongDesc := ScopeGroupConfig{
		GroupName:   "long-desc-group",
		Description: longDescription,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedScopeGroup ScopeGroupConfig
		json.Unmarshal(body, &receivedScopeGroup)

		if len(receivedScopeGroup.Description) != len(longDescription) {
			t.Errorf("Expected description length %d, got %d", len(longDescription), len(receivedScopeGroup.Description))
		}

		response := ScopeGroupResponse{
			Success: true,
			Status:  200,
			Data:    scopeGroupWithLongDesc,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	result, err := scopeGroup.Upsert(context.Background(), scopeGroupWithLongDesc)

	if err != nil {
		t.Fatalf("Upsert with long description failed: %v", err)
	}

	if len(result.Data.Description) != len(longDescription) {
		t.Errorf("Expected description length %d in response, got %d", len(longDescription), len(result.Data.Description))
	}
}

func TestScopeGroup_Upsert_BadRequest(t *testing.T) {
	server := NewMockServer(http.StatusBadRequest, `{"error": "invalid scope group configuration"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	invalidScopeGroup := ScopeGroupConfig{
		GroupName: "invalid-group",
	}

	_, err := scopeGroup.Upsert(context.Background(), invalidScopeGroup)

	if err == nil {
		t.Error("Expected error for bad request, got nil")
	}

	if !strings.Contains(err.Error(), "failed to upsert scope group") {
		t.Errorf("Expected 'failed to upsert scope group' in error, got %s", err.Error())
	}
}

// Add these tests to helpers/cidaas/scope_group_test.go

func TestScopeGroup_Get_Success(t *testing.T) {
	expectedScopeGroup := ScopeGroupConfig{
		ID:          "scope-group-123",
		GroupName:   "admin-group",
		Description: "Administrative scope group",
		CreatedTime: "2024-01-15T10:30:00Z",
		UpdatedTime: "2024-01-15T11:45:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "scopes-srv/group") {
			t.Errorf("Expected scopes-srv/group endpoint, got %s", r.URL.Path)
		}

		// Verify query parameter
		groupNameParam := r.URL.Query().Get("group_name")
		if groupNameParam != "admin-group" {
			t.Errorf("Expected group_name parameter 'admin-group', got %s", groupNameParam)
		}

		// Send success response
		response := ScopeGroupResponse{
			Success: true,
			Status:  200,
			Data:    expectedScopeGroup,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	result, err := scopeGroup.Get(context.Background(), "admin-group")

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.GroupName != expectedScopeGroup.GroupName {
		t.Errorf("Expected GroupName %s, got %s", expectedScopeGroup.GroupName, result.Data.GroupName)
	}

	if result.Data.ID != expectedScopeGroup.ID {
		t.Errorf("Expected ID %s, got %s", expectedScopeGroup.ID, result.Data.ID)
	}

	if result.Data.Description != expectedScopeGroup.Description {
		t.Errorf("Expected Description %s, got %s", expectedScopeGroup.Description, result.Data.Description)
	}
}

func TestScopeGroup_Get_WithSpecialCharacters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameter with special characters
		groupNameParam := r.URL.Query().Get("group_name")
		if groupNameParam != "test-group_123" {
			t.Errorf("Expected group_name parameter 'test-group_123', got %s", groupNameParam)
		}

		response := ScopeGroupResponse{
			Success: true,
			Status:  200,
			Data: ScopeGroupConfig{
				GroupName: "test-group_123",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	_, err := scopeGroup.Get(context.Background(), "test-group_123")

	if err != nil {
		t.Fatalf("Get with special characters failed: %v", err)
	}
}

func TestScopeGroup_Get_EmptyName(t *testing.T) {
	config := NewTestClientConfig("http://test.com")
	scopeGroup := NewScopeGroup(config)

	_, err := scopeGroup.Get(context.Background(), "")

	if err == nil {
		t.Error("Expected error for empty scope group name, got nil")
	}

	if err.Error() != "scopeGroupName cannot be empty" {
		t.Errorf("Expected 'scopeGroupName cannot be empty', got %s", err.Error())
	}
}

func TestScopeGroup_Get_NotFound(t *testing.T) {
	server := NewMockServer(http.StatusNotFound, `{"error": "scope group not found"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	_, err := scopeGroup.Get(context.Background(), "nonexistent-group")

	if err == nil {
		t.Error("Expected error for not found scope group, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get scope group") {
		t.Errorf("Expected 'failed to get scope group' in error, got %s", err.Error())
	}
}

func TestScopeGroup_Get_ServerError(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "internal server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	_, err := scopeGroup.Get(context.Background(), "test-group")

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get scope group") {
		t.Errorf("Expected 'failed to get scope group' in error, got %s", err.Error())
	}
}

func TestScopeGroup_Get_InvalidJSON(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{invalid json}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	_, err := scopeGroup.Get(context.Background(), "test-group")

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestScopeGroup_Delete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		// Verify URL path contains the group name
		if !strings.Contains(r.URL.Path, "admin-group") {
			t.Errorf("Expected 'admin-group' in URL path, got %s", r.URL.Path)
		}

		if !strings.Contains(r.URL.Path, "scopes-srv/group") {
			t.Errorf("Expected scopes-srv/group endpoint, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	err := scopeGroup.Delete(context.Background(), "admin-group")

	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestScopeGroup_Delete_WithSpecialCharacters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL path contains group name with special characters
		if !strings.Contains(r.URL.Path, "test-group_123") {
			t.Errorf("Expected 'test-group_123' in URL path, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	err := scopeGroup.Delete(context.Background(), "test-group_123")

	if err != nil {
		t.Fatalf("Delete with special characters failed: %v", err)
	}
}

func TestScopeGroup_Delete_EmptyName(t *testing.T) {
	config := NewTestClientConfig("http://test.com")
	scopeGroup := NewScopeGroup(config)

	err := scopeGroup.Delete(context.Background(), "")

	if err == nil {
		t.Error("Expected error for empty scope group name, got nil")
	}

	if err.Error() != "scopeGroupName cannot be empty" {
		t.Errorf("Expected 'scopeGroupName cannot be empty', got %s", err.Error())
	}
}

func TestScopeGroup_Delete_ServerError(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "internal server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	err := scopeGroup.Delete(context.Background(), "test-group")

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to delete scope group") {
		t.Errorf("Expected 'failed to delete scope group' in error, got %s", err.Error())
	}
}

func TestScopeGroup_Delete_NotFound(t *testing.T) {
	server := NewMockServer(http.StatusNotFound, `{"error": "scope group not found"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	err := scopeGroup.Delete(context.Background(), "nonexistent-group")

	if err == nil {
		t.Error("Expected error for not found scope group, got nil")
	}

	if !strings.Contains(err.Error(), "failed to delete scope group") {
		t.Errorf("Expected 'failed to delete scope group' in error, got %s", err.Error())
	}
}

func TestScopeGroup_GetAll_Success(t *testing.T) {
	expectedScopeGroups := []ScopeGroupConfig{
		{
			ID:          "group-1",
			GroupName:   "admin-group",
			Description: "Administrative group",
			CreatedTime: "2024-01-15T10:30:00Z",
			UpdatedTime: "2024-01-15T11:45:00Z",
		},
		{
			ID:          "group-2",
			GroupName:   "user-group",
			Description: "User group",
			CreatedTime: "2024-01-15T09:00:00Z",
			UpdatedTime: "2024-01-15T10:15:00Z",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "scopes-srv/group/list") {
			t.Errorf("Expected scopes-srv/group/list endpoint, got %s", r.URL.Path)
		}

		// Send success response
		response := AllScopeGroupResp{
			Success: true,
			Status:  200,
			Data:    expectedScopeGroups,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	result, err := scopeGroup.GetAll(context.Background())

	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 scope groups, got %d", len(result))
	}

	if result[0].GroupName != "admin-group" {
		t.Errorf("Expected first group name 'admin-group', got %s", result[0].GroupName)
	}

	if result[1].GroupName != "user-group" {
		t.Errorf("Expected second group name 'user-group', got %s", result[1].GroupName)
	}

	// Verify timestamps are preserved
	if result[0].CreatedTime != expectedScopeGroups[0].CreatedTime {
		t.Errorf("Expected CreatedTime %s, got %s", expectedScopeGroups[0].CreatedTime, result[0].CreatedTime)
	}

	if result[1].UpdatedTime != expectedScopeGroups[1].UpdatedTime {
		t.Errorf("Expected UpdatedTime %s, got %s", expectedScopeGroups[1].UpdatedTime, result[1].UpdatedTime)
	}
}

func TestScopeGroup_GetAll_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := AllScopeGroupResp{
			Success: true,
			Status:  200,
			Data:    []ScopeGroupConfig{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	result, err := scopeGroup.GetAll(context.Background())

	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 scope groups, got %d", len(result))
	}
}

func TestScopeGroup_GetAll_ServerError(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "internal server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	_, err := scopeGroup.GetAll(context.Background())

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get scope group list") {
		t.Errorf("Expected 'failed to get scope group list' in error, got %s", err.Error())
	}
}

func TestScopeGroup_GetAll_InvalidJSON(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{invalid json}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	_, err := scopeGroup.GetAll(context.Background())

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestScopeGroup_ContextCancellation_Get(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{"success": true}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := scopeGroup.Get(ctx, "test-group")

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestScopeGroup_ContextCancellation_Delete(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{"success": true}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := scopeGroup.Delete(ctx, "test-group")

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestScopeGroup_ContextCancellation_GetAll(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{"success": true}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := scopeGroup.GetAll(ctx)

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestScopeGroup_Get_WithLongGroupName(t *testing.T) {
	longGroupName := strings.Repeat("very-long-group-name-", 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		groupNameParam := r.URL.Query().Get("group_name")
		if groupNameParam != longGroupName {
			t.Errorf("Expected group_name parameter '%s', got %s", longGroupName, groupNameParam)
		}

		response := ScopeGroupResponse{
			Success: true,
			Status:  200,
			Data: ScopeGroupConfig{
				GroupName: longGroupName,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	result, err := scopeGroup.Get(context.Background(), longGroupName)

	if err != nil {
		t.Fatalf("Get with long group name failed: %v", err)
	}

	if result.Data.GroupName != longGroupName {
		t.Errorf("Expected GroupName %s, got %s", longGroupName, result.Data.GroupName)
	}
}

func TestScopeGroup_Delete_WithLongGroupName(t *testing.T) {
	longGroupName := strings.Repeat("very-long-group-name-", 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, longGroupName) {
			t.Errorf("Expected '%s' in URL path, got %s", longGroupName, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	scopeGroup := NewScopeGroup(config)

	err := scopeGroup.Delete(context.Background(), longGroupName)

	if err != nil {
		t.Fatalf("Delete with long group name failed: %v", err)
	}
}
