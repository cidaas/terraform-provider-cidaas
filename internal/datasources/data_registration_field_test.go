package datasources_test

import (
	"fmt"
	"os"
	"testing"

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRegFieldDataSource_Basic(t *testing.T) {
	t.Parallel()
	resourceName := "data.cidaas_registration_field.sample"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				data "cidaas_registration_field" "sample" {
				}
				`, os.Getenv("BASE_URL")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "registration_field.#"),
					resource.TestCheckResourceAttrSet(resourceName, "registration_field.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "registration_field.0.parent_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "registration_field.0.field_type"),
					resource.TestCheckResourceAttrSet(resourceName, "registration_field.0.data_type"),
					resource.TestCheckResourceAttrSet(resourceName, "registration_field.0.field_key"),
					resource.TestCheckResourceAttrSet(resourceName, "registration_field.0.required"),
					resource.TestCheckResourceAttrSet(resourceName, "registration_field.0.internal"),
					resource.TestCheckResourceAttrSet(resourceName, "registration_field.0.read_only"),
					resource.TestCheckResourceAttrSet(resourceName, "registration_field.0.order"),
				),
			},
		},
	})
}

func TestAccRegFieldDataSource_FieldTypeFilter(t *testing.T) {
	t.Parallel()
	resourceName := "data.cidaas_registration_field.sample"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				data "cidaas_registration_field" "sample" {
					filter {
						name   = "field_type"
						values = ["CUSTOM"]
					}
				}
				`, os.Getenv("BASE_URL")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "registration_field.0.field_type", "CUSTOM"),
				),
			},
		},
	})
}
