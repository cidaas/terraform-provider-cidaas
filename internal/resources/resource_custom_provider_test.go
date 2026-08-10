package resources_test

import (
	"context"
	"fmt"
	"os"
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

const (
	oauth2StandardType    = "OAUTH2"
	displayName           = "Sample Terraform"
	authorizationEndpoint = "https://cidaas.de/authz-srv/authz"
	tokenEndpoint         = "https://cidaas.de/token-srv/token"
	logoURL               = "https://cidaas.de/logo"
	userinfoEndpoint      = "https://cidaas.de/users-srv/userinfo"
	scopeDisplayLabel     = "terraform sample scope display name"
)

func TestAccCustomProviderResource_Basic(t *testing.T) {
	t.Parallel()

	updatedDisplayName := "Updated Sample Terraform"
	updatedOauth2StandardType := "OPENID_CONNECT"
	updatedAuthorizationEndpoint := "https://cidaas.de/authz-srv/v2/authz"
	updatedTokenEndpoint := "https://cidaas.de/token-srv/v2/token"
	updatedLogoURL := "https://cidaas.de/v2/logo"
	updatedUserinfoEndpoint := "https://cidaas.de/users-srv/v2/userinfo"
	updatedScopeDisplayLabel := "updated terraform sample scope display name"
	updatedClientID := acctest.RandString(10)
	updatedClientSecret := acctest.RandString(10)

	providerName := acctest.RandString(10)
	clientID := acctest.RandString(10)
	clientSecret := acctest.RandString(10)

	testResourceName := fmt.Sprintf("%s.%s", resources.RESOURCE_CUSTOM_PROVIDER, clientID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkCustomProviderDestroyed(testResourceName),
		Steps: []resource.TestStep{
			{
				Config: resourceCustomProviderConfig(oauth2StandardType, providerName, clientID, clientSecret),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "standard_type", oauth2StandardType),
					resource.TestCheckResourceAttr(testResourceName, "authorization_endpoint", authorizationEndpoint),
					resource.TestCheckResourceAttr(testResourceName, "token_endpoint", tokenEndpoint),
					resource.TestCheckResourceAttr(testResourceName, "provider_name", providerName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", displayName),
					resource.TestCheckResourceAttr(testResourceName, "logo_url", logoURL),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_endpoint", userinfoEndpoint),
					resource.TestCheckResourceAttr(testResourceName, "scope_display_label", scopeDisplayLabel),
					resource.TestCheckResourceAttr(testResourceName, "client_id", clientID),
					resource.TestCheckResourceAttr(testResourceName, "client_secret", clientSecret),
					resource.TestCheckResourceAttr(testResourceName, "domains.0", "cidaas.de"),
					resource.TestCheckResourceAttr(testResourceName, "domains.1", "cidaas.org"),
					resource.TestCheckResourceAttr(testResourceName, "scopes.0.recommended", strconv.FormatBool(true)),
					resource.TestCheckResourceAttr(testResourceName, "scopes.0.required", strconv.FormatBool(true)),
					resource.TestCheckResourceAttr(testResourceName, "scopes.0.scope_name", "email"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
			{
				ResourceName:            testResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           providerName,
				// API fills userinfo_source=USERINFOENDPOINT when unset; create state omits it.
				ImportStateVerifyIgnore: []string{"userinfo_source"},
			},
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_custom_provider" "%s" {
					standard_type          = "`+updatedOauth2StandardType+`"
					authorization_endpoint = "`+updatedAuthorizationEndpoint+`"
					token_endpoint         = "`+updatedTokenEndpoint+`"
					provider_name          = "`+providerName+`"
					display_name           = "`+updatedDisplayName+`"
					logo_url               =  "`+updatedLogoURL+`"
					userinfo_endpoint      =  "`+updatedUserinfoEndpoint+`"
					scope_display_label    =  "`+updatedScopeDisplayLabel+`"
					client_id              =  "`+updatedClientID+`"
					client_secret          =  "`+updatedClientSecret+`"
					domains                = ["cidaas.in", "cidaas.com"]

					scopes = [
						{
							scope_name  = "openid"
						}
					]
				}
			`, acctest.GetBaseURL(), clientID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "standard_type", updatedOauth2StandardType),
					resource.TestCheckResourceAttr(testResourceName, "authorization_endpoint", updatedAuthorizationEndpoint),
					resource.TestCheckResourceAttr(testResourceName, "token_endpoint", updatedTokenEndpoint),
					resource.TestCheckResourceAttr(testResourceName, "provider_name", providerName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedDisplayName),
					resource.TestCheckResourceAttr(testResourceName, "logo_url", updatedLogoURL),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_endpoint", updatedUserinfoEndpoint),
					resource.TestCheckResourceAttr(testResourceName, "scope_display_label", updatedScopeDisplayLabel),
					resource.TestCheckResourceAttr(testResourceName, "client_id", updatedClientID),
					resource.TestCheckResourceAttr(testResourceName, "client_secret", updatedClientSecret),
					resource.TestCheckResourceAttr(testResourceName, "domains.0", "cidaas.com"),
					resource.TestCheckResourceAttr(testResourceName, "domains.1", "cidaas.in"),
					resource.TestCheckResourceAttr(testResourceName, "scopes.0.scope_name", "openid"),
					// default value check scopes[i].recommended & scopes[i].required
					resource.TestCheckResourceAttr(testResourceName, "scopes.0.recommended", strconv.FormatBool(false)),
					resource.TestCheckResourceAttr(testResourceName, "scopes.0.required", strconv.FormatBool(false)),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
			{
				// provider_name cannot be updated
				Config:      resourceCustomProviderConfig(oauth2StandardType, "new_provider_name", clientID, clientSecret),
				ExpectError: regexp.MustCompile(`Attribute 'provider_name' can't be modified`),
			},
		},
	})
}

