package datasources_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	acctest "github.com/Cidaas/terraform-provider-cidaas/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// newMockNotificationGraphAPI mimics notification-srv graph search: rows answer 200 with the data envelope,
// no rows answer 204 No Content with an empty body (as FindTemplates / FindTemplateGroups do).
func newMockNotificationGraphAPI(t *testing.T, rows map[string][]map[string]any) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token-srv/token" {
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "mock-token"})
			return
		}
		data, ok := rows[r.URL.Path]
		if !ok || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if len(data) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "status": http.StatusOK, "data": data})
	}))
	t.Cleanup(server.Close)
	return server
}

func notificationGraphConfig(baseURL, dataSource string) string {
	return fmt.Sprintf(`
provider "cidaas" {
  base_url       = %q
  client_id      = "mock"
  client_secret  = "mock"
  cidaas_version = "4.x"
}

data %q "this" {
  graph_filter = jsonencode({ size = 10 })
}
`, baseURL, dataSource)
}

func TestNotificationTemplatesDataSourceMock_NoMatch(t *testing.T) {
	server := newMockNotificationGraphAPI(t, map[string][]map[string]any{"/notifications-srv/graph/templates": nil})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: notificationGraphConfig(server.URL, "cidaas_notification_templates"),
			Check:  resource.TestCheckResourceAttr("data.cidaas_notification_templates.this", "templates.#", "0"),
		}},
	})
}

func TestNotificationTemplatesDataSourceMock_Rows(t *testing.T) {
	server := newMockNotificationGraphAPI(t, map[string][]map[string]any{"/notifications-srv/graph/templates": {
		{"_id": "default:WELCOME_USER:sms:de", "groupId": "default", "templateKey": "WELCOME_USER", "communicationMethod": "sms", "locale": "de", "enabled": true},
		{"_id": "default:WELCOME_USER:email:de", "groupId": "default", "templateKey": "WELCOME_USER", "communicationMethod": "email", "locale": "de", "enabled": true},
	}})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: notificationGraphConfig(server.URL, "cidaas_notification_templates"),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.cidaas_notification_templates.this", "templates.#", "2"),
				resource.TestCheckResourceAttr("data.cidaas_notification_templates.this", "templates.0.id", "default:WELCOME_USER:email:de"),
			),
		}},
	})
}

func TestNotificationTemplateGroupsDataSourceMock_NoMatch(t *testing.T) {
	server := newMockNotificationGraphAPI(t, map[string][]map[string]any{"/notifications-srv/graph/templategroups": nil})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: notificationGraphConfig(server.URL, "cidaas_notification_template_groups"),
			Check:  resource.TestCheckResourceAttr("data.cidaas_notification_template_groups.this", "groups.#", "0"),
		}},
	})
}

func TestNotificationTemplateGroupsDataSourceMock_Rows(t *testing.T) {
	server := newMockNotificationGraphAPI(t, map[string][]map[string]any{"/notifications-srv/graph/templategroups": {
		{"_id": "default", "tgType": "cidaas", "description": "Default template group", "defaultLocale": "en"},
	}})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: notificationGraphConfig(server.URL, "cidaas_notification_template_groups"),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.cidaas_notification_template_groups.this", "groups.#", "1"),
				resource.TestCheckResourceAttr("data.cidaas_notification_template_groups.this", "groups.0.id", "default"),
			),
		}},
	})
}
