package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHostedPageGroupUpsertAndGet(t *testing.T) {
	t.Parallel()
	const groupID = "tf-hpgroup-1"
	var upsertBody HostedPageGroupModel
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/hostedpages-srv/hpgroup":
			if err := json.NewDecoder(r.Body).Decode(&upsertBody); err != nil {
				t.Fatal(err)
			}
			if upsertBody.ID != groupID {
				t.Fatalf("_id=%q", upsertBody.ID)
			}
			if upsertBody.DefaultLocale != "en" {
				t.Fatalf("default_locale=%q", upsertBody.DefaultLocale)
			}
			if len(upsertBody.HostedPages) != 1 || upsertBody.HostedPages[0].HostedPageID != "login" {
				t.Fatalf("hosted_pages=%+v", upsertBody.HostedPages)
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(HostedPageGroupResponse{
				Success: true,
				Status:  201,
				Data: HostedPageGroupModel{
					ID:            upsertBody.ID,
					GroupOwner:    upsertBody.GroupOwner,
					DefaultLocale: upsertBody.DefaultLocale,
					HostedPages:   upsertBody.HostedPages,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/hostedpages-srv/hpgroup/"+groupID:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(HostedPageGroupResponse{
				Success: true,
				Status:  200,
				Data: HostedPageGroupModel{
					ID:            groupID,
					GroupOwner:    "client",
					DefaultLocale: "en",
					HostedPages: []HostedPageData{
						{HostedPageID: "login", Locale: "en", URL: "https://example.com/login"},
					},
				},
			})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	svc := NewHostedPageGroupService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	out, err := svc.Upsert(context.Background(), HostedPageGroupModel{
		ID:            groupID,
		GroupOwner:    "client",
		DefaultLocale: "en",
		HostedPages: []HostedPageData{
			{HostedPageID: "login", Locale: "en", URL: "https://example.com/login"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Data.ID != groupID {
		t.Fatalf("%+v", out.Data)
	}

	got, err := svc.Get(context.Background(), groupID)
	if err != nil || got.Data.DefaultLocale != "en" {
		t.Fatalf("get: err=%v data=%+v", err, got)
	}
}

func TestHostedPageGroupDelete(t *testing.T) {
	t.Parallel()
	const groupID = "tf-hpgroup-del"
	mux := http.NewServeMux()
	mux.HandleFunc("/hostedpages-srv/hpgroup/"+groupID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	svc := NewHostedPageGroupService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	if err := svc.Delete(context.Background(), groupID); err != nil {
		t.Fatal(err)
	}
}

func TestHostedPageGroupGetNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	svc := NewHostedPageGroupService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	_, err := svc.Get(context.Background(), "missing")
	if err == nil || !strings.Contains(err.Error(), "resource not found") {
		t.Fatalf("err=%v", err)
	}
}
