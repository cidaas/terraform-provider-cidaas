package webhook_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccEventRetention_Basic needs a tenant with event retention (custom_reporting) active and
// CIDAAS_EVENT_RETENTION_EVENT_ID set to a configurable event type whose bounds allow
// CIDAAS_EVENT_RETENTION_DAYS (default 60) and CIDAAS_EVENT_RETENTION_DAYS + 1.
func TestAccEventRetention_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}
	acctest.SkipIfV3(t)
	eventID := os.Getenv("CIDAAS_EVENT_RETENTION_EVENT_ID")
	if eventID == "" {
		t.Skip("CIDAAS_EVENT_RETENTION_EVENT_ID not set; skipping event retention acceptance test")
	}
	days := 60
	if v := os.Getenv("CIDAAS_EVENT_RETENTION_DAYS"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			t.Fatalf("CIDAAS_EVENT_RETENTION_DAYS: %v", err)
		}
		days = parsed
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEventRetentionConfig(eventID, days),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(retentionResource, "id", "event_retention"),
					resource.TestCheckResourceAttr(retentionResource, "retention."+eventID, strconv.Itoa(days)),
					resource.TestCheckResourceAttr(retentionResource, "catalog."+eventID+".days_to_live", strconv.Itoa(days)),
					resource.TestCheckResourceAttrSet(retentionResource, "catalog."+eventID+".default_days_to_live"),
				),
			},
			{
				Config: testAccEventRetentionConfig(eventID, days+1),
				Check:  resource.TestCheckResourceAttr(retentionResource, "retention."+eventID, strconv.Itoa(days+1)),
			},
		},
	})
}

func testAccEventRetentionConfig(eventID string, days int) string {
	return fmt.Sprintf(`
%s
resource "cidaas_event_retention" "this" {
  retention = {
    %s = %d
  }
}
`, acctest.ProviderConfig(), eventID, days)
}
