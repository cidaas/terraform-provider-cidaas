package security_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPwdPolicyResource_Basic(t *testing.T) {
	t.Parallel()

	policyName := acctest.RandString(10)
	testResourceName := fmt.Sprintf("%s.%s", base.RESOURCE_PASSWORD_POLICY, policyName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPwdPolicyResourceConfig(policyName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "policy_name", policyName),
					resource.TestCheckResourceAttr(testResourceName, "password_policy.block_compromised", strconv.FormatBool(true)),
					resource.TestCheckResourceAttr(testResourceName, "password_policy.deny_usage_count", strconv.Itoa(0)),
					resource.TestCheckResourceAttr(testResourceName, "password_policy.change_enforcement.expiration_in_days", strconv.Itoa(0)),
					resource.TestCheckResourceAttr(testResourceName, "password_policy.change_enforcement.notify_user_before_in_days", strconv.Itoa(0)),
				),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// ImportStateId:     "id",
			},
			{
				Config: testAccPwdPolicyResourceConfig(policyName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "password_policy.block_compromised", strconv.FormatBool(false)),
				),
			},
		},
	})
}

func testAccPwdPolicyResourceConfig(policyName string, blockCompromised bool) string {
	return fmt.Sprintf(`
	provider "cidaas" {
		base_url = "%s"
	}
	resource "cidaas_password_policy" "%s" {
		policy_name 	  = "%s"
		password_policy = {
			block_compromised = "%+v"
			strength_regexes = [
				"^(?=.*[A-Za-z])(?!.*\\s).{6,15}$"
			],
		}
	}
	`, acctest.GetBaseURL(), policyName, policyName, blockCompromised)
}

func TestAccPwdPolicyResource_MissingRequired(t *testing.T) {
	t.Parallel()

	requiredParams := []string{
		"policy_name",
	}
	// TODO: to add strength_regexes in the list
	for _, param := range requiredParams {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acctest.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
						provider "cidaas" {
							base_url = "%s"
						}
						resource "cidaas_password_policy" "%s" {}
					`, acctest.GetBaseURL(), acctest.RandString(10)),
					ExpectError: regexp.MustCompile(fmt.Sprintf(`The argument "%s" is required, but no definition was found.`, param)),
				},
			},
		})
	}
}
