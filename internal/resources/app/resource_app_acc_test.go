package app_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApp_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	t.Parallel()
	acctest.SkipIfV4(t)

	clientName := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_APP, clientName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig(clientName, "https://cidaas.de"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "client_name", clientName),
					resource.TestCheckResourceAttr(testResourceName, "company_website", "https://cidaas.de"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
			{
				Config: testAccAppConfig(clientName, "https://cidaas.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "company_website", "https://cidaas.com"),
				),
			},
		},
	})
}

func testAccAppConfig(clientName, companyWebsite string) string {
	return fmt.Sprintf(`
    provider "cidaas" {
      base_url = "%s"
    }
    resource "cidaas_app" "%s" {
      client_name         = "%s"
      client_type         = "SINGLE_PAGE"
      company_address     = "01"
      company_website     = "%s"
      company_name        = "Widas ID GmbH"
      redirect_uris       = ["https://cidaas.com"]
      allow_login_with    = ["EMAIL", "MOBILE"]
      allowed_logout_urls = ["https://cidaas.com"]
      allowed_scopes      = ["openid"]
      response_types      = ["code"]
      grant_types         = ["authorization_code", "implicit", "refresh_token"]
    }`,
		os.Getenv("BASE_URL"),
		clientName,
		clientName,
		companyWebsite,
	)
}
