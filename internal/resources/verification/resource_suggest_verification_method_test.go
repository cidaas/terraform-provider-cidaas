package verification

import "testing"

func TestMethodOverlap(t *testing.T) {
	t.Parallel()
	got := methodOverlap([]string{"EMAIL", "SMS"}, []string{"TOTP", "EMAIL"})
	if len(got) != 1 || got[0] != "EMAIL" {
		t.Fatalf("%v", got)
	}
	if got := methodOverlap([]string{"EMAIL"}, []string{"SMS"}); len(got) != 0 {
		t.Fatalf("%v", got)
	}
}