func resourceCustomProviderConfig(standardType, providerName, clientID, clientSecret string) string {
	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_custom_provider" "%s" {
			standard_type          = "%s"
			authorization_endpoint = "`+authorizationEndpoint+`"
			token_endpoint         =  "`+tokenEndpoint+`"
			provider_name          = "%s"
			display_name           = "`+displayName+`"
			logo_url               =  "`+logoURL+`"
			userinfo_endpoint      =  "`+userinfoEndpoint+`"
			scope_display_label    =  "`+scopeDisplayLabel+`"
			client_id              =  "`+clientID+`"
			client_secret          =  "`+clientSecret+`"
			domains                = ["cidaas.de", "cidaas.org"]
		
			scopes = [
				{
					recommended = true
					required    = true
					scope_name  = "email"
				}
			]
		}
	`, acctest.GetBaseURL(), clientID, standardType, providerName)
}

func checkCustomProviderDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		cp := cidaas.CustomProvider{
			ClientConfig: cidaas.ClientConfig{
				BaseURL:     os.Getenv("BASE_URL"),
				AccessToken: acctest.TestToken,
			},
		}

		// Add retry logic for eventual consistency
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			res, err := cp.GetCustomProvider(context.Background(), rs.Primary.Attributes["provider_name"])

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
				return fmt.Errorf("error checking if custom provider exists: %w", err)
			}

			// If this is the last retry, return error
			if i == maxRetries-1 {
				return fmt.Errorf("custom provider still exists after %d retries: %+v", maxRetries, res)
			}

			// Wait before retrying with exponential backoff
			waitTime := time.Duration(i+1) * time.Second * 2
			time.Sleep(waitTime)
		}

		return nil
	}
}

// Invalid standard_type validation
func TestAccCustomProviderResource_InvalidStandardType(t *testing.T) {
	t.Parallel()

	providerName := acctest.RandString(10)
	invalidStandardType := "OAUTH1"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      resourceCustomProviderConfig(invalidStandardType, providerName, acctest.RandString(10), acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`Attribute standard_type value must be one of: \["OPENID_CONNECT" "OAUTH2"\]`),
			},
		},
	})
}

// missing required parameter
func TestAccCustomProviderResource_MissingRequired(t *testing.T) {
	t.Parallel()

	// client_secret is Optional (ExactlyOneOf with client_secret_wo); not schema-Required.
	requiredParams := []string{
		"provider_name", "display_name", "client_id",
		"authorization_endpoint", "token_endpoint", "userinfo_endpoint",
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
                            resource "cidaas_custom_provider" "%s" {}
                        `, acctest.GetBaseURL(), testResourceID),
						ExpectError: regexp.MustCompile(fmt.Sprintf(`The argument "%s" is required`, param)),
					},
				},
			})
		})
	}
}

