package resources

import (
	"context"
	"encoding/json"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type NotificationProviderConfigResource struct {
	BaseResource
}

func NewNotificationProviderConfigResource() resource.Resource {
	return &NotificationProviderConfigResource{
		BaseResource: NewBaseResource(BaseResourceConfig{
			Name:   RESOURCE_NOTIFICATION_PROVIDER_CONFIG,
			Schema: &notificationProviderConfigSchema,
		}),
	}
}

var notificationProviderConfigSchema = schema.Schema{
	MarkdownDescription: "Stores **provider credentials** for a service setup via **notification-srv** (`POST /{notifications_context_path}/providerconfigs`). " +
		"notification-srv proxies to mplace admin `provconfs` (upsert by `_id` = `service_setup_id`).\n\n" +
		"**Recommended:** pass wizard-shaped JSON in `config_data_wo` with top-level `commProvider`, `commMethod`, and `schemaData` " +
		"so mplace validates and maps fields server-side.\n\n" +
		"**Verification** is manual (service-desk); use `cidaas_notification_service_setup.status` after refresh.\n\n" +
		"**Scopes:** `cidaas:service_setups_read`, `cidaas:provider_config_write`.\n\n" +
		"-> **Note:** Write-Only argument `config_data_wo` is available to use in place of `config_data`. " +
		"Write-only arguments are supported in HashiCorp Terraform 1.11.0 and later.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Same as `service_setup_id`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"service_setup_id": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Service setup `_id` this config belongs to.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"config_data": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			MarkdownDescription: "Full `configData` JSON (stored in state). Prefer `config_data_wo` for secrets.",
		},
		"config_data_wo": schema.StringAttribute{
			Optional:            true,
			Sensitive:           true,
			WriteOnly:           true,
			MarkdownDescription: "Write-Only `configData` JSON. Not stored in plan or state. Must be set with `config_data_wo_version`.",
		},
		"config_data_wo_version": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Increment to push an updated `config_data_wo` to the API.",
		},
	},
}

type notificationProviderConfigModel struct {
	ID                   types.String `tfsdk:"id"`
	ServiceSetupID       types.String `tfsdk:"service_setup_id"`
	ConfigData           types.String `tfsdk:"config_data"`
	ConfigDataWO         types.String `tfsdk:"config_data_wo"`
	ConfigDataWOVersion  types.String `tfsdk:"config_data_wo_version"`
}

func (r *NotificationProviderConfigResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("config_data"),
			path.MatchRoot("config_data_wo"),
		),
		resourcevalidator.RequiredTogether(
			path.MatchRoot("config_data_wo"),
			path.MatchRoot("config_data_wo_version"),
		),
		resourcevalidator.PreferWriteOnlyAttribute(
			path.MatchRoot("config_data"),
			path.MatchRoot("config_data_wo"),
		),
	}
}

func (r *NotificationProviderConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config notificationProviderConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	raw, diags := resolveProviderConfigData(plan, config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.cidaasClient.NotificationsSrvProviderConfig.Create(ctx, plan.ServiceSetupID.ValueString(), raw)
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification provider config", util.FormatErrorMessage(err))
		return
	}
	state := providerConfigStateFromAPI(created, plan, config)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NotificationProviderConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationProviderConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.cidaasClient.NotificationsSrvProviderConfig.Get(ctx, state.ServiceSetupID.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		resp.Diagnostics.AddError("failed to read notification provider config", util.FormatErrorMessage(err))
		return
	}
	if got == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	out := providerConfigStateFromAPI(got, state, notificationProviderConfigModel{})
	// Preserve write-only usage: do not populate config_data when config_data_wo was used.
	if state.ConfigData.IsNull() && !state.ConfigDataWOVersion.IsNull() {
		out.ConfigData = types.StringNull()
	} else if !state.ConfigData.IsNull() {
		out.ConfigData = state.ConfigData
	}
	out.ConfigDataWOVersion = state.ConfigDataWOVersion
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *NotificationProviderConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config notificationProviderConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	raw, diags := resolveProviderConfigData(plan, config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.cidaasClient.NotificationsSrvProviderConfig.Create(ctx, plan.ServiceSetupID.ValueString(), raw)
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification provider config", util.FormatErrorMessage(err))
		return
	}
	out := providerConfigStateFromAPI(updated, plan, config)
	out.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *NotificationProviderConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationProviderConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// notification-srv has no DELETE for providerconfigs; remote config remains until service setup is deleted.
	tflog.Warn(ctx, "notification provider config removed from Terraform state only; remote config is not deleted by this resource", util.H{
		"service_setup_id": state.ServiceSetupID.ValueString(),
	})
}

func (r *NotificationProviderConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("service_setup_id"), req, resp)
}

func resolveProviderConfigData(plan, config notificationProviderConfigModel) (json.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics
	var jsonStr string
	if !config.ConfigDataWO.IsNull() && !config.ConfigDataWO.IsUnknown() {
		jsonStr = config.ConfigDataWO.ValueString()
	} else if !plan.ConfigDataWO.IsNull() && !plan.ConfigDataWO.IsUnknown() {
		jsonStr = plan.ConfigDataWO.ValueString()
	} else if !plan.ConfigData.IsNull() && !plan.ConfigData.IsUnknown() {
		jsonStr = plan.ConfigData.ValueString()
	}
	if jsonStr == "" {
		diags.AddError("missing config data", "set config_data or config_data_wo")
		return nil, diags
	}
	if !json.Valid([]byte(jsonStr)) {
		diags.AddError("invalid config_data JSON", "config_data / config_data_wo must be valid JSON")
		return nil, diags
	}
	return json.RawMessage(jsonStr), diags
}

func providerConfigStateFromAPI(api *cidaas.NotificationsSrvProviderConfigModel, plan, config notificationProviderConfigModel) notificationProviderConfigModel {
	state := notificationProviderConfigModel{
		ID:             types.StringValue(api.ID),
		ServiceSetupID: types.StringValue(api.ID),
	}
	if !config.ConfigDataWO.IsNull() && !config.ConfigDataWO.IsUnknown() {
		state.ConfigData = types.StringNull()
		state.ConfigDataWOVersion = plan.ConfigDataWOVersion
	} else if !plan.ConfigData.IsNull() && !plan.ConfigData.IsUnknown() {
		state.ConfigData = plan.ConfigData
	} else if len(api.ConfigData) > 0 {
		state.ConfigData = types.StringValue(string(api.ConfigData))
	}
	return state
}
