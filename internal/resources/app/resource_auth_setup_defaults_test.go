package app

import (
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAuthenticationSetupToClientIncludesDefaultsAndLoginSpi(t *testing.T) {
	t.Parallel()
	spi, diags := types.ObjectValue(loginSpiAttrTypes(), map[string]attr.Value{
		"enable_login_spi": types.BoolValue(true),
		"oauth_client_id":  types.StringValue("spi-client"),
		"spi_url":          types.StringValue("https://example/spi"),
	})
	if diags.HasError() {
		t.Fatalf("%v", diags)
	}
	cfg := &authenticationSetupConfig{
		AutoLoginAfterRegister:       types.BoolValue(true),
		RegisterWithLoginInformation: types.BoolValue(false),
		EnablePasswordLessAuth:       types.BoolValue(true),
		AllowUserLevelMultiProvider:  types.BoolValue(false),
		SocialBusinessIDs:            types.BoolValue(true),
		LoginSpi:                     spi,
		loginSpi: &loginSpiConfig{
			EnableLoginSpi: types.BoolValue(true),
			OauthClientID:  types.StringValue("spi-client"),
			SpiURL:         types.StringValue("https://example/spi"),
		},
	}
	out := authenticationSetupToClient(cfg)
	if out.AutoLoginAfterRegister == nil || !*out.AutoLoginAfterRegister {
		t.Fatalf("AutoLoginAfterRegister=%v", out.AutoLoginAfterRegister)
	}
	if out.LoginSpi == nil || out.LoginSpi.OauthClientID != "spi-client" || out.LoginSpi.SpiURL != "https://example/spi" {
		t.Fatalf("LoginSpi=%+v", out.LoginSpi)
	}
}

func TestFlattenAuthSetupDefaults(t *testing.T) {
	t.Parallel()
	trueVal := true
	falseVal := false
	m := flattenAuthSetupDefaults(client.AuthSetupDefaultsEntity{
		ID:          "default",
		Name:        "default",
		Description: "desc",
		AuthSetupDefaults: &client.AuthSetupDefaults{
			AutoLoginAfterRegister:       &trueVal,
			RegisterWithLoginInformation: &falseVal,
			EnablePasswordLessAuth:       &trueVal,
			AllowUserLevelMultiProvider:  &falseVal,
			SocialBusinessIDs:            &trueVal,
			NetID:                        &falseVal,
		},
	})
	if m.ID.ValueString() != "default" || !m.AutoLoginAfterRegister.ValueBool() || m.NetID.ValueBool() {
		t.Fatalf("%+v", m)
	}
}

func TestMergeAuthSetupDefaultsPlanOverridesKnownBools(t *testing.T) {
	t.Parallel()
	curTrue := true
	current := client.AuthSetupDefaultsEntity{
		Name:  "default",
		Owner: "SYSTEM",
		AuthSetupDefaults: &client.AuthSetupDefaults{
			EnablePasswordLessAuth: &curTrue,
			NetID:                  &curTrue,
		},
	}
	falseVal := types.BoolValue(false)
	plan := authSetupDefaultsModel{
		EnablePasswordLessAuth: falseVal,
		NetID:                  falseVal,
	}
	out := mergeAuthSetupDefaultsPlan(current, plan)
	if out.Owner != "SYSTEM" {
		t.Fatalf("owner=%q, want SYSTEM (must round-trip on PUT)", out.Owner)
	}
	if out.AuthSetupDefaults == nil || out.AuthSetupDefaults.EnablePasswordLessAuth == nil || *out.AuthSetupDefaults.EnablePasswordLessAuth {
		t.Fatalf("%+v", out.AuthSetupDefaults)
	}
	if out.AuthSetupDefaults.NetID == nil || *out.AuthSetupDefaults.NetID {
		t.Fatalf("net_id=%v", out.AuthSetupDefaults.NetID)
	}
}
