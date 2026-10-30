package notification_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const notificationTemplatesPath = "/notifications-srv/templates"

// seededVerifyUserDE mirrors a seeded notification-srv row with a language-only locale (ticket 5545).
const seededVerifyUserDE = "default:VERIFY_USER:email:de:CODE:MULTIFACTOR_AUTHENTICATION:EMAIL"

// mockNotificationTemplateAPI mimics notification-srv templates: rows are keyed by the synthetic document id,
// POST answers 409 for an existing id, GET/PUT/DELETE answer 404 for unknown ids. Locales are stored as sent.
type mockNotificationTemplateAPI struct {
	mu        sync.Mutex
	templates map[string]cidaas.NotificationsSrvTemplateModel
}

func newMockNotificationTemplateAPI(t *testing.T) (*mockNotificationTemplateAPI, *httptest.Server) {
	t.Helper()
	m := &mockNotificationTemplateAPI{templates: map[string]cidaas.NotificationsSrvTemplateModel{
		seededVerifyUserDE: {
			ID: seededVerifyUserDE, GroupID: "default", TemplateKey: "VERIFY_USER", CommunicationMethod: "email",
			Locale: "de", ProcessingType: "CODE", UsageType: "MULTIFACTOR_AUTHENTICATION", VerificationType: "EMAIL",
			Owner: "admin", MessageFormat: "html", Subject: "Ihr Code", Content: "<p>{{code}}</p>",
			Description: "Seeded VERIFY_USER email template for locale de", Enabled: true,
		},
	}}
	server := httptest.NewServer(http.HandlerFunc(m.serve))
	t.Cleanup(server.Close)
	return m, server
}

func (m *mockNotificationTemplateAPI) has(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.templates[id]
	return ok
}

func (m *mockNotificationTemplateAPI) serve(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	respond := func(status int, data any) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "status": status, "data": data})
	}
	fail := func(status int, msg string) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "status": status, "errorMsg": msg})
	}

	id := strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, notificationTemplatesPath), "/")
	switch {
	case r.URL.Path == "/token-srv/token":
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "mock-token"})
	case !strings.HasPrefix(r.URL.Path, notificationTemplatesPath):
		fail(http.StatusNotFound, "not found")
	case id == "" && r.Method == http.MethodPost:
		var body cidaas.NotificationsSrvTemplateModel
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			fail(http.StatusBadRequest, "invalid body")
			return
		}
		body.ID = cidaas.SyntheticTemplateDocumentID(body)
		if _, ok := m.templates[body.ID]; ok {
			fail(http.StatusConflict, "template already exists")
			return
		}
		m.templates[body.ID] = body
		respond(http.StatusCreated, body)
	case r.Method == http.MethodGet:
		tpl, ok := m.templates[id]
		if !ok {
			fail(http.StatusNotFound, "template not found")
			return
		}
		respond(http.StatusOK, tpl)
	case r.Method == http.MethodPut:
		if _, ok := m.templates[id]; !ok {
			fail(http.StatusNotFound, "template not found")
			return
		}
		var body cidaas.NotificationsSrvTemplateModel
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			fail(http.StatusBadRequest, "invalid body")
			return
		}
		body.ID = id
		m.templates[id] = body
		respond(http.StatusOK, body)
	case r.Method == http.MethodDelete:
		if _, ok := m.templates[id]; !ok {
			fail(http.StatusNotFound, "template not found")
			return
		}
		delete(m.templates, id)
		respond(http.StatusOK, map[string]bool{"deleted": true})
	default:
		fail(http.StatusMethodNotAllowed, "method not allowed")
	}
}

func notificationTemplateProvider(baseURL string) string {
	return fmt.Sprintf(`
provider "cidaas" {
  base_url       = %q
  client_id      = "mock"
  client_secret  = "mock"
  cidaas_version = "4.x"
}
`, baseURL)
}

