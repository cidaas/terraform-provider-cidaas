package registrationfield

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRegistrationFieldOrderChangeRequested(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		plannedOrder types.Int64
		stateOrder   types.Int64
		wantOK       bool
		wantCurrent  int64
		wantPrevious int64
	}{
		{
			name:         "explicit change",
			plannedOrder: types.Int64Value(10),
			stateOrder:   types.Int64Value(49),
			wantOK:       true,
			wantCurrent:  10,
			wantPrevious: 49,
		},
		{
			name:         "unchanged",
			plannedOrder: types.Int64Value(49),
			stateOrder:   types.Int64Value(49),
			wantOK:       false,
		},
		{
			name:         "planned null",
			plannedOrder: types.Int64Null(),
			stateOrder:   types.Int64Value(49),
			wantOK:       false,
		},
		{
			name:         "state null",
			plannedOrder: types.Int64Value(10),
			stateOrder:   types.Int64Null(),
			wantOK:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			planned := RegFieldConfig{Order: tt.plannedOrder}
			state := RegFieldConfig{Order: tt.stateOrder}

			current, previous, ok := registrationFieldOrderChangeRequested(planned, state)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if current != tt.wantCurrent || previous != tt.wantPrevious {
				t.Fatalf("current=%d previous=%d, want current=%d previous=%d", current, previous, tt.wantCurrent, tt.wantPrevious)
			}
		})
	}
}

func TestRegistrationFieldParentGroupID(t *testing.T) {
	t.Parallel()

	empty := RegFieldConfig{ParentGroupID: types.StringNull()}
	if got := registrationFieldParentGroupID(empty); got != registrationFieldDefaultParentGroupID {
		t.Fatalf("got %q, want DEFAULT", got)
	}

	custom := RegFieldConfig{ParentGroupID: types.StringValue("my_group")}
	if got := registrationFieldParentGroupID(custom); got != "my_group" {
		t.Fatalf("got %q, want my_group", got)
	}
}

func TestRegistrationFieldOrderMatchesPlan(t *testing.T) {
	t.Parallel()

	planned := RegFieldConfig{Order: types.Int64Value(10)}
	if !registrationFieldOrderMatchesPlan(planned, 10) {
		t.Fatal("expected match")
	}
	if registrationFieldOrderMatchesPlan(planned, 49) {
		t.Fatal("expected mismatch")
	}

	omitted := RegFieldConfig{Order: types.Int64Null()}
	if !registrationFieldOrderMatchesPlan(omitted, 49) {
		t.Fatal("expected match when order omitted in plan")
	}
}
