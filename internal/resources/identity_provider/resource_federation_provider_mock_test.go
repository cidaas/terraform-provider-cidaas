package identity_provider_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// mockFederationAPI is an in-memory stand-in for identity-provider-srv /federation/providers that mimics
// the behaviour Terraform depends on: wellknown_endpoint is not persisted, SAML SP certificates and
// idp_initiated_config._id are generated, PUT replaces the document, secrets are returned on GET.
type mockFederationAPI struct {
	mu        sync.Mutex
	seq       int
	providers map[string]map[string]interface{}
	bodies    []map[string]interface{} // POST/PUT bodies in order
}

func newMockFederationAPI(t *testing.T) (*mockFederationAPI, *httptest.Server) {
	t.Helper()
	m := &mockFederationAPI{providers: map[string]map[string]interface{}{}}
	server := httptest.NewServer(http.HandlerFunc(m.serve))
	t.Cleanup(server.Close)
	return m, server
}

func (m *mockFederationAPI) lastBody() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.bodies) == 0 {
		return nil
	}
	return m.bodies[len(m.bodies)-1]
}

func (m *mockFederationAPI) serve(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	respond := func(status int, data interface{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": status < 300, "status": status, "data": data})
	}

	if r.URL.Path == "/token-srv/token" {
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "mock-token"})
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/federation/providers") {
		respond(http.StatusNotFound, nil)
		return
	}
	id := strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, "/federation/providers"), "/")

	var body, sent map[string]interface{}
	if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
		_ = json.Unmarshal(raw, &body)
		_ = json.Unmarshal(raw, &sent) // untouched copy for assertions
	}

	switch {
	case r.Method == http.MethodGet && id == "":
		name := r.URL.Query().Get("provider_name")
		out := []interface{}{}
		for _, p := range m.providers {
			if name == "" || p["provider_name"] == name {
				out = append(out, p)
			}
		}
		respond(http.StatusOK, out)
	case r.Method == http.MethodGet:
		p, ok := m.providers[id]
		if !ok {
			respond(http.StatusNotFound, nil)
			return
		}
		respond(http.StatusOK, p)
	case r.Method == http.MethodPost:
		m.bodies = append(m.bodies, sent)
		if _, ok := body["id"]; ok {
			respond(http.StatusBadRequest, "id must not be sent on create")
			return
		}
		m.seq++
		body["id"] = fmt.Sprintf("fed-%d", m.seq)
		body["createdTime"] = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
		m.store(body, nil)
		respond(http.StatusCreated, m.providers[body["id"].(string)])
	case r.Method == http.MethodPut:
		m.bodies = append(m.bodies, sent)
		saved, ok := m.providers[id]
		if !ok {
			respond(http.StatusNotFound, nil)
			return
		}
		body["id"] = id
		body["createdTime"] = saved["createdTime"]
		m.store(body, saved)
		respond(http.StatusOK, m.providers[id])
	case r.Method == http.MethodDelete:
		if _, ok := m.providers[id]; !ok {
			respond(http.StatusNotFound, nil)
			return
		}
		delete(m.providers, id)
		w.WriteHeader(http.StatusNoContent)
	default:
		respond(http.StatusMethodNotAllowed, nil)
	}
}

func (m *mockFederationAPI) store(p, saved map[string]interface{}) {
	p["updatedTime"] = time.Date(2026, 1, 2, 0, 0, 0, m.seq, time.UTC).Format(time.RFC3339Nano)
	if o, ok := p["oauth2"].(map[string]interface{}); ok {
		delete(o, "wellknown_endpoint") // bson:"-"
	}
	if s, ok := p["saml"].(map[string]interface{}); ok {
		certs := []interface{}{map[string]interface{}{"certificate": "-----BEGIN CERTIFICATE-----MOCK", "sha256_fingerprint": "AA:BB"}}
		if saved != nil {
			if ss, ok := saved["saml"].(map[string]interface{}); ok {
				certs = ss["sp_signing_certificates"].([]interface{})
			}
		}
		s["sp_signing_certificates"] = certs
		if idp, ok := s["idp_initiated_config"].(map[string]interface{}); ok && idp["_id"] == nil {
			idp["_id"] = "idp-init-1"
		}
	}
	m.providers[p["id"].(string)] = p
}

