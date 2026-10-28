package webhook_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	retentionResource = "cidaas_event_retention.this"
	retentionPath     = "/webhook-srv/eventdescriptions/retention"
)

// mockRetentionAPI mimics webhook-srv event description retention: GET /retention lists only types with
// retention bounds (204 when none), PATCH applies entries one by one with the same 400/403/404 checks as
// the service, GET /eventdescriptions/{id} knows immutable types too. active=false answers 403.
type mockRetentionAPI struct {
	mu      sync.Mutex
	active  bool
	eds     map[string]map[string]any
	patches []map[string]int
}

func newMockRetentionAPI(t *testing.T, active bool) (*mockRetentionAPI, *httptest.Server) {
	t.Helper()
	m := &mockRetentionAPI{active: active, eds: map[string]map[string]any{
		"USER_CREATED": {
			"_id": "USER_CREATED", "objectType": "USER", "goodForChangelog": true, "daysToLive": 90,
			"retention": map[string]any{"defaultDaysToLive": 90, "minDaysToLive": 30, "maxDaysToLive": 365},
		},
		"LOGIN_SUCCESS": {
			"_id": "LOGIN_SUCCESS", "objectType": "USER", "goodForUserActivity": true, "daysToLive": 30,
			"retention": map[string]any{"defaultDaysToLive": 30, "minDaysToLive": 7, "maxDaysToLive": 180},
		},
		"PROVIDER_CREATED": {"_id": "PROVIDER_CREATED", "objectType": "PROVIDER", "goodForChangelog": true, "daysToLive": 365},
	}}
	server := httptest.NewServer(http.HandlerFunc(m.serve))
	t.Cleanup(server.Close)
	return m, server
}

func (m *mockRetentionAPI) days(id string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.eds[id]["daysToLive"].(int)
}

func (m *mockRetentionAPI) setDays(id string, days int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eds[id]["daysToLive"] = days
}

func (m *mockRetentionAPI) patchCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.patches)
}

func (m *mockRetentionAPI) serve(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	respond := func(status int, data any) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "status": status, "data": data})
	}
	fail := func(status int, msg string) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "status": status, "error": map[string]any{"error": msg}})
	}

	switch {
	case r.URL.Path == "/token-srv/token":
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "mock-token"})
	case r.URL.Path == retentionPath && !m.active:
		fail(http.StatusForbidden, "retention is not active")
	case r.URL.Path == retentionPath && r.Method == http.MethodGet:
		ids := make([]string, 0, len(m.eds))
		for id, ed := range m.eds {
			if ed["retention"] != nil {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		sort.Strings(ids)
		out := make([]any, 0, len(ids))
		for _, id := range ids {
			out = append(out, m.eds[id])
		}
		respond(http.StatusOK, out)
	case r.URL.Path == retentionPath && r.Method == http.MethodPatch:
		var body map[string]int
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			fail(http.StatusBadRequest, "invalid body")
			return
		}
		m.patches = append(m.patches, body)
		if len(body) == 0 {
			fail(http.StatusNotFound, "no retention values to update")
			return
		}
		var updated []any
		for id, days := range body {
			ed, ok := m.eds[id]
			if !ok {
				fail(http.StatusBadRequest, "unknown event type: "+id)
				return
			}
			bounds, ok := ed["retention"].(map[string]any)
			if !ok {
				fail(http.StatusBadRequest, fmt.Sprintf("retention for event type %s cannot be modified (immutable)", id))
				return
			}
			if days < bounds["minDaysToLive"].(int) || days > bounds["maxDaysToLive"].(int) {
				fail(http.StatusBadRequest, fmt.Sprintf("daysToLive for %s out of range", id))
				return
			}
			ed["daysToLive"] = days
			updated = append(updated, map[string]any{"eventId": id, "daysToLive": days})
		}
		respond(http.StatusOK, updated)
	case strings.HasPrefix(r.URL.Path, "/webhook-srv/eventdescriptions/") && r.Method == http.MethodGet:
		if ed, ok := m.eds[strings.TrimPrefix(r.URL.Path, "/webhook-srv/eventdescriptions/")]; ok {
			respond(http.StatusOK, ed)
			return
		}
		fail(http.StatusNotFound, "resource not found")
	default:
		fail(http.StatusNotFound, "not found")
	}
}

// retentionErr matches msg in a Terraform error even when the CLI wraps it across lines.
func retentionErr(msg string) *regexp.Regexp {
	return regexp.MustCompile(strings.ReplaceAll(regexp.QuoteMeta(msg), " ", `\s+`))
}

func retentionConfig(baseURL, retention string) string {
	return fmt.Sprintf(`
provider "cidaas" {
  base_url       = %q
  client_id      = "mock"
  client_secret  = "mock"
  cidaas_version = "4.x"
}

resource "cidaas_event_retention" "this" {
  retention = {
%s
  }
}
`, baseURL, retention)
}

