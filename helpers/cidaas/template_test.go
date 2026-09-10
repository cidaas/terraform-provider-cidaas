// helpers/cidaas/template_test.go
package cidaas

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

func TestNewTemplate(t *testing.T) {
	config := NewTestClientConfig("http://test.com")

	template := NewTemplate(config)

	if template == nil {
		t.Fatal("Expected template instance, got nil")
	}

	if template.BaseURL != config.BaseURL {
		t.Errorf("Expected BaseURL %s, got %s", config.BaseURL, template.BaseURL)
	}

	if template.AccessToken != config.AccessToken {
		t.Errorf("Expected AccessToken %s, got %s", config.AccessToken, template.AccessToken)
	}
}

func TestTemplate_Upsert_SystemTemplate_Success(t *testing.T) {
	expectedTemplate := TemplateModel{
		ID:               "template-123",
		Locale:           "en-US",
		TemplateKey:      "WELCOME_EMAIL",
		TemplateType:     "EMAIL",
		Content:          "<h1>Welcome!</h1>",
		Subject:          "Welcome to our platform",
		TemplateOwner:    "system",
		UsageType:        "NOTIFICATION",
		ProcessingType:   "SYNC",
		VerificationType: "EMAIL",
		Language:         "en",
		GroupID:          "group-123",
		Enabled:          true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint for system template
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// System template should NOT have /custom in URL
		if strings.Contains(r.URL.Path, "/custom") {
			t.Error("System template URL should not contain /custom")
		}

		if !strings.Contains(r.URL.Path, "templates-srv/template") {
			t.Errorf("Expected templates-srv/template endpoint, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedTemplate TemplateModel
		json.Unmarshal(body, &receivedTemplate)

		if receivedTemplate.TemplateKey != expectedTemplate.TemplateKey {
			t.Errorf("Expected TemplateKey %s, got %s", expectedTemplate.TemplateKey, receivedTemplate.TemplateKey)
		}

		// Send success response
		response := TemplateResponse{
			Success: true,
			Status:  200,
			Data:    expectedTemplate,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	result, err := template.Upsert(context.Background(), expectedTemplate, true) // isSystemTemplate = true

	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.TemplateKey != expectedTemplate.TemplateKey {
		t.Errorf("Expected TemplateKey %s, got %s", expectedTemplate.TemplateKey, result.Data.TemplateKey)
	}
}

func TestTemplate_Upsert_CustomTemplate_Success(t *testing.T) {
	expectedTemplate := TemplateModel{
		TemplateKey:  "CUSTOM_EMAIL",
		TemplateType: "EMAIL",
		Content:      "<h1>Custom Template</h1>",
		Enabled:      true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Custom template should have /custom in URL
		if !strings.Contains(r.URL.Path, "/custom") {
			t.Error("Custom template URL should contain /custom")
		}

		response := TemplateResponse{
			Success: true,
			Status:  200,
			Data:    expectedTemplate,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	result, err := template.Upsert(context.Background(), expectedTemplate, false) // isSystemTemplate = false

	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}
}

func TestTemplate_Upsert_EmptyResponseBody(t *testing.T) {
	server := NewMockServer(http.StatusOK, "")
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	templateModel := TemplateModel{TemplateKey: "TEST"}
	_, err := template.Upsert(context.Background(), templateModel, true)

	if err == nil {
		t.Error("Expected error for empty response body, got nil")
	}

	if !strings.Contains(err.Error(), "empty response body") {
		t.Errorf("Expected 'empty response body' in error, got %s", err.Error())
	}
}

func TestTemplate_Upsert_InvalidJSON(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{invalid json}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	templateModel := TemplateModel{TemplateKey: "TEST"}
	_, err := template.Upsert(context.Background(), templateModel, true)

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "failed to unmarshal json body") {
		t.Errorf("Expected 'failed to unmarshal json body' in error, got %s", err.Error())
	}
}

func TestTemplate_Get_SystemTemplate_Success(t *testing.T) {
	expectedTemplate := TemplateModel{
		TemplateKey:  "SYSTEM_EMAIL",
		TemplateType: "EMAIL",
		Locale:       "en-US",
		Content:      "<h1>System Template</h1>",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// System template should have /find (not /custom/find)
		if !strings.Contains(r.URL.Path, "/find") {
			t.Error("System template URL should contain /find")
		}

		if strings.Contains(r.URL.Path, "/custom/find") {
			t.Error("System template URL should not contain /custom/find")
		}

		response := TemplateResponse{
			Success: true,
			Status:  200,
			Data:    expectedTemplate,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	result, err := template.Get(context.Background(), expectedTemplate, true) // isSystemTemplate = true

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.TemplateKey != expectedTemplate.TemplateKey {
		t.Errorf("Expected TemplateKey %s, got %s", expectedTemplate.TemplateKey, result.Data.TemplateKey)
	}
}

func TestTemplate_Get_CustomTemplate_Success(t *testing.T) {
	expectedTemplate := TemplateModel{
		TemplateKey:  "CUSTOM_EMAIL",
		TemplateType: "EMAIL",
		Locale:       "en-US",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Custom template should have /custom/find
		if !strings.Contains(r.URL.Path, "/custom/find") {
			t.Error("Custom template URL should contain /custom/find")
		}

		response := TemplateResponse{
			Success: true,
			Status:  200,
			Data:    expectedTemplate,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	result, err := template.Get(context.Background(), expectedTemplate, false) // isSystemTemplate = false

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if result.Data.TemplateKey != expectedTemplate.TemplateKey {
		t.Errorf("Expected TemplateKey %s, got %s", expectedTemplate.TemplateKey, result.Data.TemplateKey)
	}
}

func TestTemplate_Get_NotFound_204(t *testing.T) {
	server := NewMockServer(http.StatusNoContent, "")
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	templateModel := TemplateModel{
		TemplateKey:  "NONEXISTENT",
		TemplateType: "EMAIL",
		Locale:       "en-US",
	}

	_, err := template.Get(context.Background(), templateModel, true)

	if err == nil {
		t.Error("Expected error for 204 status, got nil")
	}

	if !errors.Is(err, util.ErrResourceNotFound) {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "template not found for the  template_key NONEXISTENT") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTemplate_Get_EmptyResponseBody(t *testing.T) {
	server := NewMockServer(http.StatusOK, "")
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	templateModel := TemplateModel{TemplateKey: "TEST"}
	_, err := template.Get(context.Background(), templateModel, true)

	if err == nil {
		t.Error("Expected error for empty response body, got nil")
	}

	if !strings.Contains(err.Error(), "empty response body") {
		t.Errorf("Expected 'empty response body' in error, got %s", err.Error())
	}
}

func TestTemplate_Delete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		// Verify URL contains uppercase template key and type
		if !strings.Contains(r.URL.Path, "WELCOME_EMAIL") {
			t.Errorf("Expected WELCOME_EMAIL in URL path, got %s", r.URL.Path)
		}

		if !strings.Contains(r.URL.Path, "EMAIL") {
			t.Errorf("Expected EMAIL in URL path, got %s", r.URL.Path)
		}

		if !strings.Contains(r.URL.Path, "/custom") {
			t.Error("Delete URL should contain /custom")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	err := template.Delete(context.Background(), "welcome_email", "email") // lowercase input

	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestTemplate_Delete_ServerError(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "internal server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	err := template.Delete(context.Background(), "test", "email")

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestTemplate_GetMasterList_Success(t *testing.T) {
	expectedMasterList := []MasterList{
		{
			TemplateKey: "WELCOME_EMAIL",
			Requirement: "REQUIRED",
			Enabled:     true,
			TemplateTypes: []TemplateType{
				{
					TemplateType: "EMAIL",
					ProcessingTypes: []ProcessingType{
						{
							ProcessingType: "SYNC",
							VerificationTypes: []VerificationType{
								{
									VerificationType: "EMAIL",
									UsageTypes:       []string{"NOTIFICATION", "VERIFICATION"},
								},
							},
						},
					},
					Default: Default{
						UsageType:      "NOTIFICATION",
						ProcessingType: "SYNC",
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		// Verify URL contains group ID
		if !strings.Contains(r.URL.Path, "group-123") {
			t.Errorf("Expected group-123 in URL path, got %s", r.URL.Path)
		}

		if !strings.Contains(r.URL.Path, "templates-srv/master/settings") {
			t.Errorf("Expected templates-srv/master/settings endpoint, got %s", r.URL.Path)
		}

		// Send success response
		response := MasterListResponse{
			Success: true,
			Status:  200,
			Data:    expectedMasterList,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	result, err := template.GetMasterList(context.Background(), "group-123")

	if err != nil {
		t.Fatalf("GetMasterList failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if len(result.Data) != 1 {
		t.Errorf("Expected 1 master list item, got %d", len(result.Data))
	}

	if result.Data[0].TemplateKey != "WELCOME_EMAIL" {
		t.Errorf("Expected TemplateKey WELCOME_EMAIL, got %s", result.Data[0].TemplateKey)
	}

	if len(result.Data[0].TemplateTypes) != 1 {
		t.Errorf("Expected 1 template type, got %d", len(result.Data[0].TemplateTypes))
	}
}

func TestTemplate_GetMasterList_ServerError(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "internal server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	_, err := template.GetMasterList(context.Background(), "group-123")

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestTemplate_ContextCancellation(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{"success": true}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	templateModel := TemplateModel{TemplateKey: "TEST"}
	_, err := template.Upsert(ctx, templateModel, true)

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestTemplate_Upsert_ReadBodyError(t *testing.T) {
	// This test simulates a scenario where reading the response body fails
	// In practice, this is hard to simulate with httptest, but we can test the error path
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Close connection immediately to cause read error
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			_ = conn.Close()
		}
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	templateModel := TemplateModel{TemplateKey: "TEST"}
	_, err := template.Upsert(context.Background(), templateModel, true)

	if err == nil {
		t.Error("Expected error for read body failure, got nil")
	}
}

func TestTemplate_Get_InvalidJSON(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{invalid json}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	template := NewTemplate(config)

	templateModel := TemplateModel{TemplateKey: "TEST"}
	_, err := template.Get(context.Background(), templateModel, true)

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "failed to unmarshal json body") {
		t.Errorf("Expected 'failed to unmarshal json body' in error, got %s", err.Error())
	}
}
