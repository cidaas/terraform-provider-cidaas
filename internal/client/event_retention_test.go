package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

const retentionPath = "/webhook-srv/eventdescriptions/retention"

func TestEventRetentionListAndUpdate(t *testing.T) {
	t.Parallel()
	var patchBody map[string]int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == retentionPath:
			_ = json.NewEncoder(w).Encode(retentionCatalogResponse{
				Success: true,
				Status:  200,
				Data: []RetentionEventDescription{{
					ID:               "USER_CREATED",
					ObjectType:       "USER",
					GoodForChangelog: true,
					DaysToLive:       90,
					Retention:        &RetentionBounds{DefaultDaysToLive: 90, MinDaysToLive: 30, MaxDaysToLive: 365},
				}},
			})
		case r.Method == http.MethodPatch && r.URL.Path == retentionPath:
			if err := json.NewDecoder(r.Body).Decode(&patchBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(retentionUpdateResponse{
				Success: true,
				Status:  200,
				Data:    []EventRetentionItem{{EventID: "USER_CREATED", DaysToLive: patchBody["USER_CREATED"]}},
			})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	svc := NewEventRetentionService(Config{BaseURL: srv.URL, AccessToken: "t"})
	eds, err := svc.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(eds) != 1 || eds[0].Retention == nil || eds[0].Retention.MaxDaysToLive != 365 {
		t.Fatalf("catalog=%+v", eds)
	}

	items, err := svc.Update(context.Background(), map[string]int{"USER_CREATED": 120})
	if err != nil {
		t.Fatal(err)
	}
	if patchBody["USER_CREATED"] != 120 || len(items) != 1 || items[0].DaysToLive != 120 {
		t.Fatalf("patch=%v items=%+v", patchBody, items)
	}
}

func TestEventRetentionListNoContent(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	eds, err := NewEventRetentionService(Config{BaseURL: srv.URL}).List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(eds) != 0 {
		t.Fatalf("expected empty catalog, got %+v", eds)
	}
}

func TestEventRetentionInactive(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"success":false,"status":403,"error":{"error":"retention is not active"}}`))
	}))
	defer srv.Close()

	_, err := NewEventRetentionService(Config{BaseURL: srv.URL}).List(context.Background())
	if !errors.Is(err, ErrRetentionInactive) {
		t.Fatalf("expected ErrRetentionInactive, got %v", err)
	}
}
