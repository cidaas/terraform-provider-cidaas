package consent

import (
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestConsentConfigure_AllowsV4(t *testing.T) {
	t.Parallel()
	c := &client.Client{
		Capabilities: client.Capabilities{TargetVersion: "4.x"},
		CidaasClient: &cidaas.Client{},
	}
	for _, res := range []resource.Resource{
		NewConsentResource(),
		NewConsentGroupResource(),
		NewConsentVersionResource(),
	} {
		cfg, ok := res.(resource.ResourceWithConfigure)
		if !ok {
			t.Fatalf("%T does not implement ResourceWithConfigure", res)
		}
		var resp resource.ConfigureResponse
		cfg.Configure(t.Context(), resource.ConfigureRequest{ProviderData: c}, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("%T Configure on 4.x failed: %v", res, resp.Diagnostics.Errors())
		}
	}
}
