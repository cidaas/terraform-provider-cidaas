// helpers/cidaas/webhook_test.go
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

func TestNewWebhook(t *testing.T) {
	config := NewTestClientConfig("http://test.com")

	webhook := NewWebhook(config)

	if webhook == nil {
		t.Fatal("Expected webhook instance, got nil")
	}

	if webhook.BaseURL != config.BaseURL {
		t.Errorf("Expected BaseURL %s, got %s", config.BaseURL, webhook.BaseURL)
	}

	if webhook.AccessToken != config.AccessToken {
		t.Errorf("Expected AccessToken %s, got %s", config.AccessToken, webhook.AccessToken)
	}
}

func TestWebhook_Upsert_Success(t *testing.T) {
	expectedWebhook := WebhookModel{
		ID:       "webhook-123",
		AuthType: "APIKEY",
		URL:      "https://example.com/webhook",
		Events:   []string{"LOGIN", "LOGOUT"},
		APIKeyDetails: APIKeyDetails{
			ApikeyPlaceholder: "X-API-Key",
			ApikeyPlacement:   "header",
			Apikey:            "secret-key",
		},
		Disable: false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "webhook-srv/webhook") {
			t.Errorf("Expected webhook-srv/webhook endpoint, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedWebhook WebhookModel
		json.Unmarshal(body, &receivedWebhook)

		if receivedWebhook.URL != expectedWebhook.URL {
			t.Errorf("Expected URL %s, got %s", expectedWebhook.URL, receivedWebhook.URL)
		}

		if receivedWebhook.AuthType != expectedWebhook.AuthType {
			t.Errorf("Expected AuthType %s, got %s", expectedWebhook.AuthType, receivedWebhook.AuthType)
		}

		// Send success response
		response := WebhookResponse{
			Success: true,
			Status:  200,
			Data:    expectedWebhook,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	webhook := NewWebhook(config)

	result, err := webhook.Upsert(context.Background(), expectedWebhook)

	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.URL != expectedWebhook.URL {
		t.Errorf("Expected URL %s, got %s", expectedWebhook.URL, result.Data.URL)
	}

	if result.Data.AuthType != expectedWebhook.AuthType {
		t.Errorf("Expected AuthType %s, got %s", expectedWebhook.AuthType, result.Data.AuthType)
	}
}

func TestWebhook_Upsert_WithTOTPAuth(t *testing.T) {
	webhookModel := WebhookModel{
		AuthType: "TOTP",
		URL:      "https://example.com/webhook",
		Events:   []string{"LOGIN"},
		TotpDetails: TotpDetails{
			TotpPlaceholder: "X-TOTP",
			TotpPlacement:   "header",
			TotpKey:         "totp-secret",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedWebhook WebhookModel
		json.Unmarshal(body, &receivedWebhook)

		if receivedWebhook.AuthType != "TOTP" {
			t.Errorf("Expected AuthType TOTP, got %s", receivedWebhook.AuthType)
		}

		if receivedWebhook.TotpDetails.TotpKey != "totp-secret" {
			t.Errorf("Expected TotpKey totp-secret, got %s", receivedWebhook.TotpDetails.TotpKey)
		}

		response := WebhookResponse{
			Success: true,
			Status:  200,
			Data:    webhookModel,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	webhook := NewWebhook(config)

	result, err := webhook.Upsert(context.Background(), webhookModel)

	if err != nil {
		t.Fatalf("Upsert with TOTP failed: %v", err)
	}

	if result.Data.AuthType != "TOTP" {
		t.Errorf("Expected AuthType TOTP, got %s", result.Data.AuthType)
	}
}

func TestWebhook_Upsert_WithCidaasAuth(t *testing.T) {
	webhookModel := WebhookModel{
		AuthType: "CIDAAS_OAUTH2",
		URL:      "https://example.com/webhook",
		Events:   []string{"LOGIN"},
		CidaasAuthDetails: AuthDetails{
			ClientID: "oauth-client-id",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedWebhook WebhookModel
		json.Unmarshal(body, &receivedWebhook)

		if receivedWebhook.AuthType != "CIDAAS_OAUTH2" {
			t.Errorf("Expected AuthType CIDAAS_OAUTH2, got %s", receivedWebhook.AuthType)
		}

		response := WebhookResponse{
			Success: true,
			Status:  200,
			Data:    webhookModel,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	webhook := NewWebhook(config)

	result, err := webhook.Upsert(context.Background(), webhookModel)

	if err != nil {
		t.Fatalf("Upsert with Cidaas auth failed: %v", err)
	}

	if result.Data.AuthType != "CIDAAS_OAUTH2" {
		t.Errorf("Expected AuthType CIDAAS_OAUTH2, got %s", result.Data.AuthType)
	}
}

func TestWebhook_Upsert_ServerError(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "internal server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	webhook := NewWebhook(config)

	webhookModel := WebhookModel{URL: "https://example.com/webhook"}
	_, err := webhook.Upsert(context.Background(), webhookModel)

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to upsert webhook") {
		t.Errorf("Expected 'failed to upsert webhook' in error, got %s", err.Error())
	}
}

func TestWebhook_Get_Success(t *testing.T) {
	expectedWebhook := WebhookModel{
		ID:       "webhook-123",
		AuthType: "APIKEY",
		URL:      "https://example.com/webhook",
		Events:   []string{"LOGIN", "LOGOUT"},
		Disable:  false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		// Verify query parameter
		idParam := r.URL.Query().Get("id")
		if idParam != "webhook-123" {
			t.Errorf("Expected id parameter 'webhook-123', got %s", idParam)
		}

		// Send success response
		response := WebhookResponse{
			Success: true,
			Status:  200,
			Data:    expectedWebhook,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	webhook := NewWebhook(config)

	result, err := webhook.Get(context.Background(), "webhook-123")

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Data.ID != expectedWebhook.ID {
		t.Errorf("Expected ID %s, got %s", expectedWebhook.ID, result.Data.ID)
	}

	if result.Data.URL != expectedWebhook.URL {
		t.Errorf("Expected URL %s, got %s", expectedWebhook.URL, result.Data.URL)
	}
}

func TestWebhook_Get_EmptyID(t *testing.T) {
	config := ClientConfig{}
	webhook := NewWebhook(config)

	_, err := webhook.Get(context.Background(), "")

	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}

	if err.Error() != "webhook ID cannot be empty" {
		t.Errorf("Expected 'webhook ID cannot be empty', got %s", err.Error())
	}
}

func TestWebhook_Get_NotFound(t *testing.T) {
	server := NewMockServer(http.StatusNotFound, `{"error": "webhook not found"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	webhook := NewWebhook(config)

	_, err := webhook.Get(context.Background(), "nonexistent-webhook")

	if err == nil {
		t.Error("Expected error for not found webhook, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get webhook") {
		t.Errorf("Expected 'failed to get webhook' in error, got %s", err.Error())
	}
}

func TestWebhook_Delete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and endpoint
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		// Verify URL path contains the ID
		if !strings.Contains(r.URL.Path, "webhook-123") {
			t.Errorf("Expected webhook-123 in URL path, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	webhook := NewWebhook(config)

	err := webhook.Delete(context.Background(), "webhook-123")

	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestWebhook_Delete_EmptyID(t *testing.T) {
	config := ClientConfig{}
	webhook := NewWebhook(config)

	err := webhook.Delete(context.Background(), "")

	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}

	if err.Error() != "webhook ID cannot be empty" {
		t.Errorf("Expected 'webhook ID cannot be empty', got %s", err.Error())
	}
}

func TestWebhook_Delete_ServerError(t *testing.T) {
	server := NewMockServer(http.StatusInternalServerError, `{"error": "internal server error"}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	webhook := NewWebhook(config)

	err := webhook.Delete(context.Background(), "webhook-123")

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to delete webhook") {
		t.Errorf("Expected 'failed to delete webhook' in error, got %s", err.Error())
	}
}

func TestWebhook_Upsert_InvalidJSON(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{invalid json}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	webhook := NewWebhook(config)

	webhookModel := WebhookModel{URL: "https://example.com/webhook"}
	_, err := webhook.Upsert(context.Background(), webhookModel)

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestWebhook_Upsert_ContextCancellation(t *testing.T) {
	server := NewMockServer(http.StatusOK, `{"success": true}`)
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	webhook := NewWebhook(config)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	webhookModel := WebhookModel{URL: "https://example.com/webhook"}
	_, err := webhook.Upsert(ctx, webhookModel)

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestWebhook_Upsert_WithAllEvents(t *testing.T) {
	webhookModel := WebhookModel{
		AuthType: "APIKEY",
		URL:      "https://example.com/webhook",
		Events:   []string{"ACCOUNT_MODIFIED", "LOGIN_WITH_CIDAAS", "LOGOUT", "APP_CREATED", "APP_DELETED"},
		APIKeyDetails: APIKeyDetails{
			ApikeyPlaceholder: "X-API-Key",
			ApikeyPlacement:   "header",
			Apikey:            "secret-key",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedWebhook WebhookModel
		json.Unmarshal(body, &receivedWebhook)

		if len(receivedWebhook.Events) != 5 {
			t.Errorf("Expected 5 events, got %d", len(receivedWebhook.Events))
		}

		response := WebhookResponse{
			Success: true,
			Status:  200,
			Data:    webhookModel,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	webhook := NewWebhook(config)

	result, err := webhook.Upsert(context.Background(), webhookModel)

	if err != nil {
		t.Fatalf("Upsert with multiple events failed: %v", err)
	}

	if len(result.Data.Events) != 5 {
		t.Errorf("Expected 5 events in response, got %d", len(result.Data.Events))
	}
}

func TestWebhook_Upsert_DisabledWebhook(t *testing.T) {
	webhookModel := WebhookModel{
		AuthType: "APIKEY",
		URL:      "https://example.com/webhook",
		Events:   []string{"LOGIN"},
		Disable:  true, // Test disabled webhook
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var receivedWebhook WebhookModel
		json.Unmarshal(body, &receivedWebhook)

		if !receivedWebhook.Disable {
			t.Error("Expected webhook to be disabled")
		}

		response := WebhookResponse{
			Success: true,
			Status:  200,
			Data:    webhookModel,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NewTestClientConfig(server.URL)
	webhook := NewWebhook(config)

	result, err := webhook.Upsert(context.Background(), webhookModel)

	if err != nil {
		t.Fatalf("Upsert disabled webhook failed: %v", err)
	}

	if !result.Data.Disable {
		t.Error("Expected webhook to be disabled in response")
	}
}

// Test constants validation
func TestAllowedConstants(t *testing.T) {
	// Test AllowedAuthType
	expectedAuthTypes := []string{"APIKEY", "TOTP", "CIDAAS_OAUTH2"}
	if len(AllowedAuthType) != len(expectedAuthTypes) {
		t.Errorf("Expected %d auth types, got %d", len(expectedAuthTypes), len(AllowedAuthType))
	}

	for i, expected := range expectedAuthTypes {
		if AllowedAuthType[i] != expected {
			t.Errorf("Expected auth type %s at index %d, got %s", expected, i, AllowedAuthType[i])
		}
	}

	// Test AllowedKeyPlacementValue
	expectedPlacements := []string{"query", "header"}
	if len(AllowedKeyPlacementValue) != len(expectedPlacements) {
		t.Errorf("Expected %d placement values, got %d", len(expectedPlacements), len(AllowedKeyPlacementValue))
	}

	for i, expected := range expectedPlacements {
		if AllowedKeyPlacementValue[i] != expected {
			t.Errorf("Expected placement %s at index %d, got %s", expected, i, AllowedKeyPlacementValue[i])
		}
	}
}

func TestWebhook_ListEventDescriptions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "webhook-srv/eventdescriptions") {
			t.Errorf("Expected eventdescriptions endpoint, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("category") != "webhook" {
			t.Errorf("Expected category=webhook, got %s", r.URL.Query().Get("category"))
		}
		response := EventDescriptionsResponse{
			Success: true,
			Status:  200,
			Data: []EventDescriptionModel{
				{ID: "ACCOUNT_MODIFIED", ObjectType: "users", GoodForWebhook: true},
				{ID: "APP_CREATED", ObjectType: "apps", GoodForWebhook: true},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	webhook := NewWebhook(NewTestClientConfig(server.URL))
	eds, err := webhook.ListEventDescriptions(context.Background(), "webhook")
	if err != nil {
		t.Fatalf("ListEventDescriptions failed: %v", err)
	}
	if len(eds) != 2 {
		t.Fatalf("Expected 2 event descriptions, got %d", len(eds))
	}
	if eds[0].ID != "ACCOUNT_MODIFIED" {
		t.Errorf("Expected ACCOUNT_MODIFIED, got %s", eds[0].ID)
	}
}

func TestWebhook_ListEventDescriptions_NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	webhook := NewWebhook(NewTestClientConfig(server.URL))
	eds, err := webhook.ListEventDescriptions(context.Background(), "webhook")
	if err != nil {
		t.Fatalf("ListEventDescriptions failed: %v", err)
	}
	if eds != nil {
		t.Errorf("Expected nil for 204, got %v", eds)
	}
}

func TestWebhook_ListWebhookEventIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("category") != "webhook" {
			t.Errorf("Expected category=webhook, got %s", r.URL.Query().Get("category"))
		}
		response := EventDescriptionsResponse{
			Success: true,
			Status:  200,
			Data: []EventDescriptionModel{
				{ID: "ACCOUNT_MODIFIED", ObjectType: "users", GoodForWebhook: true},
				{ID: "APP_CREATED", ObjectType: "apps", GoodForWebhook: true},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	webhook := NewWebhook(NewTestClientConfig(server.URL))
	ids, err := webhook.ListWebhookEventIDs(context.Background())
	if err != nil {
		t.Fatalf("ListWebhookEventIDs failed: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("Expected 2 IDs, got %d", len(ids))
	}
	if ids[0] != "ACCOUNT_MODIFIED" || ids[1] != "APP_CREATED" {
		t.Errorf("Unexpected IDs: %v", ids)
	}
}
