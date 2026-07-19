package cidaas

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNotificationsSrvServiceSetup_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/servicesetups/ss-1") {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"status":  200,
			"data": map[string]any{
				"_id":    "ss-1",
				"status": "active",
				"name":   "Twilio SMS",
				"serviceDescInfo": map[string]any{
					"serviceId":       "twilio-sms",
					"serviceCategory": "comm_prov",
					"commProv": map[string]any{
						"commMethods": []string{"sms"},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewNotificationsSrvServiceSetup(NewTestClientConfig(server.URL))
	got, err := client.Get(context.Background(), "ss-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Status != "active" || got.ID != "ss-1" {
		t.Fatalf("unexpected %+v", got)
	}
}

func TestNotificationsSrvServiceSetup_Create_NoSaasEntity(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"status":  201,
			"data": map[string]any{
				"_id":    "ss-new",
				"status": "in-progress",
				"name":   "Twilio SMS",
				"serviceDescInfo": map[string]any{
					"serviceId": "twilio-sms",
					"commProv": map[string]any{
						"commMethods": []string{"sms"},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewNotificationsSrvServiceSetup(NewTestClientConfig(server.URL))
	_, err := client.Create(context.Background(), NotificationsSrvServiceSetupWrite{
		Name: "Twilio SMS",
		ServiceDescInfo: NotificationsSrvServiceDescInfo{
			ServiceID:       "twilio-sms",
			ServiceCategory: "comm_prov",
			CommProv: NotificationsSrvCommProvider{
				CommMethods: []string{"sms"},
			},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	setup, ok := body["serviceSetup"].(map[string]any)
	if !ok {
		t.Fatalf("missing serviceSetup in body: %v", body)
	}
	if _, has := setup["saasEntity"]; has {
		t.Fatal("terraform helper must not send saasEntity")
	}
	if setup["status"] != serviceSetupStatusInProgress {
		t.Fatalf("expected in-progress status, got %v", setup["status"])
	}
}

func TestNotificationsSrvServiceSetup_UpdateAndDelete(t *testing.T) {
	var patchBody map[string]any
	var deleteBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &patchBody)
		case http.MethodDelete:
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &deleteBody)
			w.WriteHeader(http.StatusOK)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"status":  200,
			"data": map[string]any{
				"_id":    "ss-1",
				"status": "in-progress",
				"name":   "updated",
				"serviceDescInfo": map[string]any{
					"serviceId": "twilio-sms",
				},
			},
		})
	}))
	defer server.Close()

	client := NewNotificationsSrvServiceSetup(NewTestClientConfig(server.URL))
	_, err := client.Update(context.Background(), NotificationsSrvServiceSetupUpdate{
		ID:   "ss-1",
		Name: "updated",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if patchBody["serviceSetup"].(map[string]any)["_id"] != "ss-1" {
		t.Fatalf("unexpected patch body %v", patchBody)
	}
	if err := client.Delete(context.Background(), "ss-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if deleteBody["id"] != "ss-1" {
		t.Fatalf("unexpected delete body %v", deleteBody)
	}
}

func TestNotificationsSrvServiceSetup_DeleteNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"success":false,"status":404,"error":"resource not found","code":"35010"}`))
	}))
	defer server.Close()

	client := NewNotificationsSrvServiceSetup(NewTestClientConfig(server.URL))
	err := client.Delete(context.Background(), "gone")
	if err == nil {
		t.Fatal("expected not-found error")
	}
	if !strings.Contains(err.Error(), "resource not found") {
		t.Fatalf("expected resource not found, got: %v", err)
	}
}

func TestNotificationsSrvProviderConfig_CreateSchemaData(t *testing.T) {
	var posted NotificationsSrvProviderConfigModel
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &posted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"status":  201,
			"data": map[string]any{
				"_id": "ss-1",
				"configData": map[string]any{
					"commProvider": "custom-twilio-sms",
					"commMethod":   "sms",
					"schemaData": map[string]any{
						"accountSid": "AC",
						"authToken":  "secret",
					},
				},
			},
		})
	}))
	defer server.Close()

	configJSON := `{"commProvider":"custom-twilio-sms","commMethod":"sms","schemaData":{"accountSid":"AC","authToken":"secret"}}`
	client := NewNotificationsSrvProviderConfig(NewTestClientConfig(server.URL))
	_, err := client.Create(context.Background(), "ss-1", json.RawMessage(configJSON))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if posted.ID != "ss-1" {
		t.Fatalf("expected _id ss-1, got %s", posted.ID)
	}
	if !strings.Contains(string(posted.ConfigData), "schemaData") {
		t.Fatalf("expected schemaData in configData: %s", posted.ConfigData)
	}
}

func TestNotificationsSrvServiceSetup_StatusRefresh(t *testing.T) {
	status := "in-progress"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"status":  200,
			"data": map[string]any{
				"_id":    "ss-1",
				"status": status,
				"name":   "Twilio SMS",
				"serviceDescInfo": map[string]any{
					"serviceId": "twilio-sms",
					"commProv": map[string]any{
						"commMethods": []string{"sms"},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewNotificationsSrvServiceSetup(NewTestClientConfig(server.URL))
	first, err := client.Get(context.Background(), "ss-1")
	if err != nil || first.Status != "in-progress" {
		t.Fatalf("first read: %v %+v", err, first)
	}
	status = "active"
	second, err := client.Get(context.Background(), "ss-1")
	if err != nil || second.Status != "active" {
		t.Fatalf("second read: %v %+v", err, second)
	}
}
