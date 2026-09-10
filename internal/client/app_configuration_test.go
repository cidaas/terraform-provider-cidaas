package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAppConfigurationCreateGetUpdateDelete(t *testing.T) {
	t.Parallel()
	const clientID = "550e8400-e29b-41d4-a716-446655440010"
	var createBody AppConfigurationModel
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/app-srv/apps/":
			if err := json.NewDecoder(r.Body).Decode(&createBody); err != nil {
				t.Fatal(err)
			}
			if createBody.ClientName != "tf-app-test" {
				t.Fatalf("client_name=%q", createBody.ClientName)
			}
			if createBody.Owner != OwnerClient {
				t.Fatalf("owner=%q want %q", createBody.Owner, OwnerClient)
			}
			if createBody.Scopes == nil || createBody.Scopes.AllowedScopes[0] != "openid" {
				t.Fatalf("scopes=%v", createBody.Scopes)
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(AppConfigurationResponse{
				Success: true,
				Status:  200,
				Data: AppConfigurationModel{
					ID:         "doc-id",
					ClientID:   clientID,
					ClientName: createBody.ClientName,
					ClientType: createBody.ClientType,
					Scopes:     createBody.Scopes,
					SigningKeyConfig: &SigningKeyConfig{
						ActiveKID: "kid-1",
					},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/app-srv/apps/"+clientID:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(AppConfigurationResponse{
				Success: true,
				Status:  200,
				Data: AppConfigurationModel{
					ClientID:   clientID,
					ClientName: "tf-app-test",
					ClientType: "NON_INTERACTIVE",
					Scopes:     &ScopesConfig{AllowedScopes: []string{"openid"}},
				},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/app-srv/apps/"+clientID:
			b, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(b), `"client_name":"updated"`) {
				t.Fatalf("body=%s", b)
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(AppConfigurationResponse{
				Success: true,
				Status:  200,
				Data: AppConfigurationModel{
					ClientID:   clientID,
					ClientName: "updated",
					ClientType: "NON_INTERACTIVE",
					Scopes:     &ScopesConfig{AllowedScopes: []string{"openid", "profile"}},
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/app-srv/apps/"+clientID:
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	svc := NewAppConfigurationService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	enabled := true
	out, err := svc.Create(context.Background(), AppConfigurationModel{
		ClientName: "tf-app-test",
		ClientType: "NON_INTERACTIVE",
		Owner:      OwnerClient,
		Enabled:    &enabled,
		GrantTypes: []string{"client_credentials"},
		Scopes:     &ScopesConfig{AllowedScopes: []string{"openid"}},
		OwnershipDetails: &OwnershipDetailsConfig{
			CompanyName:    "Test Corp",
			CompanyAddress: "1 Test Way",
			CompanyWebsite: "https://example.com",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Data.ClientID != clientID {
		t.Fatalf("client_id=%q", out.Data.ClientID)
	}

	got, err := svc.Get(context.Background(), clientID)
	if err != nil || got.Data.ClientName != "tf-app-test" {
		t.Fatalf("get: err=%v data=%+v", err, got)
	}

	updated, err := svc.Update(context.Background(), clientID, AppConfigurationModel{
		ClientName: "updated",
		ClientType: "NON_INTERACTIVE",
		Scopes:     &ScopesConfig{AllowedScopes: []string{"openid", "profile"}},
		OwnershipDetails: &OwnershipDetailsConfig{
			CompanyName:    "Test Corp",
			CompanyAddress: "1 Test Way",
			CompanyWebsite: "https://example.com",
		},
	})
	if err != nil || updated.Data.ClientName != "updated" {
		t.Fatalf("update: err=%v data=%+v", err, updated)
	}
	if err := svc.Delete(context.Background(), clientID); err != nil {
		t.Fatal(err)
	}
}

func TestAppConfigurationGetNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	svc := NewAppConfigurationService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	_, err := svc.Get(context.Background(), "missing")
	if err == nil || !strings.Contains(err.Error(), "resource not found") {
		t.Fatalf("err=%v", err)
	}
}
