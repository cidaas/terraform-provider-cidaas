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
