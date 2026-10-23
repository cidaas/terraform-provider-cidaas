package verification_test

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

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// mockVerificationConfigAPI mimics verification-actions-srv /config: system methods (owner "system") only
// accept active/localeText/encryptionConfig and always get localeText replaced; POST creates pluggable
// methods (owner CLIENT, 409 if the type exists); DELETE refuses system methods (403); unknown types get
// 400 "verification method not found". Write responses carry nanosecond timestamps, GET milliseconds.
type mockVerificationConfigAPI struct {
	mu      sync.Mutex
	methods map[string]map[string]any
	order   []string
	puts    []map[string]any
}

func newMockVerificationConfigAPI(t *testing.T) (*mockVerificationConfigAPI, *httptest.Server) {
	t.Helper()
	m := &mockVerificationConfigAPI{methods: map[string]map[string]any{}}
	m.add(map[string]any{
		"_id": "id-email", "verificationType": "EMAIL", "owner": "system", "active": true,
		"factor": "POSSESSION", "category": "code", "amr_values": []any{"103"},
		"code":        map[string]any{"length": 6, "validity": 1800},
		"localeText":  []any{map[string]any{"locale": "en", "name": "Email", "description": "Email code"}},
		"createdTime": "2026-01-01T00:00:00.123Z", "updatedTime": "2026-01-01T00:00:00.123Z",
	})
	m.add(map[string]any{
		"_id": "id-sms", "verificationType": "SMS", "owner": "system", "active": true,
		"encryptionConfig": map[string]any{"enabled": false, "mode": "OPTIONAL", "encryptedFields": []any{"mobile_number"}},
		"createdTime":      "2026-01-01T00:00:00.123Z",
	})
	server := httptest.NewServer(http.HandlerFunc(m.serve))
	t.Cleanup(server.Close)
	return m, server
}

func (m *mockVerificationConfigAPI) add(method map[string]any) {
	m.methods[method["verificationType"].(string)] = method
	m.order = append(m.order, method["verificationType"].(string))
}

func (m *mockVerificationConfigAPI) get(verificationType string) map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.methods[verificationType]
}

func (m *mockVerificationConfigAPI) lastPut() map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.puts) == 0 {
		return nil
	}
	return m.puts[len(m.puts)-1]
}

func withName(method map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range method {
		out[k] = v
	}
	if lt, ok := method["localeText"].([]any); ok && len(lt) > 0 {
		first := lt[0].(map[string]any)
		out["name"], out["description"] = first["name"], first["description"]
	}
	return out
}

func withNanos(method map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range method {
		out[k] = v
	}
	for _, key := range []string{"createdTime", "updatedTime"} {
		if s, ok := out[key].(string); ok {
			out[key] = strings.TrimSuffix(s, "Z") + "456789Z"
		}
	}
	return out
}

func (m *mockVerificationConfigAPI) serve(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	respond := func(status int, data any) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": status < 300, "status": status, "data": data})
	}
	fail := func(status int, msg string) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "status": status, "error": map[string]any{"error": msg}})
	}
	if r.URL.Path == "/token-srv/token" {
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "mock-token"})
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/verification-actions-srv/config") {
		fail(http.StatusNotFound, "not found")
		return
	}
	var body map[string]any
	if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
		_ = json.Unmarshal(raw, &body)
	}

	switch r.Method {
	case http.MethodGet:
		out := []any{}
		for _, vt := range m.order {
			if method, ok := m.methods[vt]; ok {
				out = append(out, withName(method))
			}
		}
		respond(http.StatusOK, out)
	case http.MethodPost:
		vt, _ := body["verificationType"].(string)
		auth, _ := body["cidaasAuthDetails"].(map[string]any)
		switch {
		case body["url"] == nil || body["auth_type"] == nil || auth == nil || auth["client_id"] == nil || vt == "":
			fail(http.StatusExpectationFailed, "url, auth_type, cidaasAuthDetails.client_id and verificationType are required")
			return
		case m.methods[vt] != nil:
			fail(http.StatusConflict, "verification method already exists")
			return
		}
		body["_id"], body["owner"] = "id-"+strings.ToLower(vt), "CLIENT"
		if body["active"] == nil {
			body["active"] = true
		}
		body["createdTime"], body["updatedTime"] = "2026-02-01T00:00:00.5Z", "2026-02-01T00:00:00.5Z"
		m.add(body)
		respond(http.StatusCreated, withNanos(body))
	case http.MethodPut:
		m.puts = append(m.puts, body)
		vt, _ := body["verificationType"].(string)
		saved := m.methods[vt]
		if saved == nil {
			fail(http.StatusBadRequest, "verification method not found")
			return
		}
		if body["active"] != nil {
			saved["active"] = body["active"]
		}
		saved["localeText"] = body["localeText"] // always replaced, like the service
		if saved["owner"] == "CLIENT" {
			for _, key := range []string{"url", "auth_type", "prerequisite", "cidaasAuthDetails"} {
				if body[key] != nil {
					saved[key] = body[key]
				}
			}
			for _, key := range []string{"factor", "category", "icon", "amr_values", "acr_pin"} {
				saved[key] = body[key]
			}
		} else if enc, ok := body["encryptionConfig"].(map[string]any); ok {
			current, configured := saved["encryptionConfig"].(map[string]any)
			if !configured {
				fail(http.StatusBadRequest, "encryptionConfig is not configured for this verification method")
				return
			}
			current["enabled"], current["mode"] = enc["enabled"], enc["mode"]
		}
		saved["updatedTime"] = "2026-03-01T00:00:00.25Z"
		respond(http.StatusOK, withNanos(saved))
	case http.MethodDelete:
		vt := strings.TrimPrefix(r.URL.Path, "/verification-actions-srv/config/")
		saved := m.methods[vt]
		switch {
		case saved == nil:
			fail(http.StatusBadRequest, "verification method not found")
		case saved["owner"] != "CLIENT":
			fail(http.StatusForbidden, "cannot delete system-owned verification method")
		default:
			delete(m.methods, vt)
			respond(http.StatusOK, true)
		}
	}
}

