package notification_test

import (
	"context"
	"fmt"
	"net/http"
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

// create, read and update test
// func TestTemplate_Basic(t *testing.T) {
// 	updatedTemplateContent := acctest.RandString(256)
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
// 		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
// 		CheckDestroy:             checkTemplateDestroyed,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testTemplateConfig(templateLocale, templateKey, templateType, templateContent),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr(resourceTemplate, "locale", templateLocale),
// 					resource.TestCheckResourceAttr(resourceTemplate, "template_key", templateKey),
// 					resource.TestCheckResourceAttr(resourceTemplate, "template_type", templateType),
// 					resource.TestCheckResourceAttr(resourceTemplate, "content", templateContent),

// 					resource.TestCheckResourceAttrSet(resourceTemplate, "id"),
// 					resource.TestCheckResourceAttrSet(resourceTemplate, "template_owner"),
// 					resource.TestCheckResourceAttrSet(resourceTemplate, "group_id"),
// 					resource.TestCheckResourceAttrSet(resourceTemplate, "is_system_template"),
// 				),
// 			},
// 			{
// 				ResourceName:      resourceTemplate,
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 				ImportStateId:     templateKey + ":" + templateType + ":" + templateLocale,
// 			},
// 			{
// 				Config: testTemplateConfig(templateLocale, templateKey, templateType, updatedTemplateContent),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr(resourceTemplate, "content", updatedTemplateContent),
// 					// check default value
// 					resource.TestCheckResourceAttr(resourceTemplate, "is_system_template", strconv.FormatBool(false)),
// 				),
// 			},
// 			// locale, template_key and template type can't be modified
// 			{
// 				Config:      testTemplateConfig("en-US", strings.ToUpper(acctest.RandString(10)), "IVR", updatedTemplateContent),
// 				ExpectError: regexp.MustCompile("can't be modified"),
// 			},
// 		},
// 	})
// }

func testTemplateConfig(locale, templateKey, templateType, content string) string {
	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_template" "%s" {
			locale        = "%s"
			template_key  = "%s"
			template_type = "%s"
			content       = "%s"
		}
		`, acctest.GetBaseURL(), templateKey, locale, templateKey, templateType, content)
}

func checkTemplateDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		template := cidaas.Template{
			ClientConfig: cidaas.ClientConfig{
				BaseURL:     os.Getenv("BASE_URL"),
				AccessToken: acctest.TestToken,
			},
		}

		templatePayload := cidaas.TemplateModel{
			Locale:       rs.Primary.Attributes["locale"],
			TemplateKey:  rs.Primary.Attributes["template_key"],
			TemplateType: rs.Primary.Attributes["template_type"],
		}

		// Add retry logic for eventual consistency
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			res, err := template.Get(context.Background(), templatePayload, false)

			// Check if resource is successfully deleted (nil or NoContent status)
			if res == nil || res.Status == http.StatusNoContent {
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
				return fmt.Errorf("error checking if template exists: %w", err)
			}

			// If this is the last retry, return error
			if i == maxRetries-1 {
				return fmt.Errorf("template still exists after %d retries: %+v", maxRetries, res)
			}

			// Wait before retrying with exponential backoff
			waitTime := time.Duration(i+1) * time.Second * 2
			time.Sleep(waitTime)
		}

		return nil
	}
}

// subject can not be empty when template type is SMS
func TestTemplate_EmailSubjectCheck(t *testing.T) {
	t.Parallel()

	templateLocale := "de-DE"
	templateContent := acctest.RandString(256)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testTemplateConfig(templateLocale, acctest.RandString(10), "EMAIL", templateContent),
				ExpectError: regexp.MustCompile("subject can not be empty when template_type is EMAIL"),
			},
		},
	})
}

// template_key must be a valid string consisting only of uppercase letters,
// digits (0-9), underscores (_), and hyphens (-)
func TestTemplate_TemplateKeyValidation(t *testing.T) {
	t.Parallel()

	templateLocale := "de-DE"
	templateType := "SMS"
	templateContent := acctest.RandString(256)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testTemplateConfig(templateLocale, acctest.RandString(10), templateType, templateContent),
				ExpectError: regexp.MustCompile("template_key must be a valid string consisting"), // TODO: full string validation
			},
		},
	})
}

// template_type must be one of "EMAIL", "SMS", "IVR" and "PUSH"
func TestTemplate_TemplateTypeValidation(t *testing.T) {
	t.Parallel()

	templateLocale := "de-DE"
	templateContent := acctest.RandString(256)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testTemplateConfig(templateLocale, acctest.RandString(10), "INVALID", templateContent),
				ExpectError: regexp.MustCompile("template_key must be a valid string consisting"), // TODO: full string validation
			},
		},
	})
}

// required params locale, template_key, teamplte_type and content
func TestTemplate_MissingRequired(t *testing.T) {
	t.Parallel()

	requiredParams := []string{"locale", "template_key", "template_type", "content"}

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
                            resource "cidaas_template" "%s" {}
                        `, acctest.GetBaseURL(), testResourceID),
						ExpectError: regexp.MustCompile(fmt.Sprintf(`"%s" is required`, param)),
					},
				},
			})
		})
	}
}

