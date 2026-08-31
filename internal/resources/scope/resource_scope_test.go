package scope_test

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
	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	scopeSecurityLevel  = "CONFIDENTIAL"
	requiredUserConsent = false
	title               = "scope title in German"
	locale              = "de-DE"
	scopeDescription    = "The description of the scope in German"
)

var (
	defaultScopeGroupName = []string{"developer"}
	localizedDescriptions = []map[string]string{
		{
			"title":       title,
			"locale":      locale,
			"description": scopeDescription,
		},
	}
)

// create, read and update test
func TestAccScopeResource_Basic(t *testing.T) {
	t.Parallel()

	updatedScopeDescription := "Updated description of the scope in German"
	updatedRequiredUserConsent := true
	localizedDesc := []map[string]string{
		{
			"title":  title,
			"locale": locale,
			// description is updated to validate update operation
			"description": updatedScopeDescription,
		},
	}
	scopeKey := acctest.RandString(10)
	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_SCOPE, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckScopeDestroyed(testResourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeResourceConfig(scopeSecurityLevel, scopeKey, testResourceID, requiredUserConsent, defaultScopeGroupName, localizedDesc),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "security_level", scopeSecurityLevel),
					resource.TestCheckResourceAttr(testResourceName, "scope_key", scopeKey),
					resource.TestCheckResourceAttr(testResourceName, "required_user_consent", strconv.FormatBool(requiredUserConsent)),
					resource.TestCheckResourceAttr(testResourceName, "group_name.0", "developer"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "scope_owner"),
				),
			},
			{
				ResourceName:      testResourceName,
				ImportStateId:     scopeKey,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// required_user_consent & description in localized_descriptions updated
				Config: testAccScopeResourceConfig(
					scopeSecurityLevel,
					scopeKey,
					testResourceID,
					updatedRequiredUserConsent,
					defaultScopeGroupName,
					localizedDesc,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "required_user_consent", strconv.FormatBool(updatedRequiredUserConsent)),
					resource.TestCheckResourceAttr(testResourceName, "localized_descriptions.0.description", updatedScopeDescription),
				),
			},
		},
	})
}

func testAccScopeResourceConfig(
	securityLevel, scopeKey, resourceID string,
	requiredUserConsent bool,
	groupName []string,
	localizedDescriptions []map[string]string,
) string {
	groupNameString := "[]"
	if len(groupName) > 0 {
		groupNameString = `["` + strings.Join(groupName, `", "`) + `"]`
	}

	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_scope" "%s" {
			security_level = "`+securityLevel+`"
			scope_key = "`+scopeKey+`"
			required_user_consent = "`+strconv.FormatBool(requiredUserConsent)+`"
			group_name = `+groupNameString+`
			localized_descriptions =[
				{
					title = "`+localizedDescriptions[0]["title"]+`"
					locale = "`+localizedDescriptions[0]["locale"]+`"
					description = "`+localizedDescriptions[0]["description"]+`"
				}
			]
		}
	`, acctest.GetBaseURL(), resourceID)
}

func testCheckScopeDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		scope := cidaas.Scope{
			ClientConfig: cidaas.ClientConfig{
				BaseURL:     os.Getenv("BASE_URL"),
				AccessToken: acctest.TestToken,
			},
		}

		// Add retry logic for eventual consistency
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			res, err := scope.Get(context.Background(), rs.Primary.Attributes["scope_key"])

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
				return fmt.Errorf("error checking if scope exists: %w", err)
			}

			// If this is the last retry, return error
			if i == maxRetries-1 {
				return fmt.Errorf("scope still exists after %d retries: %+v", maxRetries, res)
			}

			// Wait before retrying with exponential backoff
			waitTime := time.Duration(i+1) * time.Second * 2
			time.Sleep(waitTime)
		}

		return nil
	}
}

// failed validation on updating immutable proprty scope_key
func TestAccScopeResource_ImmutableScopeKeyUpdateFail(t *testing.T) {
	t.Parallel()

	scopeKey := acctest.RandString(10)
	updatedScopeKey := acctest.RandString(10)

	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_SCOPE, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScopeResourceConfig(scopeSecurityLevel, scopeKey, testResourceID, requiredUserConsent, defaultScopeGroupName, localizedDescriptions),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "scope_key", scopeKey),
				),
			},
			{
				Config:      testAccScopeResourceConfig(scopeSecurityLevel, updatedScopeKey, testResourceID, requiredUserConsent, defaultScopeGroupName, localizedDescriptions),
				ExpectError: regexp.MustCompile(`Attribute 'scope_key' can't be modified.`),
			},
		},
	})
}

// Invalid security_level validation
func TestAccScopeResource_InvalidSecurityLevel(t *testing.T) {
	t.Parallel()

	scopeKey := acctest.RandString(10)
	invalidSecurityLevel := "INVALID"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccScopeResourceConfig(invalidSecurityLevel, scopeKey, scopeKey, requiredUserConsent, defaultScopeGroupName, localizedDescriptions),
				ExpectError: regexp.MustCompile(`Attribute security_level value must be one of: \["PUBLIC" "CONFIDENTIAL"\]`),
			},
		},
	})
}

// missing required parameter scope_key
func TestAccScopeResource_MissingRequired(t *testing.T) {
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
				resource "cidaas_scope" "%s" {
				}
				`, acctest.GetBaseURL(), acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`The argument "scope_key" is required, but no definition was found.`),
			},
		},
	})
}

// check default required_user_consent is false
func TestAccScopeResource_DefaultRequiredConsent(t *testing.T) {
	t.Parallel()

	scopeKey := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_SCOPE, scopeKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_scope" "%s" {
					security_level = "PUBLIC"
					scope_key = "`+scopeKey+`"
					group_name = ["developer"]
					localized_descriptions =[
						{
							title = "`+title+`"
							locale = "`+locale+`"
							description = "`+scopeDescription+`"
						}
					]
				}
				`, acctest.GetBaseURL(), scopeKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "required_user_consent"),
					resource.TestCheckResourceAttr(testResourceName, "required_user_consent", strconv.FormatBool(false)),
				),
			},
		},
	})
}

// localized_descriptions[i].title is required
func TestAccScopeResource_TitleRequired(t *testing.T) {
	t.Parallel()

	scopeKey := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_scope" "%s" {
					security_level = "PUBLIC"
					scope_key = "`+scopeKey+`"
					group_name = ["developer"]
					localized_descriptions =[
						{
							locale = "`+locale+`"
							description = "`+scopeDescription+`"
						}
					]
				}
				`, acctest.GetBaseURL(), scopeKey),
				ExpectError: regexp.MustCompile(`attribute "title" is required`),
			},
		},
	})
}

// Invalid locale validation
func TestAccScopeResource_InvalidLocale(t *testing.T) {
	t.Parallel()

	scopeKey := acctest.RandString(10)
	invalidLocale := "ab"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_scope" "%s" {
					security_level = "PUBLIC"
					scope_key = "`+scopeKey+`"
					group_name = ["developer"]
					localized_descriptions =[
						{
							title = "`+title+`"
							locale = "`+invalidLocale+`"
							description = "`+scopeDescription+`"
						}
					]
				}
				`, acctest.GetBaseURL(), scopeKey),
				ExpectError: regexp.MustCompile(`locale value must be one of`), // TODO:full error string comparison
			},
		},
	})
}
