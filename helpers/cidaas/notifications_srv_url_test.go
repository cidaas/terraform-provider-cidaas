package cidaas

import "testing"

func TestSegmentNotificationsURL(t *testing.T) {
	cfg := ClientConfig{BaseURL: "https://example.com", NotificationsContextPath: "notifications-srv"}
	u := SegmentNotificationsURL(cfg, "templates", "x")
	want := "https://example.com/notifications-srv/templates/x"
	if u != want {
		t.Fatalf("got %q want %q", u, want)
	}
	cfg2 := ClientConfig{BaseURL: "https://example.com/", NotificationsContextPath: ""}
	u2 := SegmentNotificationsURL(cfg2, "templatetypes")
	if u2 != "https://example.com/notifications-srv/templatetypes" {
		t.Fatalf("default ctx: got %q", u2)
	}
}
