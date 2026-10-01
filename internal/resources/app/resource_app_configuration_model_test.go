package app

import (
	"context"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestToModelSetsOwnerClient(t *testing.T) {
	t.Parallel()
	cfg := appConfigurationConfig{
		ClientName: types.StringValue("app-1"),
		ClientType: types.StringValue("NON_INTERACTIVE"),
		Enabled:    types.BoolValue(true),
		Scopes: types.ObjectValueMust(scopesAttrTypes(), map[string]attr.Value{
			"allowed_scopes": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("openid")}),
			"default_scopes": types.ListValueMust(types.StringType, []attr.Value{}),
		}),
		OwnershipDetails: types.ObjectValueMust(ownershipDetailsAttrTypes(), map[string]attr.Value{
			"company_name":    types.StringValue("Co"),
			"company_address": types.StringValue("Addr"),
			"company_website": types.StringValue("https://example.com"),
		}),
	}
	diags := cfg.extract(context.Background())
	if diags.HasError() {
		t.Fatalf("extract: %v", diags)
	}
	model, diags := cfg.toModel(context.Background())
	if diags.HasError() {
		t.Fatalf("toModel: %v", diags)
	}
	if model.Owner != client.OwnerClient {
		t.Fatalf("owner=%q want %q", model.Owner, client.OwnerClient)
	}
}

func TestFlattenAppConfigurationPreservesOwner(t *testing.T) {
	t.Parallel()
	cfg, diags := flattenAppConfiguration(client.AppConfigurationModel{
		ClientID:   "id-1",
		ClientName: "app-1",
		ClientType: "NON_INTERACTIVE",
		Owner:      client.OwnerClient,
		Scopes:     &client.ScopesConfig{AllowedScopes: []string{"openid"}},
	})
	if diags.HasError() {
		t.Fatalf("flatten: %v", diags)
	}
	if cfg.Owner.ValueString() != client.OwnerClient {
		t.Fatalf("owner=%q", cfg.Owner.ValueString())
	}
}

func TestToModelSendsDisableInsecurePKCEMethod(t *testing.T) {
	t.Parallel()
	cfg := appConfigurationConfig{
		ClientName:                types.StringValue("app-pkce"),
		ClientType:                types.StringValue("SINGLE_PAGE"),
		Enabled:                   types.BoolValue(true),
		DisableInsecurePKCEMethod: types.BoolValue(true),
		Scopes: types.ObjectValueMust(scopesAttrTypes(), map[string]attr.Value{
			"allowed_scopes": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("openid")}),
			"default_scopes": types.ListValueMust(types.StringType, []attr.Value{}),
		}),
		OwnershipDetails: types.ObjectValueMust(ownershipDetailsAttrTypes(), map[string]attr.Value{
			"company_name":    types.StringValue("Co"),
			"company_address": types.StringValue("Addr"),
			"company_website": types.StringValue("https://example.com"),
		}),
	}
	diags := cfg.extract(context.Background())
	if diags.HasError() {
		t.Fatalf("extract: %v", diags)
	}
	model, diags := cfg.toModel(context.Background())
	if diags.HasError() {
		t.Fatalf("toModel: %v", diags)
	}
	if model.DisableInsecurePKCEMethod == nil || !*model.DisableInsecurePKCEMethod {
		t.Fatalf("DisableInsecurePKCEMethod=%v", model.DisableInsecurePKCEMethod)
	}
	if model.PKCE == nil || model.PKCE.DisableInsecurePKCEMethod == nil || !*model.PKCE.DisableInsecurePKCEMethod {
		t.Fatalf("PKCE=%v", model.PKCE)
	}
}

func TestMergeOmittedAppConfigurationBoolsKeepsPlannedPKCE(t *testing.T) {
	t.Parallel()
	plan := appConfigurationConfig{
		RequirePKCE:               types.BoolValue(true),
		DisableInsecurePKCEMethod: types.BoolValue(true),
	}
	// Simulate app-srv create/get response that omits PKCE flags.
	fromAPI, diags := flattenAppConfiguration(client.AppConfigurationModel{
		ClientID:   "id-1",
		ClientName: "app-1",
		ClientType: "SINGLE_PAGE",
	})
	if diags.HasError() {
		t.Fatalf("flatten: %v", diags)
	}
	if !fromAPI.DisableInsecurePKCEMethod.IsNull() {
		t.Fatalf("expected null from API flatten, got %v", fromAPI.DisableInsecurePKCEMethod)
	}
	merged := mergeOmittedAppConfigurationBools(plan, fromAPI)
	if merged.DisableInsecurePKCEMethod.IsNull() || !merged.DisableInsecurePKCEMethod.ValueBool() {
		t.Fatalf("DisableInsecurePKCEMethod=%v", merged.DisableInsecurePKCEMethod)
	}
	if merged.RequirePKCE.IsNull() || !merged.RequirePKCE.ValueBool() {
		t.Fatalf("RequirePKCE=%v", merged.RequirePKCE)
	}
}

func TestToModelPreservesExplicitClientID(t *testing.T) {
	t.Parallel()
	cfg := appConfigurationConfig{
		ClientID:   types.StringValue("custom-client-id"),
		ClientName: types.StringValue("app-1"),
		ClientType: types.StringValue("NON_INTERACTIVE"),
		Enabled:    types.BoolValue(true),
		Scopes: types.ObjectValueMust(scopesAttrTypes(), map[string]attr.Value{
			"allowed_scopes": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("openid")}),
			"default_scopes": types.ListValueMust(types.StringType, []attr.Value{}),
		}),
		OwnershipDetails: types.ObjectValueMust(ownershipDetailsAttrTypes(), map[string]attr.Value{
			"company_name":    types.StringValue("Co"),
			"company_address": types.StringValue("Addr"),
			"company_website": types.StringValue("https://example.com"),
		}),
	}
	diags := cfg.extract(context.Background())
	if diags.HasError() {
		t.Fatalf("extract: %v", diags)
	}
	model, diags := cfg.toModel(context.Background())
	if diags.HasError() {
		t.Fatalf("toModel: %v", diags)
	}
	if model.ClientID != "custom-client-id" {
		t.Fatalf("ClientID=%q", model.ClientID)
	}
}

func TestListToStringsSkipsUnknownElements(t *testing.T) {
	t.Parallel()
	l := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("openid"),
		types.StringUnknown(),
		types.StringValue("profile"),
		types.StringNull(),
	})
	got := util.ListToStrings(l)
	if len(got) != 2 || got[0] != "openid" || got[1] != "profile" {
		t.Fatalf("got %v", got)
	}
}