// wrappedErr matches msg in a Terraform error even when the CLI wraps it across lines.
func wrappedErr(msg string) *regexp.Regexp {
	return regexp.MustCompile(strings.ReplaceAll(regexp.QuoteMeta(msg), " ", `\s+`))
}

func mockVerificationProvider(baseURL string) string {
	return fmt.Sprintf(`
provider "cidaas" {
  base_url       = %q
  client_id      = "mock"
  client_secret  = "mock"
  cidaas_version = "4.x"
}
`, baseURL)
}

func TestVerificationMethodMock_SystemMethod(t *testing.T) {
	m, server := newMockVerificationConfigAPI(t)
	name := "cidaas_verification_method.email"
	config := func(body string) string {
		return mockVerificationProvider(server.URL) + `
resource "cidaas_verification_method" "email" {
  verification_type = "EMAIL"
` + body + `
}
`
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			if m.get("EMAIL") == nil {
				return fmt.Errorf("system method EMAIL must not be deleted on destroy")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				// Adopt with only active: locale texts must be sent back unchanged.
				Config: config(`  active = false`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "id", "EMAIL"),
					resource.TestCheckResourceAttr(name, "owner", "system"),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "locale_text.0.name", "Email"),
					resource.TestCheckResourceAttr(name, "name", "Email"),
					resource.TestCheckResourceAttr(name, "factor", "POSSESSION"),
					resource.TestCheckResourceAttr(name, "amr_values.0", "103"),
					resource.TestCheckResourceAttr(name, "code.length", "6"),
					resource.TestCheckResourceAttr(name, "created_time", "2026-01-01T00:00:00.123Z"),
					func(_ *terraform.State) error {
						if lt, _ := m.lastPut()["localeText"].([]any); len(lt) != 1 {
							return fmt.Errorf("PUT must send the existing localeText, got %v", m.lastPut())
						}
						return nil
					},
				),
			},
			{
				Config: config(`
  active = true
  locale_text = [
    { locale = "en", name = "E-Mail", description = "Code per e-mail" },
    { locale = "de", name = "E-Mail", description = "Code per E-Mail" },
  ]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "locale_text.#", "2"),
					resource.TestCheckResourceAttr(name, "name", "E-Mail"),
					resource.TestCheckResourceAttr(name, "updated_time", "2026-03-01T00:00:00.25Z"),
				),
			},
			{
				// Removing active/locale_text from config keeps the current values (no wipe).
				Config: config(``),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "locale_text.#", "2"),
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateId:     "EMAIL",
				ImportStateVerify: true,
			},
		},
	})
}

func TestVerificationMethodMock_SystemEncryption(t *testing.T) {
	_, server := newMockVerificationConfigAPI(t)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockVerificationProvider(server.URL) + `
resource "cidaas_verification_method" "sms" {
  verification_type = "SMS"
  encryption_config = { enabled = true, mode = "REQUIRED" }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("cidaas_verification_method.sms", "encryption_config.enabled", "true"),
					resource.TestCheckResourceAttr("cidaas_verification_method.sms", "encryption_config.mode", "REQUIRED"),
					resource.TestCheckResourceAttr("cidaas_verification_method.sms", "encryption_config.encrypted_fields.0", "mobile_number"),
				),
			},
			{
				Config: mockVerificationProvider(server.URL) + `
resource "cidaas_verification_method" "email" {
  verification_type = "EMAIL"
  encryption_config = { enabled = true, mode = "REQUIRED" }
}
`,
				ExpectError: wrappedErr("encryptionConfig is not configured for this verification method"),
			},
		},
	})
}