func TestEventRetentionMock_Lifecycle(t *testing.T) {
	m, server := newMockRetentionAPI(t, true)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			if got := m.days("LOGIN_SUCCESS"); got != 30 {
				return fmt.Errorf("destroy must restore LOGIN_SUCCESS to default 30, got %d", got)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: retentionConfig(server.URL, `    USER_CREATED = 120`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(retentionResource, "id", "event_retention"),
					resource.TestCheckResourceAttr(retentionResource, "retention.%", "1"),
					resource.TestCheckResourceAttr(retentionResource, "retention.USER_CREATED", "120"),
					resource.TestCheckResourceAttr(retentionResource, "catalog.%", "2"),
					resource.TestCheckResourceAttr(retentionResource, "catalog.USER_CREATED.days_to_live", "120"),
					resource.TestCheckResourceAttr(retentionResource, "catalog.USER_CREATED.object_type", "USER"),
					resource.TestCheckResourceAttr(retentionResource, "catalog.USER_CREATED.default_days_to_live", "90"),
					resource.TestCheckResourceAttr(retentionResource, "catalog.USER_CREATED.min_days_to_live", "30"),
					resource.TestCheckResourceAttr(retentionResource, "catalog.USER_CREATED.max_days_to_live", "365"),
					resource.TestCheckResourceAttr(retentionResource, "catalog.LOGIN_SUCCESS.days_to_live", "30"),
					func(_ *terraform.State) error {
						if got := m.days("USER_CREATED"); got != 120 {
							return fmt.Errorf("remote USER_CREATED=%d, want 120", got)
						}
						return nil
					},
				),
			},
			{
				Config: retentionConfig(server.URL, "    USER_CREATED  = 200\n    LOGIN_SUCCESS = 60"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(retentionResource, "retention.USER_CREATED", "200"),
					resource.TestCheckResourceAttr(retentionResource, "retention.LOGIN_SUCCESS", "60"),
					resource.TestCheckResourceAttr(retentionResource, "catalog.LOGIN_SUCCESS.days_to_live", "60"),
				),
			},
			{
				// Removing an entry restores its default on the tenant.
				Config: retentionConfig(server.URL, `    LOGIN_SUCCESS = 60`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(retentionResource, "retention.%", "1"),
					resource.TestCheckNoResourceAttr(retentionResource, "retention.USER_CREATED"),
					resource.TestCheckResourceAttr(retentionResource, "catalog.USER_CREATED.days_to_live", "90"),
					func(_ *terraform.State) error {
						if got := m.days("USER_CREATED"); got != 90 {
							return fmt.Errorf("remote USER_CREATED=%d, want default 90", got)
						}
						return nil
					},
				),
			},
			{
				// Import adopts every configurable event type.
				ResourceName:  retentionResource,
				ImportState:   true,
				ImportStateId: "event_retention",
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 imported instance, got %d", len(states))
					}
					attrs := states[0].Attributes
					if attrs["id"] != "event_retention" || attrs["retention.%"] != "2" ||
						attrs["retention.USER_CREATED"] != "90" || attrs["retention.LOGIN_SUCCESS"] != "60" {
						return fmt.Errorf("unexpected imported attributes: %v", attrs)
					}
					return nil
				},
			},
			{
				ResourceName:  retentionResource,
				ImportState:   true,
				ImportStateId: "something-else",
				ExpectError:   retentionErr(`import it with ID "event_retention"`),
			},
		},
	})
}

func TestEventRetentionMock_DriftAndNoopPatch(t *testing.T) {
	m, server := newMockRetentionAPI(t, true)
	config := retentionConfig(server.URL, `    USER_CREATED = 120`)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config},
			{
				// Changed outside Terraform (e.g. Trust Desk UI): refresh detects drift.
				PreConfig:          func() { m.setDays("USER_CREATED", 45) },
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: config,
				Check: func(_ *terraform.State) error {
					if got := m.days("USER_CREATED"); got != 120 {
						return fmt.Errorf("drift not corrected: USER_CREATED=%d", got)
					}
					return nil
				},
			},
			{
				// Re-applying an unchanged config sends no PATCH (webhook-srv rejects an empty one with 404).
				PreConfig: func() { m.mu.Lock(); m.patches = nil; m.mu.Unlock() },
				Config:    config,
				Check: func(_ *terraform.State) error {
					if n := m.patchCount(); n != 0 {
						return fmt.Errorf("expected no PATCH for unchanged config, got %d", n)
					}
					return nil
				},
			},
		},
	})
}

func TestEventRetentionMock_ValidationErrors(t *testing.T) {
	m, server := newMockRetentionAPI(t, true)
	cases := []struct {
		name      string
		retention string
		err       string
	}{
		{"below_min", `    USER_CREATED = 10`, "days_to_live for USER_CREATED must be between 30 and 365 (default 90), got 10."},
		{"above_max", `    LOGIN_SUCCESS = 181`, "days_to_live for LOGIN_SUCCESS must be between 7 and 180 (default 30), got 181."},
		{"immutable", `    PROVIDER_CREATED = 30`, "retention for event type PROVIDER_CREATED cannot be modified (immutable)."},
		{"unknown", `    NOT_A_REAL_EVENT = 30`, "unknown event type: NOT_A_REAL_EVENT."},
		{"empty_map", ``, "retention map must contain at least 1 elements"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config:      retentionConfig(server.URL, tc.retention),
					ExpectError: retentionErr(tc.err),
				}},
			})
		})
	}
	// Plan-time validation must stop before anything is half-applied.
	if n := m.patchCount(); n != 0 {
		t.Fatalf("expected no PATCH for invalid configs, got %d", n)
	}
}

func TestEventRetentionMock_Inactive(t *testing.T) {
	_, server := newMockRetentionAPI(t, false)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config:      retentionConfig(server.URL, `    USER_CREATED = 120`),
			ExpectError: retentionErr("Event retention is not active"),
		}},
	})
}

func TestEventRetentionMock_RequiresV4(t *testing.T) {
	_, server := newMockRetentionAPI(t, true)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config:      strings.Replace(retentionConfig(server.URL, `    USER_CREATED = 120`), `"4.x"`, `"3.x"`, 1),
			ExpectError: retentionErr("only supported on cidaas v4.x"),
		}},
	})
}
