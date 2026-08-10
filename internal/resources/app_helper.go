package resources

import (
	"fmt"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func StringValueOrNullWithPlanValue(value *string, configValue, planValue *basetypes.StringValue, attrName string, operation int) diag.Diagnostics {
	var diags diag.Diagnostics
	if value != nil && *value != "" {
		if !planValue.IsNull() && !planValue.IsUnknown() && *value != planValue.ValueString() && operation != READ && operation != IMPORT {
			msg := fmt.Sprintf(
				"The attribute '%s' can not be updated in Cidaas according to the provided configuration."+
					"This could be due to the Cidaas API no longer supporting the attribute or because the attribute value might be computed based on other attributes or dependencies."+
					"The state is updated with the value '%s' computed by Cidaas to prevent unnecessary errors in the provider. To prevent this warning, please update your configuration accordingly.",
				attrName, *value)
			summaryMsg := fmt.Sprintf("Failed to Set %s", attrName)
			diags.AddWarning(summaryMsg, msg)
		}
		*planValue = util.StringValueOrNull(value)
	} else {
		if !planValue.IsNull() && !planValue.IsUnknown() && operation != READ && operation != IMPORT {
			if operation == UPDATE && configValue.IsNull() && planValue.Equal(types.StringValue("")) {
				return diags
			}
			msg := fmt.Sprintf(
				"The attribute '%s' can not be updated in Cidaas according to the provided configuration."+
					"This could be due to the Cidaas API no longer supporting the attribute or because the attribute value might be computed based on other attributes or dependencies."+
					"The state is updated with the planned value '%s' to prevent unnecessary errors in the provider. To prevent this warning, please update your configuration accordingly.",
				attrName, planValue.ValueString())

			summaryMsg := fmt.Sprintf("Failed to Set %s", attrName)
			diags.AddWarning(summaryMsg, msg)
		} else if planValue.IsUnknown() {
			*planValue = types.StringValue("")
		}
	}
	return diags
}

func SetValueOrNullWithPlanValue(values []string, planValue *basetypes.SetValue, attrName string, operation int) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(values) > 0 {
		var temp []attr.Value
		for _, v := range values {
			temp = append(temp, types.StringValue(v))
		}
		*planValue = types.SetValueMust(types.StringType, temp)
		return diags
	} else if !planValue.IsNull() && len(planValue.Elements()) > 0 && operation != READ && operation != IMPORT {
		msg := fmt.Sprintf(
			"The attribute '%s' can not be updated in Cidaas according to the provided configuration."+
				"This could be due to the Cidaas API no longer supporting the attribute or because the attribute value might be computed based on other attributes or dependencies."+
				"The state is updated with the planned value '%t' to prevent unnecessary errors in the provider. To prevent this warning, please update your configuration accordingly.",
			attrName, planValue.Elements())
		summaryMsg := fmt.Sprintf("Failed to Set %s", attrName)
		diags.AddWarning(summaryMsg, msg)
	}

	computedSetAttributes := []string{
		"redirect_uris", "allowed_scopes", "allowed_logout_urls",
		"login_providers", "allowed_web_origins", "allowed_origins", "allowed_mfa", "allowed_roles", "default_roles", "default_scopes", "pending_scopes",
	}
	if planValue.IsUnknown() && util.Contains(computedSetAttributes, attrName) {
		*planValue = types.SetValueMust(types.StringType, []attr.Value{})
	}
	return diags
}

func BoolValueOrNullWithPlanValue(value *bool, planValue *basetypes.BoolValue, attrName string, operation int, defautlValue bool) diag.Diagnostics {
	var diags diag.Diagnostics
	if value != nil {
		if !planValue.IsNull() && !planValue.IsUnknown() && *value != planValue.ValueBool() && operation != READ && operation != IMPORT {
			msg := fmt.Sprintf(
				"The attribute '%s' can not be updated in Cidaas according to the provided configuration."+
					"This could be due to the Cidaas API no longer supporting the attribute or because the attribute value might be computed based on other attributes or dependencies."+
					"The state is updated with the value '%t' computed by Cidaas to prevent unnecessary errors in the provider. To prevent this warning, please update your configuration accordingly.",
				attrName, planValue.ValueBool())
			summaryMsg := fmt.Sprintf("Failed to Set %s", attrName)
			diags.AddWarning(summaryMsg, msg)
		}
		*planValue = types.BoolValue(*value)
	} else if planValue.IsNull() || planValue.IsUnknown() {
		*planValue = types.BoolValue(defautlValue)
	}
	return diags
}
