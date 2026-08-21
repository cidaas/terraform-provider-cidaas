package resources

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type TemplateGroupResource struct {
	BaseResource
}

func NewTemplateGroupResource() resource.Resource {
	return &TemplateGroupResource{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_TEMPLATE_GROUP,
				Schema: &templateGroupSchema,
			},
		),
	}
}

type TemplateGroupConfig struct {
	ID                types.String `tfsdk:"id"`
	GroupID           types.String `tfsdk:"group_id"`
	EmailSenderConfig types.Object `tfsdk:"email_sender_config"`
	SMSSenderConfig   types.Object `tfsdk:"sms_sender_config"`
	IVRSenderConfig   types.Object `tfsdk:"ivr_sender_config"`
	PushSenderConfig  types.Object `tfsdk:"push_sender_config"`

	emailSenderConfig *EmailSenderConfig
	smsSenderConfig   *SMSSenderConfig
	ivrSenderConfig   *IVRSenderConfig
	pushSenderConfig  *IVRSenderConfig
}

type EmailSenderConfig struct {
	FromEmail   types.String `tfsdk:"from_email"`
	FromName    types.String `tfsdk:"from_name"`
	ReplyTo     types.String `tfsdk:"reply_to"`
	SenderNames types.Set    `tfsdk:"sender_names"`
}

type SMSSenderConfig struct {
	FromName    types.String `tfsdk:"from_name"`
	SenderNames types.Set    `tfsdk:"sender_names"`
}

type IVRSenderConfig struct {
	SenderNames types.Set `tfsdk:"sender_names"`
}

var templateGroupSchema = schema.Schema{
	MarkdownDescription: "**Deprecated for new designs:** this resource uses legacy **templates-srv/groups**. " +
		"For notification-srv template groups, use **`cidaas_notifications_template_group`**.\n\n" +
		"The cidaas_template_group resource in the provider is used to define and manage templates groups within the Cidaas system." +
		" Template Groups categorize your communication templates allowing you to map preferred templates to specific clients effectively." +
		"\n\n Ensure that the below scopes are assigned to the client with the specified `client_id`:" +
		"\n- cidaas:templates_read" +
		"\n- cidaas:templates_write" +
		"\n- cidaas:templates_delete",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The ID of the resource",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"group_id": schema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.LengthAtMost(15),
			},
			MarkdownDescription: "The group_id of the Template Group. The group_id is used to import an existing template group." +
				" The maximum allowed length of a group_id is **15** characters.",
		},
		"email_sender_config": schema.SingleNestedAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			MarkdownDescription: "The `email_sender_config` is used to configure your email sender.",
			Attributes: map[string]schema.Attribute{
				"from_email": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "The email from address from which the emails will be sent when the specific group is configured.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"from_name": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "The `from_name` attribute is the display name that appears in the 'From' field of the emails.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"reply_to": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "The `reply_to` attribute is the email address where replies should be directed.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"sender_names": schema.SetAttribute{
					ElementType:         types.StringType,
					Optional:            true,
					MarkdownDescription: "The `sender_names` attribute defines the names associated with email senders.",
				},
			},
		},
		"sms_sender_config": schema.SingleNestedAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The configuration of the SMS sender.",
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: map[string]schema.Attribute{
				"from_name": schema.StringAttribute{
					Optional: true,
				},
				"sender_names": schema.SetAttribute{
					ElementType: types.StringType,
					Optional:    true,
				},
			},
		},
		"ivr_sender_config": schema.SingleNestedAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The configuration of the IVR sender.",
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: map[string]schema.Attribute{
				"sender_names": schema.SetAttribute{
					ElementType: types.StringType,
					Optional:    true,
				},
			},
		},
		"push_sender_config": schema.SingleNestedAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The configuration of the PUSH notification sender.",
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: map[string]schema.Attribute{
				"sender_names": schema.SetAttribute{
					ElementType: types.StringType,
					Optional:    true,
				},
			},
		},
	},
}

