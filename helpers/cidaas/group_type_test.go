// helpers/cidaas/group_type_test.go
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

func TestNewGroupType(t *testing.T) {
	config := NewTestClientConfig("http://test.com")

	groupType := NewGroupType(config)

	if groupType == nil {
		t.Fatal("Expected group type instance, got nil")
	}

	if groupType.BaseURL != config.BaseURL {
		t.Errorf("Expected BaseURL %s, got %s", config.BaseURL, groupType.BaseURL)
	}

	if groupType.AccessToken != config.AccessToken {
		t.Errorf("Expected AccessToken %s, got %s", config.AccessToken, groupType.AccessToken)
	}
}

func TestGroupType_Create_Success(t *testing.T) {
	expectedGroupType := GroupTypeData{
		ID:           "gt-123",
		RoleMode:     "SINGLE_ROLE",
		GroupType:    "admin-group",
		Description:  "Administrative group type",
		AllowedRoles: []string{"admin", "super-admin", "moderator"},
		CreatedTime:  "2024-01-15T10:30:00Z",
		UpdatedTime:  "2024-01-15T11:45:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "groups-srv/grouptypes") {
			t.Errorf("Expected groups-srv/grouptypes endpoint, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var receivedGroupType GroupTypeData
		_ = json.Unmarshal(body, &receivedGroupType)

		if receivedGroupType.GroupType != expectedGroupType.GroupType {
			t.Errorf("Expected GroupType %s, got %s", expectedGroupType.GroupType, receivedGroupType.GroupType)
		}

		if len(receivedGroupType.AllowedRoles) != 3 {
			t.Errorf("Expected 3 allowed roles, got %d", len(receivedGroupType.AllowedRoles))
		}

		response := GroupTypeResponse{
			Success: true,
			Status:  201,
			Data:    expectedGroupType,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	groupType := NewGroupType(config)

	result, err := groupType.Create(context.Background(), expectedGroupType)

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.GroupType != expectedGroupType.GroupType {
		t.Errorf("Expected GroupType %s, got %s", expectedGroupType.GroupType, result.Data.GroupType)
	}
}

func TestGroupType_Create_Error(t *testing.T) {
	server := NewMockServer(http.StatusBadRequest, `{"error": "invalid group type"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	groupType := NewGroupType(config)

	groupTypeData := GroupTypeData{GroupType: "invalid"}
	_, err := groupType.Create(context.Background(), groupTypeData)

	if err == nil {
		t.Error("Expected error for bad request, got nil")
	}
}

func TestGroupType_Get_Success(t *testing.T) {
	expectedGroupType := GroupTypeData{
		ID:           "gt-456",
		RoleMode:     "MULTI_ROLE",
		GroupType:    "user-group",
		Description:  "User group type",
		AllowedRoles: []string{"user", "viewer"},
		CreatedTime:  "2024-01-10T09:00:00Z",
		UpdatedTime:  "2024-01-10T09:30:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "groups-srv/grouptypes") {
			t.Errorf("Expected groups-srv/grouptypes endpoint, got %s", r.URL.Path)
		}

		groupTypeParam := r.URL.Query().Get("groupType")
		if groupTypeParam != "user-group" {
			t.Errorf("Expected groupType parameter 'user-group', got %s", groupTypeParam)
		}

		response := GroupTypeResponse{
			Success: true,
			Status:  200,
			Data:    expectedGroupType,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	groupType := NewGroupType(config)

	result, err := groupType.Get(context.Background(), "user-group")

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.RoleMode != expectedGroupType.RoleMode {
		t.Errorf("Expected RoleMode %s, got %s", expectedGroupType.RoleMode, result.Data.RoleMode)
	}
}

func TestGroupType_Get_Error(t *testing.T) {
	server := NewMockServer(http.StatusNotFound, `{"error": "group type not found"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	groupType := NewGroupType(config)

	_, err := groupType.Get(context.Background(), "nonexistent")

	if err == nil {
		t.Error("Expected error for not found group type, got nil")
	}
}

func TestGroupType_Update_Success(t *testing.T) {
	updateGroupType := GroupTypeData{
		ID:           "gt-123",
		RoleMode:     "SINGLE_ROLE",
		GroupType:    "updated-group",
		Description:  "Updated description",
		AllowedRoles: []string{"admin", "editor"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "groups-srv/grouptypes") {
			t.Errorf("Expected groups-srv/grouptypes endpoint, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var receivedGroupType GroupTypeData
		json.Unmarshal(body, &receivedGroupType)

		if receivedGroupType.ID != updateGroupType.ID {
			t.Errorf("Expected ID %s, got %s", updateGroupType.ID, receivedGroupType.ID)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	groupType := NewGroupType(config)

	err := groupType.Update(context.Background(), updateGroupType)

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
}

func TestGroupType_Update_Error(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	groupType := NewGroupType(config)

	updateGroupType := GroupTypeData{ID: "gt-123"}
	err := groupType.Update(context.Background(), updateGroupType)

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestGroupType_Delete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "groups-srv/grouptypes/admin-group") {
			t.Errorf("Expected groups-srv/grouptypes/admin-group endpoint, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	groupType := NewGroupType(config)

	err := groupType.Delete(context.Background(), "admin-group")

	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestGroupType_Delete_Error(t *testing.T) {
	server := NewMockServer(http.StatusNotFound, `{"error": "group type not found"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	groupType := NewGroupType(config)

	err := groupType.Delete(context.Background(), "nonexistent")

	if err == nil {
		t.Error("Expected error for not found group type, got nil")
	}
}

func TestGroupType_GetAll_Success(t *testing.T) {
	expectedGroupTypes := []GroupTypeData{
		{
			ID:           "gt-1",
			RoleMode:     "SINGLE_ROLE",
			GroupType:    "admin-group",
			Description:  "Admin group",
			AllowedRoles: []string{"admin"},
		},
		{
			ID:           "gt-2",
			RoleMode:     "MULTI_ROLE",
			GroupType:    "user-group",
			Description:  "User group",
			AllowedRoles: []string{"user", "viewer"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "groups-srv/graph/grouptypes") {
			t.Errorf("Expected groups-srv/graph/grouptypes endpoint, got %s", r.URL.Path)
		}

		// Verify empty struct payload
		body, _ := io.ReadAll(r.Body)
		var payload struct{}
		json.Unmarshal(body, &payload)

		response := AllGroupTypeResponse{
			Success: true,
			Status:  200,
			Data:    expectedGroupTypes,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	groupType := NewGroupType(config)

	result, err := groupType.GetAll(context.Background())

	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 group types, got %d", len(result))
	}

	if result[0].GroupType != "admin-group" {
		t.Errorf("Expected first group type 'admin-group', got %s", result[0].GroupType)
	}

	if result[1].GroupType != "user-group" {
		t.Errorf("Expected second group type 'user-group', got %s", result[1].GroupType)
	}
}

func TestGroupType_GetAll_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := AllGroupTypeResponse{
			Success: true,
			Status:  200,
			Data:    []GroupTypeData{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	groupType := NewGroupType(config)

	result, err := groupType.GetAll(context.Background())

	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 group types, got %d", len(result))
	}
}

func TestGroupType_GetAll_Error(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	groupType := NewGroupType(config)

	_, err := groupType.GetAll(context.Background())

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}
