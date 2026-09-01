package identity_provider

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//nolint:revive
type FederationProviderResource struct {
	base.BaseResource
}

func NewFederationProviderResource() resource.Resource {
	return &FederationProviderResource{
		BaseResource: base.NewBaseResource(
			base.BaseResourceConfig{
				Name: base.RESOURCE_FEDERATION_PROVIDER,
			},
		),
	}
}

func (r *FederationProviderResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Native Federated Identity Providers (v4.x) via /federation/providers.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "Unique identifier of the federation provider.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Unique identifier of the federation provider.",
			},
			"provider_name": schema.StringAttribute{
				Required:      true,
				Description:   "Unique provider name.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"display_name": schema.StringAttribute{
				Required:    true,
				Description: "Display name of the provider.",
			},
			"standard_type": schema.StringAttribute{
				Required:    true,
				Description: "Standard type (e.g. OAUTH2, OIDC, SAML, LDAP).",
			},
			"client_id": schema.StringAttribute{
				Required:    true,
				Description: "Client ID of the provider.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Client secret of the provider.",
			},
			"authorization_endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "Authorization endpoint URL.",
			},
			"token_endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "Token endpoint URL.",
			},
			"userinfo_endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "Userinfo endpoint URL.",
			},
			"logo_url": schema.StringAttribute{
				Optional:    true,
				Description: "Logo URL of the provider.",
			},
			"domains": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Allowed domains for provider.",
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Owner of the provider (defaults to client for Admin UI compatibility).",
			},
		},
	}
}

func (r *FederationProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan federationProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := prepareFederationProviderModel(ctx, plan)
	res, err := r.CidaasClient.FederationProvider.Create(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create federation provider", util.FormatErrorMessage(err))
		return
	}

	plan.ID = types.StringValue(res.Data.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

//nolint:dupl
func (r *FederationProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state federationProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.CidaasClient.FederationProvider.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read federation provider", util.FormatErrorMessage(err))
		return
	}

	state.DisplayName = types.StringValue(res.Data.DisplayName)
	state.StandardType = types.StringValue(res.Data.StandardType)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FederationProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan federationProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := prepareFederationProviderModel(ctx, plan)
	res, err := r.CidaasClient.FederationProvider.Update(ctx, plan.ID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update federation provider", util.FormatErrorMessage(err))
		return
	}

	plan.ID = types.StringValue(res.Data.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FederationProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state federationProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.CidaasClient.FederationProvider.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete federation provider", util.FormatErrorMessage(err))
		return
	}
}

func (r *FederationProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type federationProviderModel struct {
	ID                    types.String `tfsdk:"id"`
	ProviderName          types.String `tfsdk:"provider_name"`
	DisplayName           types.String `tfsdk:"display_name"`
	StandardType          types.String `tfsdk:"standard_type"`
	ClientID              types.String `tfsdk:"client_id"`
	ClientSecret          types.String `tfsdk:"client_secret"`
	AuthorizationEndpoint types.String `tfsdk:"authorization_endpoint"`
	TokenEndpoint         types.String `tfsdk:"token_endpoint"`
	UserinfoEndpoint      types.String `tfsdk:"userinfo_endpoint"`
	LogoURL               types.String `tfsdk:"logo_url"`
	Domains               types.List   `tfsdk:"domains"`
	Owner                 types.String `tfsdk:"owner"`
}

func prepareFederationProviderModel(_ context.Context, plan federationProviderModel) *cidaas.ProviderConfigModel {
	ownerVal := plan.Owner.ValueString()
	if ownerVal == "" {
		ownerVal = "client"
	}
	return &cidaas.ProviderConfigModel{
		ID:                    plan.ID.ValueString(),
		ProviderName:          plan.ProviderName.ValueString(),
		DisplayName:           plan.DisplayName.ValueString(),
		StandardType:          plan.StandardType.ValueString(),
		ClientID:              plan.ClientID.ValueString(),
		ClientSecret:          plan.ClientSecret.ValueString(),
		AuthorizationEndpoint: plan.AuthorizationEndpoint.ValueString(),
		TokenEndpoint:         plan.TokenEndpoint.ValueString(),
		UserinfoEndpoint:      plan.UserinfoEndpoint.ValueString(),
		LogoURL:               plan.LogoURL.ValueString(),
		Owner:                 ownerVal,
	}
}
