package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestToLower_PlanModifyString(t *testing.T) {
	t.Parallel()

	modifier := ToLower{}
	req := planmodifier.StringRequest{
		ConfigValue: types.StringValue("Marketing Opt-In"),
		PlanValue:   types.StringValue("Marketing Opt-In"),
	}
	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	modifier.PlanModifyString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if got := resp.PlanValue.ValueString(); got != "marketing opt-in" {
		t.Fatalf("expected lowercase plan value, got %q", got)
	}
}

func TestCaseInsensitiveUniqueIdentifier_AllowsCaseOnlyChange(t *testing.T) {
	t.Parallel()

	modifier := CaseInsensitiveUniqueIdentifier{}
	req := planmodifier.StringRequest{
		StateValue:  types.StringValue("marketing opt-in"),
		ConfigValue: types.StringValue("Marketing Opt-In"),
		PlanValue:   types.StringValue("marketing opt-in"),
	}
	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}

	modifier.PlanModifyString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected case-only change to be allowed, got: %v", resp.Diagnostics)
	}
}

func TestCaseInsensitiveUniqueIdentifier_RejectsRealChange(t *testing.T) {
	t.Parallel()

	modifier := CaseInsensitiveUniqueIdentifier{}
	req := planmodifier.StringRequest{
		StateValue:  types.StringValue("marketing opt-in"),
		ConfigValue: types.StringValue("privacy notices"),
		PlanValue:   types.StringValue("privacy notices"),
	}
	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}

	modifier.PlanModifyString(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error when immutable value changes")
	}
}
