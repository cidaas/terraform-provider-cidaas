package notification

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	commMethodValueRegexp = regexp.MustCompile(`(?i)^(email|sms|ivr|push)$`)
	allowedCategories     = []string{"cidaas", "custom"}
	templateTypeOwners    = []string{"client", "admin", "core", "system"}
)

//nolint:revive
type TemplateTypeResource struct {
	base.BaseResource
}

func NewNotificationTemplateTypeResource() resource.Resource {
	return &TemplateTypeResource{
		BaseResource: base.NewBaseResource(
			base.BaseResourceConfig{
				Name:   base.RESOURCE_NOTIFICATION_TEMPLATE_TYPE,
				Schema: &templateTypeSchema,
			},
		),
	}
}

type TemplateTypeConfig struct {
	ID                   types.String `tfsdk:"id"`
	TemplateKey          types.String `tfsdk:"template_key"`
	Category             types.String `tfsdk:"category"`
	Description          types.String `tfsdk:"description"`
	Deactivatable        types.Bool   `tfsdk:"deactivatable"`
	SystemAttributes     types.Map    `tfsdk:"system_attributes"`
	CustomAttributes     types.Map    `tfsdk:"custom_attributes"`
	ContextAttributes    types.Map    `tfsdk:"context_attributes"`
	ProcessingTypes      types.Set    `tfsdk:"processing_types"`
	UsageTypes           types.Set    `tfsdk:"usage_types"`
	VerificationTypes    types.Set    `tfsdk:"verification_types"`
	CommunicationMethods types.Set    `tfsdk:"communication_methods"`
	TemplateGroupIDs     types.Set    `tfsdk:"template_group_ids"`
	MsgFormats           types.Set    `tfsdk:"msg_formats"`
	Owner                types.String `tfsdk:"owner"`
	CreatedTime          types.String `tfsdk:"created_time"`
	UpdatedTime          types.String `tfsdk:"updated_time"`
}

var templateTypeSchema = schema.Schema{
	MarkdownDescription: "The `cidaas_notification_template_type` resource manages template type definitions in the cidaas system." +
		" Template Types define which attributes can be used in templates and which communication methods are supported." +
		"\n\n**Important Notes:**" +
		"\n- System Template Types (category: `cidaas`) are pre-provisioned and cannot be created via Terraform" +
		"\n- For System Template Types, only `custom_attributes` can be modified" +
		"\n- Custom Template Types (category: `custom`) can be fully created and managed",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The unique identifier of the template type resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"template_key": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The key of the template type (e.g. `VERIFY_USER`). Cannot be modified once created.",
			Validators: []validator.String{
				stringvalidator.RegexMatches(
					regexp.MustCompile(`^[A-Z0-9_]+$`),
					"must contain only uppercase letters, numbers, and underscores",
				),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"category": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("custom"),
			MarkdownDescription: "Category of the template type: `cidaas` (system) or `custom`.",
			Validators: []validator.String{
				stringvalidator.OneOf(allowedCategories...),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"description": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Description of what this template type is used for.",
		},
		"deactivatable": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(true),
			MarkdownDescription: "Whether templates of this type can be deactivated.",
		},
		"system_attributes": schema.MapAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Map of system attribute names to their descriptions.",
		},
		"custom_attributes": schema.MapAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Map of custom attribute names to their descriptions.",
		},
		"context_attributes": schema.MapAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Map of context attribute names to their descriptions.",
		},
		"processing_types": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Supported processing types (e.g. `SYNC`, `ASYNC`).",
		},
		"usage_types": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Supported usage types.",
		},
		"verification_types": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Supported verification types.",
		},
		"communication_methods": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Supported communication methods: `email`, `sms`, `ivr`, `push`.",
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(stringvalidator.RegexMatches(
					commMethodValueRegexp,
					"must be one of: email, sms, ivr, push (case-insensitive)",
				)),
			},
		},
		"template_group_ids": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Template group IDs associated with this template type.",
		},
		"msg_formats": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Supported message formats: `html`, `text`, `media`.",
		},
		"owner": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("client"),
			MarkdownDescription: "Owner of the template type (`client`, `admin`, `core`, `system`).",
			Validators: []validator.String{
				stringvalidator.OneOf(templateTypeOwners...),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"created_time": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Creation timestamp.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_time": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Last update timestamp.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	},
}

