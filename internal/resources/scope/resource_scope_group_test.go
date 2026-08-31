package scope_test

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

func TestAccScopeGroupResource_Basic(t *testing.T) {
	t.Parallel()

	scopeGroupName := acctest.RandString(10)
	scopeGroupdescription := "Test Scope Group Description"
	updatedDescription := "Updated Scope Group Description"

	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_SCOPE_GROUP, scopeGroupName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckScopeGroupDestroyed(testResourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeGroupResourceConfig(scopeGroupName, scopeGroupdescription, scopeGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "group_name", scopeGroupName),
					resource.TestCheckResourceAttr(testResourceName, "description", scopeGroupdescription),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testResourceName, "updated_at"),
				),
			},
			{
				ResourceName:      testResourceName,
				ImportStateId:     scopeGroupName,
				ImportState:       true,
				ImportStateVerify: true,
				// TODO: remove ImportStateVerifyIgnore
				ImportStateVerifyIgnore: []string{"updated_at", "created_at"},
			},
			{
				Config: testAccScopeGroupResourceConfig(scopeGroupName, updatedDescription, scopeGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "description", updatedDescription),
					resource.TestCheckResourceAttrSet(testResourceName, "updated_at"),
				),
			},
		},
	})
}

func testAccScopeGroupResourceConfig(groupType, description, resourceID string) string {
	return fmt.Sprintf(`
	provider "cidaas" {
		base_url = "%s"
	}
	resource "cidaas_scope_group" "%s" {
		group_name  = "%s"
		description = "%s"
	}
	`, acctest.GetBaseURL(), resourceID, groupType, description)
}

func testCheckScopeGroupDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		scopeGroup := cidaas.ScopeGroup{
			ClientConfig: cidaas.ClientConfig{
				BaseURL:     acctest.GetBaseURL(),
				AccessToken: acctest.TestToken,
			},
		}

		// Add retry logic for eventual consistency
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			res, err := scopeGroup.Get(context.Background(), rs.Primary.Attributes["group_name"])
			if err != nil {
				// If error is "not found", that's what we want
				if strings.Contains(err.Error(), "not found") ||
					strings.Contains(err.Error(), "404") ||
					strings.Contains(err.Error(), "204") {
					return nil
				}
				return fmt.Errorf("error checking if scope group exists: %w", err)
			}

			// Check if resource is successfully deleted (empty ID or nil response)
			if res == nil || res.Data.ID == "" {
				return nil // Resource successfully deleted
			}

			// If this is the last retry, return error
			if i == maxRetries-1 {
				return fmt.Errorf("scope group still exists after %d retries: %+v", maxRetries, res)
			}

			// Wait before retrying with exponential backoff
			waitTime := time.Duration(i+1) * time.Second * 2
			time.Sleep(waitTime)
		}

		return nil
	}
}

// failed validation on updating immutable proprty group_name
func TestAccScopeGroupResource_GoupNameUpdateFail(t *testing.T) {
	t.Parallel()

	scopeGroupName := acctest.RandString(10)
	updateGroupName := acctest.RandString(10)
	scopeGroupdescription := "Test Scope Group Description"

	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_SCOPE_GROUP, scopeGroupName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScopeGroupResourceConfig(scopeGroupName, scopeGroupdescription, scopeGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "group_name", scopeGroupName),
				),
			},
			{
				Config:      testAccScopeGroupResourceConfig(updateGroupName, scopeGroupdescription, scopeGroupName),
				ExpectError: regexp.MustCompile("Attribute 'group_name' can't be modified"), // TODO: full string comparison
			},
		},
	})
}

// Empty group_name validation test
func TestAccScopeGroupResource_EmptyGroupName(t *testing.T) {
	t.Parallel()

	scopeGroupdescription := "Test Scope Group Description"
	emptyScopeGroupName := ""

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccScopeGroupResourceConfig(emptyScopeGroupName, scopeGroupdescription, acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`Attribute group_name string length must be at least 1, got: 0`),
			},
		},
	})
}

// missing required parameter
func TestAccScopeGroupResource_MissingRequired(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_scope_group" "%s" {
					description = "test description"
				}
				`, acctest.GetBaseURL(), acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`The argument "group_name" is required, but no definition was found.`),
			},
		},
	})
}
