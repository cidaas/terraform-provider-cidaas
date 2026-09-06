package group_test

import (
	"fmt"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupSelection_Basic(t *testing.T) {
	t.Parallel()
	acctest.SkipIfV3(t)

	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_GROUP_SELECTION, testResourceID)
	groupType := acctest.RandString(10)
	groupID := acctest.RandString(10)
	name := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupSelectionConfig(testResourceID, groupType, groupID, name, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttr(testResourceName, "name", name),
					resource.TestCheckResourceAttr(testResourceName, "always_show_group_selection", "false"),
				),
			},
			{
				ResourceName:            testResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
			{
				Config: testAccGroupSelectionConfig(testResourceID, groupType, groupID, name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "always_show_group_selection", "true"),
				),
			},
		},
	})
}

func testAccGroupSelectionConfig(resourceID, groupType, groupID, name string, alwaysShow bool) string {
	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_group_type" "%s" {
			group_type = "%s"
			role_mode = "any_roles"
			description = "Sample Group Type Description"
		}
		resource "cidaas_user_groups" "%s" {
			group_type = cidaas_group_type.%s.group_type
			group_id = "%s"
			group_name = "Sample Group %s"
		}
		resource "cidaas_group_selection" "%s" {
			name = "%s"
			description = "Test Group Selection"
			is_group_login_selection_enabled = true
			always_show_group_selection = %t
			selectable_groups = [cidaas_user_groups.%s.group_id]
			selectable_group_types = [cidaas_group_type.%s.group_type]
		}
	`, acctest.GetBaseURL(), resourceID, groupType, resourceID, resourceID, groupID, groupID, resourceID, name, alwaysShow, resourceID, resourceID)
}
