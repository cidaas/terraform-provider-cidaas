package verification_test

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

func TestAccSuggestVerificationMethod_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	t.Parallel()
	acctest.SkipIfV3(t)

	name := "tf-svm-" + acctest.RandString(8)
	resourceName := "cidaas_suggest_verification_method.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testCheckSuggestDestroyed(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccSuggestConfig(name, "ONEOF"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "suggest_verification_method.mandatory_config.range", "ONEOF"),
					resource.TestCheckResourceAttr(resourceName, "suggest_verification_method.skip_duration_in_days", "7"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSuggestConfig(name, "ALLOF"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "suggest_verification_method.mandatory_config.range", "ALLOF"),
					resource.TestCheckResourceAttr(resourceName, "description", "acc suggest updated"),
				),
			},
		},
	})
}

func TestAccVerificationOptions_Linked(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	t.Parallel()
	acctest.SkipIfV3(t)

	suffix := acctest.RandString(8)
	suggestName := "tf-svm-" + suffix
	optionsName := "tf-vo-" + suffix
	suggestRN := "cidaas_suggest_verification_method.linked"
	optionsRN := "cidaas_verification_options.linked"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSuggestDestroyed(suggestRN),
			testCheckOptionsDestroyed(optionsRN),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccLinkedConfig(suggestName, optionsName, "ALWAYS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(optionsRN, "name", optionsName),
					resource.TestCheckResourceAttrSet(optionsRN, "id"),
					resource.TestCheckResourceAttr(optionsRN, "verification_options.setting", "ALWAYS"),
					resource.TestCheckResourceAttrPair(optionsRN, "verification_options.suggest_verification_method_id", suggestRN, "id"),
				),
			},
			{
				ResourceName:      optionsRN,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLinkedConfig(suggestName, optionsName, "OFF"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(optionsRN, "verification_options.setting", "OFF"),
				),
			},
		},
	})
}

func TestAccSuggestVerificationMethod_MethodOverlap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	t.Parallel()
	acctest.SkipIfV3(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
%s
resource "cidaas_suggest_verification_method" "bad" {
  name        = "tf-svm-bad-%s"
  description = "overlap test"
  suggest_verification_method = {
    mandatory_config = {
      methods = ["EMAIL"]
      range   = "ONEOF"
    }
    optional_config = {
      methods = ["EMAIL"]
    }
  }
}
`, acctest.ProviderConfig(), acctest.RandString(6)),
				ExpectError: regexp.MustCompile(`both mandatory and optional`),
			},
		},
	})
}

func TestAccVerificationOptions_InvalidSetting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	t.Parallel()
	acctest.SkipIfV3(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
%s
resource "cidaas_verification_options" "bad" {
  name        = "tf-vo-bad-%s"
  description = "bad setting"
  verification_options = {
    setting = "NOT_REAL"
  }
}
`, acctest.ProviderConfig(), acctest.RandString(6)),
				ExpectError: regexp.MustCompile(`setting`),
			},
		},
	})
}

func testAccSuggestConfig(name, rangeVal string) string {
	desc := "acc suggest"
	if rangeVal == "ALLOF" {
		desc = "acc suggest updated"
	}
	return fmt.Sprintf(`
%s
resource "cidaas_suggest_verification_method" "test" {
  name        = %q
  description = %q
  suggest_verification_method = {
    mandatory_config = {
      methods = ["EMAIL"]
      range   = %q
    }
    optional_config = {
      methods = ["TOTP"]
    }
    skip_duration_in_days = 7
  }
}
`, acctest.ProviderConfig(), name, desc, rangeVal)
}

func testAccLinkedConfig(suggestName, optionsName, setting string) string {
	allowed := `["EMAIL", "SMS"]`
	if setting == "OFF" {
		allowed = `[]`
	}
	return fmt.Sprintf(`
%s
resource "cidaas_suggest_verification_method" "linked" {
  name        = %q
  description = "linked suggest"
  suggest_verification_method = {
    mandatory_config = {
      methods = ["EMAIL"]
      range   = "ONEOF"
    }
    skip_duration_in_days = 7
  }
}

resource "cidaas_verification_options" "linked" {
  name        = %q
  description = "linked options"
  verification_options = {
    setting                        = %q
    allowed_methods                = %s
    suggest_verification_method_id = cidaas_suggest_verification_method.linked.id
  }
}
`, acctest.ProviderConfig(), suggestName, optionsName, setting, allowed)
}

func testCheckSuggestDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		c, err := accClient(context.Background())
		if err != nil {
			return err
		}
		_, err = c.SuggestVerificationMethod.Get(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("suggest verification method %s still exists", rs.Primary.ID)
		}
		if !errors.Is(err, client.ErrNotFound) {
			return fmt.Errorf("expected not found for %s, got: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testCheckOptionsDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		c, err := accClient(context.Background())
		if err != nil {
			return err
		}
		_, err = c.VerificationOptions.Get(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("verification options %s still exists", rs.Primary.ID)
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
