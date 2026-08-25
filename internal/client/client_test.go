package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMajorVersion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want int
		ok   bool
	}{
		{"4.0.2", 4, true},
		{"v3.102.8", 3, true},
		{"5", 5, true},
		{"", 0, false},
		{"abc", 0, false},
	}
	for _, tc := range cases {
		got, err := majorVersion(tc.in)
		if tc.ok && err != nil {
			t.Fatalf("%q: unexpected err %v", tc.in, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%q: expected error", tc.in)
		}
		if tc.ok && got != tc.want {
			t.Fatalf("%q: got %d want %d", tc.in, got, tc.want)
		}
	}
}

func TestProbeCapabilities_V4(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/public-srv/version" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data":    map[string]string{"version": "4.0.2"},
		})
	}))
	defer srv.Close()

	caps, err := probeCapabilities(context.Background(), Config{BaseURL: srv.URL, AccessToken: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if !caps.SupportsV4 || caps.Version != "4.0.2" {
		t.Fatalf("got %+v", caps)
	}
}

func TestProbeCapabilities_V3(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data":    map[string]string{"version": "3.102.8"},
		})
	}))
	defer srv.Close()

	caps, err := probeCapabilities(context.Background(), Config{BaseURL: srv.URL, AccessToken: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if caps.SupportsV4 {
		t.Fatalf("expected SupportsV4 false, got %+v", caps)
	}
}

func TestProbeCapabilities_NotFoundAssumesV4(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	caps, err := probeCapabilities(context.Background(), Config{BaseURL: srv.URL, AccessToken: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if !caps.SupportsV4 {
		t.Fatalf("expected SupportsV4 true on 404, got %+v", caps)
	}
}

func TestThemeUpload(t *testing.T) {
	t.Parallel()
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("parse multipart: %v", err)
			http.Error(w, err.Error(), 400)
			return
		}
		f, hdr, err := r.FormFile("theme")
		if err != nil {
			t.Errorf("form file: %v", err)
			http.Error(w, err.Error(), 400)
			return
		}
		defer f.Close()
		if hdr.Filename != "x.css" {
			t.Errorf("filename %q", hdr.Filename)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"status":200}`))
	}))
	defer srv.Close()

	svc := NewThemeService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	if err := svc.Upload(context.Background(), "x.css", "body{}"); err != nil {
		t.Fatal(err)
	}
	if gotCT == "" || gotCT[:19] != "multipart/form-data" {
		t.Fatalf("content-type %q", gotCT)
	}
}

func TestTranslationsCreateUsesTranslationKey(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if _, ok := body["translation"]; !ok {
			t.Fatalf("missing translation key: %#v", body)
		}
		if body["locale"] != "fr" {
			t.Fatalf("locale=%v", body["locale"])
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"status":  201,
			"data": map[string]any{
				"locale":      "fr",
				"enabled":     true,
				"translation": body["translation"],
			},
		})
	}))
	defer srv.Close()

	svc := NewTranslationsService(Config{BaseURL: srv.URL, AccessToken: "t"})
	out, err := svc.Create(context.Background(), TranslationModel{
		Locale:      "fr",
		Enabled:     true,
		Translation: map[string]any{"login.title": "Bonjour"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Data.Locale != "fr" {
		t.Fatalf("%+v", out.Data)
	}
}

func TestHostedPageGroupUpsert(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body HostedPageGroupModel
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.ID != "g1" {
			t.Fatalf("id=%q", body.ID)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(HostedPageGroupResponse{
			Success: true,
			Status:  201,
			Data:    body,
		})
	}))
	defer srv.Close()

	svc := NewHostedPageGroupService(Config{BaseURL: srv.URL, AccessToken: "t"})
	out, err := svc.Upsert(context.Background(), HostedPageGroupModel{
		ID:            "g1",
		GroupOwner:    "client",
		DefaultLocale: "en",
		HostedPages: []HostedPageData{{
			HostedPageID: "login",
			Locale:       "en",
			URL:          "https://example.com/login",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Data.ID != "g1" {
		t.Fatalf("%+v", out.Data)
	}
}
