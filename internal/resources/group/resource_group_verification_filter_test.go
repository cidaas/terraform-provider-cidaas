package group_test

import (
	"fmt"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupVerificationFilter_Basic(t *testing.T) {
	t.Parallel()
	acctest.SkipIfV3(t)

	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_GROUP_VERIFICATION_FILTER, testResourceID)
	description := "Test verification filter description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupVerificationFilterConfig(testResourceID, description, "or"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "id", testResourceID),
					resource.TestCheckResourceAttr(testResourceName, "description", description),
					resource.TestCheckResourceAttr(testResourceName, "match_condition", "or"),
				),
			},
			{
				ResourceName:            testResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
			{
				Config: testAccGroupVerificationFilterConfig(testResourceID, description, "and"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "match_condition", "and"),
				),
			},
		},
	})
}

func testAccGroupVerificationFilterConfig(resourceID, description, matchCondition string) string {
	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_group_verification_filter" "%s" {
			id = "%s"
			description = "%s"
			match_condition = "%s"
			filters {
				group_id = "CIDAAS_ADMINS"
				role_filter {
					roles = ["ADMIN"]
					match_condition = "or"
				}
			}
		}
	`, acctest.GetBaseURL(), resourceID, resourceID, description, matchCondition)
}
