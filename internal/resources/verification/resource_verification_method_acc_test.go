package verification_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/acctest"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccVerificationMethod_System adopts the EMAIL system method. It sets active = true (the tenant default)
// and does not touch locale texts; destroy must leave EMAIL on the tenant.
func TestAccVerificationMethod_System(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	t.Parallel()
	acctest.SkipIfV3(t)

	resourceName := "cidaas_verification_method.email"
	config := acctest.ProviderConfig() + `
resource "cidaas_verification_method" "email" {
  verification_type = "EMAIL"
  active            = true
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testCheckVerificationMethodExists("EMAIL"),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "EMAIL"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttrSet(resourceName, "locale_text.#"),
					resource.TestCheckResourceAttrSet(resourceName, "factor"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     "EMAIL",
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccVerificationMethod_Pluggable creates, updates, imports and deletes a pluggable method.
func TestAccVerificationMethod_Pluggable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	t.Parallel()
	acctest.SkipIfV3(t)

	suffix := acctest.RandString(8)
	verificationType := "TF_ACC_" + strings.ToUpper(suffix)
	resourceName := "cidaas_verification_method.spi"
	config := func(active bool, icon string) string {
		return acctest.ProviderConfig() + fmt.Sprintf(`
resource "cidaas_verification_method" "spi" {
  verification_type = %q
  active            = %t
  url               = "https://mfa.example.com/verify"
  auth_type         = "Bearer"
  cidaas_auth_details = {
    client_id = "tf-acc-spi-client"
  }
  factor     = "POSSESSION"
  category   = "code"
  icon       = %q
  amr_values = ["tf_acc_%s"]
  locale_text = [
    { locale = "en", name = "TF Acc SPI", description = "Acceptance test pluggable method" },
  ]
}
`, verificationType, active, icon, suffix)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testCheckVerificationMethodDeleted(verificationType),
		Steps: []resource.TestStep{
			{
				Config: config(true, "https://cdn.example.com/a.svg"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", verificationType),
					resource.TestCheckResourceAttr(resourceName, "owner", "CLIENT"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "cidaas_auth_details.client_id", "tf-acc-spi-client"),
					resource.TestCheckResourceAttr(resourceName, "amr_values.0", "tf_acc_"+suffix),
				),
			},
			{
				Config: config(false, "https://cdn.example.com/b.svg"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "active", "false"),
					resource.TestCheckResourceAttr(resourceName, "icon", "https://cdn.example.com/b.svg"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     verificationType,
				ImportStateVerify: true,
			},
		},
	})
}

// testCheckVerificationMethodExists asserts a (system) method is still on the tenant after destroy.
func testCheckVerificationMethodExists(verificationType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c, err := accClient(context.Background())
		if err != nil {
			return err
		}
		if _, err := c.VerificationMethod.Get(context.Background(), verificationType); err != nil {
			return fmt.Errorf("system verification method %s must remain after destroy: %w", verificationType, err)
		}
		return nil
	}
}

// testCheckVerificationMethodDeleted asserts a pluggable method is gone after destroy.
func testCheckVerificationMethodDeleted(verificationType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c, err := accClient(context.Background())
		if err != nil {
			return err
		}
		_, err = c.VerificationMethod.Get(context.Background(), verificationType)
		if err == nil {
			return fmt.Errorf("pluggable verification method %s still exists", verificationType)
		}
		if !errors.Is(err, client.ErrNotFound) {
			return fmt.Errorf("expected not found for %s, got: %w", verificationType, err)
		}
		return nil
	}
}
