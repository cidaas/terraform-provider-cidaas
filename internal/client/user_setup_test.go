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

func TestUserSetupCreateAndGet(t *testing.T) {
	t.Parallel()
	const setupID = "550e8400-e29b-41d4-a716-446655440001"
	var createBody UserAppSetupModel
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/user-srv/usersetup":
			if err := json.NewDecoder(r.Body).Decode(&createBody); err != nil {
				t.Fatal(err)
			}
			if createBody.UserSetup.ConsentRefs == nil || createBody.UserSetup.ConsentRefs[0] != "consent-uuid" {
				t.Fatalf("consent_refs=%v", createBody.UserSetup.ConsentRefs)
			}
			if createBody.UserSetup.AllowedFields[0] != "email" {
				t.Fatalf("allowed_fields=%v", createBody.UserSetup.AllowedFields)
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(UserAppSetupResponse{
				Success: true,
				Status:  201,
				Data: UserAppSetupModel{
					ID:          setupID,
					Name:        createBody.Name,
					Description: createBody.Description,
					UserSetup:   createBody.UserSetup,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/user-srv/usersetup/"+setupID:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(UserAppSetupResponse{
				Success: true,
				Status:  200,
				Data: UserAppSetupModel{
					ID:   setupID,
					Name: "test-setup",
					UserSetup: UserSetupDetail{
						AllowedFields:  []string{"email"},
						RequiredFields: []string{"email"},
						AllowLoginWith: []string{"EMAIL"},
					},
				},
			})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	svc := NewUserSetupService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	dedup := true
	out, err := svc.Create(context.Background(), UserAppSetupModel{
		Name:        "test-setup",
		Description: "desc",
		UserSetup: UserSetupDetail{
			AllowedFields:       []string{"email"},
			RequiredFields:      []string{"email"},
			AllowLoginWith:      []string{"EMAIL"},
			ConsentRefs:         []string{"consent-uuid"},
			EnableDeduplication: &dedup,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Data.ID != setupID {
		t.Fatalf("id=%q", out.Data.ID)
	}

	got, err := svc.Get(context.Background(), setupID)
	if err != nil || got.Data.Name != "test-setup" {
		t.Fatalf("get: err=%v data=%+v", err, got)
	}
}

func TestUserSetupPatchAndDelete(t *testing.T) {
	t.Parallel()
	const setupID = "550e8400-e29b-41d4-a716-446655440002"
	mux := http.NewServeMux()
	mux.HandleFunc("/user-srv/usersetup/"+setupID, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			b, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(b), `"name":"updated"`) {
				t.Fatalf("body=%s", b)
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(UserAppSetupResponse{
				Success: true,
				Status:  200,
				Data: UserAppSetupModel{
					ID:   setupID,
					Name: "updated",
					UserSetup: UserSetupDetail{
						AllowedFields:  []string{"email"},
						RequiredFields: []string{"email"},
					},
				},
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method", http.StatusMethodNotAllowed)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	svc := NewUserSetupService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	updated, err := svc.Update(context.Background(), setupID, UserAppSetupModel{
		Name: "updated",
		UserSetup: UserSetupDetail{
			AllowedFields:  []string{"email"},
			RequiredFields: []string{"email"},
		},
	})
	if err != nil || updated.Data.Name != "updated" {
		t.Fatalf("update: err=%v data=%+v", err, updated)
	}
	if err := svc.Delete(context.Background(), setupID); err != nil {
		t.Fatal(err)
	}
}

func TestUserSetupGetNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	svc := NewUserSetupService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	_, err := svc.Get(context.Background(), "missing")
	if err == nil || !strings.Contains(err.Error(), "resource not found") {
		t.Fatalf("err=%v", err)
	}
}