// check userinfo_fields parameters
func TestAccCustomProviderResource_UserinfoFieldsCheck(t *testing.T) {
	t.Parallel()

	providerName := acctest.RandString(10)
	clientID := acctest.RandString(10)
	clientSecret := acctest.RandString(10)

	testResourceName := fmt.Sprintf("%s.%s", resources.RESOURCE_CUSTOM_PROVIDER, clientID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_custom_provider" "%s" {
					standard_type          = "`+oauth2StandardType+`"
					authorization_endpoint = "`+authorizationEndpoint+`"
					token_endpoint         = "`+tokenEndpoint+`"
					provider_name          = "`+providerName+`"
					display_name           = "`+displayName+`"
					userinfo_endpoint      = "`+userinfoEndpoint+`"
					scope_display_label    = "`+scopeDisplayLabel+`"
					client_id              = "`+clientID+`"
					client_secret          = "`+clientSecret+`"
					scopes = [
						{
							scope_name  = "email"
						}
					]
					userinfo_fields = {
						family_name        = { "ext_field_key" = "cp_family_name" }
						address            = { "ext_field_key" = "cp_address" }
						birthdate          = { "ext_field_key" = "01-01-2000" }
						email              = { "ext_field_key" = "cp@cidaas.de" }
						email_verified     = { "ext_field_key" = "email_verified", "default" = false }
						gender             = { "ext_field_key" = "male" }
						nickname           = { "ext_field_key" = "nickname" }
						given_name         = { "ext_field_key" = "cp_given_name" }
						locale             = { "ext_field_key" = "cp_locale" }
						middle_name        = { "ext_field_key" = "cp_middle_name" }
						mobile_number      = { "ext_field_key" = "100000000" }
						phone_number       = { "ext_field_key" = "10000000" }
						picture            = { "ext_field_key" = "https://cidaas.de/image.jpg" }
						preferred_username = { "ext_field_key" = "cp_preferred_username" }
						profile            = { "ext_field_key" = "cp_profile" }
						updated_at         = { "ext_field_key" = "01-01-01" }
						website            = { "ext_field_key" = "https://cidaas.de" }
						zoneinfo           = { "ext_field_key" = "cp_zone_info" }
						custom_fields = {
							zipcode         = "123456"
							alternate_phone = "1234567890"
						}
					}
				}`, acctest.GetBaseURL(), clientID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// default value check
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.family_name.ext_field_key", "cp_family_name"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.address.ext_field_key", "cp_address"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.birthdate.ext_field_key", "01-01-2000"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.email.ext_field_key", "cp@cidaas.de"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.email_verified.ext_field_key", "email_verified"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.gender.ext_field_key", "male"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.given_name.ext_field_key", "cp_given_name"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.locale.ext_field_key", "cp_locale"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.middle_name.ext_field_key", "cp_middle_name"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.mobile_number.ext_field_key", "100000000"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.phone_number.ext_field_key", "10000000"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.picture.ext_field_key", "https://cidaas.de/image.jpg"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.preferred_username.ext_field_key", "cp_preferred_username"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.profile.ext_field_key", "cp_profile"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.updated_at.ext_field_key", "01-01-01"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.website.ext_field_key", "https://cidaas.de"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.zoneinfo.ext_field_key", "cp_zone_info"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.custom_fields.zipcode", "123456"),
					resource.TestCheckResourceAttr(testResourceName, "userinfo_fields.custom_fields.alternate_phone", "1234567890"),
				),
			},
		},
	})
}
