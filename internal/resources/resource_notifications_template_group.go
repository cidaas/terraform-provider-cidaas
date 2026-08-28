package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var notificationsTemplateGroupTGTypes = []string{"cidaas", "developer", "reminder"}

type NotificationsTemplateGroupResource struct {
	BaseResource
}

func NewNotificationsTemplateGroupResource() resource.Resource {
	return &NotificationsTemplateGroupResource{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_NOTIFICATIONS_TEMPLATE_GROUP, //nolint:revive
				Schema: &notificationsTemplateGroupSchema,
			},
		),
	}
}

// CommSettingAttrs are shared optional comm_setting_* blocks (notification-srv CommSetting).
var commSettingChannelAttrs = map[string]schema.Attribute{
	"service_setup_id": schema.StringAttribute{
		Optional:            true,
		MarkdownDescription: "Service setup (communication provider) id from notification-srv / tenant configuration.",
	},
	"sender_name": schema.StringAttribute{
		Optional:            true,
		MarkdownDescription: "Optional sender display name for this channel.",
	},
	"sender_address": schema.StringAttribute{
		Optional:            true,
		MarkdownDescription: "Optional sender address (e.g. email from / SMS originator).",
	},
	"reply_to": schema.StringAttribute{
		Optional:            true,
		MarkdownDescription: "Optional reply-to (e.g. for EMAIL).",
	},
	"has_remote_templates": schema.BoolAttribute{
		Optional:            true,
		MarkdownDescription: "Whether remote templates exist for this channel.",
	},
}

var notificationsTemplateGroupSchema = schema.Schema{
	MarkdownDescription: "Manages a **template group** via **notification-srv** (`/notifications-srv/templategroups`). " +
		"This is separate from `cidaas_template_group` (legacy `templates-srv/groups`) and does not replace it.\n\n" +
		"Template **locales** are managed with **`cidaas_notifications_template_group_locale`** (copy on create, bulk-delete on destroy). " +
		"Group create does not send `copy`; notification-srv may still seed locales from `default` per API rules.\n\n" +
		"**Scopes:** `cidaas:templates_read`, `cidaas:templates_write`, `cidaas:templates_delete` (and admin roles as enforced by notification-srv).",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Same as `group_id` (template group `_id`).",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"group_id": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Template group id (`_id`). Immutable after create.",
			Validators: []validator.String{
				stringvalidator.LengthAtMost(64),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"tg_type": schema.StringAttribute{
			Required: true,
			MarkdownDescription: "Template group type: `cidaas`, `developer`, or `reminder`. " +
				"`cidaas` — standard Cidaas platform emails (e.g. welcome, password reset). " +
				"`developer` — custom emails you design, trigger (e.g. from a flow), and fill with your payload. " +
				"`reminder` — scheduled follow-up when a user delays an action (e.g. email verification).",
			Validators: []validator.String{
				stringvalidator.OneOf(notificationsTemplateGroupTGTypes...),
			},
		},
		"description": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Description (10–600 characters per notification-srv validation).",
			Validators: []validator.String{
				stringvalidator.LengthBetween(10, 600),
			},
		},
		"default_locale": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Default locale (BCP47), e.g. `en`, `de`.",
		},
		"owner": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("client"),
			MarkdownDescription: "Object owner, e.g. `client`.",
		},
		"comm_setting_email": schema.SingleNestedAttribute{
			Optional: true,
			MarkdownDescription: "EMAIL `commSettings` entry. On **update**, other channels are merged from the existing group; you may set only the channels you change. " +
				"On **create**, notification-srv requires all four channels unless `commSettings` is omitted (server fills from copy source).",
			Attributes: commSettingChannelAttrs,
		},
		"comm_setting_sms": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "SMS `commSettings` entry.",
			Attributes:          commSettingChannelAttrs,
		},
		"comm_setting_ivr": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "IVR `commSettings` entry.",
			Attributes:          commSettingChannelAttrs,
		},
		"comm_setting_push": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: "PUSH `commSettings` entry.",
			Attributes:          commSettingChannelAttrs,
		},
	},
}

