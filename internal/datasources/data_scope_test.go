package datasources_test

import (
	"fmt"
	"os"
	"testing"

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScopeDataSource_Basic(t *testing.T) {
	t.Parallel()
	resourceName := "data.cidaas_scope.sample"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				data "cidaas_scope" "sample" {}
				`, os.Getenv("BASE_URL")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "scope.#"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.security_level"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.scope_key"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.required_user_consent"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.localized_descriptions.#"),
				),
			},
		},
	})
}

func TestAccScopeDataSource_SecurityLevelFilter(t *testing.T) {
	t.Parallel()
	resourceName := "data.cidaas_scope.sample"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				data "cidaas_scope" "sample" {
					filter {
						name = "security_level"
						values = ["PUBLIC"]
					}
				}
				`, os.Getenv("BASE_URL")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "scope.#"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.security_level", "PUBLIC"),
				),
			},
		},
	})
}
