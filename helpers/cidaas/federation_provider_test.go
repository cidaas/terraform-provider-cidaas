package cidaas

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFederationProvider_CreateAndUpdatePayload(t *testing.T) {
	var createBody map[string]interface{}
	var updateBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if r.Method == http.MethodPost {
			_ = json.Unmarshal(raw, &createBody)
		} else if r.Method == http.MethodPut {
			_ = json.Unmarshal(raw, &updateBody)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"status":  200,
			"data": map[string]interface{}{
				"_id":           "fed-id-123",
				"provider_name": "jhentraid",
				"display_name":  "Jungheinrich Employee",
				"standard_type": "OPENID_CONNECT",
				"logo_url":      "https://example.com/jh-logo.svg",
				"owner":         "client",
				"provider":      "custom",
				"oauth2": map[string]interface{}{
					"client_id":              "client_id_val",
					"client_secret":          "client_secret_val",
					"authorization_endpoint": "https://auth.example.com/oauth2/v2.0/authorize",
					"token_endpoint":         "https://auth.example.com/oauth2/v2.0/token",
					"userinfo_endpoint":      "https://graph.example.com/oidc/userinfo",
				},
			},
		})
	}))
	defer server.Close()

	client := NewFederationProvider(NewTestClientConfig(server.URL))
	model := &ProviderConfigModel{
		ProviderName:          "jhentraid",
		DisplayName:           "Jungheinrich Employee",
		StandardType:          "OPENID_CONNECT",
		ClientID:              "client_id_val",
		ClientSecret:          "client_secret_val",
		AuthorizationEndpoint: "https://auth.example.com/oauth2/v2.0/authorize",
		TokenEndpoint:         "https://auth.example.com/oauth2/v2.0/token",
		UserinfoEndpoint:      "https://graph.example.com/oidc/userinfo",
		LogoURL:               "https://example.com/jh-logo.svg",
		Owner:                 "client",
		Domains:               []string{"example.com"},
	}

	// Test Create
	created, err := client.Create(context.Background(), model)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.Data.ClientID != "client_id_val" || created.Data.AuthorizationEndpoint != "https://auth.example.com/oauth2/v2.0/authorize" {
		t.Fatalf("Create unwrap failed: %+v", created.Data)
	}

	if createBody["provider"] != "custom" {
		t.Fatalf("expected provider=custom in Create body, got: %v", createBody["provider"])
	}
	oauth2Create, ok := createBody["oauth2"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected oauth2 object in Create body, got: %v", createBody)
	}
	if oauth2Create["client_id"] != "client_id_val" || oauth2Create["authorization_endpoint"] != "https://auth.example.com/oauth2/v2.0/authorize" {
		t.Fatalf("unexpected oauth2 payload in Create: %v", oauth2Create)
	}

	// Test Update
	updated, err := client.Update(context.Background(), "fed-id-123", model)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Data.ClientID != "client_id_val" {
		t.Fatalf("Update unwrap failed: %+v", updated.Data)
	}

	if updateBody["provider"] != "custom" {
		t.Fatalf("expected provider=custom in Update body, got: %v", updateBody["provider"])
	}
	if updateBody["id"] != "fed-id-123" {
		t.Fatalf("expected id=fed-id-123 in Update body, got: %v", updateBody["id"])
	}
	oauth2Update, ok := updateBody["oauth2"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected oauth2 object in Update body, got: %v", updateBody)
	}
	if oauth2Update["client_id"] != "client_id_val" || oauth2Update["authorization_endpoint"] != "https://auth.example.com/oauth2/v2.0/authorize" {
		t.Fatalf("unexpected oauth2 payload in Update: %v", oauth2Update)
	}
}
