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
