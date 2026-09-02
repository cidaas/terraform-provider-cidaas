package consent_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccConsentResource_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	acctest.SkipIfV4(t)

	groupName := acctest.RandString(10)
	name := acctest.RandString(10)

	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_CONSENT, groupName)
	consentGroupResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_CONSENT_GROUP, "example")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckConsentDestroyed(consentGroupResourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccConsentResourceConfig(groupName, name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(testResourceName, "consent_group_id", consentGroupResourceName, "id"),
					resource.TestCheckResourceAttr(testResourceName, "name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "enabled"),
					resource.TestCheckResourceAttrSet(testResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testResourceName, "updated_at"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						testResourceName,
						tfjsonpath.New("enabled"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName: testResourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[consentGroupResourceName]
					if !ok {
						return "", fmt.Errorf("Not found: %s", consentGroupResourceName)
					}
					return rs.Primary.ID + ":" + name, nil
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at", "created_at"},
			},
		},
	})
}

func testAccConsentResourceConfig(groupName, name string, enabled bool) string {
	return fmt.Sprintf(`
	provider "cidaas" {
		base_url = "%s"
	}
	resource "cidaas_consent_group" "example" {
		group_name  = "%s"
	}
	resource "cidaas_consent" "%s" {
		consent_group_id  = cidaas_consent_group.example.id
		name = "%s"
		enabled = %t
	}
	`, acctest.GetBaseURL(), groupName, groupName, name, enabled)
}

func testCheckConsentDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		consent := cidaas.Consent{
			ClientConfig: cidaas.ClientConfig{
				BaseURL:     acctest.GetBaseURL(),
				AccessToken: acctest.TestToken,
			},
		}

		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			res, err := consent.GetConsentInstances(context.Background(), rs.Primary.ID)
			if res == nil || res.Status == http.StatusNoContent || len(res.Data) == 0 {
				return nil
			}
			if err != nil {
				if strings.Contains(err.Error(), "not found") ||
					strings.Contains(err.Error(), "404") ||
					strings.Contains(err.Error(), "204") {
					return nil
				}
				return fmt.Errorf("error checking if consent exists: %w", err)
			}
			if i == maxRetries-1 {
				return fmt.Errorf("consent still exists after %d retries: %+v", maxRetries, res)
			}
			waitTime := time.Duration(i+1) * time.Second * 2
			time.Sleep(waitTime)
		}
		return nil
	}
}

func TestAccConsentResource_GroupNameUpdateFail(t *testing.T) {
	t.Parallel()
	acctest.SkipIfV4(t)
	groupName := acctest.RandString(10)
	name := acctest.RandString(10)
	updatedName := acctest.RandString(10)

	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_CONSENT, groupName)
	consentGroupResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_CONSENT_GROUP, "example")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConsentResourceConfig(groupName, name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(testResourceName, "consent_group_id", consentGroupResourceName, "id"),
				),
			},
			{
				Config:      testAccConsentResourceConfig(groupName, updatedName, true),
				ExpectError: regexp.MustCompile(`Attribute 'name' can't be modified.`),
			},
		},
	})
}

func TestAccConsentResource_EmptyGroupName(t *testing.T) {
	t.Parallel()
	acctest.SkipIfV4(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_consent" "%s" {
					consent_group_id  = ""
					name = ""
				}
				`, acctest.GetBaseURL(), acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`Attribute consent_group_id string length must be at least 1, got: 0`),
			},
		},
	})
}

func TestAccConsentResource_MissingRequired(t *testing.T) {
	t.Parallel()
	acctest.SkipIfV4(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
				}
				resource "cidaas_consent" "%s" {
				}
				`, acctest.GetBaseURL(), acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`The argument "name" is required, but no definition was found.`),
			},
		},
	})
}

func TestAccConsentResource_V4ComingSoon(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				provider "cidaas" {
					base_url = "%s"
					cidaas_version = "4.x"
				}
				resource "cidaas_consent" "test" {
					consent_group_id = "test-group"
					name             = "test-name"
				}
				`, acctest.GetBaseURL()),
				ExpectError: regexp.MustCompile(`Consent Resource Coming Soon on v4`),
			},
		},
	})
}
