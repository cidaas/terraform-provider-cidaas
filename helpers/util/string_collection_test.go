package util //nolint:revive

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestListToStringsSkipsUnknownAndNull(t *testing.T) {
	t.Parallel()
	l := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("a"),
		types.StringUnknown(),
		types.StringValue("b"),
		types.StringNull(),
	})
	got := ListToStrings(l)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("got %v", got)
	}
}

func TestSetToStringsSkipsUnknownAndNull(t *testing.T) {
	t.Parallel()
	s := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("a"),
		types.StringUnknown(),
		types.StringValue("b"),
		types.StringNull(),
	})
	got := SetToStrings(s)
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
	seen := map[string]bool{}
	for _, v := range got {
		seen[v] = true
	}
	if !seen["a"] || !seen["b"] {
		t.Fatalf("got %v", got)
	}
}

func TestListContainsUnknown(t *testing.T) {
	t.Parallel()
	withUnknown := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("a"),
		types.StringUnknown(),
	})
	if !ListContainsUnknown(withUnknown) {
		t.Fatal("expected unknown")
	}
	known := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")})
	if ListContainsUnknown(known) {
		t.Fatal("did not expect unknown")
	}
}
