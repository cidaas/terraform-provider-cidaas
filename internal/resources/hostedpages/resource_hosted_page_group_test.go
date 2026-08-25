package hostedpages_test

import (
	"fmt"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHostedPageGroup_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	name := fmt.Sprintf("tf_v4_hpg_%s", acctest.RandString(6))
	base := acctest.BaseURL()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ProviderConfig() + fmt.Sprintf(`
resource "cidaas_hosted_page_group" "test" {
  name           = %q
  default_locale = "en"
  group_owner    = "client"
  hosted_pages = [
    {
      hosted_page_id = "login"
      locale         = "en"
      url            = "%s/login-%s"
    }
  ]
}
`, name, base, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cidaas_hosted_page_group.test", "name", name),
					resource.TestCheckResourceAttr("cidaas_hosted_page_group.test", "default_locale", "en"),
				),
			},
			{
				Config: acctest.ProviderConfig() + fmt.Sprintf(`
resource "cidaas_hosted_page_group" "test" {
  name           = %q
  default_locale = "en"
  group_owner    = "client"
  hosted_pages = [
    {
      hosted_page_id = "login"
      locale         = "en"
      url            = "%s/login-updated-%s"
    },
    {
      hosted_page_id = "register"
      locale         = "en"
      url            = "%s/register-%s"
    }
  ]
}
`, name, base, name, base, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cidaas_hosted_page_group.test", "name", name),
				),
			},
		},
	})
}
