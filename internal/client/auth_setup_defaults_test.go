package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthSetupDefaultsGetUpdate(t *testing.T) {
	t.Parallel()
	var putBody AuthSetupDefaultsEntity
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/app-srv/apps/auth-setup-defaults":
			trueVal := true
			falseVal := false
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(AuthSetupDefaultsResponse{
				Success: true,
				Status:  200,
				Data: AuthSetupDefaultsEntity{
					ID:          "default",
					Name:        "default",
					Description: "seeded",
					AuthSetupDefaults: &AuthSetupDefaults{
						AutoLoginAfterRegister:       &falseVal,
						RegisterWithLoginInformation: &falseVal,
						EnablePasswordLessAuth:       &trueVal,
						AllowUserLevelMultiProvider:  &trueVal,
						SocialBusinessIDs:            &falseVal,
						NetID:                        &falseVal,
					},
				},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/app-srv/apps/auth-setup-defaults":
			if err := json.NewDecoder(r.Body).Decode(&putBody); err != nil {
				t.Fatal(err)
			}
			if putBody.AuthSetupDefaults == nil || putBody.AuthSetupDefaults.EnablePasswordLessAuth == nil || *putBody.AuthSetupDefaults.EnablePasswordLessAuth {
				t.Fatalf("put body=%+v", putBody.AuthSetupDefaults)
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(AuthSetupDefaultsResponse{
				Success: true,
				Status:  200,
				Data:    putBody,
			})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	svc := NewAuthSetupDefaultsService(Config{BaseURL: srv.URL, AccessToken: "t"})
	got, err := svc.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Data.ID != "default" || got.Data.AuthSetupDefaults == nil || !*got.Data.AuthSetupDefaults.EnablePasswordLessAuth {
		t.Fatalf("get=%+v", got.Data)
	}

	falseVal := false
	trueVal := true
	updated, err := svc.Update(context.Background(), AuthSetupDefaultsEntity{
		Name:        "default",
		Description: "edited",
		AuthSetupDefaults: &AuthSetupDefaults{
			AutoLoginAfterRegister:       &trueVal,
			RegisterWithLoginInformation: &trueVal,
			EnablePasswordLessAuth:       &falseVal,
			AllowUserLevelMultiProvider:  &falseVal,
			SocialBusinessIDs:            &trueVal,
			NetID:                        &trueVal,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Data.Description != "edited" || putBody.Description != "edited" {
		t.Fatalf("update=%+v put=%+v", updated.Data, putBody)
	}
}
