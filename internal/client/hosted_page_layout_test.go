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

func TestHostedPageLayoutCreateJSON(t *testing.T) {
	t.Parallel()
	var body HostedPageLayoutWrite
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/hostedpages-srv/hosted-page-layouts" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if _, ok := body.Resources["default-hosted-pages-webapp"]; !ok {
			t.Fatalf("missing resources key: %#v", body.Resources)
		}
		if body.Resources["default-hosted-pages-webapp"].TranslationSet != "default" {
			t.Fatalf("translationSet=%q", body.Resources["default-hosted-pages-webapp"].TranslationSet)
		}
		if body.Layout.HostedPageGroup != "grp1" {
			t.Fatalf("hosted_page_group=%q", body.Layout.HostedPageGroup)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(HostedPageLayoutResponse{
			Success: true,
			Status:  201,
			Data: HostedPageLayoutModel{
				ID:          "layout-uuid",
				Description: body.Description,
				GroupID:     "from-token",
				Fingerprint: "fp",
				Layout:      body.Layout,
				Resources:   body.Resources,
				Owner:       "client",
			},
		})
	}))
	defer srv.Close()

	svc := NewHostedPageLayoutService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	out, err := svc.Create(context.Background(), HostedPageLayoutWrite{
		Description: "test layout",
		Layout: LayoutDetail{
			HostedPageGroup: "grp1",
			Theme:           "theme.css",
			MediaType:       "IMAGE",
			PrimaryColor:    "#ff0000",
			ContentAlign:    "CENTER",
		},
		Resources: map[string]LayoutResourceConfig{
			"default-hosted-pages-webapp": {TranslationSet: "default", Theme: "themeA"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Data.ID != "layout-uuid" {
		t.Fatalf("%+v", out.Data)
	}
	if out.Data.GroupID != "from-token" {
		t.Fatalf("groupId=%q", out.Data.GroupID)
	}
}

func TestHostedPageLayoutGetUpdateDelete(t *testing.T) {
	t.Parallel()
	const layoutID = "550e8400-e29b-41d4-a716-446655440000"
	mux := http.NewServeMux()
	mux.HandleFunc("/hostedpages-srv/hosted-page-layouts/"+layoutID, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(HostedPageLayoutResponse{
				Success: true,
				Status:  200,
				Data: HostedPageLayoutModel{
					ID:          layoutID,
					Description: "read",
					GroupID:     "g1",
					Layout:      LayoutDetail{HostedPageGroup: "grp", MediaType: "IMAGE"},
				},
			})
		case http.MethodPut:
			b, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(b), `"primary_color":"#222222"`) {
				t.Fatalf("body=%s", b)
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(HostedPageLayoutResponse{
				Success: true,
				Status:  200,
				Data: HostedPageLayoutModel{
					ID:          layoutID,
					Description: "updated",
					Layout:      LayoutDetail{PrimaryColor: "#222222", MediaType: "IMAGE"},
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

	svc := NewHostedPageLayoutService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	got, err := svc.Get(context.Background(), layoutID)
	if err != nil || got.Data.Description != "read" {
		t.Fatalf("get: err=%v data=%+v", err, got)
	}
	updated, err := svc.Update(context.Background(), layoutID, HostedPageLayoutWrite{
		Description: "updated",
		Layout:      LayoutDetail{PrimaryColor: "#222222", MediaType: "IMAGE"},
	})
	if err != nil || updated.Data.Description != "updated" {
		t.Fatalf("update: err=%v data=%+v", err, updated)
	}
	if err := svc.Delete(context.Background(), layoutID); err != nil {
		t.Fatal(err)
	}
}
