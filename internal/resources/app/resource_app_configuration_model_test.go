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

func TestToModelSendsIdentityProviders(t *testing.T) {
	t.Parallel()
	providerType := types.ObjectType{AttrTypes: identityProviderAttrTypes()}
	cfg := appConfigurationConfig{
		ClientName: types.StringValue("app-idp"),
		ClientType: types.StringValue("SINGLE_PAGE"),
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
		IdentityProviders: types.ObjectValueMust(identityProvidersAttrTypes(), map[string]attr.Value{
			"providers": types.ListValueMust(providerType, []attr.Value{
				types.ObjectValueMust(identityProviderAttrTypes(), map[string]attr.Value{
					"provider_id":         types.StringValue("idp-1"),
					"is_provider_visible": types.BoolValue(true),
				}),
				types.ObjectValueMust(identityProviderAttrTypes(), map[string]attr.Value{
					"provider_id":         types.StringValue("idp-2"),
					"is_provider_visible": types.BoolValue(false),
				}),
			}),
			"webfinger": types.StringValue("no_redirection"),
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
	ip := model.IdentityProviders
	if ip == nil || len(ip.Providers) != 2 {
		t.Fatalf("IdentityProviders=%+v", ip)
	}
	if ip.Providers[0].ProviderID != "idp-1" || ip.Providers[0].IsProviderVisible == nil || !*ip.Providers[0].IsProviderVisible {
		t.Fatalf("providers[0]=%+v", ip.Providers[0])
	}
	if ip.Providers[1].ProviderID != "idp-2" || ip.Providers[1].IsProviderVisible == nil || *ip.Providers[1].IsProviderVisible {
		t.Fatalf("providers[1]=%+v", ip.Providers[1])
	}
	if ip.Webfinger != "no_redirection" {
		t.Fatalf("webfinger=%q", ip.Webfinger)
	}
}

func TestToModelOmitsIdentityProvidersWhenUnset(t *testing.T) {
	t.Parallel()
	cfg := appConfigurationConfig{
		ClientName:        types.StringValue("app-1"),
		ClientType:        types.StringValue("NON_INTERACTIVE"),
		IdentityProviders: types.ObjectNull(identityProvidersAttrTypes()),
	}
	if diags := cfg.extract(context.Background()); diags.HasError() {
		t.Fatalf("extract: %v", diags)
	}
	model, diags := cfg.toModel(context.Background())
	if diags.HasError() {
		t.Fatalf("toModel: %v", diags)
	}
	if model.IdentityProviders != nil {
		t.Fatalf("IdentityProviders=%+v want nil", model.IdentityProviders)
	}
}

func TestFlattenAppConfigurationReadsIdentityProviders(t *testing.T) {
	t.Parallel()
	hidden := false
	cfg, diags := flattenAppConfiguration(client.AppConfigurationModel{
		ClientID:   "id-1",
		ClientName: "app-1",
		ClientType: "SINGLE_PAGE",
		IdentityProviders: &client.IdentityProvidersConfig{
			Providers: []client.IdentityProviderRef{
				{ProviderID: "idp-1"},
				{ProviderID: "idp-2", IsProviderVisible: &hidden},
			},
			Webfinger: "domain_based_redirection",
		},
	})
	if diags.HasError() {
		t.Fatalf("flatten: %v", diags)
	}
	var ip identityProvidersConfig
	if d := cfg.IdentityProviders.As(context.Background(), &ip, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("identity_providers.As: %v", d)
	}
	var providers []identityProviderRef
	if d := ip.Providers.ElementsAs(context.Background(), &providers, false); d.HasError() {
		t.Fatalf("providers.ElementsAs: %v", d)
	}
	if len(providers) != 2 {
		t.Fatalf("providers=%v", providers)
	}
	// Omitted isProviderVisible reads back as visible.
	if providers[0].ProviderID.ValueString() != "idp-1" || !providers[0].IsProviderVisible.ValueBool() {
		t.Fatalf("providers[0]=%+v", providers[0])
	}
	if providers[1].ProviderID.ValueString() != "idp-2" || providers[1].IsProviderVisible.ValueBool() {
		t.Fatalf("providers[1]=%+v", providers[1])
	}
	if ip.Webfinger.ValueString() != "domain_based_redirection" {
		t.Fatalf("webfinger=%v", ip.Webfinger)
	}
}

func TestFlattenAppConfigurationIdentityProvidersNullWhenEmpty(t *testing.T) {
	t.Parallel()
	cfg, diags := flattenAppConfiguration(client.AppConfigurationModel{
		ClientID:          "id-1",
		ClientName:        "app-1",
		ClientType:        "SINGLE_PAGE",
		IdentityProviders: &client.IdentityProvidersConfig{},
	})
	if diags.HasError() {
		t.Fatalf("flatten: %v", diags)
	}
	if !cfg.IdentityProviders.IsNull() {
		t.Fatalf("identity_providers=%v want null", cfg.IdentityProviders)
	}
}