func (tg *TemplateGroupConfig) ExtractConfigs(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	if !tg.EmailSenderConfig.IsNull() && !tg.EmailSenderConfig.IsUnknown() {
		tg.emailSenderConfig = &EmailSenderConfig{}
		diags = tg.EmailSenderConfig.As(ctx, tg.emailSenderConfig, basetypes.ObjectAsOptions{})
	}
	if !tg.SMSSenderConfig.IsNull() && !tg.SMSSenderConfig.IsUnknown() {
		tg.smsSenderConfig = &SMSSenderConfig{}
		diags = tg.SMSSenderConfig.As(ctx, tg.smsSenderConfig, basetypes.ObjectAsOptions{})
	}
	if !tg.IVRSenderConfig.IsNull() && !tg.IVRSenderConfig.IsUnknown() {
		tg.ivrSenderConfig = &IVRSenderConfig{}
		diags = tg.IVRSenderConfig.As(ctx, tg.ivrSenderConfig, basetypes.ObjectAsOptions{})
	}
	if !tg.PushSenderConfig.IsNull() && !tg.PushSenderConfig.IsUnknown() {
		tg.pushSenderConfig = &IVRSenderConfig{}
		diags = tg.PushSenderConfig.As(ctx, tg.pushSenderConfig, basetypes.ObjectAsOptions{})
	}
	return diags
}

func (r *TemplateGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TemplateGroupConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.ExtractConfigs(ctx)...)
	tgModel, diags := prepareTemplateGroupModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to prepare template group model", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	res, err := r.cidaasClient.TemplateGroup.Create(ctx, *tgModel)
	if err != nil {
		tflog.Error(ctx, "failed to create template group via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create template group", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully created template group via API", util.H{
		"group_id": res.Data.GroupID,
	})

	res, err = r.cidaasClient.TemplateGroup.Get(ctx, res.Data.GroupID)
	if err != nil {
		tflog.Error(ctx, "failed to get template group after creation", util.H{
			"group_id": res.Data.GroupID,
			"error":    err.Error(),
		})
		resp.Diagnostics.AddError("failed to get template group", util.FormatErrorMessage(err))
		return
	}

	updatedPlan := updateState(&plan, *res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedPlan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "resource template group created successfully", util.H{
		"group_id": res.Data.GroupID,
	})
}

func (r *TemplateGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TemplateGroupConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	res, err := r.cidaasClient.TemplateGroup.Get(ctx, state.GroupID.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		tflog.Error(ctx, "failed to read template group via API", util.H{
			"group_id": state.GroupID.ValueString(),
			"error":    err.Error(),
		})
		resp.Diagnostics.AddError("failed to read template group", util.FormatErrorMessage(err))
		return
	}

	updatedState := updateState(&state, *res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Debug(ctx, "resource template group read successfully", util.H{
		"group_id": res.Data.GroupID,
	})
}

func (r *TemplateGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) { //nolint:dupl
	var plan, state TemplateGroupConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.ExtractConfigs(ctx)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	templateGroupModel, diags := prepareTemplateGroupModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to prepare template group model for update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	templateGroupModel.ID = state.ID.ValueString()
	_, err := r.cidaasClient.TemplateGroup.Update(ctx, *templateGroupModel)
	if err != nil {
		tflog.Error(ctx, "failed to update template group via API", util.H{
			"group_id": state.ID.ValueString(),
			"error":    err.Error(),
		})
		resp.Diagnostics.AddError("failed to update template group", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully updated template group via API", util.H{
		"group_id": state.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state after update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Debug(ctx, "resource template group updated successfully", util.H{
		"group_id": state.ID.ValueString(),
	})
}

func (r *TemplateGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) { //nolint:dupl
	var state TemplateGroupConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data for deletion", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	err := r.cidaasClient.TemplateGroup.Delete(ctx, state.GroupID.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to delete template group via API", util.H{
			"group_id": state.GroupID.ValueString(),
			"error":    err.Error(),
		})
		resp.Diagnostics.AddError("failed to delete template group", util.FormatErrorMessage(err))
		return
	}

	tflog.Info(ctx, "resource template group deleted successfully", util.H{
		"group_id": state.GroupID.ValueString(),
	})
}

func (r *TemplateGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("group_id"), req, resp)
}

