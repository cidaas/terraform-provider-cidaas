package resources_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccGroupTypeResource_Basic(t *testing.T) {
	t.Parallel()

	roleMode := "any_roles"
	updatedDescription := "Updated Group Type Description"
	groupType := acctest.RandString(10)
	description := "Test Group Type Description"

	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", resources.RESOURCE_GROUP_TYPE, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckGroupTypeDestroyed(testResourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupTypeResourceConfig(groupType, roleMode, description, testResourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "group_type", groupType),
					resource.TestCheckResourceAttr(testResourceName, "role_mode", roleMode),
					resource.TestCheckResourceAttr(testResourceName, "description", description),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "created_at"),
				),
			},
			{
				ResourceName:      testResourceName,
				ImportStateId:     groupType,
				ImportState:       true,
				ImportStateVerify: true,
				// remove ImportStateVerifyIgnore to enhance the result
				ImportStateVerifyIgnore: []string{"updated_at", "created_at"},
			},
			{
				Config: testAccGroupTypeResourceConfig(groupType, roleMode, updatedDescription, testResourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "description", updatedDescription),
					resource.TestCheckResourceAttrSet(testResourceName, "updated_at"),
				),
			},
		},
	})
}

func testAccGroupTypeResourceConfig(groupType, roleMode, description, resourceID string) string {
	return fmt.Sprintf(`
	provider "cidaas" {
		base_url = "%s"
	}
	resource "cidaas_group_type" "%s" {
		group_type  = "%s"
		role_mode   = "%s"
		description = "%s"
	}
	`, acctest.GetBaseURL(), resourceID, groupType, roleMode, description)
}

func testCheckGroupTypeDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		groupType := cidaas.GroupType{
			ClientConfig: cidaas.ClientConfig{
				BaseURL:     os.Getenv("BASE_URL"),
				AccessToken: acctest.TestToken,
			},
		}

		// Add retry logic for eventual consistency
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			res, err := groupType.Get(context.Background(), rs.Primary.Attributes["group_type"])

			// Check if resource is successfully deleted (nil response)
			if res == nil {
				return nil // Resource successfully deleted
			}

			// Handle other errors
			if err != nil {
				// If error is "not found", that's what we want
				if strings.Contains(err.Error(), "not found") ||
					strings.Contains(err.Error(), "404") ||
					strings.Contains(err.Error(), "204") {
					return nil
				}
				return fmt.Errorf("error checking if group type exists: %w", err)
			}

			// If this is the last retry, return error
			if i == maxRetries-1 {
				return fmt.Errorf("group type still exists after %d retries: %s", maxRetries, res.Data.GroupType)
			}

			// Wait before retrying with exponential backoff
			waitTime := time.Duration(i+1) * time.Second * 2
			time.Sleep(waitTime)
		}

		return nil
	}
}

// validation test for role_mode
func TestAccGroupTypeResource_InvalidRoleMode(t *testing.T) {
	t.Parallel()

	invalidRoleMode := "invalid_role_mode"
	groupType := acctest.RandString(10)
	description := "Test Group Type Description"

	testResourceID := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupTypeResourceConfig(groupType, invalidRoleMode, description, testResourceID),
				ExpectError: regexp.MustCompile(`Attribute role_mode value must be one of:`), // TODO: full string comparison
			},
		},
	})
}

// group_type can't be modified
func TestAccGroupTypeResource_UpdateFails(t *testing.T) {
	t.Parallel()

	updatedGroupType := acctest.RandString(10)
	roleMode := "any_roles"
	groupType := acctest.RandString(10)
	description := "Test Group Type Description"

	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", resources.RESOURCE_GROUP_TYPE, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupTypeResourceConfig(groupType, roleMode, description, testResourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "group_type", groupType),
				),
			},
			{
				Config:      testAccGroupTypeResourceConfig(updatedGroupType, roleMode, description, testResourceID),
				ExpectError: regexp.MustCompile("Attribute 'group_type' can't be modified"), // TODO: full string comparison
			},
		},
	})
}

// allowed_roles must have value when role_mode is allowed_roles or roles_required
func TestAccGroupTypeResource_EmptyAllowedRolesError(t *testing.T) {
	t.Parallel()

	roleMode := "allowed_roles"
	groupType := acctest.RandString(10)
	description := "Test Group Type Description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_group_type" "%s" {
					group_type  = "%s"
					role_mode   = "%s"
					description = "%s"
					allowed_roles = []
				}
				`, acctest.GetBaseURL(), groupType, groupType, roleMode, description),
				ExpectError: regexp.MustCompile("The attribute allowed_roles cannot be empty when role_mode is set to"), // TODO: full string comparison
			},
		},
	})
}

// allowed_roles with unknown resource reference should not crash with value conversion error
func TestAccGroupTypeResource_UnknownAllowedRolesReference(t *testing.T) {
	t.Parallel()

	roleMode := "allowed_roles"
	groupType := acctest.RandString(10)
	description := "Test Group Type Description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_role" "test_role" {
					role = "ROLE_%s"
					description = "test role"
				}
				resource "cidaas_group_type" "%s" {
					group_type  = "%s"
					role_mode   = "%s"
					description = "%s"
					allowed_roles = [cidaas_role.test_role.id]
				}
				`, acctest.GetBaseURL(), groupType, groupType, groupType, roleMode, description),
			},
		},
	})
}
