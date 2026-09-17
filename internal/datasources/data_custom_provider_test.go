package datasources_test

import (
	"fmt"
	"os"
	"testing"

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceCustomProvider_basic(t *testing.T) {
	t.Parallel()
	resourceName := "data.cidaas_custom_provider.sample"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				data "cidaas_custom_provider" "sample" {
				}
				`, os.Getenv("BASE_URL")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "custom_provider.#"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_provider.0.provider_name"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_provider.0.standard_type"),
				),
			},
		},
	})
}
