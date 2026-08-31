package consent_test

import (
	"context"
	"fmt"
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

func TestAccConsentGroupResource_Basic(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandString(10)
	description := "Test consent Description"
	updatedDescription := "Updated consent Description"

	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_CONSENT_GROUP, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckConsentGroupDestroyed(testResourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccConsentGroupResourceConfig(groupName, description, testResourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "group_name", groupName),
					resource.TestCheckResourceAttr(testResourceName, "description", description),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testResourceName, "updated_at"),
				),
			},
			{
				ResourceName:                         testResourceName,
				ImportStateVerifyIdentifierAttribute: "id",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"updated_at", "created_at"},
			},
			{
				Config: testAccConsentGroupResourceConfig(groupName, updatedDescription, testResourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "description", updatedDescription),
					resource.TestCheckResourceAttrSet(testResourceName, "updated_at"),
				),
			},
		},
	})
}

func testAccConsentGroupResourceConfig(groupName, description, resourceID string) string {
	return fmt.Sprintf(`
	provider "cidaas" {
		base_url = "%s"
		cidaas_version = "3.x"
	}
	resource "cidaas_consent_group" "%s" {
		group_name  = "%s"
		description = "%s"
	}
	`, acctest.GetBaseURL(), resourceID, groupName, description)
}

func testCheckConsentGroupDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		consentGroup := cidaas.ConsentGroup{
			ClientConfig: cidaas.ClientConfig{
				BaseURL:     os.Getenv("BASE_URL"),
				AccessToken: acctest.TestToken,
			},
		}

		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			res, err := consentGroup.Get(context.Background(), rs.Primary.ID)
			if res == nil {
				return nil
			}
			if err != nil {
				if strings.Contains(err.Error(), "not found") ||
					strings.Contains(err.Error(), "404") ||
					strings.Contains(err.Error(), "204") {
					return nil
				}
				return fmt.Errorf("error checking if consent group exists: %w", err)
			}
			if i == maxRetries-1 {
				return fmt.Errorf("consent group still exists after %d retries: %+v", maxRetries, res)
			}
			waitTime := time.Duration(i+1) * time.Second * 2
			time.Sleep(waitTime)
		}

		return nil
	}
}

func TestAccConsentGroupResource_GoupNameUpdateFail(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandString(10)
	description := "Test consent Description"
	updateGroupName := acctest.RandString(10)

	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_CONSENT_GROUP, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConsentGroupResourceConfig(groupName, description, testResourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "group_name", groupName),
				),
			},
			{
				Config:      testAccConsentGroupResourceConfig(updateGroupName, description, testResourceID),
				ExpectError: regexp.MustCompile(`Attribute 'group_name' can't be modified.`),
			},
		},
	})
}

func TestAccConsentGroupResource_EmptyGroupName(t *testing.T) {
	t.Parallel()

	description := "Test consent Description"
	emptyGroupName := ""

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccConsentGroupResourceConfig(emptyGroupName, description, acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`Attribute group_name string length must be at least 1, got: 0`),
			},
		},
	})
}

func TestAccConsentGroupResource_MissingRequired(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
					cidaas_version = "3.x"
				}
				resource "cidaas_consent_group" "%s" {
					description = "test description"
				}
				`, acctest.GetBaseURL(), acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`The argument "group_name" is required, but no definition was found.`),
			},
		},
	})
}