// Custom (non-system) template: create and update via templates-srv /template/custom.
// Opt-in only: shared CI tenants often return HTTP 500 (code 35001) on custom template POST as well.
//
//	RUN_TEMPLATE_CUSTOM_ACC_TEST=1 TF_ACC=1 BASE_URL=... TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID=... TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET=... \
//	  go test ./internal/resources/ -run TestTemplate_CustomTemplateBasic -v
func TestTemplate_CustomTemplateBasic(t *testing.T) {
	if os.Getenv("RUN_TEMPLATE_CUSTOM_ACC_TEST") != "1" {
		t.Skip("set RUN_TEMPLATE_CUSTOM_ACC_TEST=1 to run; templates-srv custom template create often returns 500 on shared tenants (code 35001)")
	}
	t.Parallel()

	templateKey := strings.ToUpper(acctest.RandString(12))
	templateLocale := "en-US"
	initialContent := acctest.RandString(128)
	updatedContent := acctest.RandString(128)
	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_TEMPLATE, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if os.Getenv("TF_ACC") == "" {
				t.Skip("set TF_ACC=1 for acceptance tests (along with BASE_URL and TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID / TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET)")
			}
			if os.Getenv("BASE_URL") == "" || os.Getenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID") == "" || os.Getenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET") == "" {
				t.Skip("set BASE_URL, TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID, and TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET for acceptance tests")
			}
			acctest.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkTemplateDestroyed(testResourceName),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_template" "%s" {
					locale        = "%s"
					template_key  = "%s"
					template_type = "SMS"
					content       = "%s"
				}
				`, acctest.GetBaseURL(), testResourceID, templateLocale, templateKey, initialContent),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "locale", templateLocale),
					resource.TestCheckResourceAttr(testResourceName, "template_key", templateKey),
					resource.TestCheckResourceAttr(testResourceName, "content", initialContent),
					resource.TestCheckResourceAttr(testResourceName, "is_system_template", "false"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_template" "%s" {
					locale        = "%s"
					template_key  = "%s"
					template_type = "SMS"
					content       = "%s"
				}
				`, acctest.GetBaseURL(), testResourceID, templateLocale, templateKey, updatedContent),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "content", updatedContent),
				),
			},
		},
	})
}

// System Template basic create, update and delete, system template can not be imported.
// Opt-in only: templates-srv frequently returns HTTP 500 (code 35001) for system template
// POST in CI / shared tenants (VERIFY_USER constraints, master list, tenant policy).
// Run manually (all required for a real apply):
//
//	RUN_TEMPLATE_SYSTEM_ACC_TEST=1 TF_ACC=1 BASE_URL=... TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID=... TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET=... \
//	  go test ./internal/resources/ -run TestTemplate_SystemTemplateBasic -v
func TestTemplate_SystemTemplateBasic(t *testing.T) {
	if os.Getenv("RUN_TEMPLATE_SYSTEM_ACC_TEST") != "1" {
		t.Skip("set RUN_TEMPLATE_SYSTEM_ACC_TEST=1 to run; system template create is flaky on shared tenants (templates-srv 35001)")
	}
	t.Parallel()

	testResourceID := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_TEMPLATE, testResourceID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			// Skip instead of Fatal when env is incomplete (TestAccPreCheck uses t.Fatal).
			if os.Getenv("TF_ACC") == "" {
				t.Skip("set TF_ACC=1 for acceptance tests (along with BASE_URL and TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID / TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET)")
			}
			if os.Getenv("BASE_URL") == "" || os.Getenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID") == "" || os.Getenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET") == "" {
				t.Skip("set BASE_URL, TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID, and TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET for acceptance tests")
			}
			acctest.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkTemplateDestroyed(testResourceName),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_template" "%s" {
					locale             = "en-US"
					template_key       = "VERIFY_USER"
					template_type      = "SMS"
					content            = "Hi {{name}}, here is the {{code}} to verify the user"
					is_system_template = true
					group_id           = "default"
					processing_type    = "GENERAL"
					verification_type  = "SMS"
					usage_type         = "VERIFICATION_CONFIGURATION"
				}
				`, acctest.GetBaseURL(), testResourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "is_system_template", strconv.FormatBool(true)),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_template" "%s" {
					locale             = "en-US"
					template_key       = "VERIFY_USER"
					template_type      = "SMS"
					content            = "Hi {{name}}, here is the {{code}} to verify the user updated"
					is_system_template = true
					group_id           = "default"
					processing_type    = "GENERAL"
					verification_type  = "SMS"
					usage_type         = "VERIFICATION_CONFIGURATION"
				}
				`, acctest.GetBaseURL(), testResourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "content", "Hi {{name}}, here is the {{code}} to verify the user updated"),
				),
			},
			// templated reverted back to the old state
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_template" "%s" {
					locale             = "en-US"
					template_key       = "VERIFY_USER"
					template_type      = "SMS"
					content            = "Hi {{name}}, here is the {{code}} to verify the user"
					is_system_template = true
					group_id           = "default"
					processing_type    = "GENERAL"
					verification_type  = "SMS"
					usage_type         = "VERIFICATION_CONFIGURATION"
				}
				`, acctest.GetBaseURL(), testResourceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "content", "Hi {{name}}, here is the {{code}} to verify the user"),
				),
			},
		},
	})
}
