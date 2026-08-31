package role_test

import (
	"context"
	"fmt"
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

func TestAccRoleResource(t *testing.T) {
	t.Parallel()
	role := acctest.RandString(10)
	name := "Terraform Admin Role"
	updatedName := "Updated Terraform Admin Role"
	description := "This is a test terraform admin role"
	updatedDescription := "This is a test terraform admin updated role"

	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_ROLE, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		// verify check destroy
		CheckDestroy: testAccCheckRoleResourceDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRoleResourceConfig(role, name, description, testResourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "name", name),
					resource.TestCheckResourceAttr(testResourceName, "role", role),
					resource.TestCheckResourceAttr(testResourceName, "description", description),
					resource.TestCheckResourceAttr(testResourceName, "id", role),
				),
			},
			// ImportState testing(read after create)
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "role", role),
					resource.TestCheckResourceAttr(testResourceName, "name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", description),
				),
			},
			// Update and Read testing
			{
				Config: testAccRoleResourceConfig(role, updatedName, updatedDescription, testResourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", updatedDescription),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRoleResource_validateAttrSet(t *testing.T) {
	t.Parallel()
	role := acctest.RandString(10)
	name := "Test Terraform Admin Role"
	description := "This is a test terraform admin role"
	updatedDescription := "This is a test terraform admin updated role"
	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_ROLE, testResourceID)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccRoleResourceConfig(role, name, description, testResourceID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "role"),
					resource.TestCheckResourceAttrSet(testResourceName, "name"),
					resource.TestCheckResourceAttrSet(testResourceName, "description"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
			{
				ResourceName:                         testResourceName,
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[testResourceName]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					id, ok := rs.Primary.Attributes["role"]
					if !ok {
						return "", fmt.Errorf("role not set")
					}
					return id, nil
				},
				ImportStateVerify: true,
			},
			{
				Config: testAccRoleResourceConfig(role, name, updatedDescription, testResourceID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "role"),
					resource.TestCheckResourceAttrSet(testResourceName, "name"),
					resource.TestCheckResourceAttrSet(testResourceName, "description"),
				),
			},
		},
	})
}

func TestAccRoleResource_updateRoleFails(t *testing.T) {
	t.Parallel()
	role := acctest.RandString(10)
	updatedRole := acctest.RandString(10)
	name := "Test Role"
	description := "This is a test role"

	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_ROLE, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleResourceConfig(role, name, description, testResourceID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleResourceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "role", role),
				),
			},
			{
				Config:      testAccRoleResourceConfig(updatedRole, name, description, testResourceID),
				ExpectError: regexp.MustCompile("Attribute 'role' can't be modified"), // TODO: full string comparison
			},
		},
	})
}

func TestAccRoleResource_readAfterUpdate(t *testing.T) {
	t.Parallel()
	role := acctest.RandString(10)
	name := "Test Role"
	description := "This is a test role"

	updatedName := "Updated Test Role"
	updatedDescription := "This is an updated test role"

	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_ROLE, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleResourceConfig(role, name, description, testResourceID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleResourceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "role", role),
					resource.TestCheckResourceAttr(testResourceName, "name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", description),
				),
			},
			{
				Config: testAccRoleResourceConfig(role, updatedName, updatedDescription, testResourceID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleResourceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", updatedDescription),
				),
			},
			{
				ResourceName: testResourceName,
				ImportState:  true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "role", role),
					resource.TestCheckResourceAttr(testResourceName, "name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", updatedDescription),
				),
			},
		},
	})
}

func TestAccRoleResource_createMissingFields(t *testing.T) {
	t.Parallel()
	testResourceID := acctest.RandString(10)

	missingRoleConfig := fmt.Sprintf(`
	provider "cidaas" {
		base_url = "%s"
	}
	resource "cidaas_role" "%s" {
		name = "Test Name"
		description = "Test Description"
	}`, acctest.GetBaseURL(), testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      missingRoleConfig,
				ExpectError: regexp.MustCompile(`The argument "role" is required, but no definition was found.`), // TODO: full string comparison
			},
		},
	})
}

func testAccRoleResourceConfig(role, name, description, resourceID string) string {
	return fmt.Sprintf(`
	provider "cidaas" {
		base_url = "%s"
	}
	resource "cidaas_role" "%s" {
		role = "%s"
		name = "%s"
		description = "%s"
	}`, acctest.GetBaseURL(), resourceID, role, name, description)
}

func testAccCheckRoleResourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Role ID is set")
		}
		return nil
	}
}

func testAccCheckRoleResourceDestroyed(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != base.RESOURCE_ROLE {
			continue
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		role := cidaas.Role{
			ClientConfig: cidaas.ClientConfig{
				BaseURL:     acctest.GetBaseURL(),
				AccessToken: acctest.TestToken,
			},
		}

		// Add retry logic for eventual consistency
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			res, err := role.GetRole(context.Background(), rs.Primary.ID)
			if err != nil {
				// If error is "not found", that's what we want
				if strings.Contains(err.Error(), "not found") ||
					strings.Contains(err.Error(), "404") ||
					strings.Contains(err.Error(), "204") {

					break // Role successfully deleted, continue to next resource
				}
				return fmt.Errorf("error checking if role exists: %w", err)
			}

			// Check if resource is successfully deleted (nil response)
			if res == nil {
				break // Role successfully deleted, continue to next resource
			}

			// If this is the last retry, return error
			if i == maxRetries-1 {
				return fmt.Errorf("role still exists after %d retries: %+v", maxRetries, res)
			}

			// Wait before retrying with exponential backoff
			waitTime := time.Duration(i+1) * time.Second * 2
			time.Sleep(waitTime)
		}
	}
	return nil
}
