package app

import (
	"context"
	"testing"

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
