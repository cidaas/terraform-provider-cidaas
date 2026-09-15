package identity_provider

import (
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestFederationConfigure_RequiresV4(t *testing.T) {
	t.Parallel()
	res := NewFederationProviderResource().(resource.ResourceWithConfigure)

	var reject resource.ConfigureResponse
	res.Configure(t.Context(), resource.ConfigureRequest{
		ProviderData: &client.Client{Capabilities: client.Capabilities{TargetVersion: "3.x"}},
	}, &reject)
	if !reject.Diagnostics.HasError() {
		t.Fatal("expected Configure to fail on 3.x")
	}

	var allow resource.ConfigureResponse
	res.Configure(t.Context(), resource.ConfigureRequest{
		ProviderData: &client.Client{
			Capabilities: client.Capabilities{TargetVersion: "4.x"},
			CidaasClient: &cidaas.Client{},
		},
	}, &allow)
	if allow.Diagnostics.HasError() {
		t.Fatalf("expected Configure to succeed on 4.x: %v", allow.Diagnostics.Errors())
	}
}
