package usersetup_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/acctest"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUserSetup_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	t.Parallel()

	name := "tf-user-setup-" + acctest.RandString(8)
	resourceName := "cidaas_user_setup.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testCheckUserSetupDestroyed(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccUserSetupConfig(name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "acc test user setup"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "user_setup.allowed_fields.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "user_setup.required_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "user_setup.enable_deduplication", "true"),
					resource.TestCheckResourceAttr(resourceName, "user_setup.validate_email", "true"),
					resource.TestCheckResourceAttr(resourceName, "user_setup.auto_activate_user", "false"),
					resource.TestCheckResourceAttr(resourceName, "user_setup.communication_medium_verification", "email_verification_required"),
					resource.TestCheckResourceAttr(resourceName, "user_setup.auto_confirm_communication_method.0", "email"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserSetupConfig(name, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user_setup.enable_deduplication", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", "acc test user setup updated"),
				),
			},
		},
	})
}

func TestAccUserSetup_RequiredFieldsSubset(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
%s
resource "cidaas_user_setup" "bad" {
  name = "tf-user-setup-bad-%s"
  user_setup = {
    allowed_fields  = ["email"]
    required_fields = ["email", "mobile_number"]
  }
}
`, acctest.ProviderConfig(), acctest.RandString(6)),
				ExpectError: regexp.MustCompile(`required_fields must be a subset of allowed_fields`),
			},
		},
	})
}

func TestAccUserSetup_InvalidEnum(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
%s
resource "cidaas_user_setup" "bad" {
  name = "tf-user-setup-enum-%s"
  user_setup = {
    allowed_fields                    = ["email"]
    required_fields                   = ["email"]
    communication_medium_verification = "not_a_real_value"
  }
}
`, acctest.ProviderConfig(), acctest.RandString(6)),
				ExpectError: regexp.MustCompile(`communication_medium_verification`),
			},
		},
	})
}

func testAccUserSetupConfig(name string, enableDedup bool) string {
	desc := "acc test user setup"
	if !enableDedup {
		desc = "acc test user setup updated"
	}
	return fmt.Sprintf(`
%s
resource "cidaas_user_setup" "test" {
  name        = %q
  description = %q
  user_setup = {
    allowed_fields                    = ["email", "given_name", "family_name"]
    required_fields                   = ["email", "given_name"]
    allow_login_with                  = ["EMAIL"]
    enable_deduplication              = %t
    validate_phone_number             = true
    validate_email                    = true
    auto_activate_user                = false
    allow_disposable_email            = false
    accept_roles_in_the_registration  = false
    send_welcome_notification         = true
    birthdate_as_date                 = true
    communication_medium_verification = "email_verification_required"
    auto_confirm_communication_method = ["email"]
    verification_for_medium           = []
  }
}
`, acctest.ProviderConfig(), name, desc, enableDedup)
}

func testCheckUserSetupDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		c, err := accClient(context.Background())
		if err != nil {
			return err
		}
		_, err = c.UserSetup.Get(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("user setup %s still exists", rs.Primary.ID)
		}
		if !errors.Is(err, client.ErrNotFound) {
			return fmt.Errorf("expected not found for %s, got: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func accClient(ctx context.Context) (*client.Client, error) {
	return client.NewClient(ctx, client.Config{
		ClientID:     os.Getenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID"),
		ClientSecret: os.Getenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET"),
		BaseURL:      acctest.BaseURL(),
	})
}
