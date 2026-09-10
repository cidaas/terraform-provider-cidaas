package identity_provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFederationProvider_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	t.Parallel()
	acctest.SkipIfV3(t)

	testResourceID := acctest.RandString(8)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_FEDERATION_PROVIDER, testResourceID)
	providerName := fmt.Sprintf("tf_acc_fed_%s", testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFederationProviderConfig(testResourceID, providerName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttr(testResourceName, "provider_name", providerName),
					resource.TestCheckResourceAttr(testResourceName, "standard_type", "OAUTH2"),
				),
			},
		},
	})
}

func testAccFederationProviderConfig(resourceID, providerName string) string {
	return fmt.Sprintf(`
provider "cidaas" {
  base_url = "%s"
}

resource "cidaas_federation_provider" "%s" {
  provider_name          = "%s"
  display_name           = "TF Acc Federation Provider"
  standard_type          = "OAUTH2"
  client_id              = "acc_fed_client_id_456"
  client_secret          = "acc_fed_client_secret_456"
  authorization_endpoint = "https://auth.example.com/oauth/authorize"
  token_endpoint         = "https://auth.example.com/oauth/token"
  userinfo_endpoint      = "https://auth.example.com/oauth/userinfo"
  logo_url               = "https://cdn.example.com/logo.svg"
  domains                = ["example.com"]
  owner                  = "client"
}
`, acctest.GetBaseURL(), resourceID, providerName)
}
