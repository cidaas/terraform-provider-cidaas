package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidateIsRequiredMsgAvailable_AllowsRequiredMsgWhenOptional(t *testing.T) {
	t.Parallel()
	testValidateIsRequiredMsgAvailable(t, false, "The field is required", false)
}

func TestValidateIsRequiredMsgAvailable_RequiresRequiredMsgWhenRequired(t *testing.T) {
	t.Parallel()
	testValidateIsRequiredMsgAvailable(t, true, "", true)
}

func testValidateIsRequiredMsgAvailable(t *testing.T, required bool, requiredMsg string, expectError bool) {
	t.Helper()
	ctx := context.Background()

	req := validator.BoolRequest{
		Config:      buildValidatorTestConfig(ctx, t, required, requiredMsg),
		ConfigValue: types.BoolValue(required),
		Path:        path.Root("required"),
	}
	resp := &validator.BoolResponse{}
	validateIsRequiredMsgAvailable{}.ValidateBool(ctx, req, resp)

	if expectError && !resp.Diagnostics.HasError() {
		t.Fatal("expected validation error, got none")
	}
	if !expectError && resp.Diagnostics.HasError() {
		t.Fatalf("expected no validation error, got: %v", resp.Diagnostics.Errors())
	}
}

func buildValidatorTestConfig(ctx context.Context, t *testing.T, required bool, requiredMsg string) tfsdk.Config {
	t.Helper()

	attributeElemType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	}}
	consentLabelType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"label":      types.StringType,
		"label_text": types.StringType,
	}}
	localTextAttrs := map[string]attr.Type{
		"locale":         types.StringType,
		"name":           types.StringType,
		"max_length_msg": types.StringType,
		"min_length_msg": types.StringType,
		"required_msg":   types.StringType,
		"match_with_msg": types.StringType,
		"attributes":     types.ListType{ElemType: attributeElemType},
		"consent_label":  consentLabelType,
	}
	localText := types.ObjectValueMust(localTextAttrs, map[string]attr.Value{
		"locale":         types.StringValue("en-US"),
		"name":           types.StringValue("Sample Field"),
		"max_length_msg": types.StringNull(),
		"min_length_msg": types.StringNull(),
		"required_msg":   types.StringValue(requiredMsg),
		"match_with_msg": types.StringNull(),
		"attributes":     types.ListNull(attributeElemType),
		"consent_label":  types.ObjectNull(consentLabelType.AttrTypes),
	})
	localTexts := types.ListValueMust(types.ObjectType{AttrTypes: localTextAttrs}, []attr.Value{localText})

	fieldDefinition := types.ObjectValueMust(fieldDefinitionAttrTypes(), map[string]attr.Value{
		"max_length":        types.Int64Null(),
		"min_length":        types.Int64Null(),
		"min_date":          types.StringNull(),
		"max_date":          types.StringNull(),
		"initial_date_view": types.StringNull(),
		"initial_date":      types.StringNull(),
		"regex":             types.StringNull(),
		"regexes":           types.ListNull(types.StringType),
		"match_with":        types.StringNull(),
	})

	cfg := RegFieldConfig{
		ID:                                  types.StringNull(),
		BaseDataType:                        types.StringNull(),
		ParentGroupID:                       types.StringNull(),
		FieldType:                           types.StringNull(),
		DataType:                            types.StringValue("TEXT"),
		FieldKey:                            types.StringValue("sample"),
		Required:                            types.BoolValue(required),
		Internal:                            types.BoolNull(),
		Claimable:                           types.BoolNull(),
		IsSearchable:                        types.BoolNull(),
		Enabled:                             types.BoolNull(),
		Unique:                              types.BoolNull(),
		OverwriteWithNullFromSocialProvider: types.BoolNull(),
		ReadOnly:                            types.BoolNull(),
		IsList:                              types.BoolNull(),
		Order:                               types.Int64Null(),
		Scopes:                              types.SetNull(types.StringType),
		ConsentRefs:                         types.SetNull(types.StringType),
		LocalTexts:                          localTexts,
		FieldDefinition:                     fieldDefinition,
		RemoteFieldSettings:                 types.ObjectNull(remoteFieldSettingsAttrTypes()),
	}

	objType, ok := regFieldSchema.Type().(types.ObjectType)
	if !ok {
		t.Fatal("expected regFieldSchema to be an object type")
	}
	obj, diags := types.ObjectValueFrom(ctx, objType.AttrTypes, cfg)
	if diags.HasError() {
		t.Fatalf("ObjectValueFrom: %v", diags)
	}
	raw, err := obj.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("ToTerraformValue: %v", err)
	}
	return tfsdk.Config{Schema: regFieldSchema, Raw: raw}
}

func TestValidateMatchWith_AllowsOnPasswordEcho(t *testing.T) {
	t.Parallel()
	testValidateMatchWith(t, registrationFieldPasswordEchoKey, "password", "password must match", false)
}

