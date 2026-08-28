package resources_test

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// create, read and update test
func TestUserGroup_Basic(t *testing.T) {
	t.Parallel()

	userGroupType := acctest.RandString(10)
	groupID := acctest.RandString(10)
	userGroupName := acctest.RandString(10)
	userGroupDescription := "sample user groups description"
	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", resources.RESOURCE_USER_GROUP, testResourceID)
	resourceGroupTypeName := fmt.Sprintf("%s.%s", resources.RESOURCE_GROUP_TYPE, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckUserGroupDestroyed(testResourceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupResourceConfig(userGroupType, groupID, userGroupDescription, testResourceID, userGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(testResourceName, "group_type", resourceGroupTypeName, "group_type"),
					resource.TestCheckResourceAttr(testResourceName, "group_id", groupID),
					resource.TestCheckResourceAttr(testResourceName, "group_name", userGroupName),
					resource.TestCheckResourceAttr(testResourceName, "logo_url", "https://cidaas.de/logo"),
					resource.TestCheckResourceAttr(testResourceName, "description", userGroupDescription),
					resource.TestCheckResourceAttr(testResourceName, "custom_fields.first_name", "cidaas"),
					resource.TestCheckResourceAttr(testResourceName, "custom_fields.family_name", "widas"),
					// default value check
					resource.TestCheckResourceAttr(testResourceName, "make_first_user_admin", strconv.FormatBool(true)),
					resource.TestCheckResourceAttr(testResourceName, "member_profile_visibility", "full"),
					resource.TestCheckResourceAttr(testResourceName, "none_member_profile_visibility", "public"),
					resource.TestCheckResourceAttr(testResourceName, "parent_id", "root"),
					// computed properties check
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testResourceName, "updated_at"),
				),
			},
			{
				ResourceName:            testResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           groupID,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
			{
				Config: testAccUserGroupResourceConfig(userGroupType, groupID, "updated user group description", testResourceID, userGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "group_type", userGroupType),
					resource.TestCheckResourceAttr(testResourceName, "description", "updated user group description"),
				),
			},
		},
	})
}

