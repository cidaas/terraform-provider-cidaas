package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var (
	_ planmodifier.String  = UniqueIdentifier{}
	_ planmodifier.String  = CaseInsensitiveUniqueIdentifier{}
	_ planmodifier.Float64 = ImmutableInt64Identifier{}
	_ planmodifier.Set     = ImmutableSetIdentifier{}
)

type (
	UniqueIdentifier                struct{}
	CaseInsensitiveUniqueIdentifier struct{}
	ImmutableInt64Identifier        struct{}
	ImmutableSetIdentifier          struct{}
)

func (v UniqueIdentifier) Description(_ context.Context) string {
	return "Checks if a immutable attribute has been changed."
}

func (v UniqueIdentifier) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v UniqueIdentifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	if !req.ConfigValue.Equal(req.StateValue) {
		resp.Diagnostics.AddError("Unexpected Resource Configuration",
			fmt.Sprintf("Attribute '%s' can't be modified. Existing value %s, got %s", req.Path.String(), req.StateValue.ValueString(), req.ConfigValue.ValueString()))
	}
}

func (v CaseInsensitiveUniqueIdentifier) Description(_ context.Context) string {
	return "Checks if an immutable attribute has been changed, ignoring letter case."
}

func (v CaseInsensitiveUniqueIdentifier) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v CaseInsensitiveUniqueIdentifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	if !strings.EqualFold(req.ConfigValue.ValueString(), req.StateValue.ValueString()) {
		resp.Diagnostics.AddError("Unexpected Resource Configuration",
			fmt.Sprintf("Attribute '%s' can't be modified. Existing value %s, got %s", req.Path.String(), req.StateValue.ValueString(), req.ConfigValue.ValueString()))
	}
}

func (v ImmutableInt64Identifier) Description(_ context.Context) string {
	return "Checks if an immutable int64 attribute has been changed."
}

func (v ImmutableInt64Identifier) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ImmutableInt64Identifier) PlanModifyFloat64(_ context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	if !req.ConfigValue.Equal(req.StateValue) {
		resp.Diagnostics.AddError("Unexpected Resource Configuration",
			fmt.Sprintf("Attribute '%s' can't be modified. Existing value %v, got %v", req.Path.String(), req.StateValue.ValueFloat64(), req.ConfigValue.ValueFloat64()))
	}
}

func (v ImmutableSetIdentifier) Description(_ context.Context) string {
	return "Checks if an immutable set attribute has been changed."
}

func (v ImmutableSetIdentifier) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ImmutableSetIdentifier) PlanModifySet(_ context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	if !req.ConfigValue.Equal(req.StateValue) {
		resp.Diagnostics.AddError("Unexpected Resource Configuration",
			fmt.Sprintf("Attribute '%s' can't be modified. Existing value %v, got %v", req.Path.String(), req.StateValue.Elements(), req.ConfigValue.Elements()))
	}
}
