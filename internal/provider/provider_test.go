package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
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
