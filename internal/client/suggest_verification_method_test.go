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

func TestSuggestVerificationMethodCreateGet(t *testing.T) {
	t.Parallel()
	const id = "svm-uuid-1"
	var createBody SuggestVerificationMethodModel
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/suggest-verification-configs/"):
			if err := json.NewDecoder(r.Body).Decode(&createBody); err != nil {
				t.Fatal(err)
			}
			if createBody.SuggestVerificationMethod == nil || createBody.SuggestVerificationMethod.MandatoryConfig == nil {
				t.Fatalf("missing suggest_verification_method: %+v", createBody)
			}
			if createBody.SuggestVerificationMethod.MandatoryConfig.Range != "ONEOF" {
				t.Fatalf("range=%q", createBody.SuggestVerificationMethod.MandatoryConfig.Range)
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(SuggestVerificationMethodResponse{
				Success: true,
				Status:  201,
				Data: SuggestVerificationMethodModel{
					ID:                        id,
					Name:                      createBody.Name,
					Description:               createBody.Description,
					Owner:                     "client",
					SuggestVerificationMethod: createBody.SuggestVerificationMethod,
				},
			})
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, id):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(SuggestVerificationMethodResponse{
				Success: true,
				Status:  200,
				Data: SuggestVerificationMethodModel{
					ID:   id,
					Name: "tf-suggest",
					SuggestVerificationMethod: &SuggestVerificationMethodSetup{
						MandatoryConfig:    &SuggestVerificationMandatoryConfig{Methods: []string{"EMAIL"}, Range: "ONEOF"},
						SkipDurationInDays: 7,
					},
				},
			})
		default:
			http.Error(w, "not found "+r.URL.Path, http.StatusNotFound)
		}
	}))
	defer srv.Close()

	svc := NewSuggestVerificationMethodService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	out, err := svc.Create(context.Background(), SuggestVerificationMethodModel{
		Name:        "tf-suggest",
		Description: "acc suggest",
		SuggestVerificationMethod: &SuggestVerificationMethodSetup{
			MandatoryConfig:    &SuggestVerificationMandatoryConfig{Methods: []string{"EMAIL"}, Range: "ONEOF"},
			SkipDurationInDays: 7,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Data.ID != id || out.Data.Owner != "client" {
		t.Fatalf("%+v", out.Data)
	}
	got, err := svc.Get(context.Background(), id)
	if err != nil || got.Data.Name != "tf-suggest" {
		t.Fatalf("get: err=%v data=%+v", err, got)
	}
}

func TestSuggestVerificationMethodPutDelete(t *testing.T) {
	t.Parallel()
	const id = "svm-uuid-2"
	mux := http.NewServeMux()
	mux.HandleFunc("/verification-actions-srv/suggest-verification-configs/"+id, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			b, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(b), `"name":"updated"`) {
				t.Fatalf("body=%s", b)
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(SuggestVerificationMethodResponse{
				Success: true,
				Status:  200,
				Data: SuggestVerificationMethodModel{
					ID:   id,
					Name: "updated",
					SuggestVerificationMethod: &SuggestVerificationMethodSetup{
						MandatoryConfig: &SuggestVerificationMandatoryConfig{Methods: []string{"EMAIL"}, Range: "ONEOF"},
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

	svc := NewSuggestVerificationMethodService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	updated, err := svc.Update(context.Background(), id, SuggestVerificationMethodModel{
		Name: "updated",
		SuggestVerificationMethod: &SuggestVerificationMethodSetup{
			MandatoryConfig: &SuggestVerificationMandatoryConfig{Methods: []string{"EMAIL"}, Range: "ONEOF"},
		},
	})
	if err != nil || updated.Data.Name != "updated" {
		t.Fatalf("update: err=%v data=%+v", err, updated)
	}
	if err := svc.Delete(context.Background(), id); err != nil {
		t.Fatal(err)
	}
}
