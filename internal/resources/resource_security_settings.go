package resources

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Synthetic Terraform state id for the per-tenant singleton (also used for terraform import).
const securitySettingsResourceID = "security_settings"

type securitySettingsResource struct {
	BaseResource
}

// NewSecuritySettings returns the fraud-detection / security settings resource.
func NewSecuritySettings() resource.Resource {
	return &securitySettingsResource{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_SECURITY_SETTINGS,
				Schema: &securitySettingsSchema,
			},
		),
	}
}

type securitySettingsModel struct {
	ID                             types.String `tfsdk:"id"`
	BlockingSetting                types.Object `tfsdk:"blocking_setting"`
	RepeatedLoginBlockingMechanism types.Object `tfsdk:"repeated_login_blocking_mechanism"`
	RuleConfiguration              types.Object `tfsdk:"rule_configuration"`
}

type blockingSettingModel struct {
	Enabled                     types.Bool `tfsdk:"enabled"`
	BlackListedEmailDomains     types.Set  `tfsdk:"black_listed_email_domains"`
	BlackListedIPs              types.Set  `tfsdk:"black_listed_ips"`
	ExcludedEmailsFromBlackList types.Set  `tfsdk:"excluded_emails_from_black_list"`
	ExcludedIPsFromBlackList    types.Set  `tfsdk:"excluded_ips_from_black_list"`
	Subs                        types.Set  `tfsdk:"subs"`
	WhiteListedEmailDomains     types.Set  `tfsdk:"white_listed_email_domains"`
	WhiteListedIPs              types.Set  `tfsdk:"white_listed_ips"`
	BlackListedIdentifiers      types.Set  `tfsdk:"black_listed_identifiers"`
}

type repeatedLoginBlockingMechanismModel struct {
	BlockingDurationInMin     types.Int64 `tfsdk:"blocking_duration_in_min"`
	BlockedCount              types.Int64 `tfsdk:"blocked_count"`
	BlockedCountUnknownDevice types.Int64 `tfsdk:"blocked_count_unknown_device"`
}

type ruleConfigurationModel struct {
	RepeatedLoginBlockingMechanismEnabled types.Bool `tfsdk:"repeated_login_blocking_mechanism_enabled"`
}

var securitySettingsSchema = schema.Schema{
	MarkdownDescription: "Tenant-wide fraud-detection settings via `fraud-detection-srv/settings`. " +
		"This resource exposes **`blocking_setting`**, **`repeated_login_blocking_mechanism`**, and **`rule_configuration.repeated_login_blocking_mechanism_enabled`** only; other fraud-detection API fields are not configurable here. " +
		"Updates use HTTP PATCH (partial merge). Destroy only removes the resource from Terraform state; it does **not** reset settings in Cidaas.\n\n" +
		"State only stores attributes you set in Terraform. Remote values for omitted fields are not written to state (so apply matches partial configuration). " +
		"After `terraform import`, the first refresh may load a full API snapshot into state until your `.tf` matches or you use `ignore_changes`.\n\n" +
		"Required OAuth scopes on the Terraform client: `cidaas:fds_settings_read`, `cidaas:fds_settings_write`.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Stable id for this singleton resource. After the first apply, import with " +
				"`terraform import cidaas_security_settings.<name> " + securitySettingsResourceID + "`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"blocking_setting": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "Blocking / allowlist configuration (`blockingSetting` in the API).",
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					Optional:            true,
					MarkdownDescription: "Whether blocking is enabled.",
				},
				"black_listed_email_domains": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Email domains on the block list. Set to an empty set to send an empty list in PATCH (clear), if the API supports it.",
				},
				"black_listed_ips": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "IPs or CIDRs on the block list.",
				},
				"excluded_emails_from_black_list": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Emails excluded from domain blocking.",
				},
				"excluded_ips_from_black_list": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "IPs excluded from IP blocking.",
				},
				"subs": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Subscription identifiers used by blocking rules.",
				},
				"white_listed_email_domains": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Email domains on the allow list.",
				},
				"white_listed_ips": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "IPs on the allow list.",
				},
				"black_listed_identifiers": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: "Blocked identifiers (e.g. usernames or emails).",
				},
			},
		},
		"repeated_login_blocking_mechanism": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "Brute-force / repeated-login thresholds (`repeatedLoginBlockingMechanism`).",
			Attributes: map[string]schema.Attribute{
				"blocking_duration_in_min": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "How long logins stay blocked, in minutes.",
				},
				"blocked_count": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "Failed attempts before block (known device).",
				},
				"blocked_count_unknown_device": schema.Int64Attribute{
					Optional:            true,
					MarkdownDescription: "Failed attempts before block (unknown device).",
				},
			},
		},
		"rule_configuration": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "Subset of `ruleConfiguration` supported by this resource.",
			Attributes: map[string]schema.Attribute{
				"repeated_login_blocking_mechanism_enabled": schema.BoolAttribute{
					Optional:            true,
					MarkdownDescription: "`repeatedLoginBlockingMechanismEnabled`.",
				},
			},
		},
	},
}

func blockingSettingAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":                         types.BoolType,
		"black_listed_email_domains":      types.SetType{ElemType: types.StringType},
		"black_listed_ips":                types.SetType{ElemType: types.StringType},
		"excluded_emails_from_black_list": types.SetType{ElemType: types.StringType},
		"excluded_ips_from_black_list":    types.SetType{ElemType: types.StringType},
		"subs":                            types.SetType{ElemType: types.StringType},
		"white_listed_email_domains":      types.SetType{ElemType: types.StringType},
		"white_listed_ips":                types.SetType{ElemType: types.StringType},
		"black_listed_identifiers":        types.SetType{ElemType: types.StringType},
	}
}

func repeatedLoginAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"blocking_duration_in_min":     types.Int64Type,
		"blocked_count":                types.Int64Type,
		"blocked_count_unknown_device": types.Int64Type,
	}
}

func ruleConfigurationAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"repeated_login_blocking_mechanism_enabled": types.BoolType,
	}
}

func setFromOptionalStringSlice(p *[]string) (types.Set, diag.Diagnostics) {
	if p == nil {
		return types.SetNull(types.StringType), nil
	}
	if len(*p) == 0 {
		v, d := types.SetValue(types.StringType, []attr.Value{})
		return v, d
	}
	elements := make([]attr.Value, 0, len(*p))
	for _, s := range *p {
		elements = append(elements, types.StringValue(s))
	}
	v, d := types.SetValue(types.StringType, elements)
	return v, d
}

