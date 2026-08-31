// helpers/cidaas/template_group_test.go
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

func TestNewTemplateGroup(t *testing.T) {
	config := ClientConfig{
		BaseURL:     "http://test.com",
		AccessToken: "test-token",
		ClientID:    "test-client-id",
	}

	templateGroup := NewTemplateGroup(config)

	if templateGroup == nil {
		t.Fatal("Expected template group instance, got nil")
	}

	if templateGroup.BaseURL != config.BaseURL {
		t.Errorf("Expected BaseURL %s, got %s", config.BaseURL, templateGroup.BaseURL)
	}

	if templateGroup.AccessToken != config.AccessToken {
		t.Errorf("Expected AccessToken %s, got %s", config.AccessToken, templateGroup.AccessToken)
	}
}

func TestTemplateGroup_Create_Success(t *testing.T) {
	expectedTemplate := TemplateGroupModel{
		ID:      "template-123",
		GroupID: "group-456",
		SenderConfig: &SenderConfig{
			ID:        "sender-123",
			FromEmail: "test@example.com",
			FromName:  "Test Sender",
		},
		EmailSenderConfig: &EmailSenderConfig{
			ID:          "email-sender-123",
			FromEmail:   "noreply@example.com",
			FromName:    "No Reply",
			ReplyTo:     "support@example.com",
			SenderNames: []string{"sender1", "sender2"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "templates-srv/groups") {
			t.Errorf("Expected templates-srv/groups endpoint, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedTemplate TemplateGroupModel
		json.Unmarshal(body, &receivedTemplate)

		if receivedTemplate.GroupID != expectedTemplate.GroupID {
			t.Errorf("Expected GroupID %s, got %s", expectedTemplate.GroupID, receivedTemplate.GroupID)
		}

		// Send success response
		response := TemplateGroupResponse{
			Success: true,
			Status:  200,
			Data:    expectedTemplate,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	templateGroup := NewTemplateGroup(config)

	result, err := templateGroup.Create(context.Background(), expectedTemplate)

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.GroupID != expectedTemplate.GroupID {
		t.Errorf("Expected GroupID %s, got %s", expectedTemplate.GroupID, result.Data.GroupID)
	}
}

func TestTemplateGroup_Create_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	templateGroup := NewTemplateGroup(config)

	templateData := TemplateGroupModel{GroupID: "test-group"}
	_, err := templateGroup.Create(context.Background(), templateData)

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestTemplateGroup_Update_Success(t *testing.T) {
	updatedTemplate := TemplateGroupModel{
		ID:      "template-123",
		GroupID: "group-456",
		SenderConfig: &SenderConfig{
			ID:        "sender-123",
			FromEmail: "updated@example.com",
			FromName:  "Updated Sender",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}

		// Verify URL contains group ID
		if !strings.Contains(r.URL.Path, "group-456") {
			t.Errorf("Expected group-456 in URL path, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedTemplate TemplateGroupModel
		json.Unmarshal(body, &receivedTemplate)

		if receivedTemplate.SenderConfig.FromEmail != updatedTemplate.SenderConfig.FromEmail {
			t.Errorf("Expected FromEmail %s, got %s", updatedTemplate.SenderConfig.FromEmail, receivedTemplate.SenderConfig.FromEmail)
		}

		// Send success response
		response := TemplateGroupResponse{
			Success: true,
			Status:  200,
			Data:    updatedTemplate,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	templateGroup := NewTemplateGroup(config)

	result, err := templateGroup.Update(context.Background(), updatedTemplate)

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.SenderConfig.FromEmail != updatedTemplate.SenderConfig.FromEmail {
		t.Errorf("Expected FromEmail %s, got %s", updatedTemplate.SenderConfig.FromEmail, result.Data.SenderConfig.FromEmail)
	}
}

func TestTemplateGroup_Update_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	templateGroup := NewTemplateGroup(config)

	templateData := TemplateGroupModel{GroupID: "test-group"}
	_, err := templateGroup.Update(context.Background(), templateData)

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestTemplateGroup_Get_Success(t *testing.T) {
	expectedTemplate := TemplateGroupModel{
		ID:      "template-123",
		GroupID: "group-456",
		SMSSenderConfig: &SMSSenderConfig{
			ID:          "sms-sender-123",
			FromName:    "SMS Sender",
			SenderNames: []string{"sms1", "sms2"},
		},
		IVRSenderConfig: &IVRSenderConfig{
			ID:          "ivr-sender-123",
			SenderNames: []string{"ivr1", "ivr2"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		// Verify URL contains group ID
		if !strings.Contains(r.URL.Path, "group-456") {
			t.Errorf("Expected group-456 in URL path, got %s", r.URL.Path)
		}

		// Send success response
		response := TemplateGroupResponse{
			Success: true,
			Status:  200,
			Data:    expectedTemplate,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	templateGroup := NewTemplateGroup(config)

	result, err := templateGroup.Get(context.Background(), "group-456")

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.GroupID != expectedTemplate.GroupID {
		t.Errorf("Expected GroupID %s, got %s", expectedTemplate.GroupID, result.Data.GroupID)
	}
}

func TestTemplateGroup_Get_NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	templateGroup := NewTemplateGroup(config)

	result, err := templateGroup.Get(context.Background(), "nonexistent-group")

	if err == nil {
		t.Error("Expected error for no content, got nil")
	}

	if !strings.Contains(err.Error(), "template group not found") {
		t.Errorf("Expected 'template group not found' in error, got %s", err.Error())
	}

	if result.Status != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, result.Status)
	}
}

func TestTemplateGroup_Get_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	templateGroup := NewTemplateGroup(config)

	_, err := templateGroup.Get(context.Background(), "test-group")

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestTemplateGroup_Delete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		// Verify URL contains group ID
		if !strings.Contains(r.URL.Path, "group-456") {
			t.Errorf("Expected group-456 in URL path, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	templateGroup := NewTemplateGroup(config)

	err := templateGroup.Delete(context.Background(), "group-456")

	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestTemplateGroup_Delete_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	templateGroup := NewTemplateGroup(config)

	err := templateGroup.Delete(context.Background(), "test-group")

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestTemplateGroup_Create_WithAllSenderConfigs(t *testing.T) {
	templateWithAllConfigs := TemplateGroupModel{
		GroupID: "comprehensive-group",
		SenderConfig: &SenderConfig{
			ID:        "sender-123",
			FromEmail: "test@example.com",
			FromName:  "Test Sender",
		},
		EmailSenderConfig: &EmailSenderConfig{
			ID:          "email-123",
			FromEmail:   "email@example.com",
			FromName:    "Email Sender",
			ReplyTo:     "reply@example.com",
			SenderNames: []string{"email1", "email2"},
		},
		SMSSenderConfig: &SMSSenderConfig{
			ID:          "sms-123",
			FromName:    "SMS Sender",
			SenderNames: []string{"sms1", "sms2"},
		},
		IVRSenderConfig: &IVRSenderConfig{
			ID:          "ivr-123",
			SenderNames: []string{"ivr1", "ivr2"},
		},
		PushSenderConfig: &IVRSenderConfig{
			ID:          "push-123",
			SenderNames: []string{"push1", "push2"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedTemplate TemplateGroupModel
		json.Unmarshal(body, &receivedTemplate)

		// Verify all sender configs are present
		if receivedTemplate.SenderConfig == nil {
			t.Error("Expected SenderConfig to be present")
		}
		if receivedTemplate.EmailSenderConfig == nil {
			t.Error("Expected EmailSenderConfig to be present")
		}
		if receivedTemplate.SMSSenderConfig == nil {
			t.Error("Expected SMSSenderConfig to be present")
		}
		if receivedTemplate.IVRSenderConfig == nil {
			t.Error("Expected IVRSenderConfig to be present")
		}
		if receivedTemplate.PushSenderConfig == nil {
			t.Error("Expected PushSenderConfig to be present")
		}

		response := TemplateGroupResponse{
			Success: true,
			Status:  200,
			Data:    templateWithAllConfigs,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	templateGroup := NewTemplateGroup(config)

	result, err := templateGroup.Create(context.Background(), templateWithAllConfigs)

	if err != nil {
		t.Fatalf("Create with all configs failed: %v", err)
	}

	if result.Data.SenderConfig == nil {
		t.Error("Expected SenderConfig in response")
	}
	if result.Data.EmailSenderConfig == nil {
		t.Error("Expected EmailSenderConfig in response")
	}
	if result.Data.SMSSenderConfig == nil {
		t.Error("Expected SMSSenderConfig in response")
	}
}

func TestTemplateGroup_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This shouldn't be reached due to context cancellation
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	templateGroup := NewTemplateGroup(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	templateData := TemplateGroupModel{GroupID: "test-group"}
	_, err := templateGroup.Create(ctx, templateData)

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestTemplateGroup_EmptyGroupID_Update(t *testing.T) {
	config := ClientConfig{
		BaseURL:     "http://test.com",
		AccessToken: "test-token",
	}
	templateGroup := NewTemplateGroup(config)

	// Test update with empty GroupID
	templateData := TemplateGroupModel{GroupID: ""}
	_, err := templateGroup.Update(context.Background(), templateData)

	if err == nil {
		t.Error("Expected error for empty GroupID in update, got nil")
	}
}

func TestTemplateGroup_InvalidJSON_Response(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL:     server.URL,
		AccessToken: "test-token",
	}
	templateGroup := NewTemplateGroup(config)

	templateData := TemplateGroupModel{GroupID: "test-group"}
	_, err := templateGroup.Create(context.Background(), templateData)

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
