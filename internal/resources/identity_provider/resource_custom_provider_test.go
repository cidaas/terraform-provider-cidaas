package identity_provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCustomProvider_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	acctest.SkipIfV4(t)

	testResourceID := acctest.RandString(8)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_CUSTOM_PROVIDER, testResourceID)
	providerName := fmt.Sprintf("tf_acc_custom_%s", testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomProviderConfig(testResourceID, providerName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttr(testResourceName, "provider_name", providerName),
					resource.TestCheckResourceAttr(testResourceName, "standard_type", "OAUTH2"),
				),
			},
		},
	})
}

func testAccCustomProviderConfig(resourceID, providerName string) string {
	return fmt.Sprintf(`
provider "cidaas" {
  base_url = "%s"
}

resource "cidaas_custom_provider" "%s" {
  provider_name          = "%s"
  display_name           = "TF Acc Custom Provider"
  standard_type          = "OAUTH2"
  client_id              = "acc_client_id_123"
  client_secret          = "acc_client_secret_123"
  authorization_endpoint = "https://idp.example.com/oauth2/v1/authorize"
  token_endpoint         = "https://idp.example.com/oauth2/v1/token"
  userinfo_endpoint      = "https://idp.example.com/oauth2/v1/userinfo"
  logo_url               = "https://cdn.example.com/logo.png"
  domains                = ["example.com"]
  owner                  = "client"
}
`, acctest.GetBaseURL(), resourceID, providerName)
}
