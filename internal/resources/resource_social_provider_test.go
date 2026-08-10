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

var (
	spName         = acctest.RandString(10)
	spProviderName = "google"
	spClientID     = acctest.RandString(10)
	spClientSecret = acctest.RandString(10)
)

// create, read and update test
func TestSocialProvider_Basic(t *testing.T) {
	t.Parallel()
	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", resources.RESOURCE_SOCIAL_PROVIDER, testResourceID)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checksocialProviderDestroyed(testResourceName),
		Steps: []resource.TestStep{
			{
				Config: socialProviderConfig(spName, spProviderName, spClientID, spClientSecret, testResourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "name", spName),
					resource.TestCheckResourceAttr(testResourceName, "provider_name", spProviderName),
					resource.TestCheckResourceAttr(testResourceName, "client_id", spClientID),
					resource.TestCheckResourceAttr(testResourceName, "client_secret", spClientSecret),
					resource.TestCheckResourceAttr(testResourceName, "claims.required_claims.user_info.0", "name"),

					// default value check
					resource.TestCheckResourceAttr(testResourceName, "enabled", strconv.FormatBool(false)),
					resource.TestCheckResourceAttr(testResourceName, "enabled_for_admin_portal", strconv.FormatBool(false)),

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
					return rs.Primary.Attributes["provider_name"] + ":" + rs.Primary.ID, nil
				},
				// API may return userinfo_fields in a different order than config.
				ImportStateVerifyIgnore: []string{"userinfo_fields"},
			},
			{
				Config: socialProviderConfig(spName, spProviderName, spClientID, spClientSecret, testResourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
			// immutable name and provider_name validation
			{
				Config:      socialProviderConfig(spName, "facebook", spClientID, spClientSecret, testResourceID),
				ExpectError: regexp.MustCompile(`Attribute 'provider_name' can't be modified`),
			},
			{
				Config:      socialProviderConfig(acctest.RandString(5), spProviderName, spClientID, spClientSecret, testResourceID),
				ExpectError: regexp.MustCompile(`Attribute 'name' can't be modified`),
			},
		},
	})
}

func socialProviderConfig(name, providerName, clientID, clientSecret, resourceID string) string {
	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_social_provider" "%s" {
			name                     = "%s"
			provider_name            = "%s"
			client_id                = "%s"
			client_secret            = "%s"
			scopes                   = ["profile", "email"]
			claims = {
				required_claims = {
					user_info = ["name"]
					id_token  = ["phone_number"]
				}
				optional_claims = {
					user_info = ["website"]
					id_token  = ["street_address"]
				}
			}
			userinfo_fields = [
				{
					inner_key       = "sample_custom_field"
					external_key    = "external_sample_cf"
					is_custom_field = true
					is_system_field = false
				},
				{
					inner_key       = "sample_system_field"
					external_key    = "external_sample_sf"
					is_custom_field = false
					is_system_field = true
				}
			]
		}
	`, acctest.GetBaseURL(), resourceID, name, providerName, clientID, clientSecret)
}

func checksocialProviderDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		sp := cidaas.SocialProvider{
			ClientConfig: cidaas.ClientConfig{
				BaseURL:     acctest.GetBaseURL(),
				AccessToken: acctest.TestToken,
			},
		}

		// Add retry logic for eventual consistency
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			res, err := sp.Get(context.Background(), rs.Primary.Attributes["provider_name"], rs.Primary.Attributes["id"])

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
				return fmt.Errorf("error checking if social provider exists: %w", err)
			}

			// If this is the last retry, return error
			if i == maxRetries-1 {
				return fmt.Errorf("social provider still exists after %d retries: %+v", maxRetries, res)
			}

			// Wait before retrying with exponential backoff
			waitTime := time.Duration(i+1) * time.Second * 2
			time.Sleep(waitTime)
		}

		return nil
	}
}

// Invalid provider_name validation
func TestSocialProvider_InvalidProviderName(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      socialProviderConfig(spName, acctest.RandString(10), spClientID, spClientSecret, acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`Attribute provider_name value must be one of:`),
			},
		},
	})
}

// missing required parameter
func TestSocialProvider_MissingRequired(t *testing.T) {
	t.Parallel()

	// client_secret is Optional (ExactlyOneOf with client_secret_wo); not schema-Required.
	requiredParams := []string{
		"name", "provider_name", "client_id",
	}

	for _, param := range requiredParams {
		param := param // Capture loop variable
		t.Run(fmt.Sprintf("missing_%s", param), func(t *testing.T) {
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
                            resource "cidaas_social_provider" "%s" {}
                        `, acctest.GetBaseURL(), testResourceID),
						ExpectError: regexp.MustCompile(fmt.Sprintf(`The argument "%s" is required`, param)),
					},
				},
			})
		})
	}
}
