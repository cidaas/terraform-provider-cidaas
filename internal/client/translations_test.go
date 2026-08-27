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

func TestTranslationsCreateAndGet(t *testing.T) {
	t.Parallel()
	const locale = "fr"
	var createBody TranslationModel
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/hostedpages-srv/translations":
			if err := json.NewDecoder(r.Body).Decode(&createBody); err != nil {
				t.Fatal(err)
			}
			if createBody.Locale != locale {
				t.Fatalf("locale=%q", createBody.Locale)
			}
			if createBody.Translation["login.title"] != "Connexion" {
				t.Fatalf("translation=%v", createBody.Translation)
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(TranslationResponse{
				Success: true,
				Status:  201,
				Data:    createBody,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/hostedpages-srv/translations/"+locale:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(TranslationResponse{
				Success: true,
				Status:  200,
				Data: TranslationModel{
					Locale:      locale,
					Enabled:     true,
					Translation: map[string]any{"login.title": "Connexion"},
				},
			})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	svc := NewTranslationsService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	out, err := svc.Create(context.Background(), TranslationModel{
		Locale:      locale,
		Enabled:     true,
		Translation: map[string]any{"login.title": "Connexion"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Data.Locale != locale {
		t.Fatalf("%+v", out.Data)
	}

	got, err := svc.Get(context.Background(), locale)
	if err != nil || got.Data.Translation["login.title"] != "Connexion" {
		t.Fatalf("get: err=%v data=%+v", err, got)
	}
}

func TestTranslationsUpdateAndDelete(t *testing.T) {
	t.Parallel()
	const locale = "de"
	mux := http.NewServeMux()
	mux.HandleFunc("/hostedpages-srv/translations/"+locale, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			b, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(b), `"locale":"de"`) {
				t.Fatalf("body=%s", b)
			}
			if !strings.Contains(string(b), `"login.title":"Anmelden"`) {
				t.Fatalf("body=%s", b)
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(TranslationResponse{
				Success: true,
				Status:  200,
				Data: TranslationModel{
					Locale:      locale,
					Enabled:     true,
					Translation: map[string]any{"login.title": "Anmelden"},
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

	svc := NewTranslationsService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	updated, err := svc.Update(context.Background(), locale, TranslationModel{
		Enabled:     true,
		Translation: map[string]any{"login.title": "Anmelden"},
	})
	if err != nil || updated.Data.Translation["login.title"] != "Anmelden" {
		t.Fatalf("update: err=%v data=%+v", err, updated)
	}
	if err := svc.Delete(context.Background(), locale); err != nil {
		t.Fatal(err)
	}
}
