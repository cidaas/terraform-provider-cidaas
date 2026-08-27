// helpers.go provides Terraform Framework type conversions for cidaas_app_configuration.
package app

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func listToStrings(ctx context.Context, l types.List) ([]string, diag.Diagnostics) {
	if l.IsNull() || l.IsUnknown() {
		return nil, nil
	}
	var out []string
	diags := l.ElementsAs(ctx, &out, false)
	return out, diags
}

func stringList(values []string) types.List {
	elems := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elems = append(elems, types.StringValue(v))
	}
	return types.ListValueMust(types.StringType, elems)
}

func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func boolPtr(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

func boolValueOrNull(v *bool) types.Bool {
	if v == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*v)
}

func int64Ptr(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := v.ValueInt64()
	return &i
}

func int64ValueOrNull(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*v)
}

func redirectURIsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"redirect_uris":              types.ListType{ElemType: types.StringType},
		"allowed_logout_urls":        types.ListType{ElemType: types.StringType},
		"post_logout_redirect_uris":  types.ListType{ElemType: types.StringType},
		"allowed_web_origins":        types.ListType{ElemType: types.StringType},
	}
}

func scopesAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"allowed_scopes": types.ListType{ElemType: types.StringType},
		"default_scopes": types.ListType{ElemType: types.StringType},
	}
}

func tokenLifetimesAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"token_lifetime_in_seconds":         types.Int64Type,
		"refresh_token_lifetime_in_seconds": types.Int64Type,
		"id_token_lifetime_in_seconds":      types.Int64Type,
		"code_lifetime_in_seconds":          types.Int64Type,
		"default_max_age":                   types.Int64Type,
	}
}

func ownershipDetailsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"company_name":    types.StringType,
		"company_address": types.StringType,
		"company_website": types.StringType,
	}
}

func authenticationSetupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"verification_options_id":          types.StringType,
		"group_selection_id":               types.StringType,
		"group_verification_request_id":    types.StringType,
		"template_group_id":                types.StringType,
		"allow_guest_login":                types.BoolType,
		"is_remember_me_selected":          types.BoolType,
		"admin_client":                     types.BoolType,
		"is_login_success_page_enabled":    types.BoolType,
		"is_register_success_page_enabled": types.BoolType,
	}
}

func clientAuthConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"token_endpoint_auth_method": types.StringType,
	}
}

func signingKeyConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"active_kid": types.StringType,
		"next_kid":   types.StringType,
	}
}

func emptyStringList() types.List {
	return types.ListValueMust(types.StringType, []attr.Value{})
}