func mockProviderBlock(baseURL string) string {
	return fmt.Sprintf(`
provider "cidaas" {
  base_url       = %q
  client_id      = "mock"
  client_secret  = "mock"
  cidaas_version = "4.x"
}
`, baseURL)
}

func checkSentOAuth2(m *mockFederationAPI, key, want string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		body := m.lastBody()
		o, _ := body["oauth2"].(map[string]interface{})
		if got, _ := o[key].(string); got != want {
			return fmt.Errorf("sent oauth2.%s = %q, want %q (body: %v)", key, got, want, body)
		}
		return nil
	}
}

func checkNotSent(m *mockFederationAPI, nest, key string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		body := m.lastBody()
		if n, ok := body[nest].(map[string]interface{}); ok {
			if _, present := n[key]; present {
				return fmt.Errorf("%s.%s must not be sent (body: %v)", nest, key, body)
			}
		}
		return nil
	}
}

func TestFederationProviderMock_CatalogSocial(t *testing.T) {
	m, server := newMockFederationAPI(t)
	name := "cidaas_federation_provider.apple"

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockProviderBlock(server.URL) + `
resource "cidaas_federation_provider" "apple" {
  provider_name = "default_apple"
  provider_type = "apple"
  display_name  = "Apple"
  standard_type = "OAUTH2"

  oauth2 = {
    client_id     = "apple-client"
    client_secret = "apple-secret"
    scopes        = ["openid", "email"]
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "id", "fed-1"),
					resource.TestCheckResourceAttr(name, "provider_type", "apple"),
					resource.TestCheckResourceAttr(name, "owner", "client"),
					resource.TestCheckResourceAttr(name, "oauth2.client_secret", "apple-secret"),
					resource.TestCheckResourceAttr(name, "oauth2.pkce", "false"),
					resource.TestCheckNoResourceAttr(name, "oauth2.authorization_endpoint"),
					resource.TestCheckNoResourceAttr(name, "oauth2.userinfo_source"),
					resource.TestCheckResourceAttrSet(name, "created_time"),
					checkNotSent(m, "oauth2", "userinfoSource"),
					checkNotSent(m, "oauth2", "authorization_endpoint"),
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

func TestFederationProviderMock_CustomOIDCWriteOnly(t *testing.T) {
	m, server := newMockFederationAPI(t)
	name := "cidaas_federation_provider.oidc"
	config := func(displayName, secret, version string) string {
		return mockProviderBlock(server.URL) + fmt.Sprintf(`
resource "cidaas_federation_provider" "oidc" {
  provider_name = "my_oidc"
  provider_type = "custom"
  display_name  = %q
  standard_type = "OPENID_CONNECT"
  domains       = ["example.com"]

  attribute_mapping = {
    email = {
      field_key    = "email"
      external_key = "mail"
    }
    "customFields.department" = {
      field_key    = "customFields.department"
      external_key = "dept"
      data_type    = "string"
      default      = "none"
    }
  }

  oauth2 = {
    client_id                = "cid"
    client_secret_wo         = %q
    client_secret_wo_version = %q
    wellknown_endpoint       = "https://idp.example.com/.well-known/openid-configuration"
    userinfo_source          = "USERINFOENDPOINT"
    token_endpoint_auth      = jsonencode({ apiAccessType = "BASIC" })
    authorization_parameters = { prompt = "login" }
    scopes_config = {
      display_label = "Permission"
      scopes        = [{ scope_name = "email", required = true }]
    }
    claims = {
      required_claims = { id_token = ["sub"] }
    }
    amr_config = [{ amr_value = "pwd", ext_amr_value = "password" }]
  }
}
`, displayName, secret, version)
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config("My OIDC", "secret-v1", "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(name, "oauth2.client_secret"),
					resource.TestCheckNoResourceAttr(name, "oauth2.client_secret_wo"),
					resource.TestCheckResourceAttr(name, "oauth2.client_secret_wo_version", "1"),
					resource.TestCheckResourceAttr(name, "oauth2.wellknown_endpoint", "https://idp.example.com/.well-known/openid-configuration"),
					resource.TestCheckResourceAttr(name, "oauth2.scopes_config.scopes.0.required", "true"),
					resource.TestCheckResourceAttr(name, "oauth2.scopes_config.scopes.0.recommended", "false"),
					resource.TestCheckResourceAttr(name, "attribute_mapping.email.ignore", "false"),
					resource.TestCheckResourceAttr(name, "attribute_mapping.customFields.department.default", "none"),
					checkSentOAuth2(m, "client_secret", "secret-v1"),
					checkSentOAuth2(m, "wellknown_endpoint", "https://idp.example.com/.well-known/openid-configuration"),
				),
			},
			{
				Config: config("My OIDC v2", "secret-v2", "2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "display_name", "My OIDC v2"),
					resource.TestCheckNoResourceAttr(name, "oauth2.client_secret"),
					checkSentOAuth2(m, "client_secret", "secret-v2"),
				),
			},
			{
				ResourceName:            name,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oauth2.client_secret", "oauth2.client_secret_wo_version", "oauth2.wellknown_endpoint"},
			},
		},
	})
}

func TestFederationProviderMock_SAML(t *testing.T) {
	_, server := newMockFederationAPI(t)
	name := "cidaas_federation_provider.saml"
	config := func(signAuth bool) string {
		return mockProviderBlock(server.URL) + fmt.Sprintf(`
resource "cidaas_federation_provider" "saml" {
  provider_name = "corp_saml"
  provider_type = "custom"
  display_name  = "Corp SAML"
  standard_type = "SAML"

  attribute_mapping = {
    email = {
      field_key    = "email"
      external_key = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
      data_type    = "xsd:string"
    }
  }

  saml = {
    meta_data         = "<EntityDescriptor/>"
    sign_auth_request = %t
    idp_initiated_config = {
      client_id = "app-client"
      protocol  = "SAML"
      enabled   = true
    }
    signature_config = {
      signing_algorithm = "RSA-SHA256"
      sign_targets      = ["authnrequest"]
    }
    encryption_config = {
      require_encrypted_for        = ["assertion"]
      content_encryption_algorithm = "AES-256-GCM"
    }
  }
}
`, signAuth)
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "saml.sp_signing_certificates.#", "1"),
					resource.TestCheckResourceAttr(name, "saml.sp_signing_certificates.0.sha256_fingerprint", "AA:BB"),
					resource.TestCheckResourceAttr(name, "saml.sp_encryption_certificates.#", "0"),
					resource.TestCheckResourceAttr(name, "saml.idp_initiated_config.id", "idp-init-1"),
					resource.TestCheckResourceAttr(name, "saml.encryption_config.encrypt_targets.#", "0"),
				),
			},
			{
				Config: config(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "saml.sign_auth_request", "true"),
					resource.TestCheckResourceAttr(name, "saml.idp_initiated_config.id", "idp-init-1"),
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

func TestFederationProviderMock_LDAPAndDataSource(t *testing.T) {
	m, server := newMockFederationAPI(t)
	name := "cidaas_federation_provider.ldap"
	config := mockProviderBlock(server.URL) + `
resource "cidaas_federation_provider" "ldap" {
  provider_name = "corp_ldap"
  provider_type = "custom"
  display_name  = "Corp LDAP"
  standard_type = "LDAP"

  ldap = {
    connection_string           = "ldaps://ldap.example.com"
    port                        = 636
    use_ssl                     = true
    bind_dn                     = "cn=admin,dc=example,dc=com"
    bind_credentials_wo         = "ldap-pw"
    bind_credentials_wo_version = "1"
    base_dn                     = "dc=example,dc=com"
    default_groups              = [{ group_id = "CIDAAS_USERS", roles = ["USER"] }]
  }
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "ldap.port", "636"),
					resource.TestCheckNoResourceAttr(name, "ldap.bind_credentials"),
					resource.TestCheckResourceAttr(name, "ldap.default_groups.0.roles.0", "USER"),
					func(_ *terraform.State) error {
						l, _ := m.lastBody()["ldap"].(map[string]interface{})
						if l["bind_credentials"] != "ldap-pw" {
							return fmt.Errorf("write-only bind credentials not sent: %v", l)
						}
						return nil
					},
				),
			},
			{
				Config: config + `
data "cidaas_federation_provider" "by_id" {
  id = cidaas_federation_provider.ldap.id
}

data "cidaas_federation_provider" "by_name" {
  provider_name = cidaas_federation_provider.ldap.provider_name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.cidaas_federation_provider.by_id", "id", name, "id"),
					resource.TestCheckResourceAttr("data.cidaas_federation_provider.by_id", "provider_name", "corp_ldap"),
					resource.TestCheckResourceAttr("data.cidaas_federation_provider.by_id", "ldap.bind_credentials", "ldap-pw"),
					resource.TestCheckNoResourceAttr("data.cidaas_federation_provider.by_id", "ldap.bind_credentials_wo_version"),
					resource.TestCheckResourceAttrPair("data.cidaas_federation_provider.by_name", "id", name, "id"),
					resource.TestCheckResourceAttr("data.cidaas_federation_provider.by_name", "ldap.default_groups.0.group_id", "CIDAAS_USERS"),
				),
			},
			{
				Config: config + `
data "cidaas_federation_provider" "missing" {
  provider_name = "does_not_exist"
}
`,
				ExpectError: regexp.MustCompile(`no federation provider found with provider_name "does_not_exist"`),
			},
		},
	})
}

func TestFederationProviderMock_ValidationErrors(t *testing.T) {
	_, server := newMockFederationAPI(t)
	cases := map[string]struct {
		hcl  string
		want string
	}{
		"custom without endpoints": {
			hcl: `provider_type = "custom"
  standard_type = "OAUTH2"
  oauth2 = {
    client_id       = "c"
    client_secret   = "s"
    userinfo_source = "USERINFOENDPOINT"
  }`,
			want: "`authorization_endpoint` is required for custom providers",
		},
		"custom without userinfo_source": {
			hcl: `provider_type = "custom"
  standard_type = "OAUTH2"
  oauth2 = {
    client_id          = "c"
    client_secret      = "s"
    wellknown_endpoint = "https://idp.example.com/.well-known/openid-configuration"
  }`,
			want: "`userinfo_source` is required for custom providers",
		},
		"missing secret without pkce": {
			hcl: `provider_type = "google"
  standard_type = "OAUTH2"
  oauth2 = { client_id = "c" }`,
			want: "One of `client_secret` or `client_secret_wo` is required",
		},
		"nest does not match": {
			hcl: `provider_type = "google"
  standard_type = "SAML"
  display_name  = "x"
  oauth2 = {
    client_id     = "c"
    client_secret = "s"
  }`,
			want: "`saml` is required when standard_type is \"SAML\"",
		},
		"saml without display_name": {
			hcl: `provider_type = "custom"
  standard_type = "SAML"
  saml = { meta_data = "<x/>" }`,
			want: "`display_name` is required when standard_type is \"SAML\"",
		},
		"mapping key mismatch": {
			hcl: `provider_type = "google"
  standard_type = "OAUTH2"
  oauth2 = {
    client_id = "c"
    pkce      = true
  }
  attribute_mapping = {
    email = {
      field_key    = "mail"
      external_key = "email"
    }
  }`,
			want: `field_key "mail" must equal the map key "email"`,
		},
		"ldap without host": {
			hcl: `provider_type = "custom"
  standard_type = "LDAP"
  ldap = {
    bind_dn = "cn=a"
    base_dn = "dc=b"
  }`,
			want: "At least one of `connection_string` or `host` is required",
		},
		"http endpoint": {
			hcl: `provider_type = "custom"
  standard_type = "OAUTH2"
  oauth2 = {
    client_id              = "c"
    client_secret          = "s"
    userinfo_source        = "IDTOKEN"
    authorization_endpoint = "http://idp.example.com/authorize"
    token_endpoint         = "https://idp.example.com/token"
  }`,
			want: "must be an https:// URL",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config: mockProviderBlock(server.URL) + `
resource "cidaas_federation_provider" "invalid" {
  provider_name = "invalid"
  ` + tc.hcl + `
}
`,
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(regexp.QuoteMeta(tc.want)),
				}},
			})
		})
	}
}
