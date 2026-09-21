package datasources_test

import (
	"fmt"
	"os"
	"testing"

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScopeGroupDataSource_Basic(t *testing.T) {
	t.Parallel()
	resourceName := "data.cidaas_scope_group.sample"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				data "cidaas_scope_group" "sample" {}
				`, os.Getenv("BASE_URL")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "scope_group.#"),
					resource.TestCheckResourceAttrSet(resourceName, "scope_group.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "scope_group.0.group_name"),
				),
			},
		},
	})
}