func (m *securitySettingsModel) toPatch(ctx context.Context) (cidaas.SecuritySettingsPatch, diag.Diagnostics) {
	var diags diag.Diagnostics
	var patch cidaas.SecuritySettingsPatch

	if !m.BlockingSetting.IsNull() && !m.BlockingSetting.IsUnknown() {
		var bs blockingSettingModel
		diags.Append(m.BlockingSetting.As(ctx, &bs, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return patch, diags
		}
		apiBS := &cidaas.BlockingSetting{}
		hasAny := false
		if !bs.Enabled.IsNull() {
			v := bs.Enabled.ValueBool()
			apiBS.Enabled = &v
			hasAny = true
		}
		if !bs.BlackListedEmailDomains.IsNull() {
			var sl []string
			diags.Append(bs.BlackListedEmailDomains.ElementsAs(ctx, &sl, false)...)
			apiBS.BlackListedEmailDomains = &sl
			hasAny = true
		}
		if !bs.BlackListedIPs.IsNull() {
			var sl []string
			diags.Append(bs.BlackListedIPs.ElementsAs(ctx, &sl, false)...)
			apiBS.BlackListedIPs = &sl
			hasAny = true
		}
		if !bs.ExcludedEmailsFromBlackList.IsNull() {
			var sl []string
			diags.Append(bs.ExcludedEmailsFromBlackList.ElementsAs(ctx, &sl, false)...)
			apiBS.ExcludedEmailsFromBlackList = &sl
			hasAny = true
		}
		if !bs.ExcludedIPsFromBlackList.IsNull() {
			var sl []string
			diags.Append(bs.ExcludedIPsFromBlackList.ElementsAs(ctx, &sl, false)...)
			apiBS.ExcludedIPsFromBlackList = &sl
			hasAny = true
		}
		if !bs.Subs.IsNull() {
			var sl []string
			diags.Append(bs.Subs.ElementsAs(ctx, &sl, false)...)
			apiBS.Subs = &sl
			hasAny = true
		}
		if !bs.WhiteListedEmailDomains.IsNull() {
			var sl []string
			diags.Append(bs.WhiteListedEmailDomains.ElementsAs(ctx, &sl, false)...)
			apiBS.WhiteListedEmailDomains = &sl
			hasAny = true
		}
		if !bs.WhiteListedIPs.IsNull() {
			var sl []string
			diags.Append(bs.WhiteListedIPs.ElementsAs(ctx, &sl, false)...)
			apiBS.WhiteListedIPs = &sl
			hasAny = true
		}
		if !bs.BlackListedIdentifiers.IsNull() {
			var sl []string
			diags.Append(bs.BlackListedIdentifiers.ElementsAs(ctx, &sl, false)...)
			apiBS.BlackListedIdentifiers = &sl
			hasAny = true
		}
		if hasAny {
			patch.BlockingSetting = apiBS
		}
	}

	if !m.RepeatedLoginBlockingMechanism.IsNull() && !m.RepeatedLoginBlockingMechanism.IsUnknown() {
		var rl repeatedLoginBlockingMechanismModel
		diags.Append(m.RepeatedLoginBlockingMechanism.As(ctx, &rl, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return patch, diags
		}
		apiRL := &cidaas.RepeatedLoginBlockingMechanism{}
		hasAny := false
		if !rl.BlockingDurationInMin.IsNull() {
			v := rl.BlockingDurationInMin.ValueInt64()
			apiRL.BlockingDurationInMin = &v
			hasAny = true
		}
		if !rl.BlockedCount.IsNull() {
			v := rl.BlockedCount.ValueInt64()
			apiRL.BlockedCount = &v
			hasAny = true
		}
		if !rl.BlockedCountUnknownDevice.IsNull() {
			v := rl.BlockedCountUnknownDevice.ValueInt64()
			apiRL.BlockedCountUnknownDevice = &v
			hasAny = true
		}
		if hasAny {
			patch.RepeatedLoginBlockingMechanism = apiRL
		}
	}

	if !m.RuleConfiguration.IsNull() && !m.RuleConfiguration.IsUnknown() {
		var rc ruleConfigurationModel
		diags.Append(m.RuleConfiguration.As(ctx, &rc, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return patch, diags
		}
		if !rc.RepeatedLoginBlockingMechanismEnabled.IsNull() {
			v := rc.RepeatedLoginBlockingMechanismEnabled.ValueBool()
			patch.RuleConfiguration = &cidaas.RuleConfiguration{
				RepeatedLoginBlockingMechanismEnabled: &v,
			}
		}
	}

	return patch, diags
}

func isEmptyPatch(p cidaas.SecuritySettingsPatch) bool {
	return p.BlockingSetting == nil &&
		p.RepeatedLoginBlockingMechanism == nil &&
		p.RuleConfiguration == nil
}

func dataToModel(_ context.Context, data *cidaas.SecuritySettingsData) (securitySettingsModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var m securitySettingsModel
	if data == nil {
		diags.AddError("invalid API response", "fraud-detection settings data was empty")
		return m, diags
	}

	if data.BlockingSetting != nil {
		bs := data.BlockingSetting
		enabled := util.BoolValueOrNull(bs.Enabled)
		blackListedEmailDomains, d := setFromOptionalStringSlice(bs.BlackListedEmailDomains)
		diags.Append(d...)
		blackListedIPs, d := setFromOptionalStringSlice(bs.BlackListedIPs)
		diags.Append(d...)
		exclEmails, d := setFromOptionalStringSlice(bs.ExcludedEmailsFromBlackList)
		diags.Append(d...)
		exclIPs, d := setFromOptionalStringSlice(bs.ExcludedIPsFromBlackList)
		diags.Append(d...)
		subs, d := setFromOptionalStringSlice(bs.Subs)
		diags.Append(d...)
		whiteDomains, d := setFromOptionalStringSlice(bs.WhiteListedEmailDomains)
		diags.Append(d...)
		whiteIPs, d := setFromOptionalStringSlice(bs.WhiteListedIPs)
		diags.Append(d...)
		blackIdents, d := setFromOptionalStringSlice(bs.BlackListedIdentifiers)
		diags.Append(d...)
		if diags.HasError() {
			return m, diags
		}
		obj, d := types.ObjectValue(blockingSettingAttrTypes(), map[string]attr.Value{
			"enabled":                         enabled,
			"black_listed_email_domains":      blackListedEmailDomains,
			"black_listed_ips":                blackListedIPs,
			"excluded_emails_from_black_list": exclEmails,
			"excluded_ips_from_black_list":    exclIPs,
			"subs":                            subs,
			"white_listed_email_domains":      whiteDomains,
			"white_listed_ips":                whiteIPs,
			"black_listed_identifiers":        blackIdents,
		})
		diags.Append(d...)
		m.BlockingSetting = obj
	} else {
		m.BlockingSetting = types.ObjectNull(blockingSettingAttrTypes())
	}

	if data.RepeatedLoginBlockingMechanism != nil {
		rl := data.RepeatedLoginBlockingMechanism
		obj, d := types.ObjectValue(repeatedLoginAttrTypes(), map[string]attr.Value{
			"blocking_duration_in_min":     util.Int64ValueOrNull(rl.BlockingDurationInMin),
			"blocked_count":                util.Int64ValueOrNull(rl.BlockedCount),
			"blocked_count_unknown_device": util.Int64ValueOrNull(rl.BlockedCountUnknownDevice),
		})
		diags.Append(d...)
		m.RepeatedLoginBlockingMechanism = obj
	} else {
		m.RepeatedLoginBlockingMechanism = types.ObjectNull(repeatedLoginAttrTypes())
	}

	if data.RuleConfiguration != nil {
		rc := data.RuleConfiguration
		obj, d := types.ObjectValue(ruleConfigurationAttrTypes(), map[string]attr.Value{
			"repeated_login_blocking_mechanism_enabled": util.BoolValueOrNull(rc.RepeatedLoginBlockingMechanismEnabled),
		})
		diags.Append(d...)
		m.RuleConfiguration = obj
	} else {
		m.RuleConfiguration = types.ObjectNull(ruleConfigurationAttrTypes())
	}

	return m, diags
}

// priorSecuritySettingsIsBare is true when state only has id (e.g. after terraform import) — refresh should store the full API view.
func priorSecuritySettingsIsBare(prior securitySettingsModel) bool {
	return prior.BlockingSetting.IsNull() &&
		prior.RepeatedLoginBlockingMechanism.IsNull() &&
		prior.RuleConfiguration.IsNull()
}

func mergeStringSet(_ context.Context, mask, api types.Set, _ *diag.Diagnostics) types.Set {
	if mask.IsNull() {
		return types.SetNull(types.StringType)
	}
	if !api.IsNull() && !api.IsUnknown() {
		return api
	}
	return mask
}

func mergeBool(mask, api types.Bool) types.Bool {
	if mask.IsNull() {
		return types.BoolNull()
	}
	if !api.IsNull() && !api.IsUnknown() {
		return api
	}
	return mask
}

func mergeInt64(mask, api types.Int64) types.Int64 {
	if mask.IsNull() {
		return types.Int64Null()
	}
	if !api.IsNull() && !api.IsUnknown() {
		return api
	}
	return mask
}

func mergeBlockingSettingState(ctx context.Context, mask, api blockingSettingModel, diags *diag.Diagnostics) blockingSettingModel {
	return blockingSettingModel{
		Enabled:                     mergeBool(mask.Enabled, api.Enabled),
		BlackListedEmailDomains:     mergeStringSet(ctx, mask.BlackListedEmailDomains, api.BlackListedEmailDomains, diags),
		BlackListedIPs:              mergeStringSet(ctx, mask.BlackListedIPs, api.BlackListedIPs, diags),
		ExcludedEmailsFromBlackList: mergeStringSet(ctx, mask.ExcludedEmailsFromBlackList, api.ExcludedEmailsFromBlackList, diags),
		ExcludedIPsFromBlackList:    mergeStringSet(ctx, mask.ExcludedIPsFromBlackList, api.ExcludedIPsFromBlackList, diags),
		Subs:                        mergeStringSet(ctx, mask.Subs, api.Subs, diags),
		WhiteListedEmailDomains:     mergeStringSet(ctx, mask.WhiteListedEmailDomains, api.WhiteListedEmailDomains, diags),
		WhiteListedIPs:              mergeStringSet(ctx, mask.WhiteListedIPs, api.WhiteListedIPs, diags),
		BlackListedIdentifiers:      mergeStringSet(ctx, mask.BlackListedIdentifiers, api.BlackListedIdentifiers, diags),
	}
}

func blockingModelToObject(_ context.Context, m blockingSettingModel, diags *diag.Diagnostics) types.Object {
	obj, d := types.ObjectValue(blockingSettingAttrTypes(), map[string]attr.Value{
		"enabled":                         m.Enabled,
		"black_listed_email_domains":      m.BlackListedEmailDomains,
		"black_listed_ips":                m.BlackListedIPs,
		"excluded_emails_from_black_list": m.ExcludedEmailsFromBlackList,
		"excluded_ips_from_black_list":    m.ExcludedIPsFromBlackList,
		"subs":                            m.Subs,
		"white_listed_email_domains":      m.WhiteListedEmailDomains,
		"white_listed_ips":                m.WhiteListedIPs,
		"black_listed_identifiers":        m.BlackListedIdentifiers,
	})
	diags.Append(d...)
	return obj
}

func mergeRepeatedLoginState(mask, api repeatedLoginBlockingMechanismModel) repeatedLoginBlockingMechanismModel {
	return repeatedLoginBlockingMechanismModel{
		BlockingDurationInMin:     mergeInt64(mask.BlockingDurationInMin, api.BlockingDurationInMin),
		BlockedCount:              mergeInt64(mask.BlockedCount, api.BlockedCount),
		BlockedCountUnknownDevice: mergeInt64(mask.BlockedCountUnknownDevice, api.BlockedCountUnknownDevice),
	}
}

func mergeRuleConfigurationState(mask, api ruleConfigurationModel) ruleConfigurationModel {
	return ruleConfigurationModel{
		RepeatedLoginBlockingMechanismEnabled: mergeBool(mask.RepeatedLoginBlockingMechanismEnabled, api.RepeatedLoginBlockingMechanismEnabled),
	}
}

// mergeSecuritySettingsState aligns refreshed API data with the mask (plan or prior state) so Terraform does not report
// "inconsistent result after apply" when the config only sets a subset of fields or when the API returns null for an empty list.
func mergeSecuritySettingsState(ctx context.Context, mask, refreshed securitySettingsModel) (securitySettingsModel, diag.Diagnostics) {
	var out securitySettingsModel
	var diags diag.Diagnostics

	if mask.BlockingSetting.IsNull() {
		out.BlockingSetting = types.ObjectNull(blockingSettingAttrTypes())
	} else {
		var mbs, abs blockingSettingModel
		diags.Append(mask.BlockingSetting.As(ctx, &mbs, basetypes.ObjectAsOptions{})...)
		if !refreshed.BlockingSetting.IsNull() && !refreshed.BlockingSetting.IsUnknown() {
			diags.Append(refreshed.BlockingSetting.As(ctx, &abs, basetypes.ObjectAsOptions{})...)
		}
		if diags.HasError() {
			return out, diags
		}
		merged := mergeBlockingSettingState(ctx, mbs, abs, &diags)
		out.BlockingSetting = blockingModelToObject(ctx, merged, &diags)
	}

	if mask.RepeatedLoginBlockingMechanism.IsNull() {
		out.RepeatedLoginBlockingMechanism = types.ObjectNull(repeatedLoginAttrTypes())
	} else {
		var mrl, arl repeatedLoginBlockingMechanismModel
		diags.Append(mask.RepeatedLoginBlockingMechanism.As(ctx, &mrl, basetypes.ObjectAsOptions{})...)
		if !refreshed.RepeatedLoginBlockingMechanism.IsNull() && !refreshed.RepeatedLoginBlockingMechanism.IsUnknown() {
			diags.Append(refreshed.RepeatedLoginBlockingMechanism.As(ctx, &arl, basetypes.ObjectAsOptions{})...)
		}
		if diags.HasError() {
			return out, diags
		}
		merged := mergeRepeatedLoginState(mrl, arl)
		obj, d := types.ObjectValue(repeatedLoginAttrTypes(), map[string]attr.Value{
			"blocking_duration_in_min":     merged.BlockingDurationInMin,
			"blocked_count":                merged.BlockedCount,
			"blocked_count_unknown_device": merged.BlockedCountUnknownDevice,
		})
		diags.Append(d...)
		out.RepeatedLoginBlockingMechanism = obj
	}

	if mask.RuleConfiguration.IsNull() {
		out.RuleConfiguration = types.ObjectNull(ruleConfigurationAttrTypes())
	} else {
		var mrc, arc ruleConfigurationModel
		diags.Append(mask.RuleConfiguration.As(ctx, &mrc, basetypes.ObjectAsOptions{})...)
		if !refreshed.RuleConfiguration.IsNull() && !refreshed.RuleConfiguration.IsUnknown() {
			diags.Append(refreshed.RuleConfiguration.As(ctx, &arc, basetypes.ObjectAsOptions{})...)
		}
		if diags.HasError() {
			return out, diags
		}
		merged := mergeRuleConfigurationState(mrc, arc)
		obj, d := types.ObjectValue(ruleConfigurationAttrTypes(), map[string]attr.Value{
			"repeated_login_blocking_mechanism_enabled": merged.RepeatedLoginBlockingMechanismEnabled,
		})
		diags.Append(d...)
		out.RuleConfiguration = obj
	}

	return out, diags
}

func (r *securitySettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securitySettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patch, diags := plan.toPatch(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if isEmptyPatch(patch) {
		resp.Diagnostics.AddError(
			"invalid configuration",
			"at least one of blocking_setting, repeated_login_blocking_mechanism, or rule_configuration must be set",
		)
		return
	}

	if err := r.cidaasClient.SecuritySettings.Patch(ctx, patch); err != nil {
		tflog.Error(ctx, "failed to PATCH fraud-detection settings", map[string]interface{}{"error": err.Error()})
		resp.Diagnostics.AddError("failed to update fraud-detection settings", err.Error())
		return
	}

	refreshed, err := r.cidaasClient.SecuritySettings.Get(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to read fraud-detection settings after update", err.Error())
		return
	}
	if refreshed.Data == nil {
		resp.Diagnostics.AddError("invalid API response", "fraud-detection settings GET returned no data")
		return
	}

	full, diags := dataToModel(ctx, refreshed.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state, mergeDiags := mergeSecuritySettingsState(ctx, plan, full)
	resp.Diagnostics.Append(mergeDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ID = types.StringValue(securitySettingsResourceID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *securitySettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securitySettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	refreshed, err := r.cidaasClient.SecuritySettings.Get(ctx)
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		resp.Diagnostics.AddError("failed to read fraud-detection settings", err.Error())
		return
	}
	if refreshed.Data == nil {
		resp.Diagnostics.AddError("invalid API response", "fraud-detection settings GET returned no data")
		return
	}

	full, diags := dataToModel(ctx, refreshed.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var newState securitySettingsModel
	if priorSecuritySettingsIsBare(state) {
		newState = full
	} else {
		var mergeDiags diag.Diagnostics
		newState, mergeDiags = mergeSecuritySettingsState(ctx, state, full)
		resp.Diagnostics.Append(mergeDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !state.ID.IsNull() && state.ID.ValueString() != "" {
		newState.ID = state.ID
	} else {
		newState.ID = types.StringValue(securitySettingsResourceID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *securitySettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan securitySettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patch, diags := plan.toPatch(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if isEmptyPatch(patch) {
		resp.Diagnostics.AddError(
			"invalid configuration",
			"at least one of blocking_setting, repeated_login_blocking_mechanism, or rule_configuration must be set",
		)
		return
	}

	if err := r.cidaasClient.SecuritySettings.Patch(ctx, patch); err != nil {
		resp.Diagnostics.AddError("failed to update fraud-detection settings", err.Error())
		return
	}

	refreshed, err := r.cidaasClient.SecuritySettings.Get(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to read fraud-detection settings after update", err.Error())
		return
	}
	if refreshed.Data == nil {
		resp.Diagnostics.AddError("invalid API response", "fraud-detection settings GET returned no data")
		return
	}

	full, diags := dataToModel(ctx, refreshed.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state, mergeDiags := mergeSecuritySettingsState(ctx, plan, full)
	resp.Diagnostics.Append(mergeDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ID = types.StringValue(securitySettingsResourceID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *securitySettingsResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Info(ctx, "cidaas_security_settings destroy: removing from Terraform state only; fraud-detection settings in Cidaas are unchanged")
}

func (r *securitySettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
