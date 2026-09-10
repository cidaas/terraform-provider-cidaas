package hostedpages_test

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

const (
	hostedPageURL     = "https://cidaad.de/register_success"
	hostedPageID      = "register_success"
	hostedPageContent = "<html>Register Success</html>"
	defaultLocale     = "en-US"
)

var hostedPages = []map[string]string{
	{
		"hosted_page_id": hostedPageID,
		"locale":         defaultLocale,
		"url":            hostedPageURL,
		"content":        hostedPageContent,
	},
}

func TestAccHostedPageResource_Basic(t *testing.T) {
	t.Parallel()

	updatedHostedPageURL := "https://cidaad.de/updated_register_success"
	updatedHostedPages := []map[string]string{
		{
			"hosted_page_id": hostedPageID,
			"locale":         defaultLocale,
			"url":            updatedHostedPageURL,
			"content":        "<html>Updated Success</html>",
		},
	}

	resourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_HOSTED_PAGE, resourceID)
	hostedPageGroupName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckHostedPageDestroyed(testResourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedPageResourceConfig(hostedPageGroupName, defaultLocale, resourceID, hostedPages),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckHostedPageExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "hosted_page_group_name", hostedPageGroupName),
					resource.TestCheckResourceAttr(testResourceName, "default_locale", defaultLocale),
					resource.TestCheckResourceAttr(testResourceName, "hosted_pages.0.hosted_page_id", hostedPageID),
					resource.TestCheckResourceAttr(testResourceName, "hosted_pages.0.locale", defaultLocale),
					resource.TestCheckResourceAttr(testResourceName, "hosted_pages.0.url", hostedPageURL),
					resource.TestCheckResourceAttr(testResourceName, "hosted_pages.0.content", hostedPageContent),
				),
			},
			{
				ResourceName:            testResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
			{
				Config: testAccHostedPageResourceConfig(hostedPageGroupName, defaultLocale, resourceID, updatedHostedPages),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "hosted_pages.0.url", updatedHostedPageURL),
					resource.TestCheckResourceAttr(testResourceName, "hosted_pages.0.content", "<html>Updated Success</html>"),
				),
			},
		},
	})
}

func TestAccHostedPageResource_InvalidLocale(t *testing.T) {
	t.Parallel()

	invalidLocale := "invalid-locale"
	hostedPageGroupName := acctest.RandString(10)
	resourceID := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccHostedPageResourceConfig(hostedPageGroupName, invalidLocale, resourceID, hostedPages),
				ExpectError: regexp.MustCompile("Attribute default_locale value must be one of"),
			},
		},
	})
}

func TestAccHostedPageResource_MissingRequiredFields(t *testing.T) {
	t.Parallel()

	config1 := fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_hosted_page" "%s" {
			hosted_page_group_name = ""
			default_locale = "en-US"
			hosted_pages =[{
				hosted_page_id = "register_success"
				url = ""
			}]
		}
		`, acctest.GetBaseURL(), acctest.RandString(10))
	config2 := fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_hosted_page" "%s" {
			hosted_page_group_name = ""
			default_locale = "en-US"
		}
		`, acctest.GetBaseURL(), acctest.RandString(10))
	config3 := fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_hosted_page" "%s" {
			hosted_page_group_name = ""
			default_locale = "en-US"
			hosted_pages =[]
		}
		`, acctest.GetBaseURL(), acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      config1,
				ExpectError: regexp.MustCompile("Attribute hosted_page_group_name string length must be at least 1, got: 0"),
			},
			{
				Config:      config2,
				ExpectError: regexp.MustCompile(`The argument "hosted_pages" is required, but no definition was found.`),
			},
			{
				Config:      config3,
				ExpectError: regexp.MustCompile(`Attribute hosted_pages set must contain at least 1 elements, got: 0`),
			},
		},
	})
}

func TestAccHostedPageResource_UniqueIdentifier(t *testing.T) {
	t.Parallel()

	updatedHostedPageGroupName := "Updated Hosted Page Group"
	resourceID := acctest.RandString(10)
	hostedPageGroupName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostedPageResourceConfig(hostedPageGroupName, defaultLocale, resourceID, hostedPages),
			},
			{
				Config:      testAccHostedPageResourceConfig(updatedHostedPageGroupName, defaultLocale, resourceID, hostedPages),
				ExpectError: regexp.MustCompile("Attribute 'hosted_page_group_name' can't be modified"),
			},
		},
	})
}

func testAccHostedPageResourceConfig(hostedPageGroupName, defaultLocale, resourceID string, pages []map[string]string) string {
	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_hosted_page" "%s" {
			hosted_page_group_name = "`+hostedPageGroupName+`"
			default_locale = "`+defaultLocale+`"
			hosted_pages =[{
				hosted_page_id = "`+pages[0]["hosted_page_id"]+`"
				locale = "`+pages[0]["locale"]+`"
				url = "`+pages[0]["url"]+`"
				content = "`+pages[0]["content"]+`"
			}]
		}
	`, acctest.GetBaseURL(), resourceID)
}

func testCheckHostedPageExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, ok := s.RootModule().Resources[resourceName]; !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		return nil
	}
}

func testCheckHostedPageDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		hp := cidaas.HostedPage{
			ClientConfig: cidaas.ClientConfig{
				BaseURL:     os.Getenv("BASE_URL"),
				AccessToken: acctest.TestToken,
			},
		}

		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			res, err := hp.Get(context.Background(), rs.Primary.Attributes["hosted_page_group_name"])
			if res == nil {
				return nil
			}
			if err != nil {
				if strings.Contains(err.Error(), "not found") ||
					strings.Contains(err.Error(), "404") ||
					strings.Contains(err.Error(), "204") {
					return nil
				}
				return fmt.Errorf("error checking if hosted page exists: %w", err)
			}
			if i == maxRetries-1 {
				return fmt.Errorf("hosted page still exists after %d retries", maxRetries)
			}
			time.Sleep(time.Duration(i+1) * time.Second * 2)
		}
		return nil
	}
}
