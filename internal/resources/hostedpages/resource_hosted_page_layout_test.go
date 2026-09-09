package hostedpages_test

import (
	"fmt"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHostedPageLayout_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	acctest.SkipIfV3(t)
	suffix := acctest.RandString(6)
	groupName := fmt.Sprintf("tf_v4_hpg_%s", suffix)
	themeFile := fmt.Sprintf("tf-v4-theme-%s.css", suffix)
	base := acctest.BaseURL()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ProviderConfig() + fmt.Sprintf(`
resource "cidaas_hosted_page" "grp" {
  hosted_page_group_name = %q
  default_locale         = "en"
  hosted_pages = [
    {
      hosted_page_id = "login"
      locale         = "en"
      url            = "%s/login-%s"
    }
  ]
}

resource "cidaas_theme" "css" {
  filename    = %q
  css_content = "body { background-color: #ffffff; }"
}

resource "cidaas_hosted_page_layout" "layout" {
  description = "acc-test-layout-%s"
  layout = {
    hosted_page_group = cidaas_hosted_page.grp.hosted_page_group_name
    theme             = cidaas_theme.css.filename
    primary_color     = "#123456"
    accent_color      = "#abcdef"
    content_align     = "CENTER"
    media_type        = "IMAGE"
  }
  resources = {
    "default-hosted-pages-webapp" = {
      translation_set = "default"
      theme           = cidaas_theme.css.filename
    }
  }
}
`, groupName, base, suffix, themeFile, suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cidaas_hosted_page_layout.layout", "description", fmt.Sprintf("acc-test-layout-%s", suffix)),
					resource.TestCheckResourceAttr("cidaas_hosted_page_layout.layout", "layout.primary_color", "#123456"),
					resource.TestCheckResourceAttrSet("cidaas_hosted_page_layout.layout", "id"),
					resource.TestCheckResourceAttrSet("cidaas_hosted_page_layout.layout", "fingerprint"),
				),
			},
			{
				Config: acctest.ProviderConfig() + fmt.Sprintf(`
resource "cidaas_hosted_page" "grp" {
  hosted_page_group_name = %q
  default_locale         = "en"
  hosted_pages = [
    {
      hosted_page_id = "login"
      locale         = "en"
      url            = "%s/login-%s"
    }
  ]
}

resource "cidaas_theme" "css" {
  filename    = %q
  css_content = "body { background-color: #ffffff; }"
}

resource "cidaas_hosted_page_layout" "layout" {
  description = "acc-test-layout-updated-%s"
  layout = {
    hosted_page_group = cidaas_hosted_page.grp.hosted_page_group_name
    theme             = cidaas_theme.css.filename
    primary_color     = "#654321"
    accent_color      = "#fedcba"
    content_align     = "CENTER"
    media_type        = "IMAGE"
  }
  resources = {
    "default-hosted-pages-webapp" = {
      translation_set = "default"
      theme           = cidaas_theme.css.filename
    }
  }
}
`, groupName, base, suffix, themeFile, suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cidaas_hosted_page_layout.layout", "description", fmt.Sprintf("acc-test-layout-updated-%s", suffix)),
					resource.TestCheckResourceAttr("cidaas_hosted_page_layout.layout", "layout.primary_color", "#654321"),
				),
			},
		},
	})
}
