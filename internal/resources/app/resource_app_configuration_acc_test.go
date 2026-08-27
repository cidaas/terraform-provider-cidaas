package app_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/acctest"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAppConfiguration_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	t.Parallel()

	name := "tf-app-" + acctest.RandString(8)
	resourceName := "cidaas_app_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testCheckAppConfigurationDestroyed(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfigurationConfig(name, []string{"openid"}, []string{"openid"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_name", name),
					resource.TestCheckResourceAttrSet(resourceName, "client_id"),
					resource.TestCheckResourceAttr(resourceName, "client_type", "NON_INTERACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "owner", "client"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "scopes.allowed_scopes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scopes.allowed_scopes.0", "openid"),
					resource.TestCheckResourceAttrSet(resourceName, "signing_key_config.active_kid"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource %s not found in state", resourceName)
					}
					id := rs.Primary.Attributes["client_id"]
					if id == "" {
						return "", fmt.Errorf("client_id not in state for %s", resourceName)
					}
					return id, nil
				},
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "client_id",
			},
			{
				Config: testAccAppConfigurationConfig(name, []string{"openid", "profile"}, []string{"openid"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "scopes.allowed_scopes.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "scopes.allowed_scopes.1", "profile"),
				),
			},
		},
	})
}

func testAccAppConfigurationConfig(name string, allowed, defaults []string) string {
	allowedHCL := hclStringList(allowed)
	defaultsHCL := hclStringList(defaults)
	return fmt.Sprintf(`
%s
resource "cidaas_app_configuration" "test" {
  client_name = %q
  client_type = "NON_INTERACTIVE"
  enabled     = true

  grant_types = ["client_credentials"]

  client_auth_config = {
    token_endpoint_auth_method = "none"
  }

  ownership_details = {
    company_name    = "Terraform Acc Test Corp"
    company_address = "1 Test Way"
    company_website = "https://example.com"
  }

  scopes = {
    allowed_scopes = %s
    default_scopes = %s
  }
}
`, acctest.ProviderConfig(), name, allowedHCL, defaultsHCL)
}

func hclStringList(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	out := "["
	for i, v := range values {
		if i > 0 {
			out += ", "
		}
		out += fmt.Sprintf("%q", v)
	}
	out += "]"
	return out
}

func testCheckAppConfigurationDestroyed(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		c, err := accClient(context.Background())
		if err != nil {
			return err
		}
		_, err = c.AppConfiguration.Get(context.Background(), rs.Primary.Attributes["client_id"])
		if err == nil {
			return fmt.Errorf("app configuration %s still exists", rs.Primary.Attributes["client_id"])
		}
		if !errors.Is(err, client.ErrNotFound) {
			return fmt.Errorf("expected not found for %s, got: %w", rs.Primary.Attributes["client_id"], err)
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
