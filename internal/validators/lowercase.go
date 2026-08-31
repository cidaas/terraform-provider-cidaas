package validators

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.String = ToLower{}

// ToLower normalizes a string plan value to lowercase so Terraform matches
// Cidaas APIs that store identifiers in lowercase (e.g. consent group_name).
type ToLower struct{}

func (v ToLower) Description(_ context.Context) string {
	return "Normalizes the attribute value to lowercase."
}

func (v ToLower) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ToLower) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	resp.PlanValue = types.StringValue(strings.ToLower(req.ConfigValue.ValueString()))
}
