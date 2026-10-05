package app

import (
	"context"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

func TestToModelSendsPKCECodeChallengeMethod(t *testing.T) {
	t.Parallel()
	cfg := appConfigurationConfig{
		ClientName: types.StringValue("app-pkce"),
		ClientType: types.StringValue("SINGLE_PAGE"),
		Enabled:    types.BoolValue(true),
		PKCE: types.ObjectValueMust(pkceAttrTypes(), map[string]attr.Value{
			"require_pkce":          types.BoolValue(true),
			"code_challenge_method": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("S256")}),
		}),
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
	if model.PKCE == nil || model.PKCE.RequirePKCE == nil || !*model.PKCE.RequirePKCE {
		t.Fatalf("PKCE.RequirePKCE=%v", model.PKCE)
	}
	if len(model.PKCE.CodeChallengeMethod) != 1 || model.PKCE.CodeChallengeMethod[0] != "S256" {
		t.Fatalf("PKCE.CodeChallengeMethod=%v", model.PKCE.CodeChallengeMethod)
	}
}

func TestFlattenAppConfigurationReadsPKCEFromAPI(t *testing.T) {
	t.Parallel()
	reqPKCE := true
	cfg, diags := flattenAppConfiguration(client.AppConfigurationModel{
		ClientID:   "id-1",
		ClientName: "app-1",
		ClientType: "SINGLE_PAGE",
		PKCE: &client.PKCEConfig{
			RequirePKCE:         &reqPKCE,
			CodeChallengeMethod: []string{"S256"},
		},
	})
	if diags.HasError() {
		t.Fatalf("flatten: %v", diags)
	}
	if cfg.PKCE.IsNull() {
		t.Fatal("expected pkce from API")
	}
	var pkce pkceConfig
	if d := cfg.PKCE.As(context.Background(), &pkce, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("pkce.As: %v", d)
	}
	if pkce.RequirePKCE.IsNull() || !pkce.RequirePKCE.ValueBool() {
		t.Fatalf("require_pkce=%v", pkce.RequirePKCE)
	}
	methods := util.ListToStrings(pkce.CodeChallengeMethod)
	if len(methods) != 1 || methods[0] != "S256" {
		t.Fatalf("code_challenge_method=%v", methods)
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
