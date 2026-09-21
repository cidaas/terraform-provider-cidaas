package identity_provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUserInfoFieldsListToSet(t *testing.T) {
	t.Parallel()
	elemType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"inner_key":       types.StringType,
			"external_key":    types.StringType,
			"is_custom_field": types.BoolType,
			"is_system_field": types.BoolType,
		},
	}
	list := types.ListValueMust(elemType, []attr.Value{
		types.ObjectValueMust(elemType.AttrTypes, map[string]attr.Value{
			"inner_key":       types.StringValue("email"),
			"external_key":    types.StringValue("mail"),
			"is_custom_field": types.BoolValue(true),
			"is_system_field": types.BoolValue(false),
		}),
	})
	set, diags := userInfoFieldsListToSet(list)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags.Errors())
	}
	if set.IsNull() || len(set.Elements()) != 1 {
		t.Fatalf("unexpected set: %#v", set)
	}
}

func TestSocialSchemaParityWith35(t *testing.T) {
	t.Parallel()
	attrs := socialProviderSchema.Attributes
	for _, name := range []string{"claims", "userinfo_fields", "client_secret_wo", "scopes", "enabled_for_admin_portal"} {
		if _, ok := attrs[name]; !ok {
			t.Fatalf("missing attribute %q", name)
		}
	}
	wo, ok := attrs["client_secret_wo"].(schema.StringAttribute)
	if !ok || !wo.WriteOnly {
		t.Fatal("client_secret_wo must be WriteOnly")
	}
	if socialProviderSchema.Version != 1 {
		t.Fatalf("schema version = %d, want 1 for UpgradeState", socialProviderSchema.Version)
	}
	r := &SocialProviderResource{}
	upgraders := r.UpgradeState(t.Context())
	if _, ok := upgraders[0]; !ok {
		t.Fatal("missing state upgrader for version 0")
	}
}

func TestCustomSchemaParityWith35(t *testing.T) {
	t.Parallel()
	attrs := customProviderSchema.Attributes
	for _, name := range []string{"userinfo_fields", "scopes", "amr_config", "apikey_details", "totp_details", "cidaas_auth_details", "client_secret_wo", "domains"} {
		if _, ok := attrs[name]; !ok {
			t.Fatalf("missing attribute %q", name)
		}
	}
	wo, ok := attrs["client_secret_wo"].(schema.StringAttribute)
	if !ok || !wo.WriteOnly {
		t.Fatal("client_secret_wo must be WriteOnly")
	}
}

func TestListStringOrNull(t *testing.T) {
	t.Parallel()
	if !listStringOrNull(nil).IsNull() {
		t.Fatal("expected null for empty")
	}
	got := listStringOrNull([]string{"a"})
	if got.IsNull() || len(got.Elements()) != 1 {
		t.Fatalf("unexpected list: %#v", got)
	}
}
