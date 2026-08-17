package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestEnsureFieldDefinitionRegexKnown_UnknownWithoutRegexesBecomesNull(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := RegFieldConfig{
		fieldDefinition: &FieldDefinition{
			MaxLength:       types.Int64Null(),
			MinLength:       types.Int64Null(),
			MinDate:         types.StringNull(),
			MaxDate:         types.StringNull(),
			InitialDateView: types.StringNull(),
			InitialDate:     types.StringNull(),
			Regex:           types.StringUnknown(),
			Regexes:         types.ListNull(types.StringType),
			MatchWith:       types.StringNull(),
		},
	}

	diags := ensureFieldDefinitionRegexKnown(ctx, &plan)
	if diags.HasError() {
		t.Fatalf("ensure: %v", diags)
	}
	if plan.fieldDefinition.Regex.IsUnknown() {
		t.Fatal("regex still unknown; want known null")
	}
	if !plan.fieldDefinition.Regex.IsNull() {
		t.Fatalf("regex = %q, want null", plan.fieldDefinition.Regex.ValueString())
	}
	if plan.FieldDefinition.IsNull() || plan.FieldDefinition.IsUnknown() {
		t.Fatal("expected FieldDefinition object to be known")
	}
	var fd FieldDefinition
	if d := plan.FieldDefinition.As(ctx, &fd, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("as: %v", d)
	}
	if fd.Regex.IsUnknown() || !fd.Regex.IsNull() {
		t.Fatalf("object regex unknown=%v null=%v, want known null", fd.Regex.IsUnknown(), fd.Regex.IsNull())
	}
}

func TestEnsureFieldDefinitionRegexKnown_LeavesRegexesPathAlone(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := RegFieldConfig{
		fieldDefinition: &FieldDefinition{
			Regex:   types.StringUnknown(),
			Regexes: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("^.{1,2}$")}),
		},
	}
	diags := ensureFieldDefinitionRegexKnown(ctx, &plan)
	if diags.HasError() {
		t.Fatalf("ensure: %v", diags)
	}
	if !plan.fieldDefinition.Regex.IsUnknown() {
		t.Fatal("regexes path should leave unknown regex for syncComposedRegexIntoPlan")
	}
}

func TestSyncComposedRegexIntoPlan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	patterns := []string{"^[A-Za-z]*$", "^.{3,6}$"}
	regexes, diags := types.ListValueFrom(ctx, types.StringType, patterns)
	if diags.HasError() {
		t.Fatalf("list: %v", diags)
	}
	composed, err := composeANDRegexes(patterns)
	if err != nil {
		t.Fatal(err)
	}

	plan := RegFieldConfig{
		fieldDefinition: &FieldDefinition{
			Regex:   types.StringValue("^[A-Za-z]{2,5}$"), // stale prior composition
			Regexes: regexes,
		},
	}
	diags = syncComposedRegexIntoPlan(ctx, &plan, composed)
	if diags.HasError() {
		t.Fatalf("sync: %v", diags)
	}
	if got := plan.fieldDefinition.Regex.ValueString(); got != composed {
		t.Fatalf("regex = %q, want %q", got, composed)
	}
	if plan.FieldDefinition.IsNull() || plan.FieldDefinition.IsUnknown() {
		t.Fatal("expected FieldDefinition object to be set")
	}
	var fd FieldDefinition
	if d := plan.FieldDefinition.As(ctx, &fd, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("as: %v", d)
	}
	if fd.Regex.ValueString() != composed {
		t.Fatalf("object regex = %q, want %q", fd.Regex.ValueString(), composed)
	}
}

func TestValidateRegexDataTypeAndMessages(t *testing.T) {
	t.Parallel()

	cfg := RegFieldConfig{
		DataType: types.StringValue("NUMBER"),
		localTexts: []*LocalTexts{{
			MinLengthMsg: types.StringValue("min"),
			MaxLengthMsg: types.StringValue("max"),
		}},
	}
	var d diag.Diagnostics
	validateRegexDataTypeAndMessages(context.Background(), cfg, "field_definition.regexes", &d)
	if !d.HasError() {
		t.Fatal("expected error for non TEXT/URL data_type")
	}

	cfg.DataType = types.StringValue("TEXT")
	cfg.localTexts[0].MinLengthMsg = types.StringNull()
	d = diag.Diagnostics{}
	validateRegexDataTypeAndMessages(context.Background(), cfg, "field_definition.regexes", &d)
	if !d.HasError() {
		t.Fatal("expected error when min_length_msg missing")
	}

	cfg.localTexts[0].MinLengthMsg = types.StringValue("min")
	cfg.localTexts[0].MaxLengthMsg = types.StringNull()
	d = diag.Diagnostics{}
	validateRegexDataTypeAndMessages(context.Background(), cfg, "field_definition.regexes", &d)
	if !d.HasError() {
		t.Fatal("expected error when max_length_msg missing")
	}

	cfg.localTexts[0].MaxLengthMsg = types.StringValue("max")
	d = diag.Diagnostics{}
	validateRegexDataTypeAndMessages(context.Background(), cfg, "field_definition.regexes", &d)
	if d.HasError() {
		t.Fatalf("unexpected errors: %v", d.Errors())
	}
}

func TestFieldDefinitionMutualExclusiveHelpers(t *testing.T) {
	t.Parallel()

	if fieldDefinitionHasRegex(nil) || fieldDefinitionHasRegexes(nil) {
		t.Fatal("nil should be false")
	}

	fd := &FieldDefinition{
		Regex:   types.StringNull(),
		Regexes: types.ListNull(types.StringType),
	}
	if fieldDefinitionHasRegex(fd) || fieldDefinitionHasRegexes(fd) {
		t.Fatal("null values should be false")
	}

	fd.Regex = types.StringValue("^a$")
	fd.Regexes = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("^b$")})
	if !fieldDefinitionHasRegex(fd) || !fieldDefinitionHasRegexes(fd) {
		t.Fatal("expected both set")
	}
}