func (r *TemplateTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TemplateTypeConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Category.ValueString() == "cidaas" {
		resp.Diagnostics.AddError(
			"Cannot create System Template Type",
			"Template types with category 'cidaas' are system template types and cannot be created via Terraform.",
		)
		return
	}

	model := prepareTemplateTypeModel(ctx, plan)
	response, err := r.CidaasClient.TemplateType.Upsert(*model)
	if err != nil {
		resp.Diagnostics.AddError("failed to create template type", util.FormatErrorMessage(err))
		return
	}

	updateStateFromModel(ctx, &plan, response.Data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TemplateTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TemplateTypeConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.CidaasClient.TemplateType.Get(state.TemplateKey.ValueString())
	if err != nil {
		if base.ReadHandleNotFound(ctx, resp, err) {
			return
		}
		resp.Diagnostics.AddError("failed to read template type", util.FormatErrorMessage(err))
		return
	}

	updateStateFromModel(ctx, &state, response.Data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TemplateTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state TemplateTypeConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	isSystemType := state.Category.ValueString() == "cidaas" || state.Owner.ValueString() == "SYSTEM"

	if isSystemType {
		if !plan.CustomAttributes.IsNull() && !plan.CustomAttributes.Equal(state.CustomAttributes) {
			customAttrs := map[string]string{}
			if !plan.CustomAttributes.IsNull() {
				diags := plan.CustomAttributes.ElementsAs(ctx, &customAttrs, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}

			patch := cidaas.TemplateTypePatchModel{
				ID:               plan.TemplateKey.ValueString(),
				CustomAttributes: &customAttrs,
			}

			response, err := r.CidaasClient.TemplateType.Patch(patch)
			if err != nil {
				resp.Diagnostics.AddError("failed to update template type", util.FormatErrorMessage(err))
				return
			}

			updateStateFromModel(ctx, &plan, response.Data)
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}

		resp.Diagnostics.AddWarning(
			"No changes allowed",
			fmt.Sprintf("Template Type '%s' is a system template type. Only custom_attributes can be modified.",
				plan.TemplateKey.ValueString()),
		)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	model := prepareTemplateTypeModel(ctx, plan)
	model.ID = state.TemplateKey.ValueString()
	response, err := r.CidaasClient.TemplateType.Upsert(*model)
	if err != nil {
		resp.Diagnostics.AddError("failed to update template type", util.FormatErrorMessage(err))
		return
	}

	updateStateFromModel(ctx, &plan, response.Data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TemplateTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TemplateTypeConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Category.ValueString() == "cidaas" || state.Owner.ValueString() == "SYSTEM" {
		resp.Diagnostics.AddWarning(
			"System Template Type cannot be deleted",
			fmt.Sprintf("Template Type '%s' is a system template type and cannot be deleted via Terraform.",
				state.TemplateKey.ValueString()),
		)
		return
	}

	err := r.CidaasClient.TemplateType.Delete(state.TemplateKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete template type", util.FormatErrorMessage(err))
		return
	}
}

func (r *TemplateTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := strings.ToUpper(req.ID)
	if id == "" {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			"Expected import identifier to be the template_key (e.g., VERIFY_USER)",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("template_key"), id)...)
}

func prepareTemplateTypeModel(ctx context.Context, config TemplateTypeConfig) *cidaas.TemplateTypeModel {
	model := &cidaas.TemplateTypeModel{
		ID:          config.TemplateKey.ValueString(),
		Category:    config.Category.ValueString(),
		Description: config.Description.ValueString(),
	}

	if !config.Deactivatable.IsNull() {
		model.Deactivatable = config.Deactivatable.ValueBool()
	}

	if !config.SystemAttributes.IsNull() {
		config.SystemAttributes.ElementsAs(ctx, &model.SystemAttributes, false)
	}
	if !config.CustomAttributes.IsNull() {
		config.CustomAttributes.ElementsAs(ctx, &model.CustomAttributes, false)
	}
	if !config.ContextAttributes.IsNull() {
		config.ContextAttributes.ElementsAs(ctx, &model.ContextAttributes, false)
	}

	if !config.ProcessingTypes.IsNull() {
		config.ProcessingTypes.ElementsAs(ctx, &model.ProcessingTypes, false)
	}
	if !config.UsageTypes.IsNull() {
		config.UsageTypes.ElementsAs(ctx, &model.UsageTypes, false)
	}
	if !config.VerificationTypes.IsNull() {
		config.VerificationTypes.ElementsAs(ctx, &model.VerificationTypes, false)
	}
	if !config.CommunicationMethods.IsNull() {
		var comm []string
		config.CommunicationMethods.ElementsAs(ctx, &comm, false)
		for i := range comm {
			comm[i] = strings.ToLower(strings.TrimSpace(comm[i]))
		}
		model.CommunicationMethods = comm
	}
	if !config.TemplateGroupIDs.IsNull() {
		config.TemplateGroupIDs.ElementsAs(ctx, &model.TemplateGroupIDs, false)
	}
	if !config.MsgFormats.IsNull() {
		var mf []string
		config.MsgFormats.ElementsAs(ctx, &mf, false)
		for i := range mf {
			mf[i] = strings.ToLower(strings.TrimSpace(mf[i]))
		}
		model.MsgFormats = mf
	}
	if !config.Owner.IsNull() && !config.Owner.IsUnknown() {
		model.Owner = strings.ToLower(strings.TrimSpace(config.Owner.ValueString()))
	} else {
		model.Owner = "client"
	}

	return model
}

func updateStateFromModel(ctx context.Context, state *TemplateTypeConfig, model cidaas.TemplateTypeModel) {
	state.ID = util.StringValueOrNull(&model.ID)
	state.TemplateKey = types.StringValue(model.ID)
	state.Category = util.StringValueOrNull(&model.Category)
	state.Description = util.StringValueOrNull(&model.Description)
	state.Deactivatable = types.BoolValue(model.Deactivatable)
	if model.Owner != "" {
		state.Owner = types.StringValue(strings.ToLower(model.Owner))
	} else {
		state.Owner = types.StringNull()
	}
	state.CreatedTime = util.StringValueOrNull(&model.CreatedTime)
	state.UpdatedTime = util.StringValueOrNull(&model.UpdatedTime)

	if model.SystemAttributes != nil {
		sysAttrs, diags := types.MapValueFrom(ctx, types.StringType, model.SystemAttributes)
		if !diags.HasError() {
			state.SystemAttributes = sysAttrs
		}
	}
	if model.CustomAttributes != nil {
		custAttrs, diags := types.MapValueFrom(ctx, types.StringType, model.CustomAttributes)
		if !diags.HasError() {
			state.CustomAttributes = custAttrs
		}
	}
	if model.ContextAttributes != nil {
		ctxAttrs, diags := types.MapValueFrom(ctx, types.StringType, model.ContextAttributes)
		if !diags.HasError() {
			state.ContextAttributes = ctxAttrs
		}
	}

	if model.ProcessingTypes != nil {
		procTypes, diags := types.SetValueFrom(ctx, types.StringType, model.ProcessingTypes)
		if !diags.HasError() {
			state.ProcessingTypes = procTypes
		}
	}
	if model.UsageTypes != nil {
		usageTypes, diags := types.SetValueFrom(ctx, types.StringType, model.UsageTypes)
		if !diags.HasError() {
			state.UsageTypes = usageTypes
		}
	}
	if model.VerificationTypes != nil {
		verifTypes, diags := types.SetValueFrom(ctx, types.StringType, model.VerificationTypes)
		if !diags.HasError() {
			state.VerificationTypes = verifTypes
		}
	}
	if model.CommunicationMethods != nil {
		commLower := make([]string, len(model.CommunicationMethods))
		for i, m := range model.CommunicationMethods {
			commLower[i] = strings.ToLower(strings.TrimSpace(m))
		}
		commMethods, diags := types.SetValueFrom(ctx, types.StringType, commLower)
		if !diags.HasError() {
			state.CommunicationMethods = commMethods
		}
	}
	if model.TemplateGroupIDs != nil {
		tgIDs, diags := types.SetValueFrom(ctx, types.StringType, model.TemplateGroupIDs)
		if !diags.HasError() {
			state.TemplateGroupIDs = tgIDs
		}
	}
	if model.MsgFormats != nil {
		msgFormats, diags := types.SetValueFrom(ctx, types.StringType, model.MsgFormats)
		if !diags.HasError() {
			state.MsgFormats = msgFormats
		}
	}
}
