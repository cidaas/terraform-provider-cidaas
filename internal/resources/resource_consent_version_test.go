package resources_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestConsentVersion_Basic(t *testing.T) {
	// ponytail: not Parallel — shared-tenant consent-management eventual consistency
	// races with other consent acc tests (job.md: 400/30001 "consent version not found" on create).

	testResourceID := acctest.RandString(10)
	groupName := acctest.RandString(10)
	consentName := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", resources.RESOURCE_CONSENT_VERSION, testResourceID)
	consentResourceName := "cidaas_consent.sample"
	consentGroupResourceName := "cidaas_consent_group.sample"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// ponytail: apply group alone first — consent create races "group not found" if both land in one apply.
			{
				Config: testConsentVersionGroupOnlyConfig(groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(consentGroupResourceName, "id"),
					func(s *terraform.State) error {
						groupRS, ok := s.RootModule().Resources[consentGroupResourceName]
						if !ok {
							return fmt.Errorf("resource %s not found", consentGroupResourceName)
						}
						return waitUntilConsentGroupReadable(groupRS.Primary.ID)
					},
				),
			},
			// Then consent; wait until the instance API sees it. Version create retries 30001.
			{
				Config: testConsentVersionDepsConfig(groupName, consentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(consentResourceName, "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[consentResourceName]
						if !ok {
							return fmt.Errorf("resource %s not found", consentResourceName)
						}
						groupRS, ok := s.RootModule().Resources[consentGroupResourceName]
						if !ok {
							return fmt.Errorf("resource %s not found", consentGroupResourceName)
						}
						return waitUntilConsentReadyForVersionCreate(groupRS.Primary.ID, rs.Primary.ID, consentName)
					},
				),
			},
			{
				Config: testConsentVersionConfig("consent version in German", testResourceID, groupName, consentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "consent_type", "SCOPES"),
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
					return rs.Primary.Attributes["consent_id"] + ":" + rs.Primary.ID + ":de:en", nil
				},
			},
			{
				Config: testConsentVersionConfig("updated consent version in German", testResourceID, groupName, consentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "consent_type", "SCOPES"),
				),
			},
		},
	})
}

func waitUntilConsentGroupReadable(consentGroupID string) error {
	groupClient := cidaas.ConsentGroup{
		ClientConfig: cidaas.ClientConfig{
			BaseURL:     acctest.GetBaseURL(),
			AccessToken: acctest.TestToken,
		},
	}

	const maxRetries = 10
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		_, err := groupClient.Get(context.Background(), consentGroupID)
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	return fmt.Errorf("consent group %s not readable after %d retries: %v", consentGroupID, maxRetries, lastErr)
}

func waitUntilConsentReadyForVersionCreate(consentGroupID, consentID, consentName string) error {
	consentClient := cidaas.Consent{
		ClientConfig: cidaas.ClientConfig{
			BaseURL:     acctest.GetBaseURL(),
			AccessToken: acctest.TestToken,
		},
	}

	const maxRetries = 10
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		res, err := consentClient.GetConsentInstances(context.Background(), consentGroupID)
		if err != nil {
			lastErr = err
		} else if res != nil {
			for _, c := range res.Data {
				if c.ID == consentID || c.ConsentName == consentName {
					return nil
				}
			}
			lastErr = fmt.Errorf("consent %s (%s) not in group %s yet (%d instances)", consentID, consentName, consentGroupID, len(res.Data))
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	return fmt.Errorf("consent not ready for version create after %d retries: %v", maxRetries, lastErr)
}

func testConsentVersionGroupOnlyConfig(groupName string) string {
	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_consent_group" "sample" {
			group_name  = "%s"
			description = "sample description"
		}
	`, acctest.GetBaseURL(), groupName)
}

func testConsentVersionDepsConfig(groupName, consentName string) string {
	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_consent_group" "sample" {
			group_name  = "%s"
			description = "sample description"
		}
		resource "cidaas_consent" "sample" {
			consent_group_id = cidaas_consent_group.sample.id
			name             = "%s"
			enabled          = true
		}
	`, acctest.GetBaseURL(), groupName, consentName)
}

func testConsentVersionConfig(content, resourceID, groupName, consentName string) string {
	return fmt.Sprintf(`
		provider "cidaas" {
			base_url = "%s"
		}
		resource "cidaas_consent_group" "sample" {
			group_name  = "%s"
			description = "sample description"
		}
		resource "cidaas_consent" "sample" {
			consent_group_id = cidaas_consent_group.sample.id
			name             = "%s"
			enabled          = true
		}
		resource "cidaas_consent_version" "%s" {
			version         = 1
			consent_id      = cidaas_consent.sample.id
			consent_type    = "SCOPES"
			scopes          = ["openid", "profile"]
			required_fields = ["name"]
			consent_locales = [
				{
					content = "%s"
					locale  = "de"
				},
				{
					content = "consent version in English"
					locale  = "en"
				}
			]
		}
	`, acctest.GetBaseURL(), groupName, consentName, resourceID, content)
}
