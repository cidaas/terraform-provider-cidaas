package resources

import (
	"encoding/json"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNotificationServiceSetupFromAPI_StatusComputed(t *testing.T) {
	t.Parallel()
	api := &cidaas.NotificationsSrvServiceSetupModel{
		ID:     "ss-1",
		Status: "active",
		Name:   "Twilio SMS",
		ServiceDescInfo: cidaas.NotificationsSrvServiceDescInfo{
			ServiceID:       "twilio-sms",
			ServiceCategory: "comm_prov",
			CommProv: cidaas.NotificationsSrvCommProvider{
				CommMethods: []string{"sms"},
			},
		},
	}
	state := notificationServiceSetupFromAPI(api, notificationServiceSetupModel{})
	if state.Status.ValueString() != "active" {
		t.Fatalf("status %q", state.Status.ValueString())
	}
	if state.ServiceID.ValueString() != "twilio-sms" {
		t.Fatalf("service_id %q", state.ServiceID.ValueString())
	}
}

func TestResolveProviderConfigData_WriteOnly(t *testing.T) {
	t.Parallel()
	raw, diags := resolveProviderConfigData(
		notificationProviderConfigModel{
			ConfigDataWOVersion: types.StringValue("1"),
		},
		notificationProviderConfigModel{
			ConfigDataWO: types.StringValue(`{"commProvider":"x","commMethod":"sms","schemaData":{}}`),
		},
	)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if string(raw) == "" {
		t.Fatal("expected raw json")
	}
}

func TestEnsureProviderConfigDataID_InjectsWhenMissing(t *testing.T) {
	t.Parallel()
	raw, diags := ensureProviderConfigDataID("ss-1", []byte(`{"commProvider":"custom-twilio-sms","commMethod":"sms","schemaData":{}}`))
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["id"] != "ss-1" {
		t.Fatalf("expected id ss-1, got %#v", m["id"])
	}
}

func TestEnsureProviderConfigDataID_PreservesExisting(t *testing.T) {
	t.Parallel()
	raw, diags := ensureProviderConfigDataID("ss-1", []byte(`{"id":"keep-me","commProvider":"x"}`))
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["id"] != "keep-me" {
		t.Fatalf("expected keep-me, got %#v", m["id"])
	}
}
