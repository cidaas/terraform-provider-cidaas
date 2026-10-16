package identity_provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func federationAccTest(t *testing.T) (string, string) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	t.Parallel()
	acctest.SkipIfV3(t)
	id := acctest.RandString(8)
	return id, fmt.Sprintf("%s.%s", base.RESOURCE_FEDERATION_PROVIDER, id)
}

func federationProviderBlock() string {
	return fmt.Sprintf(`
provider "cidaas" {
  base_url = "%s"
}
`, acctest.GetBaseURL())
}

func TestAccFederationProvider_CatalogSocial(t *testing.T) {
	id, name := federationAccTest(t)
	providerName := "tf_acc_apple_" + id
	config := func(displayName string) string {
		return federationProviderBlock() + fmt.Sprintf(`
resource "cidaas_federation_provider" "%s" {
  provider_name = "%s"
  provider_type = "apple"
  display_name  = "%s"
  standard_type = "OAUTH2"

  oauth2 = {
    client_id     = "acc_apple_client"
    client_secret = "acc_apple_secret"
  }
}
`, id, providerName, displayName)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config("Apple"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(name, "id"),
					resource.TestCheckResourceAttr(name, "provider_type", "apple"),
					resource.TestCheckNoResourceAttr(name, "oauth2.authorization_endpoint"),
					resource.TestCheckNoResourceAttr(name, "oauth2.token_endpoint"),
				),
			},
			{
				Config: config("Apple Login"),
				Check:  resource.TestCheckResourceAttr(name, "display_name", "Apple Login"),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFederationProvider_CustomOAuth2(t *testing.T) {
	id, name := federationAccTest(t)
	providerName := "tf_acc_oauth2_" + id

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: federationProviderBlock() + fmt.Sprintf(`
resource "cidaas_federation_provider" "%s" {
  provider_name = "%s"
  provider_type = "custom"
  display_name  = "TF Acc Custom OAuth2"
  standard_type = "OAUTH2"
  logo_url      = "https://cdn.example.com/logo.svg"
  domains       = ["example.com"]

  attribute_mapping = {
    email = {
      field_key    = "email"
      external_key = "email"
    }
  }

  oauth2 = {
    client_id                = "acc_client"
    client_secret_wo         = "acc_secret"
    client_secret_wo_version = "1"
    userinfo_source          = "USERINFOENDPOINT"
    authorization_endpoint   = "https://auth.example.com/oauth/authorize"
    token_endpoint           = "https://auth.example.com/oauth/token"
    userinfo_endpoint        = "https://auth.example.com/oauth/userinfo"
    scopes                   = ["openid", "email"]
  }
}
`, id, providerName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(name, "id"),
					resource.TestCheckResourceAttr(name, "standard_type", "OAUTH2"),
					resource.TestCheckNoResourceAttr(name, "oauth2.client_secret"),
					resource.TestCheckResourceAttr(name, "oauth2.userinfo_source", "USERINFOENDPOINT"),
				),
			},
			{
				ResourceName:            name,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oauth2.client_secret", "oauth2.client_secret_wo_version"},
			},
		},
	})
}

func TestAccFederationProvider_CustomOIDCWellknown(t *testing.T) {
	id, name := federationAccTest(t)
	providerName := "tf_acc_oidc_" + id

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: federationProviderBlock() + fmt.Sprintf(`
resource "cidaas_federation_provider" "%s" {
  provider_name = "%s"
  provider_type = "custom"
  display_name  = "TF Acc OIDC"
  standard_type = "OPENID_CONNECT"

  oauth2 = {
    client_id          = "acc_oidc_client"
    client_secret      = "acc_oidc_secret"
    wellknown_endpoint = "https://accounts.google.com/.well-known/openid-configuration"
    userinfo_source    = "IDTOKEN"
  }
}
`, id, providerName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "standard_type", "OPENID_CONNECT"),
					resource.TestCheckResourceAttr(name, "oauth2.wellknown_endpoint", "https://accounts.google.com/.well-known/openid-configuration"),
				),
			},
			{
				ResourceName:            name,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oauth2.wellknown_endpoint"},
			},
		},
	})
}

func TestAccFederationProvider_SAML(t *testing.T) {
	id, name := federationAccTest(t)
	providerName := "tf_acc_saml_" + id

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: federationProviderBlock() + fmt.Sprintf(`
resource "cidaas_federation_provider" "%s" {
  provider_name = "%s"
  provider_type = "custom"
  display_name  = "TF Acc SAML"
  standard_type = "SAML"

  attribute_mapping = {
    email = {
      field_key    = "email"
      external_key = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
      data_type    = "xsd:string"
    }
  }

  saml = {
    meta_data = <<-EOT
      <EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://idp.example.com">
        <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
          <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://idp.example.com/sso"/>
        </IDPSSODescriptor>
      </EntityDescriptor>
    EOT
  }
}
`, id, providerName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "standard_type", "SAML"),
					resource.TestCheckResourceAttrSet(name, "saml.sp_signing_certificates.#"),
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFederationProvider_LDAP(t *testing.T) {
	id, name := federationAccTest(t)
	providerName := "tf_acc_ldap_" + id

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: federationProviderBlock() + fmt.Sprintf(`
resource "cidaas_federation_provider" "%s" {
  provider_name = "%s"
  provider_type = "custom"
  display_name  = "TF Acc LDAP"
  standard_type = "LDAP"

  ldap = {
    connection_string = "ldaps://ldap.example.com:636"
    bind_dn           = "cn=admin,dc=example,dc=com"
    bind_credentials  = "acc_ldap_password"
    base_dn           = "dc=example,dc=com"
  }
}
`, id, providerName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "standard_type", "LDAP"),
					resource.TestCheckResourceAttr(name, "ldap.bind_dn", "cn=admin,dc=example,dc=com"),
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