func prepareTemplateGroupModel(ctx context.Context, plan TemplateGroupConfig) (*cidaas.TemplateGroupModel, diag.Diagnostics) {
	var tgModel cidaas.TemplateGroupModel
	tgModel.GroupID = plan.GroupID.ValueString()
	if !plan.EmailSenderConfig.IsNull() && !plan.EmailSenderConfig.IsUnknown() {
		tgModel.EmailSenderConfig = &cidaas.EmailSenderConfig{
			FromEmail: plan.emailSenderConfig.FromEmail.ValueString(),
			FromName:  plan.emailSenderConfig.FromName.ValueString(),
			ReplyTo:   plan.emailSenderConfig.ReplyTo.ValueString(),
		}
		diag := plan.emailSenderConfig.SenderNames.ElementsAs(ctx, &tgModel.EmailSenderConfig.SenderNames, false)
		if diag.HasError() {
			return nil, diag
		}
	}
	if !plan.SMSSenderConfig.IsNull() && !plan.SMSSenderConfig.IsUnknown() {
		tgModel.SMSSenderConfig = &cidaas.SMSSenderConfig{
			FromName: plan.smsSenderConfig.FromName.ValueString(),
		}
		diag := plan.smsSenderConfig.SenderNames.ElementsAs(ctx, &tgModel.SMSSenderConfig.SenderNames, false)
		if diag.HasError() {
			return nil, diag
		}
	}
	if !plan.IVRSenderConfig.IsNull() && !plan.IVRSenderConfig.IsUnknown() {
		tgModel.IVRSenderConfig = &cidaas.IVRSenderConfig{}
		diag := plan.ivrSenderConfig.SenderNames.ElementsAs(ctx, &tgModel.IVRSenderConfig.SenderNames, false)
		if diag.HasError() {
			return nil, diag
		}
	}
	if !plan.PushSenderConfig.IsNull() && !plan.PushSenderConfig.IsUnknown() {
		tgModel.PushSenderConfig = &cidaas.IVRSenderConfig{}
		diag := plan.pushSenderConfig.SenderNames.ElementsAs(ctx, &tgModel.PushSenderConfig.SenderNames, false)
		if diag.HasError() {
			return nil, diag
		}
	}
	return &tgModel, nil
}

func updateState(state *TemplateGroupConfig, res cidaas.TemplateGroupResponse) *TemplateGroupConfig {
	state.ID = util.StringValueOrNull(&res.Data.GroupID)
	if res.Data.EmailSenderConfig != nil {
		state.EmailSenderConfig = types.ObjectValueMust(
			map[string]attr.Type{
				"from_email":   types.StringType,
				"from_name":    types.StringType,
				"reply_to":     types.StringType,
				"sender_names": types.SetType{ElemType: types.StringType},
			},
			map[string]attr.Value{
				"from_email": util.StringValueOrNull(&res.Data.EmailSenderConfig.FromEmail),
				"from_name":  util.StringValueOrNull(&res.Data.EmailSenderConfig.FromName),
				"reply_to": func() basetypes.StringValue {
					if res.Data.EmailSenderConfig.ReplyTo == "" {
						return util.StringValueOrNull(nil)
					}
					return util.StringValueOrNull(&res.Data.EmailSenderConfig.ReplyTo)
				}(),
				"sender_names": util.SetValueOrNull(res.Data.EmailSenderConfig.SenderNames),
			},
		)
	}
	if res.Data.SMSSenderConfig != nil {
		state.SMSSenderConfig = types.ObjectValueMust(
			map[string]attr.Type{
				"from_name":    types.StringType,
				"sender_names": types.SetType{ElemType: types.StringType},
			},
			map[string]attr.Value{
				"from_name":    util.StringValueOrNull(&res.Data.SMSSenderConfig.FromName),
				"sender_names": util.SetValueOrNull(res.Data.SMSSenderConfig.SenderNames),
			},
		)
	}
	if res.Data.IVRSenderConfig != nil {
		state.IVRSenderConfig = types.ObjectValueMust(
			map[string]attr.Type{
				"sender_names": types.SetType{ElemType: types.StringType},
			},
			map[string]attr.Value{
				"sender_names": util.SetValueOrNull(res.Data.IVRSenderConfig.SenderNames),
			},
		)
	}
	if res.Data.PushSenderConfig != nil {
		state.PushSenderConfig = types.ObjectValueMust(
			map[string]attr.Type{
				"sender_names": types.SetType{ElemType: types.StringType},
			},
			map[string]attr.Value{
				"sender_names": util.SetValueOrNull(res.Data.PushSenderConfig.SenderNames),
			},
		)
	}
	return state
}
