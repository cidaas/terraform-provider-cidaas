package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestProvider_Metadata(t *testing.T) {
	t.Parallel()
	p := New("test")()
	var resp provider.MetadataResponse
	p.Metadata(context.Background(), provider.MetadataRequest{}, &resp)
	if resp.TypeName != "cidaas" {
		t.Fatalf("TypeName=%q", resp.TypeName)
	}
}

func TestProvider_Configure_MissingCredentials(t *testing.T) {
	t.Setenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID", "")
	t.Setenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET", "")

	p := New("test")().(*cidaasProvider)
	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	raw := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"base_url": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"base_url": tftypes.NewValue(tftypes.String, "https://example.cidaas.eu"),
	})

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    raw,
	}

	var resp provider.ConfigureResponse
	p.Configure(context.Background(), provider.ConfigureRequest{Config: config}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected missing credentials error")
	}
}

func TestProvider_ResourceRegistrationParity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p := New("test")().(*cidaasProvider)

	expectedResources := map[string]bool{
		"cidaas_app":                                         true,
		"cidaas_app_configuration":                           true,
		"cidaas_consent":                                     true,
		"cidaas_consent_group":                               true,
		"cidaas_consent_version":                             true,
		"cidaas_custom_provider":                             true,
		"cidaas_federation_provider":                         true,
		"cidaas_group_selection":                             true,
		"cidaas_group_type":                                  true,
		"cidaas_group_verification_filter":                   true,
		"cidaas_hosted_page":                                 true,
		"cidaas_hosted_page_layout":                          true,
		"cidaas_notification_provider_config":                true,
		"cidaas_notification_service_setup":                  true,
		"cidaas_notification_template":                       true,
		"cidaas_notification_template_type":                  true,
		"cidaas_notifications_template_group":                true,
		"cidaas_notifications_template_group_locale":         true,
		"cidaas_password_policy":                             true,
		"cidaas_registration_field":                          true,
		"cidaas_role":                                        true,
		"cidaas_scope":                                       true,
		"cidaas_scope_group":                                 true,
		"cidaas_security_settings":                           true,
		"cidaas_social_provider":                             true,
		"cidaas_suggest_verification_method":                 true,
		"cidaas_template":                                    true,
		"cidaas_template_group":                              true,
		"cidaas_theme":                                       true,
		"cidaas_translations":                                true,
		"cidaas_user_groups":                                 true,
		"cidaas_user_setup":                                  true,
		"cidaas_verification_options":                        true,
		"cidaas_webhook":                                     true,
	}

	registered := make(map[string]bool)
	for _, resFunc := range p.Resources(ctx) {
		res := resFunc()
		var metaResp resource.MetadataResponse
		res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cidaas"}, &metaResp)
		registered[metaResp.TypeName] = true
	}

	for expected := range expectedResources {
		if !registered[expected] {
			t.Errorf("Missing expected resource registration in provider.go: %q", expected)
		}
	}
}

