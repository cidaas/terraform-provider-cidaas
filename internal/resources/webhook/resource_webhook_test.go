package webhook_test

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
	apiKey         = "APIKEY"
	apiConfigKey   = "api-key"
	apiPlaceholder = "key"
	apiPlacement   = "query"
	totp           = "TOTP"
	oauth2         = "CIDAAS_OAUTH2"
)

var events = []string{"ACCOUNT_MODIFIED"}

func testWebhookURL(unique string) string {
	return fmt.Sprintf("https://cidaas.de/webhook-srv/webhook/%s", unique)
}

func getDefaultAPIKeyConfig() map[string]string {
	return map[string]string{
		"key":         apiConfigKey,
		"placeholder": apiPlaceholder,
		"placement":   apiPlacement,
	}
}

func TestAccWebhookResource_Basic(t *testing.T) {
	t.Parallel()
	urlSuffix := acctest.RandString(10)
	initialURL := testWebhookURL(urlSuffix)
	updatedURL := fmt.Sprintf("https://cidaas.de/webhook-srv/v2/webhook/%s", urlSuffix)

	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_WEBHOOK, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckWebhookDestroyed(testResourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookResourceConfig(apiKey, initialURL, testResourceID, events, getDefaultAPIKeyConfig()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "disable"),
					resource.TestCheckResourceAttrSet(testResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testResourceName, "updated_at"),
				),
			},
			{
				ResourceName:            testResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
			{
				Config: testAccWebhookResourceConfig(apiKey, updatedURL, testResourceID, events, getDefaultAPIKeyConfig()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "url", updatedURL),
				),
			},
		},
	})
}

func testAccWebhookResourceConfig(
	authType, url, resourceID string,
	events []string,
	apikeyConfig map[string]string,
) string {
	eventsString := "[]"
	if len(events) > 0 {
		eventsString = `["` + strings.Join(events, `", "`) + `"]`
	}

	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_webhook" "%s" {
			auth_type = "`+authType+`"
			url = "`+url+`"
			events = `+eventsString+`
			apikey_config = {
				key = "`+apikeyConfig["key"]+`"
				placeholder = "`+apikeyConfig["placeholder"]+`"
				placement = "`+apikeyConfig["placement"]+`"
			}
		}
	`, acctest.GetBaseURL(), resourceID)
}

func testCheckWebhookDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		wb := cidaas.Webhook{
			ClientConfig: cidaas.ClientConfig{
				BaseURL:     os.Getenv("BASE_URL"),
				AccessToken: acctest.TestToken,
			},
		}

		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			res, err := wb.Get(context.Background(), rs.Primary.Attributes["id"])
			if err != nil {
				if strings.Contains(err.Error(), "not found") ||
					strings.Contains(err.Error(), "404") ||
					strings.Contains(err.Error(), "204") {
					return nil
				}
				return fmt.Errorf("error checking if resource exists: %w", err)
			}

			if res == nil {
				return nil
			}

			if i == maxRetries-1 {
				return fmt.Errorf("resource still exists after %d retries: %+v", maxRetries, res)
			}

			waitTime := time.Duration(i+1) * time.Second * 2
			time.Sleep(waitTime)
		}

		return nil
	}
}

func TestAccWebhookResource_InvalidAllowedValue(t *testing.T) {
	t.Parallel()
	invalidAuthType := "INVALID"
	invalidEvents := []string{"INVALID"}
	invalidPlacementConfig := getDefaultAPIKeyConfig()
	invalidPlacementConfig["placement"] = "body"

	testResourceID := acctest.RandString(10)
	testURL := testWebhookURL(acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccWebhookResourceConfig(invalidAuthType, testURL, testResourceID, events, getDefaultAPIKeyConfig()),
				ExpectError: regexp.MustCompile(`Attribute auth_type value must be one of: \["APIKEY" "TOTP" "CIDAAS_OAUTH2"`),
			},
			{
				Config:      testAccWebhookResourceConfig(apiKey, testURL, testResourceID, invalidEvents, getDefaultAPIKeyConfig()),
				ExpectError: regexp.MustCompile(`(is not a webhook-capable event|failed to list webhook-capable events|00100)`),
			},
			{
				Config:      testAccWebhookResourceConfig(apiKey, testURL, testResourceID, events, invalidPlacementConfig),
				ExpectError: regexp.MustCompile(`placement value must be one of: \["query" "header"\]`),
			},
		},
	})
}

func TestAccWebhookResource_InvalidAuthType(t *testing.T) {
	t.Parallel()
	testResourceID := acctest.RandString(10)
	testURL := testWebhookURL(acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccWebhookResourceConfig(totp, testURL, testResourceID, events, getDefaultAPIKeyConfig()),
				ExpectError: regexp.MustCompile(`The attribute totp_config cannot be empty when the auth_type is TOTP`),
			},
			{
				Config:      testAccWebhookResourceConfig(oauth2, testURL, testResourceID, events, getDefaultAPIKeyConfig()),
				ExpectError: regexp.MustCompile(`The attribute cidaas_auth_config cannot be empty when the auth_type is`),
			},
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_webhook" "%s" {
					auth_type = "APIKEY"
					url = "%s"
					events = ["ACCOUNT_MODIFIED"]
					totp_config = {
						key = "api-key"
						placeholder = "key"
						placement = "query"
					}
				}
			`, acctest.GetBaseURL(), testResourceID, testURL),
				ExpectError: regexp.MustCompile(`The attribute apikey_config cannot be empty when the auth_type is APIKEY`),
			},
		},
	})
}

func TestAccWebhookResource_SwitchAuthType(t *testing.T) {
	t.Parallel()
	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_WEBHOOK, testResourceID)
	testURL := testWebhookURL(acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckWebhookDestroyed(testResourceName),
		Steps: []resource.TestStep{
			{
				Config: webhookResouceFullConfig(apiKey, testResourceID, testURL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "auth_type", apiKey),
				),
			},
			{
				Config: webhookResouceFullConfig(totp, testResourceID, testURL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "auth_type", totp),
				),
			},
			{
				Config: webhookResouceFullConfig(oauth2, testResourceID, testURL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "auth_type", oauth2),
				),
			},
		},
	})
}

func webhookResouceFullConfig(authType, resourceID, url string) string {
	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_webhook" "%s" {
			auth_type = "%s"
			url = "%s"
			events = ["ACCOUNT_MODIFIED"]
			apikey_config = {
				key = "api-key"
				placeholder = "key"
				placement = "query"
			}
			totp_config = {
				key = "totp-key"
				placeholder = "key"
				placement = "header"
			}
			cidaas_auth_config = {
				client_id = "ce90d6ba-9a5a-49b6-9a50"
			}
		}`, acctest.GetBaseURL(), resourceID, authType, url)
}
