package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVerificationMethodListAndGet(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/verification-actions-srv/config/" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"success":true,"status":200,"data":[
			{"_id":"1","verificationType":"EMAIL","owner":"system","active":true,"factor":"POSSESSION","code":{"length":6,"validity":1800},"name":"Email"},
			{"_id":"2","verificationType":"MY_SPI","owner":"CLIENT","active":false,"url":"https://spi.example.com","auth_type":"Bearer","cidaasAuthDetails":{"client_id":"c1"}}
		]}`))
	}))
	defer srv.Close()
	svc := NewVerificationMethodService(Config{BaseURL: srv.URL, AccessToken: "t"})

	list, err := svc.List(context.Background())
	if err != nil || len(list) != 2 {
		t.Fatalf("List: %v, %d items", err, len(list))
	}
	email, err := svc.Get(context.Background(), "EMAIL")
	if err != nil {
		t.Fatal(err)
	}
	if email.IsPluggable() || email.Code == nil || email.Code.Length != 6 || email.Name != "Email" {
		t.Fatalf("unexpected EMAIL: %+v", email)
	}
	spi, err := svc.Get(context.Background(), "MY_SPI")
	if err != nil {
		t.Fatal(err)
	}
	if !spi.IsPluggable() || spi.CidaasAuthDetails.ClientID != "c1" || *spi.Active {
		t.Fatalf("unexpected MY_SPI: %+v", spi)
	}
	if _, err := svc.Get(context.Background(), "NOPE"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestVerificationMethodListNoContentIsError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	svc := NewVerificationMethodService(Config{BaseURL: srv.URL, AccessToken: "t"})
	if _, err := svc.List(context.Background()); err == nil || !strings.Contains(err.Error(), "204") {
		t.Fatalf("expected 204 error, got %v", err)
	}
}

func TestVerificationMethodCreateUpdateDelete(t *testing.T) {
	t.Parallel()
	var posted, put map[string]any
	var deletedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			_ = json.NewDecoder(r.Body).Decode(&posted)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"success":true,"status":201,"data":{"verificationType":"MY_SPI","owner":"CLIENT"}}`))
		case http.MethodPut:
			_ = json.NewDecoder(r.Body).Decode(&put)
			_, _ = w.Write([]byte(`{"success":true,"status":200,"data":{"verificationType":"EMAIL","owner":"system","active":false}}`))
		case http.MethodDelete:
			deletedPath = r.URL.Path
			_, _ = w.Write([]byte(`{"success":true,"status":200,"data":true}`))
		}
	}))
	defer srv.Close()
	svc := NewVerificationMethodService(Config{BaseURL: srv.URL, AccessToken: "t"})

	active := true
	created, err := svc.Create(context.Background(), VerificationMethodModel{
		VerificationType:  "MY_SPI",
		Active:            &active,
		URL:               "https://spi.example.com",
		AuthType:          "Bearer",
		CidaasAuthDetails: &VerificationMethodAuth{ClientID: "c1"},
		AMR:               []string{"spi"},
	})
	if err != nil || !created.IsPluggable() {
		t.Fatalf("Create: %v %+v", err, created)
	}
	if posted["url"] != "https://spi.example.com" || posted["auth_type"] != "Bearer" || posted["cidaasAuthDetails"].(map[string]any)["client_id"] != "c1" {
		t.Fatalf("unexpected POST body: %v", posted)
	}
	for _, key := range []string{"_id", "owner", "code", "requiresAuthenticatorApp", "name", "createdTime"} {
		if _, ok := posted[key]; ok {
			t.Fatalf("read-only %q must not be sent: %v", key, posted)
		}
	}

	inactive := false
	if _, err := svc.Update(context.Background(), VerificationMethodModel{
		VerificationType: "EMAIL",
		Active:           &inactive,
		LocaleText:       []VerificationMethodLocale{{Locale: "en", Name: "Email"}},
	}); err != nil {
		t.Fatal(err)
	}
	if put["verificationType"] != "EMAIL" || put["active"] != false || len(put["localeText"].([]any)) != 1 {
		t.Fatalf("unexpected PUT body: %v", put)
	}

	if err := svc.Delete(context.Background(), "MY SPI"); err != nil {
		t.Fatal(err)
	}
	if deletedPath != "/verification-actions-srv/config/MY SPI" {
		t.Fatalf("unexpected DELETE path %q", deletedPath)
	}
}

func TestVerificationMethodMissingMapsToNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"success":false,"status":400,"error":{"error":"verification method not found"}}`))
	}))
	defer srv.Close()
	svc := NewVerificationMethodService(Config{BaseURL: srv.URL, AccessToken: "t"})
	if _, err := svc.Update(context.Background(), VerificationMethodModel{VerificationType: "X"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Update: expected ErrNotFound, got %v", err)
	}
	if err := svc.Delete(context.Background(), "X"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete: expected ErrNotFound, got %v", err)
	}
}
