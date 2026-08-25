package hostedpages_test

import (
	"fmt"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTheme_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	name := fmt.Sprintf("tf-v4-theme-%s.css", acctest.RandString(6))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ProviderConfig() + fmt.Sprintf(`
resource "cidaas_theme" "test" {
  filename    = %q
  css_content = "body { background: #111111; }"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cidaas_theme.test", "filename", name),
					resource.TestCheckResourceAttrSet("cidaas_theme.test", "css_hash"),
				),
			},
			{
				Config: acctest.ProviderConfig() + fmt.Sprintf(`
resource "cidaas_theme" "test" {
  filename    = %q
  css_content = "body { background: #222222; }"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cidaas_theme.test", "css_content", "body { background: #222222; }"),
				),
			},
		},
	})
}
