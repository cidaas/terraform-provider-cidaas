package datasources_test

import (
	"fmt"
	"os"
	"testing"

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupTypeDataSource_Basic(t *testing.T) {
	t.Parallel()
	resourceName := "data.cidaas_group_type.sample"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				data "cidaas_group_type" "sample" {
				}
				`, os.Getenv("BASE_URL")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "group_type.#"),
					resource.TestCheckResourceAttrSet(resourceName, "group_type.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "group_type.0.group_type"),
					resource.TestCheckResourceAttrSet(resourceName, "group_type.0.role_mode"),
				),
			},
		},
	})
}

func TestAccGroupTypeDataSource_RoleModeFilter(t *testing.T) {
	t.Parallel()
	resourceName := "data.cidaas_group_type.sample"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				data "cidaas_group_type" "sample" {
					filter {
						name = "role_mode"
						values = ["allowed_roles"]
					}
				}
				`, os.Getenv("BASE_URL")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "group_type.#"),
					resource.TestCheckResourceAttr(resourceName, "group_type.0.role_mode", "allowed_roles"),
				),
			},
		},
	})
}
