//nolint:revive
package identity_provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//nolint:revive
type SocialProviderResource struct {
	base.BaseResource
}

func NewSocialProviderResource() resource.Resource {
	return &SocialProviderResource{
		BaseResource: base.NewBaseResource(
			base.BaseResourceConfig{
				Name: base.RESOURCE_SOCIAL_PROVIDER,
			},
		),
	}
}

func (r *SocialProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}
	if c.Capabilities.TargetVersion == "4.x" {
		resp.Diagnostics.AddError(
			"Incompatible Resource",
			"The resource `cidaas_social_provider` is a legacy v3 resource and is not supported on cidaas v4.x. Use `cidaas_federation_provider` instead.",
		)
		return
	}
	r.BaseResource.Configure(ctx, req, resp)
}

func (r *SocialProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Social Identity Providers in Cidaas.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "Unique identifier of the social provider.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Unique identifier of the social provider.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the social provider.",
			},
			"provider_name": schema.StringAttribute{
				Required:      true,
				Description:   "Provider identifier name (e.g. google, facebook, apple).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"client_id": schema.StringAttribute{
				Required:    true,
				Description: "Client ID of the social provider.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Client secret of the social provider.",
			},
			"client_secret_wo": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Write-only client secret.",
			},
			"client_secret_wo_version": schema.StringAttribute{
				Optional:    true,
				Description: "Version of write-only client secret.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether the provider is enabled.",
			},
			"enabled_for_admin_portal": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether the provider is enabled for Admin Portal login.",
			},
			"scopes": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Scopes requested from social provider.",
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Owner of the provider (defaults to client for Admin UI compatibility).",
			},
		},
	}
}

//nolint:dupl
func (r *SocialProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan socialProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := prepareSocialProviderModel(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.CidaasClient.SocialProvider.Upsert(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create social provider", util.FormatErrorMessage(err))
		return
	}

	plan.ID = types.StringValue(res.Data.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SocialProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state socialProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.CidaasClient.SocialProvider.Get(ctx, state.ProviderName.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read social provider", util.FormatErrorMessage(err))
		return
	}

	state.Name = types.StringValue(res.Data.Name)
	state.Enabled = types.BoolValue(res.Data.Enabled)
	state.EnabledForAdminPortal = types.BoolValue(res.Data.EnabledForAdminPortal)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//nolint:dupl
func (r *SocialProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan socialProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := prepareSocialProviderModel(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.CidaasClient.SocialProvider.Upsert(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update social provider", util.FormatErrorMessage(err))
		return
	}

	plan.ID = types.StringValue(res.Data.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SocialProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state socialProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.CidaasClient.SocialProvider.Delete(ctx, state.ProviderName.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete social provider", util.FormatErrorMessage(err))
		return
	}
}

func (r *SocialProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: 'provider_name:provider_id', got: %s", req.ID),
		)
		return
	}
	resp.State.SetAttribute(ctx, path.Root("provider_name"), parts[0])
	resp.State.SetAttribute(ctx, path.Root("id"), parts[1])
}

type socialProviderModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	ProviderName          types.String `tfsdk:"provider_name"`
	ClientID              types.String `tfsdk:"client_id"`
	ClientSecret          types.String `tfsdk:"client_secret"`
	ClientSecretWO        types.String `tfsdk:"client_secret_wo"`
	ClientSecretWOVersion types.String `tfsdk:"client_secret_wo_version"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	EnabledForAdminPortal types.Bool   `tfsdk:"enabled_for_admin_portal"`
	Scopes                types.Set    `tfsdk:"scopes"`
	Owner                 types.String `tfsdk:"owner"`
}

func prepareSocialProviderModel(ctx context.Context, plan socialProviderModel, diags *diag.Diagnostics) *cidaas.SocialProviderModel {
	sp := &cidaas.SocialProviderModel{
		ID:                    plan.ID.ValueString(),
		Name:                  plan.Name.ValueString(),
		ProviderName:          plan.ProviderName.ValueString(),
		ClientID:              plan.ClientID.ValueString(),
		ClientSecret:          plan.ClientSecret.ValueString(),
		Enabled:               plan.Enabled.ValueBool(),
		EnabledForAdminPortal: plan.EnabledForAdminPortal.ValueBool(),
	}

	if !plan.Scopes.IsNull() && !plan.Scopes.IsUnknown() {
		var scopes []string
		diags.Append(plan.Scopes.ElementsAs(ctx, &scopes, false)...)
		sp.Scopes = scopes
	}

	return sp
}
