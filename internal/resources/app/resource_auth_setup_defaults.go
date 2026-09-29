package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const authSetupDefaultsResourceID = "default"

var (
	_ resource.Resource                = &authSetupDefaultsResource{}
	_ resource.ResourceWithConfigure   = &authSetupDefaultsResource{}
	_ resource.ResourceWithImportState = &authSetupDefaultsResource{}
)

type authSetupDefaultsResource struct {
	client *client.Client
}

// NewAuthSetupDefaultsResource returns the cidaas_auth_setup_defaults singleton resource.
func NewAuthSetupDefaultsResource() resource.Resource {
	return &authSetupDefaultsResource{}
}

type authSetupDefaultsModel struct {
	ID                           types.String `tfsdk:"id"`
	Name                         types.String `tfsdk:"name"`
	Description                  types.String `tfsdk:"description"`
	AutoLoginAfterRegister       types.Bool   `tfsdk:"auto_login_after_register"`
	RegisterWithLoginInformation types.Bool   `tfsdk:"register_with_login_information"`
	EnablePasswordLessAuth       types.Bool   `tfsdk:"enable_password_less_auth"`
	AllowUserLevelMultiProvider  types.Bool   `tfsdk:"allow_user_level_multi_provider"`
	SocialBusinessIDs            types.Bool   `tfsdk:"social_business_ids"`
	NetID                        types.Bool   `tfsdk:"net_id"`
	CreatedTime                  types.String `tfsdk:"created_time"`
	UpdatedTime                  types.String `tfsdk:"updated_time"`
}

func (r *authSetupDefaultsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_setup_defaults"
}

func (r *authSetupDefaultsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the tenant **Default Authentication Setup** (Trustdesk) via `app-srv/apps/auth-setup-defaults`. " +
			"This is a singleton (`id` is always `default`); Create updates the seeded record, Destroy only removes it from Terraform state.\n\n" +
			"Apps may override the same bools under `cidaas_app_configuration.authentication_setup` (nil inherits these defaults).\n\n" +
			"Requires admin roles and scopes `cidaas:apps_read` / `cidaas:apps_write`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Always `default`. Import with `terraform import cidaas_auth_setup_defaults.<name> default`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Display name of the defaults record (usually `default`).",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional description.",
			},
			"auto_login_after_register": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Automatically log the user in after registration.",
			},
			"register_with_login_information": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Allow registration using login information.",
			},
			"enable_password_less_auth": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Enable passwordless authentication methods (magic link / OTP).",
			},
			"allow_user_level_multi_provider": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Allow users to link multiple identity providers.",
			},
			"social_business_ids": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Enable social business IDs.",
			},
			"net_id": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Enable the NetID special social provider (Active flag).",
			},
			"created_time": schema.StringAttribute{Computed: true},
			"updated_time": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *authSetupDefaultsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	if !c.ValidateResourceVersion("cidaas_auth_setup_defaults", &resp.Diagnostics) {
		return
	}
	r.client = c
}

func (r *authSetupDefaultsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var plan authSetupDefaultsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.AuthSetupDefaults.Get(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read auth setup defaults failed", err.Error())
		return
	}
	entity := mergeAuthSetupDefaultsPlan(current.Data, plan)
	res, err := r.client.AuthSetupDefaults.Update(ctx, entity)
	if err != nil {
		resp.Diagnostics.AddError("Update auth setup defaults failed", err.Error())
		return
	}
	state := flattenAuthSetupDefaults(res.Data)
	state = mergeOmittedAuthSetupDefaultsBools(plan, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *authSetupDefaultsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var state authSetupDefaultsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.AuthSetupDefaults.Get(ctx)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read auth setup defaults failed", err.Error())
		return
	}
	next := flattenAuthSetupDefaults(res.Data)
	next = mergeOmittedAuthSetupDefaultsBools(state, next)
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}

func (r *authSetupDefaultsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Configure the provider before managing resources.")
		return
	}
	var plan authSetupDefaultsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.client.AuthSetupDefaults.Get(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read auth setup defaults failed", err.Error())
		return
	}
	entity := mergeAuthSetupDefaultsPlan(current.Data, plan)
	res, err := r.client.AuthSetupDefaults.Update(ctx, entity)
	if err != nil {
		resp.Diagnostics.AddError("Update auth setup defaults failed", err.Error())
		return
	}
	state := flattenAuthSetupDefaults(res.Data)
	state = mergeOmittedAuthSetupDefaultsBools(plan, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *authSetupDefaultsResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Singleton is seeded by app-srv; destroy only drops Terraform state.
	resp.State.RemoveResource(ctx)
}

