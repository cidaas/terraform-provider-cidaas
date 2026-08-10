package resources_test

import (
	"fmt"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/resources"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestConsentVersion_Basic(t *testing.T) {
	t.Parallel()

	testResourceID := acctest.RandString(10)
	groupName := acctest.RandString(10)
	consentName := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", resources.RESOURCE_CONSENT_VERSION, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testConsentVersionConfig("consent version in German", testResourceID, groupName, consentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "consent_type", "SCOPES"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[testResourceName]
					if !ok {
						return "", fmt.Errorf("Not found: %s", testResourceName)
					}
					return rs.Primary.Attributes["consent_id"] + ":" + rs.Primary.ID + ":de:en", nil
				},
			},
			{
				Config: testConsentVersionConfig("updated consent version in German", testResourceID, groupName, consentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "consent_type", "SCOPES"),
				),
			},
		},
	})
}

func testConsentVersionConfig(content, resourceID, groupName, consentName string) string {
	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_consent_group" "sample" {
			group_name  = "%s"
			description = "sample description"
		}
		resource "cidaas_consent" "sample" {
			consent_group_id = cidaas_consent_group.sample.id
			name             = "%s"
			enabled          = true
		}
		resource "cidaas_consent_version" "%s" {
			version         = 1
			consent_id      = cidaas_consent.sample.id
			consent_type    = "SCOPES"
			scopes          = ["developer"]
			required_fields = ["name"]
			consent_locales = [
				{
					content = "%s"
					locale  = "de"
				},
				{
					content = "consent version in English"
					locale  = "en"
				}
			]
		}		
	`, acctest.GetBaseURL(), groupName, consentName, resourceID, content)
}
