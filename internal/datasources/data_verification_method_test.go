package datasources_test

import (
	"fmt"
	"os"
	"testing"

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVerificationMethodDataSource_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	t.Parallel()
	acctest.SkipIfV3(t)

	name := "data.cidaas_verification_method.email"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "cidaas" {
  base_url = "%s"
}

data "cidaas_verification_method" "email" {
  verification_type = "EMAIL"
}
`, acctest.GetBaseURL()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "id", "EMAIL"),
					resource.TestCheckResourceAttrSet(name, "owner"),
					resource.TestCheckResourceAttrSet(name, "active"),
					resource.TestCheckResourceAttrSet(name, "locale_text.#"),
				),
			},
		},
	})
}
