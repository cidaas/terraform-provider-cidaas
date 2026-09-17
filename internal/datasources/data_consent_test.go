package datasources_test

import (
	"fmt"
	"os"
	"testing"

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceConsent_basic(t *testing.T) {
	t.Parallel()
	resourceName := "data.cidaas_consent.sample"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				data "cidaas_consent" "sample" {
				}
				`, os.Getenv("BASE_URL")), // replace with acctest.BaseURL or have a init func to set the base URL
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "consent.#"),
					resource.TestCheckResourceAttrSet(resourceName, "consent.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "consent.0.consent_name"),
				),
			},
		},
	})
}
