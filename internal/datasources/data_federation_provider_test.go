package datasources_test

import (
	"fmt"
	"os"
	"testing"

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFederationProviderDataSource_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	t.Parallel()
	acctest.SkipIfV3(t)

	providerName := "tf_acc_ds_fed_" + acctest.RandString(8)
	resourceName := "cidaas_federation_provider.sample"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "cidaas" {
  base_url = "%s"
}

resource "cidaas_federation_provider" "sample" {
  provider_name = "%s"
  provider_type = "google"
  display_name  = "TF Acc DS Google"
  standard_type = "OAUTH2"

  oauth2 = {
    client_id     = "acc_ds_client"
    client_secret = "acc_ds_secret"
  }
}

data "cidaas_federation_provider" "by_id" {
  id = cidaas_federation_provider.sample.id
}

data "cidaas_federation_provider" "by_name" {
  provider_name = cidaas_federation_provider.sample.provider_name
}
`, acctest.GetBaseURL(), providerName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.cidaas_federation_provider.by_id", "id", resourceName, "id"),
					resource.TestCheckResourceAttr("data.cidaas_federation_provider.by_id", "provider_type", "google"),
					resource.TestCheckResourceAttr("data.cidaas_federation_provider.by_id", "oauth2.client_id", "acc_ds_client"),
					resource.TestCheckResourceAttrPair("data.cidaas_federation_provider.by_name", "id", resourceName, "id"),
					resource.TestCheckResourceAttr("data.cidaas_federation_provider.by_name", "display_name", "TF Acc DS Google"),
				),
			},
		},
	})
}