func TestVerificationMethodMock_Pluggable(t *testing.T) {
	m, server := newMockVerificationConfigAPI(t)
	name := "cidaas_verification_method.spi"
	config := func(icon, amr string) string {
		return mockVerificationProvider(server.URL) + fmt.Sprintf(`
resource "cidaas_verification_method" "spi" {
  verification_type = "MY_SPI"
  url               = "https://spi.example.com/verify"
  auth_type         = "Bearer"
  cidaas_auth_details = { client_id = "spi-client" }
  factor     = "POSSESSION"
  category   = "code"
  icon       = %q
  amr_values = [%q]
  acr_pin    = "1"
  locale_text = [{ locale = "en", name = "My SPI", description = "" }]
}
`, icon, amr)
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			if m.get("MY_SPI") != nil {
				return fmt.Errorf("pluggable method MY_SPI must be deleted on destroy")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config("https://cdn.example.com/a.svg", "spi1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "owner", "CLIENT"),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "cidaas_auth_details.client_id", "spi-client"),
					resource.TestCheckResourceAttr(name, "name", "My SPI"),
					resource.TestCheckResourceAttr(name, "created_time", "2026-02-01T00:00:00.5Z"),
				),
			},
			{
				Config: config("https://cdn.example.com/b.svg", "spi2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "icon", "https://cdn.example.com/b.svg"),
					resource.TestCheckResourceAttr(name, "amr_values.0", "spi2"),
					func(_ *terraform.State) error {
						put := m.lastPut()
						if put["url"] != "https://spi.example.com/verify" || put["factor"] != "POSSESSION" || put["acr_pin"] != "1" {
							return fmt.Errorf("pluggable PUT must send the full config, got %v", put)
						}
						return nil
					},
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateId:     "MY_SPI",
				ImportStateVerify: true,
			},
		},
	})
}

func TestVerificationMethodMock_Errors(t *testing.T) {
	m, server := newMockVerificationConfigAPI(t)
	m.add(map[string]any{"verificationType": "EXISTING_SPI", "owner": "CLIENT", "url": "https://x.example.com", "auth_type": "Bearer"})
	cases := map[string]struct {
		hcl  string
		want string
	}{
		"pluggable field on system method": {
			hcl:  `verification_type = "EMAIL"` + "\n  factor = \"KNOWLEDGE\"",
			want: "`factor` can only be set on pluggable methods",
		},
		"unknown type without url": {
			hcl:  `verification_type = "DOES_NOT_EXIST"`,
			want: `Verification method "DOES_NOT_EXIST" does not exist`,
		},
		"pluggable already exists": {
			hcl: `verification_type = "EXISTING_SPI"
  url       = "https://x.example.com"
  auth_type = "Bearer"
  cidaas_auth_details = { client_id = "c" }`,
			want: "terraform import <resource address> EXISTING_SPI",
		},
		"url without auth": {
			hcl:  `verification_type = "NEW_SPI"` + "\n  url = \"https://x.example.com\"",
			want: "`auth_type` is required with `url`",
		},
		"relative url": {
			hcl: `verification_type = "NEW_SPI"
  url       = "/verify"
  auth_type = "Bearer"
  cidaas_auth_details = { client_id = "c" }`,
			want: "must be an absolute URL",
		},
		"encryption on pluggable": {
			hcl: `verification_type = "NEW_SPI"
  url       = "https://x.example.com"
  auth_type = "Bearer"
  cidaas_auth_details = { client_id = "c" }
  encryption_config = { enabled = true, mode = "OPTIONAL" }`,
			want: "`encryption_config` can only be set on system methods",
		},
		"empty locale": {
			hcl:  `verification_type = "EMAIL"` + "\n  locale_text = [{ locale = \"\", name = \"x\", description = \"\" }]",
			want: "needs a non-empty `locale`",
		},
		"invalid auth_type": {
			hcl: `verification_type = "NEW_SPI"
  url       = "https://x.example.com"
  auth_type = "basic"
  cidaas_auth_details = { client_id = "c" }`,
			want: `value must be one of`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config:      mockVerificationProvider(server.URL) + "\nresource \"cidaas_verification_method\" \"x\" {\n  " + tc.hcl + "\n}\n",
					ExpectError: wrappedErr(tc.want),
				}},
			})
		})
	}
}

func TestVerificationMethodMock_DataSource(t *testing.T) {
	_, server := newMockVerificationConfigAPI(t)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockVerificationProvider(server.URL) + `
data "cidaas_verification_method" "email" {
  verification_type = "EMAIL"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.cidaas_verification_method.email", "owner", "system"),
					resource.TestCheckResourceAttr("data.cidaas_verification_method.email", "active", "true"),
					resource.TestCheckResourceAttr("data.cidaas_verification_method.email", "code.validity", "1800"),
					resource.TestCheckResourceAttr("data.cidaas_verification_method.email", "locale_text.0.locale", "en"),
				),
			},
			{
				Config: mockVerificationProvider(server.URL) + `
data "cidaas_verification_method" "missing" {
  verification_type = "NOPE"
}
`,
				ExpectError: wrappedErr("available verification types: EMAIL, SMS"),
			},
		},
	})
}
