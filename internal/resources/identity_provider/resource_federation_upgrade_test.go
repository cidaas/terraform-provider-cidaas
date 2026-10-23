package identity_provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUpgradeFederationProviderStateV0(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	priorSchema := federationProviderSchemaV0()
	prior := tfsdk.State{Schema: priorSchema, Raw: tftypes.NewValue(priorSchema.Type().TerraformType(ctx), nil)}
	diags := prior.Set(ctx, &federationProviderModelV0{
		ID:                    types.StringValue("fed-1"),
		ProviderName:          types.StringValue("my_oauth"),
		DisplayName:           types.StringValue("My OAuth"),
		StandardType:          types.StringValue("OAUTH2"),
		ClientID:              types.StringValue("cid"),
		ClientSecret:          types.StringValue("secret"),
		ClientSecretWO:        types.StringNull(),
		ClientSecretWOVersion: types.StringNull(),
		AuthorizationEndpoint: types.StringValue("https://idp.example.com/authorize"),
		TokenEndpoint:         types.StringValue("https://idp.example.com/token"),
		UserinfoEndpoint:      types.StringValue("https://idp.example.com/userinfo"),
		LogoURL:               types.StringNull(),
		Domains:               types.ListValueMust(types.StringType, []attr.Value{types.StringValue("example.com")}),
		Owner:                 types.StringValue("client"),
	})
	if diags.HasError() {
		t.Fatalf("setting v0 state: %v", diags.Errors())
	}

	currentSchema := federationProviderResourceSchema()
	resp := resource.UpgradeStateResponse{
		State: tfsdk.State{Schema: currentSchema, Raw: tftypes.NewValue(currentSchema.Type().TerraformType(ctx), nil)},
	}
	upgradeFederationProviderStateV0(ctx, resource.UpgradeStateRequest{State: &prior}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("upgrade failed: %v", resp.Diagnostics.Errors())
	}

	for attrPath, want := range map[string]string{
		"provider_type":                 "custom",
		"oauth2.client_id":              "cid",
		"oauth2.client_secret":          "secret",
		"oauth2.authorization_endpoint": "https://idp.example.com/authorize",
		"oauth2.userinfo_source":        "USERINFOENDPOINT",
	} {
		var got types.String
		p := path.Root(attrPath)
		if parent, child, ok := strings.Cut(attrPath, "."); ok {
			p = path.Root(parent).AtName(child)
		}
		resp.Diagnostics.Append(resp.State.GetAttribute(ctx, p, &got)...)
		if got.ValueString() != want {
			t.Errorf("%s = %q, want %q", attrPath, got.ValueString(), want)
		}
	}
	var saml types.Object
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("saml"), &saml)...)
	if !saml.IsNull() {
		t.Error("saml must be null after upgrading an OAuth2 provider")
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("reading upgraded state: %v", resp.Diagnostics.Errors())
	}
}

func TestFederationProviderSchemaIsValid(t *testing.T) {
	t.Parallel()
	s := federationProviderResourceSchema()
	if diags := s.ValidateImplementation(t.Context()); diags.HasError() {
		t.Fatalf("invalid schema: %v", diags.Errors())
	}
	if s.Version != 1 {
		t.Fatalf("schema version = %d, want 1 (state upgrade from flat v0)", s.Version)
	}
}

func TestFederationTimestamp(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		"2026-10-16T18:02:21.713792576Z": "2026-10-16T18:02:21.713Z", // create/update response
		"2026-10-16T18:02:21.713Z":       "2026-10-16T18:02:21.713Z", // GET response
		"2026-10-16T18:02:23.520456945Z": "2026-10-16T18:02:23.52Z",
		"2026-10-16T18:02:23.52Z":        "2026-10-16T18:02:23.52Z",
		"2026-10-16T20:02:21.7137+02:00": "2026-10-16T18:02:21.713Z",
		"not-a-time":                     "not-a-time",
	} {
		if got := federationTimestamp(in).ValueString(); got != want {
			t.Errorf("federationTimestamp(%q) = %q, want %q", in, got, want)
		}
	}
	if !federationTimestamp("").IsNull() {
		t.Error("empty timestamp must be null")
	}
}
