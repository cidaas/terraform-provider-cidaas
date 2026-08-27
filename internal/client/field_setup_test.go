package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFieldSetupList(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/fieldsetup-srv/graph/fields" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(fieldSetupListResponse{
			Success: true,
			Status:  200,
			Data: []FieldSetupEntry{
				{FieldKey: "email", Enabled: true},
				{FieldKey: "given_name", Enabled: true},
				{FieldKey: "disabled_field", Enabled: false},
			},
		})
	}))
	defer srv.Close()

	svc := NewFieldSetupService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	fields, err := svc.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 3 || fields[0].FieldKey != "email" {
		t.Fatalf("%+v", fields)
	}
}
