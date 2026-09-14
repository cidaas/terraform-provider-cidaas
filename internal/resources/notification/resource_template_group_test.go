package notification_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func checkTemplateGroupDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		tg := cidaas.TemplateGroup{
			ClientConfig: cidaas.ClientConfig{
				BaseURL:     os.Getenv("BASE_URL"),
				AccessToken: acctest.TestToken,
			},
		}

		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			res, err := tg.Get(context.Background(), rs.Primary.Attributes["group_id"])

			if res == nil || res.Status == http.StatusNoContent {
				return nil
			}

			if err != nil {
				if strings.Contains(err.Error(), "not found") ||
					strings.Contains(err.Error(), "404") ||
					strings.Contains(err.Error(), "204") {
					return nil
				}
				return fmt.Errorf("error checking if template group exists: %w", err)
			}

			if i == maxRetries-1 {
				return fmt.Errorf("template group still exists after %d retries: %+v", maxRetries, res)
			}

			waitTime := time.Duration(i+1) * time.Second * 2
			time.Sleep(waitTime)
		}

		return nil
	}
}

func TestTemplateGroup_GroupIDLengthCheck(t *testing.T) {
	t.Parallel()

	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_TEMPLATE_GROUP, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkTemplateGroupDestroyed(testResourceName),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					provider "cidaas" {
						base_url = "%s"
					}
					resource "cidaas_template_group" "%s" {
						group_id  = "`+acctest.RandString(16)+`"
					}		
				`, acctest.GetBaseURL(), testResourceID),
				ExpectError: regexp.MustCompile("group_id string length must be at most 15"),
			},
		},
	})
}