func testAccUserGroupResourceConfig(groupType, groupID, groupDescription, resourceID, groupName string) string {
	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_group_type" "%s" {
			group_type  = "`+groupType+`"
			role_mode   = "no_roles"
			description = "group type description"
		}
		resource "cidaas_user_groups" "%s" {
			group_type                     = cidaas_group_type.%s.group_type
			group_id                       = "`+groupID+`"
			group_name                     = "`+groupName+`"
			logo_url                       = "https://cidaas.de/logo"
			description                    = "`+groupDescription+`"
			custom_fields = {
				first_name  = "cidaas"
				family_name = "widas"
			}
			make_first_user_admin          = true
			member_profile_visibility      = "full"
			none_member_profile_visibility = "public"
		}		
	`, acctest.GetBaseURL(), resourceID, resourceID, resourceID)
}

func testCheckUserGroupDestroyed(resourceID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Check User Group destruction with retry logic
		resourceGroupName := fmt.Sprintf("%s.%s", resources.RESOURCE_USER_GROUP, resourceID)
		rs, ok := s.RootModule().Resources[resourceGroupName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceGroupName)
		}

		ug := cidaas.UserGroup{
			ClientConfig: cidaas.ClientConfig{
				BaseURL:     acctest.GetBaseURL(),
				AccessToken: acctest.TestToken,
			},
		}

		// Add retry logic for user group eventual consistency
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			res, err := ug.Get(context.Background(), rs.Primary.Attributes["group_id"])
			if err != nil {
				// If error is "not found", that's what we want
				if strings.Contains(err.Error(), "not found") ||
					strings.Contains(err.Error(), "404") ||
					strings.Contains(err.Error(), "204") {
					break // User group successfully deleted, continue to group type check
				}
				return fmt.Errorf("error checking if user group exists: %w", err)
			}

			// Check if resource is nil
			if res == nil {
				break // User group successfully deleted, continue to group type check
			}

			// If this is the last retry, return error
			if i == maxRetries-1 {
				return fmt.Errorf("user group still exists after %d retries: %+v", maxRetries, res)
			}

			// Wait before retrying with exponential backoff
			waitTime := time.Duration(i+1) * time.Second * 2
			time.Sleep(waitTime)
		}

		// Check Group Type destruction with retry logic
		resourceGroupTypeName := fmt.Sprintf("%s.%s", resources.RESOURCE_GROUP_TYPE, resourceID)
		rs, ok = s.RootModule().Resources[resourceGroupTypeName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceGroupTypeName)
		}

		groupType := cidaas.GroupType{
			ClientConfig: cidaas.ClientConfig{
				BaseURL:     acctest.GetBaseURL(),
				AccessToken: acctest.TestToken,
			},
		}

		// Add retry logic for group type eventual consistency
		for i := 0; i < maxRetries; i++ {
			resp, err := groupType.Get(context.Background(), rs.Primary.Attributes["group_type"])
			if err != nil {
				if strings.Contains(err.Error(), "not found") ||
					strings.Contains(err.Error(), "404") ||
					strings.Contains(err.Error(), "204") {
					return nil // Both resources successfully deleted
				}
				return fmt.Errorf("error checking if group type exists: %w", err)
			}

			// Check if resource is nil
			if resp == nil {
				return nil // Both resources successfully deleted
			}

			// If this is the last retry, return error
			if i == maxRetries-1 {
				return fmt.Errorf("group type still exists after %d retries: %s", maxRetries, resp.Data.GroupType)
			}

			// Wait before retrying with exponential backoff
			waitTime := time.Duration(i+1) * time.Second * 2
			time.Sleep(waitTime)
		}

		return nil
	}
}

// missing required fields group_type, group_name and group_id
func TestUserGroup_MissingRequired(t *testing.T) {
	t.Parallel()
	requiredParams := []string{"group_name", "group_id"}

	for _, v := range requiredParams {
		v := v // Capture loop variable
		t.Run(fmt.Sprintf("missing_%s", v), func(t *testing.T) {
			t.Parallel()
			testResourceID := acctest.RandString(10)

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acctest.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`
                        provider "cidaas" {
                            base_url = "%s"
                        }
                        resource "cidaas_user_groups" "%s" {}
                    `, acctest.GetBaseURL(), testResourceID),
						ExpectError: regexp.MustCompile(fmt.Sprintf(`The argument "%s" is required`, v)),
					},
				},
			})
		})
	}
}

// check if group_type, group_name and group_id are empty string
func TestUserGroup_CheckEmptyString(t *testing.T) {
	t.Parallel()
	testResourceID := acctest.RandString(10)
	requiredParams := []string{"group_type", "group_name", "group_id"}
	for _, v := range requiredParams {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acctest.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
					provider "cidaas" {
						base_url = "%s"
					}
					resource "cidaas_user_groups" "%s" {
						group_type  =""
						group_id    = ""
						group_name  = ""
					}
				`, acctest.GetBaseURL(), testResourceID),
					ExpectError: regexp.MustCompile(fmt.Sprintf(`Attribute %s string length must be at least 1`, v)),
				},
			},
		})
	}
}

// group_id can not be modified
func TestUserGroup_GroupIDIsImmutable(t *testing.T) {
	t.Parallel()

	groupType := acctest.RandString(10)
	groupID := acctest.RandString(10)
	updatedGroupID := acctest.RandString(10)
	groupName := acctest.RandString(10)
	groupDescription := "sample user groups description"

	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", resources.RESOURCE_USER_GROUP, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create group type first so user group create does not 404 on a missing type.
			{
				Config: testAccGroupTypeOnlyConfig(groupType, testResourceID),
			},
			{
				Config: testAccUserGroupResourceConfig(groupType, groupID, groupDescription, testResourceID, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "group_id", groupID),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
			{
				Config:      testAccUserGroupResourceConfig(groupType, updatedGroupID, groupDescription, testResourceID, groupName),
				ExpectError: regexp.MustCompile("Attribute 'group_id' can't be modified"),
			},
		},
	})
}

func testAccGroupTypeOnlyConfig(groupType, resourceID string) string {
	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_group_type" "%s" {
			group_type  = "%s"
			role_mode   = "no_roles"
			description = "group type description"
		}
	`, acctest.GetBaseURL(), resourceID, groupType)
}
