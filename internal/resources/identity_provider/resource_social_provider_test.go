package identity_provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSocialProvider_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	testResourceID := acctest.RandString(8)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_SOCIAL_PROVIDER, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSocialProviderConfig(testResourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttr(testResourceName, "provider_name", "google"),
				),
			},
		},
	})
}

func testAccSocialProviderConfig(resourceID string) string {
	return fmt.Sprintf(`
provider "cidaas" {
  base_url       = "%s"
  cidaas_version = "4.x"
}

resource "cidaas_social_provider" "%s" {
  provider_name            = "google"
  name                     = "TF Acc Google Login"
  client_id                = "acc-google-client-id.apps.googleusercontent.com"
  client_secret            = "acc-google-client-secret"
  enabled                  = true
  enabled_for_admin_portal = false
  scopes                   = ["openid", "email", "profile"]
  owner                    = "client"
}
`, acctest.GetBaseURL(), resourceID)
}
