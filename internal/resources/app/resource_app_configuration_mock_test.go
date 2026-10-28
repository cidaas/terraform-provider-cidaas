package app_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// mockAppConfigurationAPI mimics app-srv /app-srv/apps: PUT is a full replace (writable fields absent
// from the body are reset, like apikit.ApplyReplace) and additional_access_token_payload is rejected
// for NON_INTERACTIVE and SAML_SP (app.additional_access_token_payload.allowed).
type mockAppConfigurationAPI struct {
	mu   sync.Mutex
	apps map[string]map[string]any
	puts []map[string]any
}

// serverOwnedKeys survive a PUT regardless of the request body.
var serverOwnedKeys = []string{"_id", "client_id", "created_time", "signing_key_config"}

func newMockAppConfigurationAPI(t *testing.T) (*mockAppConfigurationAPI, *httptest.Server) {
	t.Helper()
	m := &mockAppConfigurationAPI{apps: map[string]map[string]any{}}
	server := httptest.NewServer(http.HandlerFunc(m.serve))
	t.Cleanup(server.Close)
	return m, server
}

// setPayload simulates an Admin UI change made outside Terraform.
func (m *mockAppConfigurationAPI) setPayload(clientID string, payload ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	values := make([]any, 0, len(payload))
	for _, p := range payload {
		values = append(values, p)
	}
	m.apps[clientID]["additional_access_token_payload"] = values
}

func (m *mockAppConfigurationAPI) storedPayload(clientID string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	raw, _ := m.apps[clientID]["additional_access_token_payload"].([]any)
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		out = append(out, v.(string))
	}
	slices.Sort(out)
	return out
}

func (m *mockAppConfigurationAPI) putCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.puts)
}

func (m *mockAppConfigurationAPI) serve(w http.ResponseWriter, r *http.Request) {
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
	if !strings.HasPrefix(r.URL.Path, "/app-srv/apps") {
		fail(http.StatusNotFound, "not found")
		return
	}
	clientID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/app-srv/apps"), "/")
	var body map[string]any
	if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
		_ = json.Unmarshal(raw, &body)
	}
	validate := func(app map[string]any) bool {
		if payload, _ := app["additional_access_token_payload"].([]any); len(payload) > 0 {
			if ct := app["client_type"]; ct == "NON_INTERACTIVE" || ct == "SAML_SP" {
				fail(http.StatusBadRequest, "additional_access_token_payload is not allowed for client_type "+ct.(string))
				return false
			}
		}
		return true
	}

	switch r.Method {
	case http.MethodPost:
		if !validate(body) {
			return
		}
		id := fmt.Sprintf("client-%d", len(m.apps)+1)
		body["_id"], body["client_id"], body["created_time"] = id, id, "2026-01-01T00:00:00Z"
		body["signing_key_config"] = map[string]any{"active_kid": "kid-1"}
		m.apps[id] = body
		respond(http.StatusOK, body)
	case http.MethodGet:
		app, ok := m.apps[clientID]
		if !ok {
			fail(http.StatusNotFound, "app not found")
			return
		}
		respond(http.StatusOK, app)
	case http.MethodPut:
		saved, ok := m.apps[clientID]
		if !ok {
			fail(http.StatusNotFound, "app not found")
			return
		}
		m.puts = append(m.puts, body)
		for _, key := range serverOwnedKeys {
			body[key] = saved[key]
		}
		if !validate(body) {
			return
		}
		m.apps[clientID] = body
		respond(http.StatusOK, body)
	case http.MethodDelete:
		delete(m.apps, clientID)
		respond(http.StatusOK, true)
	}
}

func mockAppConfigurationConfig(baseURL, clientType, scopes, payload string) string {
	return fmt.Sprintf(`
provider "cidaas" {
  base_url       = %q
  client_id      = "mock"
  client_secret  = "mock"
  cidaas_version = "4.x"
}

resource "cidaas_app_configuration" "test" {
  client_name = "payload-app"
  client_type = %q
  %s
  ownership_details = {
    company_name    = "Co"
    company_address = "Addr"
    company_website = "https://example.com"
  }
  scopes = {
    allowed_scopes = %s
  }
}
`, baseURL, clientType, payload, scopes)
}

func TestAppConfigurationMock_AdditionalAccessTokenPayload(t *testing.T) {
	m, server := newMockAppConfigurationAPI(t)
	name := "cidaas_app_configuration.test"
	config := func(scopes, payload string) string {
		return mockAppConfigurationConfig(server.URL, "SINGLE_PAGE", scopes, payload)
	}
	expectStored := func(want ...string) resource.TestCheckFunc {
		return func(_ *terraform.State) error {
			if got := m.storedPayload("client-1"); !slices.Equal(got, want) {
				return fmt.Errorf("stored additional_access_token_payload = %v, want %v", got, want)
			}
			return nil
		}
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config(`["openid"]`, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "additional_access_token_payload.#", "0"),
					expectStored(),
				),
			},
			{
				// Value set in the Admin UI is kept by an unrelated update when the attribute is omitted.
				PreConfig: func() { m.setPayload("client-1", "email") },
				Config:    config(`["openid", "profile"]`, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "scopes.allowed_scopes.#", "2"),
					resource.TestCheckResourceAttr(name, "additional_access_token_payload.#", "1"),
					resource.TestCheckTypeSetElemAttr(name, "additional_access_token_payload.*", "email"),
					expectStored("email"),
					func(_ *terraform.State) error {
						if m.putCount() != 1 {
							return fmt.Errorf("expected exactly one PUT, got %d", m.putCount())
						}
						return nil
					},
				),
			},
			{
				Config: config(`["openid", "profile"]`, `additional_access_token_payload = ["roles", "email"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "additional_access_token_payload.#", "2"),
					resource.TestCheckTypeSetElemAttr(name, "additional_access_token_payload.*", "roles"),
					resource.TestCheckTypeSetElemAttr(name, "additional_access_token_payload.*", "email"),
					expectStored("email", "roles"),
				),
			},
			{
				ResourceName:                         name,
				ImportState:                          true,
				ImportStateId:                        "client-1",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "client_id",
			},
			{
				Config: config(`["openid", "profile"]`, `additional_access_token_payload = []`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "additional_access_token_payload.#", "0"),
					expectStored(),
				),
			},
		},
	})
}

func TestAppConfigurationMock_AdditionalAccessTokenPayloadRejectedAtPlan(t *testing.T) {
	m, server := newMockAppConfigurationAPI(t)
	for _, clientType := range []string{"NON_INTERACTIVE", "SAML_SP"} {
		t.Run(clientType, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      mockAppConfigurationConfig(server.URL, clientType, `["openid"]`, `additional_access_token_payload = ["email"]`),
						PlanOnly:    true,
						ExpectError: regexp.MustCompile(`additional_access_token_payload\s+not\s+allowed`),
					},
					{
						// An explicit empty set is the documented way to clear the value, so it stays valid.
						Config:   mockAppConfigurationConfig(server.URL, clientType, `["openid"]`, `additional_access_token_payload = []`),
						PlanOnly: true,
						// PlanOnly with a new resource always shows a diff.
						ExpectNonEmptyPlan: true,
					},
				},
			})
		})
	}
	if len(m.apps) != 0 {
		t.Fatalf("plan-time validation must not reach the API, apps created: %d", len(m.apps))
	}
}
