package resources

import (
	"context"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var notificationServiceSetupCommMethods = []string{"email", "sms", "ivr", "push"}

type NotificationServiceSetupResource struct {
	BaseResource
}

func NewNotificationServiceSetupResource() resource.Resource {
	return &NotificationServiceSetupResource{
		BaseResource: NewBaseResource(BaseResourceConfig{
			Name:   RESOURCE_NOTIFICATION_SERVICE_SETUP,
			Schema: &notificationServiceSetupSchema,
		}),
	}
}

var notificationServiceSetupSchema = schema.Schema{
	MarkdownDescription: "Manages a **communication provider service setup** via **notification-srv** (`POST/PATCH/DELETE /{notifications_context_path}/servicesetups`). " +
		"notification-srv proxies to mplace-srv; the tenant is taken from your instance token — do **not** set `saas_instance_id`.\n\n" +
		"**`status`** is computed from `GET` and reflects manual verification in service-desk (e.g. `in-progress` → `active`). Terraform does **not** call verify.\n\n" +
		"Pair with **`cidaas_notification_provider_config`** for credentials (`config_data_wo` + `schemaData`).\n\n" +
		"**Scopes:** `cidaas:service_setups_read`, `cidaas:service_setups_write`, `cidaas:service_setups_delete`.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Service setup `_id` from notification-srv.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Human-readable name.",
		},
		"description": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Optional description.",
		},
		"status": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Setup status from the API (`in-progress`, `active`, `inactive`). Updated on read/plan refresh after manual verify.",
		},
		"service_id": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Service id (`serviceDescInfo.serviceId`). Immutable after create.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"service_category": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("comm_prov"),
			MarkdownDescription: "Service category (`serviceDescInfo.serviceCategory`). Default: `comm_prov`.",
		},
		"communication_methods": schema.SetAttribute{
			Required:    true,
			ElementType: types.StringType,
			MarkdownDescription: "Communication methods for this setup (lowercase: `email`, `sms`, `ivr`, `push`). " +
				"Stored in `serviceDescInfo.commProv.commMethods`.",
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(
					stringvalidator.OneOf(notificationServiceSetupCommMethods...),
				),
			},
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.RequiresReplace(),
			},
		},
		"parent_service_setup_id": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Optional parent service setup id.",
		},
		"has_remote_templates": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Whether templates are remote.",
		},
	},
}

type notificationServiceSetupModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	Status                types.String `tfsdk:"status"`
	ServiceID             types.String `tfsdk:"service_id"`
	ServiceCategory       types.String `tfsdk:"service_category"`
	CommunicationMethods  types.Set    `tfsdk:"communication_methods"`
	ParentServiceSetupID  types.String `tfsdk:"parent_service_setup_id"`
	HasRemoteTemplates    types.Bool   `tfsdk:"has_remote_templates"`
}

func (r *NotificationServiceSetupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationServiceSetupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiReq, diags := buildNotificationServiceSetupCreate(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.cidaasClient.NotificationsSrvServiceSetup.Create(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification service setup", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "created notification service setup", util.H{"id": created.ID})
	state := notificationServiceSetupFromAPI(created, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NotificationServiceSetupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationServiceSetupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.cidaasClient.NotificationsSrvServiceSetup.Get(ctx, state.ID.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		resp.Diagnostics.AddError("failed to read notification service setup", util.FormatErrorMessage(err))
		return
	}
	if got == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	out := notificationServiceSetupFromAPI(got, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *NotificationServiceSetupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state notificationServiceSetupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	update := cidaas.NotificationsSrvServiceSetupUpdate{
		ID: state.ID.ValueString(),
	}
	if !plan.Name.IsUnknown() {
		update.Name = plan.Name.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		update.Description = plan.Description.ValueString()
	}
	updated, err := r.cidaasClient.NotificationsSrvServiceSetup.Update(ctx, update)
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification service setup", util.FormatErrorMessage(err))
		return
	}
	out := notificationServiceSetupFromAPI(updated, plan)
	out.ID = state.ID
	out.ServiceID = state.ServiceID
	out.CommunicationMethods = state.CommunicationMethods
	if out.ServiceCategory.IsNull() || out.ServiceCategory.ValueString() == "" {
		out.ServiceCategory = state.ServiceCategory
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *NotificationServiceSetupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationServiceSetupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.cidaasClient.NotificationsSrvServiceSetup.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("failed to delete notification service setup", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "deleted notification service setup", util.H{"id": state.ID.ValueString()})
}

func (r *NotificationServiceSetupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildNotificationServiceSetupCreate(plan notificationServiceSetupModel) (cidaas.NotificationsSrvServiceSetupWrite, diag.Diagnostics) {
	var diags diag.Diagnostics
	methods := make([]string, 0)
	if !plan.CommunicationMethods.IsNull() && !plan.CommunicationMethods.IsUnknown() {
		var elems []types.String
		diags.Append(plan.CommunicationMethods.ElementsAs(context.Background(), &elems, false)...)
		if diags.HasError() {
			return cidaas.NotificationsSrvServiceSetupWrite{}, diags
		}
		for _, m := range elems {
			methods = append(methods, strings.ToLower(strings.TrimSpace(m.ValueString())))
		}
	}
	write := cidaas.NotificationsSrvServiceSetupWrite{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		ServiceDescInfo: cidaas.NotificationsSrvServiceDescInfo{
			ServiceID:       plan.ServiceID.ValueString(),
			ServiceCategory: plan.ServiceCategory.ValueString(),
			CommProv: cidaas.NotificationsSrvCommProvider{
				CommMethods: methods,
			},
		},
		HasRemoteTemplates: plan.HasRemoteTemplates.ValueBool(),
	}
	if !plan.ParentServiceSetupID.IsNull() && !plan.ParentServiceSetupID.IsUnknown() {
		write.ParentServiceSetupID = plan.ParentServiceSetupID.ValueString()
	}
	return write, diags
}

func notificationServiceSetupFromAPI(api *cidaas.NotificationsSrvServiceSetupModel, plan notificationServiceSetupModel) notificationServiceSetupModel {
	methods := make([]string, 0, len(api.ServiceDescInfo.CommProv.CommMethods))
	for _, m := range api.ServiceDescInfo.CommProv.CommMethods {
		methods = append(methods, strings.ToLower(strings.TrimSpace(m)))
	}
	commSet, _ := types.SetValueFrom(context.Background(), types.StringType, methods)
	category := api.ServiceDescInfo.ServiceCategory
	if category == "" {
		category = plan.ServiceCategory.ValueString()
	}
	if category == "" {
		category = "comm_prov"
	}
	return notificationServiceSetupModel{
		ID:                   types.StringValue(api.ID),
		Name:                 types.StringValue(api.Name),
		Description:          types.StringValue(api.Description),
		Status:               types.StringValue(api.Status),
		ServiceID:            types.StringValue(api.ServiceDescInfo.ServiceID),
		ServiceCategory:      types.StringValue(category),
		CommunicationMethods: commSet,
		ParentServiceSetupID: types.StringValue(api.ParentServiceSetupID),
		HasRemoteTemplates:   types.BoolValue(api.HasRemoteTemplates),
	}
}
