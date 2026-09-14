package registrationfield

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPrepareRegFieldModel_IncludesOrder(t *testing.T) {
	t.Parallel()

	model, diags := prepareRegFieldModel(context.Background(), testRegFieldPlanWithOrder(types.Int64Value(3)))
	if diags.HasError() {
		t.Fatalf("prepareRegFieldModel: %v", diags.Errors())
	}
	if model.Order != 3 {
		t.Fatalf("expected Order 3, got %d", model.Order)
	}
}

func TestPrepareRegFieldModel_OmitsOrderWhenUnset(t *testing.T) {
	t.Parallel()

	model, diags := prepareRegFieldModel(context.Background(), testRegFieldPlanWithOrder(types.Int64Null()))
	if diags.HasError() {
		t.Fatalf("prepareRegFieldModel: %v", diags.Errors())
	}
	if model.Order != 0 {
		t.Fatalf("expected Order 0 when unset, got %d", model.Order)
	}
}

func testRegFieldPlanWithOrder(order types.Int64) RegFieldConfig {
	attributeElemType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	}}
	localTextAttrs := map[string]attr.Type{
		"locale":         types.StringType,
		"name":           types.StringType,
		"max_length_msg": types.StringType,
		"min_length_msg": types.StringType,
		"required_msg":   types.StringType,
		"attributes":     types.ListType{ElemType: attributeElemType},
		"consent_label": types.ObjectType{AttrTypes: map[string]attr.Type{
			"label":      types.StringType,
			"label_text": types.StringType,
		}},
	}
	localText := types.ObjectValueMust(localTextAttrs, map[string]attr.Value{
		"locale":         types.StringValue("en-US"),
		"name":           types.StringValue("Sample Field"),
		"max_length_msg": types.StringNull(),
		"min_length_msg": types.StringNull(),
		"required_msg":   types.StringNull(),
		"attributes":     types.ListNull(attributeElemType),
		"consent_label": types.ObjectNull(map[string]attr.Type{
			"label":      types.StringType,
			"label_text": types.StringType,
		}),
	})

	return RegFieldConfig{
		DataType:      types.StringValue("TEXT"),
		FieldKey:      types.StringValue("sample_field"),
		FieldType:     types.StringValue("CUSTOM"),
		ParentGroupID: types.StringValue("DEFAULT"),
		Order:         order,
		Scopes:        types.SetNull(types.StringType),
		LocalTexts:    types.ListValueMust(types.ObjectType{AttrTypes: localTextAttrs}, []attr.Value{localText}),
	}
}
