package hostedpages_test

import (
	"fmt"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTranslations_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	acctest.SkipIfV3(t)
	// Prefer a fresh synthetic locale suffix via custom key content; locale itself must be valid (fr).
	suffix := acctest.RandString(4)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ProviderConfig() + fmt.Sprintf(`
resource "cidaas_translations" "test" {
  locale_id = "fr"
  enabled   = true
  translations = {
    "login.title"  = "Welcome %s"
    "login.submit" = "Submit"
  }
}
`, suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cidaas_translations.test", "locale_id", "fr"),
					resource.TestCheckResourceAttr("cidaas_translations.test", "translations.login.submit", "Submit"),
				),
			},
			{
				Config: acctest.ProviderConfig() + fmt.Sprintf(`
resource "cidaas_translations" "test" {
  locale_id = "fr"
  enabled   = true
  translations = {
    "login.title"  = "Updated %s"
    "login.submit" = "Go"
  }
}
`, suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cidaas_translations.test", "translations.login.submit", "Go"),
				),
			},
		},
	})
}
