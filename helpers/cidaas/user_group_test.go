// helpers/cidaas/user_group_test.go
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

func TestNewUserGroup(t *testing.T) {
	config := ClientConfig{
		BaseURL:     "http://test.com",
		AccessToken: "test-token",
		ClientID:    "test-client-id",
	}

	userGroup := NewUserGroup(config)

	if userGroup == nil {
		t.Fatal("Expected user group instance, got nil")
	}

	if userGroup.BaseURL != config.BaseURL {
		t.Errorf("Expected BaseURL %s, got %s", config.BaseURL, userGroup.BaseURL)
	}

	if userGroup.AccessToken != config.AccessToken {
		t.Errorf("Expected AccessToken %s, got %s", config.AccessToken, userGroup.AccessToken)
	}
}

func TestUserGroup_Create_Success(t *testing.T) {
	expectedGroup := UserGroupData{
		ID:                          "group-123",
		GroupType:                   "DYNAMIC",
		GroupID:                     "test-group",
		GroupName:                   "Test Group",
		ParentID:                    "parent-123",
		LogoURL:                     "https://example.com/logo.png",
		Description:                 "Test Description",
		MakeFirstUserAdmin:          true,
		MemberProfileVisibility:     "PUBLIC",
		NoneMemberProfileVisibility: "PRIVATE",
		GroupOwner:                  "owner-123",
		CustomFields: map[string]string{
			"field1": "value1",
			"field2": "value2",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "groups-srv/usergroups") {
			t.Errorf("Expected groups-srv/usergroups endpoint, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedGroup UserGroupData
		json.Unmarshal(body, &receivedGroup)

		if receivedGroup.GroupName != expectedGroup.GroupName {
			t.Errorf("Expected GroupName %s, got %s", expectedGroup.GroupName, receivedGroup.GroupName)
		}

		if receivedGroup.GroupType != expectedGroup.GroupType {
			t.Errorf("Expected GroupType %s, got %s", expectedGroup.GroupType, receivedGroup.GroupType)
		}

		// Send success response
		response := UserGroupResponse{
			Success: true,
			Status:  200,
			Data:    expectedGroup,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	userGroup := NewUserGroup(config)

	result, err := userGroup.Create(context.Background(), expectedGroup)

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.GroupName != expectedGroup.GroupName {
		t.Errorf("Expected GroupName %s, got %s", expectedGroup.GroupName, result.Data.GroupName)
	}

	if result.Data.GroupType != expectedGroup.GroupType {
		t.Errorf("Expected GroupType %s, got %s", expectedGroup.GroupType, result.Data.GroupType)
	}
}

func TestUserGroup_Create_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	userGroup := NewUserGroup(config)

	groupData := UserGroupData{GroupName: "Test Group"}
	_, err := userGroup.Create(context.Background(), groupData)

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to create user group") {
		t.Errorf("Expected 'failed to create user group' in error, got %s", err.Error())
	}
}

func TestUserGroup_Get_Success(t *testing.T) {
	expectedGroup := UserGroupData{
		ID:          "group-123",
		GroupName:   "Test Group",
		GroupType:   "STATIC",
		Description: "Test Description",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		// Verify URL path contains the group ID
		if !strings.Contains(r.URL.Path, "group-123") {
			t.Errorf("Expected group-123 in URL path, got %s", r.URL.Path)
		}

		// Send success response
		response := UserGroupResponse{
			Success: true,
			Status:  200,
			Data:    expectedGroup,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	userGroup := NewUserGroup(config)

	result, err := userGroup.Get(context.Background(), "group-123")

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.ID != expectedGroup.ID {
		t.Errorf("Expected ID %s, got %s", expectedGroup.ID, result.Data.ID)
	}

	if result.Data.GroupName != expectedGroup.GroupName {
		t.Errorf("Expected GroupName %s, got %s", expectedGroup.GroupName, result.Data.GroupName)
	}
}

func TestUserGroup_Get_EmptyGroupID(t *testing.T) {
	config := ClientConfig{}
	userGroup := NewUserGroup(config)

	_, err := userGroup.Get(context.Background(), "")

	if err == nil {
		t.Error("Expected error for empty group ID, got nil")
	}

	if err.Error() != "groupID cannot be empty" {
		t.Errorf("Expected 'groupID cannot be empty', got %s", err.Error())
	}
}

func TestUserGroup_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "user group not found"}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	userGroup := NewUserGroup(config)

	_, err := userGroup.Get(context.Background(), "nonexistent-group")

	if err == nil {
		t.Error("Expected error for not found group, got nil")
	}

	if !strings.Contains(err.Error(), "failed to ger user group") {
		t.Errorf("Expected 'failed to ger user group' in error, got %s", err.Error())
	}
}

func TestUserGroup_Update_Success(t *testing.T) {
	updatedGroup := UserGroupData{
		ID:          "group-123",
		GroupName:   "Updated Group",
		GroupType:   "DYNAMIC",
		Description: "Updated Description",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "groups-srv/usergroups") {
			t.Errorf("Expected groups-srv/usergroups endpoint, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedGroup UserGroupData
		json.Unmarshal(body, &receivedGroup)

		if receivedGroup.GroupName != updatedGroup.GroupName {
			t.Errorf("Expected GroupName %s, got %s", updatedGroup.GroupName, receivedGroup.GroupName)
		}

		// Send success response
		response := UserGroupResponse{
			Success: true,
			Status:  200,
			Data:    updatedGroup,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	userGroup := NewUserGroup(config)

	result, err := userGroup.Update(context.Background(), updatedGroup)

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.GroupName != updatedGroup.GroupName {
		t.Errorf("Expected GroupName %s, got %s", updatedGroup.GroupName, result.Data.GroupName)
	}
}

func TestUserGroup_Update_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	userGroup := NewUserGroup(config)

	groupData := UserGroupData{ID: "group-123", GroupName: "Test Group"}
	_, err := userGroup.Update(context.Background(), groupData)

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to update user group") {
		t.Errorf("Expected 'failed to update user group' in error, got %s", err.Error())
	}
}

func TestUserGroup_Delete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		// Verify URL path contains the group ID
		if !strings.Contains(r.URL.Path, "group-123") {
			t.Errorf("Expected group-123 in URL path, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	userGroup := NewUserGroup(config)

	err := userGroup.Delete(context.Background(), "group-123")

	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestUserGroup_Delete_EmptyGroupID(t *testing.T) {
	config := ClientConfig{}
	userGroup := NewUserGroup(config)

	err := userGroup.Delete(context.Background(), "")

	if err == nil {
		t.Error("Expected error for empty group ID, got nil")
	}

	if err.Error() != "groupID cannot be empty" {
		t.Errorf("Expected 'groupID cannot be empty', got %s", err.Error())
	}
}

func TestUserGroup_Delete_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	userGroup := NewUserGroup(config)

	err := userGroup.Delete(context.Background(), "group-123")

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to delete user group") {
		t.Errorf("Expected 'failed to delete user group' in error, got %s", err.Error())
	}
}

func TestUserGroup_GetSubGroups_Success(t *testing.T) {
	expectedSubGroups := []UserGroupData{
		{
			ID:        "subgroup-1",
			GroupName: "Sub Group 1",
			ParentID:  "parent-123",
		},
		{
			ID:        "subgroup-2",
			GroupName: "Sub Group 2",
			ParentID:  "parent-123",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "groups-srv/graph/usergroups") {
			t.Errorf("Expected groups-srv/graph/usergroups endpoint, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var payload map[string]string
		json.Unmarshal(body, &payload)

		if payload["parentId"] != "parent-123" {
			t.Errorf("Expected parentId 'parent-123', got %s", payload["parentId"])
		}

		// Send success response
		response := SubGroupResponse{
			Success: true,
			Status:  200,
			Data: struct {
				Groups []UserGroupData `json:"groups"`
			}{
				Groups: expectedSubGroups,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	userGroup := NewUserGroup(config)

	result, err := userGroup.GetSubGroups(context.Background(), "parent-123")

	if err != nil {
		t.Fatalf("GetSubGroups failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 sub groups, got %d", len(result))
	}

	if result[0].ID != "subgroup-1" {
		t.Errorf("Expected first subgroup ID 'subgroup-1', got %s", result[0].ID)
	}

	if result[1].ID != "subgroup-2" {
		t.Errorf("Expected second subgroup ID 'subgroup-2', got %s", result[1].ID)
	}
}

func TestUserGroup_GetSubGroups_NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	userGroup := NewUserGroup(config)

	result, err := userGroup.GetSubGroups(context.Background(), "parent-123")

	if err != nil {
		t.Fatalf("GetSubGroups failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 sub groups for no content, got %d", len(result))
	}
}

func TestUserGroup_GetSubGroups_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	userGroup := NewUserGroup(config)

	_, err := userGroup.GetSubGroups(context.Background(), "parent-123")

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "failed to decode subgroup response") {
		t.Errorf("Expected 'failed to decode subgroup response' in error, got %s", err.Error())
	}
}

func TestUserGroup_Create_WithCustomFields(t *testing.T) {
	groupWithCustomFields := UserGroupData{
		GroupName: "Test Group",
		CustomFields: map[string]string{
			"department": "engineering",
			"location":   "remote",
			"priority":   "high",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedGroup UserGroupData
		json.Unmarshal(body, &receivedGroup)

		if len(receivedGroup.CustomFields) != 3 {
			t.Errorf("Expected 3 custom fields, got %d", len(receivedGroup.CustomFields))
		}

		if receivedGroup.CustomFields["department"] != "engineering" {
			t.Errorf("Expected department 'engineering', got %s", receivedGroup.CustomFields["department"])
		}

		response := UserGroupResponse{
			Success: true,
			Status:  200,
			Data:    groupWithCustomFields,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	userGroup := NewUserGroup(config)

	result, err := userGroup.Create(context.Background(), groupWithCustomFields)

	if err != nil {
		t.Fatalf("Create with custom fields failed: %v", err)
	}

	if len(result.Data.CustomFields) != 3 {
		t.Errorf("Expected 3 custom fields in response, got %d", len(result.Data.CustomFields))
	}
}

func TestUserGroup_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This shouldn't be reached due to context cancellation
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	userGroup := NewUserGroup(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	groupData := UserGroupData{GroupName: "Test Group"}
	_, err := userGroup.Create(ctx, groupData)

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}
