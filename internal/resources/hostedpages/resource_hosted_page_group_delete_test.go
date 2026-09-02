package hostedpages

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestHostedPageGroup_DeleteSystemGroup_Warning(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &hostedPageGroupResource{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema: %v", schemaResp.Diagnostics)
	}

	tests := []struct {
		name      string
		groupName string
	}{
		{name: "DEFAULT uppercase", groupName: "DEFAULT"},
		{name: "admin lowercase", groupName: "admin"},
		{name: "default lowercase", groupName: "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			state := tfsdk.State{Schema: schemaResp.Schema}
			hostedPages, setDiags := types.SetValue(
				types.ObjectType{AttrTypes: hostedPageAttrTypes()},
				[]attr.Value{
					types.ObjectValueMust(hostedPageAttrTypes(), map[string]attr.Value{
						"hosted_page_id": types.StringValue("login"),
						"locale":         types.StringValue("en"),
						"url":            types.StringValue("https://example.com/login"),
						"content":        types.StringNull(),
					}),
				},
			)
			if setDiags.HasError() {
				t.Fatalf("hosted pages set: %v", setDiags)
			}
			setDiags = state.Set(ctx, &hostedPageGroupModel{
				Name:          types.StringValue(tt.groupName),
				DefaultLocale: types.StringValue("en"),
				HostedPages:   hostedPages,
			})
			if setDiags.HasError() {
				t.Fatalf("set state: %v", setDiags)
			}

			var resp resource.DeleteResponse
			r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("expected warning only, got errors: %v", resp.Diagnostics)
			}
			warnings := filterWarnings(resp.Diagnostics)
			if len(warnings) != 1 {
				t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
			}
			if warnings[0].Summary() != "Cannot delete system/default hosted page group" {
				t.Fatalf("unexpected warning summary: %q", warnings[0].Summary())
			}
			if !strings.Contains(warnings[0].Detail(), tt.groupName) {
				t.Fatalf("warning detail should mention %q: %q", tt.groupName, warnings[0].Detail())
			}
		})
	}
}

func filterWarnings(diags diag.Diagnostics) diag.Diagnostics {
	var out diag.Diagnostics
	for _, d := range diags {
		if d.Severity() == diag.SeverityWarning {
			out = append(out, d)
		}
	}
	return out
}
