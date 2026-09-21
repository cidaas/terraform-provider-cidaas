package datasources_test

import (
	"fmt"
	"os"
	"testing"

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSysTemplateOptionDataSource_Basic(t *testing.T) {
	t.Parallel()
	resourceName := "data.cidaas_system_template_option.sample"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				data "cidaas_system_template_option" "sample" {}
				`, os.Getenv("BASE_URL")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "system_template_option.#"),
					resource.TestCheckResourceAttrSet(resourceName, "system_template_option.0.template_key"),
					resource.TestCheckResourceAttrSet(resourceName, "system_template_option.0.enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "system_template_option.0.template_types.#"),
					resource.TestCheckResourceAttrSet(resourceName, "system_template_option.0.template_types.0.template_type"),
				),
			},
		},
	})
}

func TestAccSysTemplateOptionDataSource_TemplateKeyFilter(t *testing.T) {
	t.Parallel()
	resourceName := "data.cidaas_system_template_option.sample"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				data "cidaas_system_template_option" "sample" {
					filter {
						name   = "template_key"
						values = ["UN_REGISTER_USER_ALERT"]
					}
				}
				`, os.Getenv("BASE_URL")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "system_template_option.#"),
					resource.TestCheckResourceAttr(resourceName, "system_template_option.0.template_key", "UN_REGISTER_USER_ALERT"),
				),
			},
		},
	})
}
