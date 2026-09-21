package datasources_test

import (
	"fmt"
	"os"
	"testing"

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSocialProviderDataSource_Basic(t *testing.T) {
	t.Parallel()
	resourceName := "data.cidaas_social_provider.sample"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				data "cidaas_social_provider" "sample" {}
				`, os.Getenv("BASE_URL")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "social_provider.#"),
					resource.TestCheckResourceAttrSet(resourceName, "social_provider.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "social_provider.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "social_provider.0.provider_name"),
					resource.TestCheckResourceAttrSet(resourceName, "social_provider.0.enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "social_provider.0.client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "social_provider.0.client_secret"),
					resource.TestCheckResourceAttrSet(resourceName, "social_provider.0.enabled_for_admin_portal"),
					resource.TestCheckResourceAttrSet(resourceName, "social_provider.0.scopes.#"),
				),
			},
		},
	})
}
