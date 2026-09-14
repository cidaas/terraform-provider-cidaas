package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVerificationOptionsCreateGet(t *testing.T) {
	t.Parallel()
	const id = "vo-uuid-1"
	var createBody VerificationOptionsModel
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/verification-options/"):
			if err := json.NewDecoder(r.Body).Decode(&createBody); err != nil {
				t.Fatal(err)
			}
			if createBody.VerificationOptions == nil {
				t.Fatal("missing verification_options")
			}
			if createBody.VerificationOptions.SuggestVerificationMethodID != "svm-1" {
				t.Fatalf("suggest id=%q", createBody.VerificationOptions.SuggestVerificationMethodID)
			}
			if createBody.VerificationOptions.Setting != "ALWAYS" {
				t.Fatalf("setting=%q", createBody.VerificationOptions.Setting)
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(VerificationOptionsResponse{
				Success: true,
				Status:  201,
				Data: VerificationOptionsModel{
					ID:                  id,
					Name:                createBody.Name,
					Description:         createBody.Description,
					Owner:               "client",
					VerificationOptions: createBody.VerificationOptions,
				},
			})
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, id):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(VerificationOptionsResponse{
				Success: true,
				Status:  200,
				Data: VerificationOptionsModel{
					ID:   id,
					Name: "tf-options",
					VerificationOptions: &VerificationOptions{
						Setting:                     "ALWAYS",
						AllowedMethods:              []string{"EMAIL", "SMS"},
						SuggestVerificationMethodID: "svm-1",
					},
				},
			})
		default:
			http.Error(w, "not found "+r.URL.Path, http.StatusNotFound)
		}
	}))
	defer srv.Close()

	svc := NewVerificationOptionsService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	out, err := svc.Create(context.Background(), VerificationOptionsModel{
		Name:        "tf-options",
		Description: "acc options",
		VerificationOptions: &VerificationOptions{
			Setting:                     "ALWAYS",
			AllowedMethods:              []string{"EMAIL", "SMS"},
			SuggestVerificationMethodID: "svm-1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Data.ID != id {
		t.Fatalf("%+v", out.Data)
	}
	got, err := svc.Get(context.Background(), id)
	if err != nil || got.Data.Name != "tf-options" {
		t.Fatalf("get: err=%v data=%+v", err, got)
	}
}

func TestVerificationOptionsPutDelete(t *testing.T) {
	t.Parallel()
	const id = "vo-uuid-2"
	mux := http.NewServeMux()
	mux.HandleFunc("/verification-actions-srv/verification-options/"+id, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			b, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(b), `"setting":"OFF"`) {
				t.Fatalf("body=%s", b)
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(VerificationOptionsResponse{
				Success: true,
				Status:  200,
				Data: VerificationOptionsModel{
					ID:   id,
					Name: "updated",
					VerificationOptions: &VerificationOptions{
						Setting:        "OFF",
						AllowedMethods: []string{},
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

	svc := NewVerificationOptionsService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	updated, err := svc.Update(context.Background(), id, VerificationOptionsModel{
		Name: "updated",
		VerificationOptions: &VerificationOptions{
			Setting:        "OFF",
			AllowedMethods: []string{},
		},
	})
	if err != nil || updated.Data.VerificationOptions.Setting != "OFF" {
		t.Fatalf("update: err=%v data=%+v", err, updated)
	}
	if err := svc.Delete(context.Background(), id); err != nil {
		t.Fatal(err)
	}
}

func TestVerificationOptionsGetMissingAsNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"success":false,"status":400,"error":"no verificationOptions found","code":"VAC1158"}`))
	}))
	defer srv.Close()

	svc := NewVerificationOptionsService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	_, err := svc.Get(context.Background(), "missing-id")
	if err == nil || !strings.Contains(err.Error(), "resource not found") {
		t.Fatalf("err=%v", err)
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
