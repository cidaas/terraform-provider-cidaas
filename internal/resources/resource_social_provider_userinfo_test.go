package resources

import (
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSetUserInfoFields_OrderIndependent(t *testing.T) {
	custom := cidaas.UserInfoFieldsModel{
		InnerKey:      "sample_custom_field",
		ExternalKey:   "external_sample_cf",
		IsCustomField: true,
		IsSystemField: false,
	}
	system := cidaas.UserInfoFieldsModel{
		InnerKey:      "sample_system_field",
		ExternalKey:   "external_sample_sf",
		IsCustomField: false,
		IsSystemField: true,
	}

	var customFirst, systemFirst SocialProviderConfig
	diags := setUserInfoFields(&customFirst, []cidaas.UserInfoFieldsModel{custom, system})
	if diags.HasError() {
		t.Fatalf("custom-first: %v", diags.Errors())
	}
	diags = setUserInfoFields(&systemFirst, []cidaas.UserInfoFieldsModel{system, custom})
	if diags.HasError() {
		t.Fatalf("system-first: %v", diags.Errors())
	}

	if !customFirst.UserInfoFields.Equal(systemFirst.UserInfoFields) {
		t.Fatalf("set equality failed: custom-first=%v system-first=%v",
			customFirst.UserInfoFields, systemFirst.UserInfoFields)
	}

	if customFirst.UserInfoFields.Elements() == nil || len(customFirst.UserInfoFields.Elements()) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(customFirst.UserInfoFields.Elements()))
	}
}

func TestSetUserInfoFields_Empty(t *testing.T) {
	var state SocialProviderConfig
	diags := setUserInfoFields(&state, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags.Errors())
	}
	want := types.SetValueMust(userInfoFieldsType, []attr.Value{})
	if !state.UserInfoFields.Equal(want) {
		t.Fatalf("expected empty set, got %v", state.UserInfoFields)
	}
}

func TestUserInfoFieldsListToSet_PreservesElements(t *testing.T) {
	objType := userInfoFieldsType
	a := types.ObjectValueMust(objType.AttrTypes, map[string]attr.Value{
		"inner_key":       types.StringValue("sample_system_field"),
		"external_key":    types.StringValue("external_sample_sf"),
		"is_custom_field": types.BoolValue(false),
		"is_system_field": types.BoolValue(true),
	})
	b := types.ObjectValueMust(objType.AttrTypes, map[string]attr.Value{
		"inner_key":       types.StringValue("sample_custom_field"),
		"external_key":    types.StringValue("external_sample_cf"),
		"is_custom_field": types.BoolValue(true),
		"is_system_field": types.BoolValue(false),
	})

	list := types.ListValueMust(objType, []attr.Value{a, b})
	got, diags := userInfoFieldsListToSet(list)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags.Errors())
	}

	want := types.SetValueMust(objType, []attr.Value{b, a}) // reverse order, set equality
	if !got.Equal(want) {
		t.Fatalf("list→set mismatch: got=%v want=%v", got, want)
	}
}

func TestUserInfoFieldsListToSet_NullAndEmpty(t *testing.T) {
	emptyWant := types.SetValueMust(userInfoFieldsType, []attr.Value{})

	got, diags := userInfoFieldsListToSet(types.ListNull(userInfoFieldsType))
	if diags.HasError() {
		t.Fatalf("null: %v", diags.Errors())
	}
	if !got.Equal(emptyWant) {
		t.Fatalf("null list should become empty set, got %v", got)
	}

	got, diags = userInfoFieldsListToSet(types.ListValueMust(userInfoFieldsType, []attr.Value{}))
	if diags.HasError() {
		t.Fatalf("empty: %v", diags.Errors())
	}
	if !got.Equal(emptyWant) {
		t.Fatalf("empty list should become empty set, got %v", got)
	}
}

func TestSocialProviderUpgradeState_HasV0(t *testing.T) {
	r := &SocialProvider{}
	upgraders := r.UpgradeState(t.Context())
	if _, ok := upgraders[0]; !ok {
		t.Fatal("expected StateUpgrader for schema version 0")
	}
	if upgraders[0].PriorSchema == nil {
		t.Fatal("expected PriorSchema for version 0")
	}
	if socialProviderSchema.Version != 1 {
		t.Fatalf("expected Schema.Version 1, got %d", socialProviderSchema.Version)
	}
}

func TestSocialProviderSchemaV0_FrozenSnapshot(t *testing.T) {
	v0 := socialProviderSchemaV0()
	wantKeys := []string{
		"id", "name", "provider_name", "enabled", "client_id", "client_secret",
		"client_secret_wo", "client_secret_wo_version", "scopes", "claims",
		"userinfo_fields", "enabled_for_admin_portal",
	}
	if len(v0.Attributes) != len(wantKeys) {
		t.Fatalf("v0 attribute count = %d, want %d (frozen snapshot must not pick up new v1 attrs)",
			len(v0.Attributes), len(wantKeys))
	}
	for _, k := range wantKeys {
		if _, ok := v0.Attributes[k]; !ok {
			t.Fatalf("v0 missing attribute %q", k)
		}
	}
	if len(socialProviderSchema.Attributes) != len(wantKeys) {
		t.Fatalf("unexpected: v1 attribute count %d differs from v0 snapshot %d",
			len(socialProviderSchema.Attributes), len(wantKeys))
	}
}