type notificationsTemplateGroupModel struct {
	ID               types.String `tfsdk:"id"`
	GroupID          types.String `tfsdk:"group_id"`
	TGType           types.String `tfsdk:"tg_type"`
	Description      types.String `tfsdk:"description"`
	DefaultLocale    types.String `tfsdk:"default_locale"`
	Owner            types.String `tfsdk:"owner"`
	CommSettingEmail types.Object `tfsdk:"comm_setting_email"`
	CommSettingSMS   types.Object `tfsdk:"comm_setting_sms"`
	CommSettingIVR   types.Object `tfsdk:"comm_setting_ivr"`
	CommSettingPush  types.Object `tfsdk:"comm_setting_push"`
}

type commSettingModel struct {
	ServiceSetupID     types.String `tfsdk:"service_setup_id"`
	SenderName         types.String `tfsdk:"sender_name"`
	SenderAddress      types.String `tfsdk:"sender_address"`
	ReplyTo            types.String `tfsdk:"reply_to"`
	HasRemoteTemplates types.Bool   `tfsdk:"has_remote_templates"`
}

func (r *NotificationsTemplateGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationsTemplateGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiReq, diags := buildNotificationsTemplateGroupRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.cidaasClient.NotificationsSrvTemplateGroup.Create(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create notifications template group", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "created notifications template group", util.H{"group_id": plan.GroupID.ValueString()})

	res, err := r.cidaasClient.NotificationsSrvTemplateGroup.Get(ctx, plan.GroupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read notifications template group after create", util.FormatErrorMessage(err))
		return
	}
	state := notificationsDataToModel(res, plan)
	mergeCommSettingsFromPlan(ctx, &state, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NotificationsTemplateGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationsTemplateGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.cidaasClient.NotificationsSrvTemplateGroup.Get(ctx, state.GroupID.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		resp.Diagnostics.AddError("failed to read notifications template group", util.FormatErrorMessage(err))
		return
	}
	out := notificationsDataToModel(res, state)
	// Same merge as after create/update: refresh must not re-inject API-only comm defaults into state
	// when config/state omits channels or optional fields (avoids perpetual plan churn).
	mergeCommSettingsFromPlan(ctx, &out, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *NotificationsTemplateGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan notificationsTemplateGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, diags := buildNotificationsTemplateGroupRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.cidaasClient.NotificationsSrvTemplateGroup.Update(ctx, plan.GroupID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("failed to update notifications template group", util.FormatErrorMessage(err))
		return
	}

	res, err := r.cidaasClient.NotificationsSrvTemplateGroup.Get(ctx, plan.GroupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read notifications template group after update", util.FormatErrorMessage(err))
		return
	}
	state := notificationsDataToModel(res, plan)
	mergeCommSettingsFromPlan(ctx, &state, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NotificationsTemplateGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationsTemplateGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cidaasClient.NotificationsSrvTemplateGroup.Delete(ctx, state.GroupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete notifications template group", util.FormatErrorMessage(err))
		return
	}
}

func (r *NotificationsTemplateGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("group_id"), req, resp)
}

func buildNotificationsTemplateGroupRequest(ctx context.Context, m notificationsTemplateGroupModel) (cidaas.NotificationsSrvTemplateGroupRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := cidaas.NotificationsSrvTemplateGroupRequest{
		ID:            m.GroupID.ValueString(),
		TGType:        m.TGType.ValueString(),
		Description:   m.Description.ValueString(),
		Owner:         m.Owner.ValueString(),
		DefaultLocale: m.DefaultLocale.ValueString(),
	}

	comm, d := commSettingsFromModel(ctx, m)
	diags.Append(d...)
	if len(comm) > 0 {
		req.CommSettings = comm
	}

	return req, diags
}

func commSettingsFromModel(ctx context.Context, m notificationsTemplateGroupModel) (map[string]cidaas.NotificationsSrvCommSetting, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := make(map[string]cidaas.NotificationsSrvCommSetting)
	// Map keys must match notification-srv JSON (lowercase channel names); values still set communicationMethod per API.
	channels := []struct {
		mapKey string
		obj    types.Object
		method string
	}{
		{"email", m.CommSettingEmail, "EMAIL"},
		{"sms", m.CommSettingSMS, "SMS"},
		{"ivr", m.CommSettingIVR, "IVR"},
		{"push", m.CommSettingPush, "PUSH"},
	}
	anySet := false
	for _, ch := range channels {
		if ch.obj.IsNull() || ch.obj.IsUnknown() {
			continue
		}
		anySet = true
		var cm commSettingModel
		diags.Append(ch.obj.As(ctx, &cm, basetypes.ObjectAsOptions{})...)
		cs := cidaas.NotificationsSrvCommSetting{
			CommunicationMethod: ch.method,
			ServiceSetupID:      cm.ServiceSetupID.ValueString(),
			SenderName:          cm.SenderName.ValueString(),
			SenderAddress:       cm.SenderAddress.ValueString(),
			ReplyTo:             cm.ReplyTo.ValueString(),
		}
		if !cm.HasRemoteTemplates.IsNull() {
			v := cm.HasRemoteTemplates.ValueBool()
			cs.HasRemoteTemplates = &v
		}
		out[ch.mapKey] = cs
	}
	if anySet {
		for k, v := range out {
			if v.ServiceSetupID == "" {
				diags.AddError(
					"Invalid configuration",
					fmt.Sprintf("comm_setting_* for %s requires service_setup_id", k),
				)
			}
		}
	}
	return out, diags
}

func notificationsDataToModel(data *cidaas.NotificationsSrvTemplateGroupData, _ notificationsTemplateGroupModel) notificationsTemplateGroupModel {
	m := notificationsTemplateGroupModel{
		ID:            util.StringValueOrNull(&data.ID),
		GroupID:       util.StringValueOrNull(&data.ID),
		TGType:        util.StringValueOrNull(&data.TGType),
		Description:   util.StringValueOrNull(&data.Description),
		DefaultLocale: util.StringValueOrNull(&data.DefaultLocale),
		Owner:         util.StringValueOrNull(&data.Owner),
	}

	if data.CommSettings != nil {
		// JSON uses lowercase map keys (e.g. "email"); older code used "EMAIL" and never matched → null state / inconsistent apply.
		m.CommSettingEmail = commSettingObjectFromAPI(commSettingFromMap(data.CommSettings, "email"))
		m.CommSettingSMS = commSettingObjectFromAPI(commSettingFromMap(data.CommSettings, "sms"))
		m.CommSettingIVR = commSettingObjectFromAPI(commSettingFromMap(data.CommSettings, "ivr"))
		m.CommSettingPush = commSettingObjectFromAPI(commSettingFromMap(data.CommSettings, "push"))
	}
	return m
}

func commSettingFromMap(m map[string]cidaas.NotificationsSrvCommSetting, channel string) cidaas.NotificationsSrvCommSetting {
	if m == nil {
		return cidaas.NotificationsSrvCommSetting{}
	}
	ch := strings.TrimSpace(channel)
	lower := strings.ToLower(ch)
	upper := strings.ToUpper(ch)
	for _, k := range []string{lower, upper, ch} {
		if cs, ok := m[k]; ok {
			return cs
		}
	}
	for _, cs := range m {
		if strings.EqualFold(strings.TrimSpace(cs.CommunicationMethod), ch) {
			return cs
		}
	}
	return cidaas.NotificationsSrvCommSetting{}
}

func commSettingObjectAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"service_setup_id":     types.StringType,
		"sender_name":          types.StringType,
		"sender_address":       types.StringType,
		"reply_to":             types.StringType,
		"has_remote_templates": types.BoolType,
	}
}

func commSettingObjectFromAPI(cs cidaas.NotificationsSrvCommSetting) types.Object {
	cm := commSettingModel{
		ServiceSetupID:     util.StringValueOrNull(nonEmptyStr(cs.ServiceSetupID)),
		SenderName:         util.StringValueOrNull(nonEmptyStr(cs.SenderName)),
		SenderAddress:      util.StringValueOrNull(nonEmptyStr(cs.SenderAddress)),
		ReplyTo:            util.StringValueOrNull(nonEmptyStr(cs.ReplyTo)),
		HasRemoteTemplates: boolValueOrNull(cs.HasRemoteTemplates),
	}
	// Empty API struct → null object (matches prior behavior).
	if cs.ServiceSetupID == "" && cs.SenderName == "" && cs.SenderAddress == "" && cs.ReplyTo == "" && cs.HasRemoteTemplates == nil {
		return types.ObjectNull(commSettingObjectAttrTypes())
	}
	return commSettingModelToObject(cm)
}

func commSettingModelToObject(cm commSettingModel) types.Object {
	attrTypes := commSettingObjectAttrTypes()
	if cm.ServiceSetupID.IsNull() && cm.SenderName.IsNull() && cm.SenderAddress.IsNull() && cm.ReplyTo.IsNull() && cm.HasRemoteTemplates.IsNull() {
		return types.ObjectNull(attrTypes)
	}
	vals := map[string]attr.Value{
		"service_setup_id":     cm.ServiceSetupID,
		"sender_name":          cm.SenderName,
		"sender_address":       cm.SenderAddress,
		"reply_to":             cm.ReplyTo,
		"has_remote_templates": cm.HasRemoteTemplates,
	}
	return types.ObjectValueMust(attrTypes, vals)
}

// mergeCommSettingsFromPlan aligns the model with the apply plan (or prior state on Read) so API-enriched comm
// settings do not contradict Terraform: omitted channels stay null; optional fields that are null in plan/state
// stay null; configured ids are kept when the API returns a canonical replacement.
func mergeCommSettingsFromPlan(ctx context.Context, m *notificationsTemplateGroupModel, plan notificationsTemplateGroupModel) {
	m.CommSettingEmail = mergeCommSettingObject(ctx, m.CommSettingEmail, plan.CommSettingEmail)
	m.CommSettingSMS = mergeCommSettingObject(ctx, m.CommSettingSMS, plan.CommSettingSMS)
	m.CommSettingIVR = mergeCommSettingObject(ctx, m.CommSettingIVR, plan.CommSettingIVR)
	m.CommSettingPush = mergeCommSettingObject(ctx, m.CommSettingPush, plan.CommSettingPush)
}

func mergeCommSettingObject(ctx context.Context, api, plan types.Object) types.Object {
	attrTypes := commSettingObjectAttrTypes()
	// Planned: no comm_setting_* block → state must stay null (not API-filled defaults for other channels).
	if plan.IsNull() {
		return types.ObjectNull(attrTypes)
	}
	if plan.IsUnknown() {
		return api
	}
	if api.IsNull() || api.IsUnknown() {
		return plan
	}
	var a, p commSettingModel
	if diags := api.As(ctx, &a, basetypes.ObjectAsOptions{}); diags.HasError() {
		return plan
	}
	if diags := plan.As(ctx, &p, basetypes.ObjectAsOptions{}); diags.HasError() {
		return plan
	}
	merged := commSettingModel{
		ServiceSetupID:     mergeCommAttrStringPreferPlan(p.ServiceSetupID, a.ServiceSetupID),
		SenderName:         mergeCommAttrStringPreferPlan(p.SenderName, a.SenderName),
		SenderAddress:      mergeCommAttrStringPreferPlan(p.SenderAddress, a.SenderAddress),
		ReplyTo:            mergeCommAttrStringPreferPlan(p.ReplyTo, a.ReplyTo),
		HasRemoteTemplates: mergeCommAttrBoolPreferPlan(p.HasRemoteTemplates, a.HasRemoteTemplates),
	}
	return commSettingModelToObject(merged)
}

func mergeCommAttrStringPreferPlan(plan, api types.String) types.String {
	if !plan.IsUnknown() {
		return plan
	}
	return api
}

func mergeCommAttrBoolPreferPlan(plan, api types.Bool) types.Bool {
	if !plan.IsUnknown() {
		return plan
	}
	return api
}

func nonEmptyStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func boolValueOrNull(b *bool) types.Bool {
	if b == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*b)
}