func (r *authSetupDefaultsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	if id == "" {
		id = authSetupDefaultsResourceID
	}
	if id != authSetupDefaultsResourceID {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("cidaas_auth_setup_defaults is a singleton; import id must be %q (got %q).", authSetupDefaultsResourceID, req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func mergeAuthSetupDefaultsPlan(current client.AuthSetupDefaultsEntity, plan authSetupDefaultsModel) client.AuthSetupDefaultsEntity {
	name := current.Name
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		name = plan.Name.ValueString()
	}
	if name == "" {
		name = authSetupDefaultsResourceID
	}
	desc := current.Description
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		desc = plan.Description.ValueString()
	}
	defaults := &client.AuthSetupDefaults{}
	if current.AuthSetupDefaults != nil {
		*defaults = *current.AuthSetupDefaults
	}
	defaults.AutoLoginAfterRegister = preferBoolPtr(plan.AutoLoginAfterRegister, defaults.AutoLoginAfterRegister)
	defaults.RegisterWithLoginInformation = preferBoolPtr(plan.RegisterWithLoginInformation, defaults.RegisterWithLoginInformation)
	defaults.EnablePasswordLessAuth = preferBoolPtr(plan.EnablePasswordLessAuth, defaults.EnablePasswordLessAuth)
	defaults.AllowUserLevelMultiProvider = preferBoolPtr(plan.AllowUserLevelMultiProvider, defaults.AllowUserLevelMultiProvider)
	defaults.SocialBusinessIDs = preferBoolPtr(plan.SocialBusinessIDs, defaults.SocialBusinessIDs)
	defaults.NetID = preferBoolPtr(plan.NetID, defaults.NetID)

	return client.AuthSetupDefaultsEntity{
		ID:                authSetupDefaultsResourceID,
		Name:              name,
		Description:       desc,
		AuthSetupDefaults: defaults,
	}
}

func preferBoolPtr(plan types.Bool, current *bool) *bool {
	if !plan.IsNull() && !plan.IsUnknown() {
		v := plan.ValueBool()
		return &v
	}
	return current
}

func flattenAuthSetupDefaults(entity client.AuthSetupDefaultsEntity) authSetupDefaultsModel {
	m := authSetupDefaultsModel{
		ID:          types.StringValue(authSetupDefaultsResourceID),
		Name:        stringOrNull(entity.Name),
		Description: stringOrNull(entity.Description),
		CreatedTime: stringOrNull(entity.CreatedTime),
		UpdatedTime: stringOrNull(entity.UpdatedTime),
	}
	if entity.AuthSetupDefaults != nil {
		d := entity.AuthSetupDefaults
		m.AutoLoginAfterRegister = boolValueOrNull(d.AutoLoginAfterRegister)
		m.RegisterWithLoginInformation = boolValueOrNull(d.RegisterWithLoginInformation)
		m.EnablePasswordLessAuth = boolValueOrNull(d.EnablePasswordLessAuth)
		m.AllowUserLevelMultiProvider = boolValueOrNull(d.AllowUserLevelMultiProvider)
		m.SocialBusinessIDs = boolValueOrNull(d.SocialBusinessIDs)
		m.NetID = boolValueOrNull(d.NetID)
	}
	return m
}

func mergeOmittedAuthSetupDefaultsBools(prior, state authSetupDefaultsModel) authSetupDefaultsModel {
	state.AutoLoginAfterRegister = preferKnownBool(prior.AutoLoginAfterRegister, state.AutoLoginAfterRegister)
	state.RegisterWithLoginInformation = preferKnownBool(prior.RegisterWithLoginInformation, state.RegisterWithLoginInformation)
	state.EnablePasswordLessAuth = preferKnownBool(prior.EnablePasswordLessAuth, state.EnablePasswordLessAuth)
	state.AllowUserLevelMultiProvider = preferKnownBool(prior.AllowUserLevelMultiProvider, state.AllowUserLevelMultiProvider)
	state.SocialBusinessIDs = preferKnownBool(prior.SocialBusinessIDs, state.SocialBusinessIDs)
	state.NetID = preferKnownBool(prior.NetID, state.NetID)
	state.Name = preferKnownString(prior.Name, state.Name)
	state.Description = preferKnownString(prior.Description, state.Description)
	return state
}

// preferKnownBool keeps prior when the API omitted a bool (nil → null).
func preferKnownBool(prior, fromAPI types.Bool) types.Bool {
	if fromAPI.IsNull() && !prior.IsNull() && !prior.IsUnknown() {
		return prior
	}
	return fromAPI
}

func preferKnownString(prior, fromAPI types.String) types.String {
	if fromAPI.IsNull() && !prior.IsNull() && !prior.IsUnknown() {
		return prior
	}
	return fromAPI
}