func TestValidateMatchWith_RejectsOtherFieldKeys(t *testing.T) {
	t.Parallel()
	testValidateMatchWith(t, "custom_field", "password", "password must match", true)
}

func TestValidateMatchWith_RequiresMatchWithMsg(t *testing.T) {
	t.Parallel()
	testValidateMatchWith(t, registrationFieldPasswordEchoKey, "password", "", true)
}

func TestValidateMatchWithMsg_RejectsOtherFieldKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	req := validator.StringRequest{
		Config:      buildMatchWithValidatorTestConfig(ctx, t, "custom_field", "password", "password must match"),
		ConfigValue: types.StringValue("password must match"),
		Path:        path.Root("local_texts").AtListIndex(0).AtName("match_with_msg"),
	}
	resp := &validator.StringResponse{}
	validateMatchWithMsg{}.ValidateString(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected validation error, got none")
	}
}

func testValidateMatchWith(t *testing.T, fieldKey, matchWith, matchWithMsg string, expectError bool) {
	t.Helper()
	ctx := context.Background()
	req := validator.StringRequest{
		Config:      buildMatchWithValidatorTestConfig(ctx, t, fieldKey, matchWith, matchWithMsg),
		ConfigValue: types.StringValue(matchWith),
		Path:        path.Root("field_definition").AtName("match_with"),
	}
	resp := &validator.StringResponse{}
	validateMatchWith{}.ValidateString(ctx, req, resp)

	if expectError && !resp.Diagnostics.HasError() {
		t.Fatal("expected validation error, got none")
	}
	if !expectError && resp.Diagnostics.HasError() {
		t.Fatalf("expected no validation error, got: %v", resp.Diagnostics.Errors())
	}
}

func buildMatchWithValidatorTestConfig(ctx context.Context, t *testing.T, fieldKey, matchWith, matchWithMsg string) tfsdk.Config {
	t.Helper()

	attributeElemType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	}}
	consentLabelType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"label":      types.StringType,
		"label_text": types.StringType,
	}}
	localTextAttrs := map[string]attr.Type{
		"locale":         types.StringType,
		"name":           types.StringType,
		"max_length_msg": types.StringType,
		"min_length_msg": types.StringType,
		"required_msg":   types.StringType,
		"match_with_msg": types.StringType,
		"attributes":     types.ListType{ElemType: attributeElemType},
		"consent_label":  consentLabelType,
	}
	matchWithMsgValue := types.StringNull()
	if matchWithMsg != "" {
		matchWithMsgValue = types.StringValue(matchWithMsg)
	}
	localText := types.ObjectValueMust(localTextAttrs, map[string]attr.Value{
		"locale":         types.StringValue("en-US"),
		"name":           types.StringValue("Confirm Password"),
		"max_length_msg": types.StringNull(),
		"min_length_msg": types.StringNull(),
		"required_msg":   types.StringNull(),
		"match_with_msg": matchWithMsgValue,
		"attributes":     types.ListNull(attributeElemType),
		"consent_label":  types.ObjectNull(consentLabelType.AttrTypes),
	})
	localTexts := types.ListValueMust(types.ObjectType{AttrTypes: localTextAttrs}, []attr.Value{localText})

	fieldDefinition := types.ObjectValueMust(fieldDefinitionAttrTypes(), map[string]attr.Value{
		"max_length":        types.Int64Null(),
		"min_length":        types.Int64Null(),
		"min_date":          types.StringNull(),
		"max_date":          types.StringNull(),
		"initial_date_view": types.StringNull(),
		"initial_date":      types.StringNull(),
		"regex":             types.StringNull(),
		"regexes":           types.ListNull(types.StringType),
		"match_with":        types.StringValue(matchWith),
	})

	cfg := RegFieldConfig{
		DataType:            types.StringValue("PASSWORD"),
		FieldKey:            types.StringValue(fieldKey),
		FieldType:           types.StringValue("SYSTEM"),
		ParentGroupID:       types.StringValue("DEFAULT"),
		Required:            types.BoolValue(true),
		Scopes:              types.SetNull(types.StringType),
		ConsentRefs:         types.SetNull(types.StringType),
		LocalTexts:          localTexts,
		FieldDefinition:     fieldDefinition,
		RemoteFieldSettings: types.ObjectNull(remoteFieldSettingsAttrTypes()),
	}

	objType, ok := regFieldSchema.Type().(types.ObjectType)
	if !ok {
		t.Fatal("expected regFieldSchema to be an object type")
	}
	obj, diags := types.ObjectValueFrom(ctx, objType.AttrTypes, cfg)
	if diags.HasError() {
		t.Fatalf("ObjectValueFrom: %v", diags)
	}
	raw, err := obj.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("ToTerraformValue: %v", err)
	}
	return tfsdk.Config{Schema: regFieldSchema, Raw: raw}
}