func notificationTemplateConfig(name, locale string) string {
	return fmt.Sprintf(`
resource "cidaas_notification_template" %q {
  group_id             = "developer"
  template_key         = "TF_LOCALE_5545"
  communication_method = "email"
  locale               = %q
  message_format       = "html"
  owner                = "client"
  description          = "Terraform mock template for locale %s"
  subject              = "Locale %s"
  content              = "<p>locale %s</p>"
}
`, name, locale, locale, locale, locale)
}

// Language-only (de, en) and language-REGION (de-DE) locales are created, re-planned without drift and imported.
func TestNotificationTemplateMock_LocaleFormats(t *testing.T) {
	m, server := newMockNotificationTemplateAPI(t)
	config := notificationTemplateProvider(server.URL) +
		notificationTemplateConfig("de", "de") +
		notificationTemplateConfig("en", "en") +
		notificationTemplateConfig("de_de", "de-DE")

	ids := map[string]string{
		"de":    "developer:TF_LOCALE_5545:email:de",
		"en":    "developer:TF_LOCALE_5545:email:en",
		"de_de": "developer:TF_LOCALE_5545:email:de-de",
	}
	steps := []resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("cidaas_notification_template.de", "locale", "de"),
			resource.TestCheckResourceAttr("cidaas_notification_template.de", "id", ids["de"]),
			resource.TestCheckResourceAttr("cidaas_notification_template.en", "locale", "en"),
			resource.TestCheckResourceAttr("cidaas_notification_template.en", "id", ids["en"]),
			resource.TestCheckResourceAttr("cidaas_notification_template.de_de", "locale", "de-DE"),
			resource.TestCheckResourceAttr("cidaas_notification_template.de_de", "id", ids["de_de"]),
		),
	}, {
		Config: config,
		ConfigPlanChecks: resource.ConfigPlanChecks{
			PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
		},
	}}
	for name, id := range ids {
		steps = append(steps, resource.TestStep{
			Config:            config,
			ResourceName:      "cidaas_notification_template." + name,
			ImportState:       true,
			ImportStateId:     id,
			ImportStateVerify: true,
		})
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			for _, id := range ids {
				if m.has(id) {
					return fmt.Errorf("template %s still exists after destroy", id)
				}
			}
			return nil
		},
		Steps: steps,
	})
}

// Customer scenario (ticket 5545): an import block adopts a seeded template whose locale is "de" without changes.
func TestNotificationTemplateMock_ImportSeededShortLocale(t *testing.T) {
	_, server := newMockNotificationTemplateAPI(t)
	config := notificationTemplateProvider(server.URL) + fmt.Sprintf(`
import {
  to = cidaas_notification_template.verify_user_email_de
  id = %q
}

resource "cidaas_notification_template" "verify_user_email_de" {
  group_id             = "default"
  template_key         = "VERIFY_USER"
  communication_method = "email"
  locale               = "de"
  message_format       = "html"
  processing_type      = "CODE"
  usage_type           = "MULTIFACTOR_AUTHENTICATION"
  verification_type    = "EMAIL"
  owner                = "admin"
  description          = "Seeded VERIFY_USER email template for locale de"
  subject              = "Ihr Code"
  content              = "<p>{{code}}</p>"
}
`, seededVerifyUserDE)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: config,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("cidaas_notification_template.verify_user_email_de", plancheck.ResourceActionNoop),
				},
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("cidaas_notification_template.verify_user_email_de", "id", seededVerifyUserDE),
				resource.TestCheckResourceAttr("cidaas_notification_template.verify_user_email_de", "locale", "de"),
			),
		}, {
			Config: config,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
			},
		}},
	})
}

func TestNotificationTemplateMock_InvalidLocale(t *testing.T) {
	_, server := newMockNotificationTemplateAPI(t)
	for _, locale := range []string{"de_DE", "deu", "DE", "en-us", "xx-YY", ""} {
		t.Run(locale, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config:      notificationTemplateProvider(server.URL) + notificationTemplateConfig("bad", locale),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
				}},
			})
		})
	}
}
