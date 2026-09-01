package identity_provider

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//nolint:revive
type CustomProviderResource struct {
	base.BaseResource
}

func NewCustomProviderResource() resource.Resource {
	return &CustomProviderResource{
		BaseResource: base.NewBaseResource(
			base.BaseResourceConfig{
				Name: base.RESOURCE_CUSTOM_PROVIDER,
			},
		),
	}
}

func (r *CustomProviderResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Custom Identity Providers in Cidaas.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "Unique identifier of the custom provider.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Unique identifier of the custom provider.",
			},
			"client_id": schema.StringAttribute{
				Required:    true,
				Description: "Client ID for authentication with the custom provider.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Client Secret for authentication with the custom provider.",
			},
			"client_secret_wo": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Write-only client secret for authentication.",
			},
			"client_secret_wo_version": schema.StringAttribute{
				Optional:    true,
				Description: "Version of write-only client secret.",
			},
			"display_name": schema.StringAttribute{
				Required:    true,
				Description: "Display name of the custom provider.",
			},
			"standard_type": schema.StringAttribute{
				Required:    true,
				Description: "Standard type of custom provider (e.g. OAUTH2, OIDC).",
				Validators:  []validator.String{stringvalidator.OneOf("OAUTH2", "OIDC")},
			},
			"authorization_endpoint": schema.StringAttribute{
				Required:    true,
				Description: "Authorization endpoint URL.",
			},
			"token_endpoint": schema.StringAttribute{
				Required:    true,
				Description: "Token endpoint URL.",
			},
			"provider_name": schema.StringAttribute{
				Required:      true,
				Description:   "Unique provider name.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"logo_url": schema.StringAttribute{
				Required:    true,
				Description: "Logo URL of the provider.",
			},
			"userinfo_endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "Userinfo endpoint URL.",
			},
			"scope_display_label": schema.StringAttribute{
				Optional:    true,
				Description: "Display label for scopes.",
			},
			"domains": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Allowed domain names for provider.",
			},
			"userinfo_source": schema.StringAttribute{
				Optional:    true,
				Description: "Userinfo source setting.",
			},
			"pkce": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable PKCE for provider flow.",
			},
			"auth_type": schema.StringAttribute{
				Optional:    true,
				Description: "Auth type setting.",
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Owner of the provider (defaults to client for Admin UI compatibility).",
			},
		},
	}
}

func (r *CustomProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := prepareCustomProviderModel(plan)
	res, err := r.CidaasClient.CustomProvider.CreateCustomProvider(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create custom provider", util.FormatErrorMessage(err))
		return
	}

	plan.ID = types.StringValue(res.Data.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

//nolint:dupl
func (r *CustomProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.CidaasClient.CustomProvider.GetCustomProvider(ctx, state.ProviderName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read custom provider", util.FormatErrorMessage(err))
		return
	}

	state.ID = types.StringValue(res.Data.ID)
	state.DisplayName = types.StringValue(res.Data.DisplayName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CustomProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := prepareCustomProviderModel(plan)
	err := r.CidaasClient.CustomProvider.UpdateCustomProvider(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update custom provider", util.FormatErrorMessage(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CustomProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.CidaasClient.CustomProvider.DeleteCustomProvider(ctx, state.ProviderName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete custom provider", util.FormatErrorMessage(err))
		return
	}
}

func (r *CustomProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("provider_name"), req, resp)
}

type customProviderModel struct {
	ID                    types.String `tfsdk:"id"`
	ClientID              types.String `tfsdk:"client_id"`
	ClientSecret          types.String `tfsdk:"client_secret"`
	ClientSecretWO        types.String `tfsdk:"client_secret_wo"`
	ClientSecretWOVersion types.String `tfsdk:"client_secret_wo_version"`
	DisplayName           types.String `tfsdk:"display_name"`
	StandardType          types.String `tfsdk:"standard_type"`
	AuthorizationEndpoint types.String `tfsdk:"authorization_endpoint"`
	TokenEndpoint         types.String `tfsdk:"token_endpoint"`
	ProviderName          types.String `tfsdk:"provider_name"`
	LogoURL               types.String `tfsdk:"logo_url"`
	UserinfoEndpoint      types.String `tfsdk:"userinfo_endpoint"`
	ScopeDisplayLabel     types.String `tfsdk:"scope_display_label"`
	Domains               types.List   `tfsdk:"domains"`
	UserinfoSource        types.String `tfsdk:"userinfo_source"`
	Pkce                  types.Bool   `tfsdk:"pkce"`
	AuthType              types.String `tfsdk:"auth_type"`
	Owner                 types.String `tfsdk:"owner"`
}

func prepareCustomProviderModel(plan customProviderModel) *cidaas.CustomProviderModel {
	return &cidaas.CustomProviderModel{
		ClientID:              plan.ClientID.ValueString(),
		ClientSecret:          plan.ClientSecret.ValueString(),
		DisplayName:           plan.DisplayName.ValueString(),
		StandardType:          plan.StandardType.ValueString(),
		AuthorizationEndpoint: plan.AuthorizationEndpoint.ValueString(),
		TokenEndpoint:         plan.TokenEndpoint.ValueString(),
		ProviderName:          plan.ProviderName.ValueString(),
		LogoURL:               plan.LogoURL.ValueString(),
		UserinfoEndpoint:      plan.UserinfoEndpoint.ValueString(),
		UserInfoSource:        plan.UserinfoSource.ValueString(),
		Pkce:                  plan.Pkce.ValueBool(),
		AuthType:              plan.AuthType.ValueString(),
	}
}
