//nolint:revive
package identity_provider

import (
	"context"
	"fmt"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                     = (*FederationProviderResource)(nil)
	_ resource.ResourceWithConfigure        = (*FederationProviderResource)(nil)
	_ resource.ResourceWithImportState      = (*FederationProviderResource)(nil)
	_ resource.ResourceWithConfigValidators = (*FederationProviderResource)(nil)
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

func (r *FederationProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	if !c.RequireV4("cidaas_federation_provider", &resp.Diagnostics) {
		return
	}
	r.BaseResource.Configure(ctx, req, resp)
}

func (r *FederationProviderResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("client_secret"),
			path.MatchRoot("client_secret_wo"),
		),
		resourcevalidator.RequiredTogether(
			path.MatchRoot("client_secret_wo"),
			path.MatchRoot("client_secret_wo_version"),
		),
		resourcevalidator.PreferWriteOnlyAttribute(
			path.MatchRoot("client_secret"),
			path.MatchRoot("client_secret_wo"),
		),
	}
}

func (r *FederationProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages federated identity providers on cidaas v4 (Trustdesk) via `/federation/providers` " +
			"(OAuth2, OpenID Connect, SAML, LDAP). Preferred replacement for deprecated `cidaas_social_provider` and `cidaas_custom_provider`. " +
			"Exactly one of `client_secret` or `client_secret_wo` must be set.",
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
				Description: "Standard type (e.g. OAUTH2, OPENID_CONNECT, SAML, LDAP).",
			},
			"client_id": schema.StringAttribute{
				Required:    true,
				Description: "Client ID of the provider.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Client secret of the provider. Exactly one of `client_secret` or `client_secret_wo` must be set. Stored in state.",
			},
			"client_secret_wo": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Description: "Write-only client secret. Sent on create/update without saving to state. Requires `client_secret_wo_version`.",
			},
			"client_secret_wo_version": schema.StringAttribute{
				Optional:    true,
				Description: "Used together with client_secret_wo to trigger an update.",
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
				Default:     stringdefault.StaticString("client"),
				Description: "Owner of the provider (defaults to client for Admin UI compatibility).",
			},
		},
	}
}

func (r *FederationProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config federationProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := prepareFederationProviderModel(ctx, plan, config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.CidaasClient.FederationProvider.Create(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create federation provider", util.FormatErrorMessage(err))
		return
	}

	plan.ID = types.StringValue(res.Data.ID)
	if usingFederationWriteOnlySecret(config) {
		plan.ClientSecret = types.StringNull()
	}
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
		if util.IsResourceNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read federation provider", util.FormatErrorMessage(err))
		return
	}

	isImport := state.ClientID.IsNull()
	state.ID = types.StringValue(res.Data.ID)
	state.ProviderName = types.StringValue(res.Data.ProviderName)
	state.DisplayName = types.StringValue(res.Data.DisplayName)
	state.StandardType = types.StringValue(res.Data.StandardType)
	state.ClientID = types.StringValue(res.Data.ClientID)
	if !state.ClientSecret.IsNull() || isImport {
		state.ClientSecret = util.StringValueOrNull(&res.Data.ClientSecret)
	}
	state.AuthorizationEndpoint = util.StringValueOrNull(&res.Data.AuthorizationEndpoint)
	state.TokenEndpoint = util.StringValueOrNull(&res.Data.TokenEndpoint)
	state.UserinfoEndpoint = util.StringValueOrNull(&res.Data.UserinfoEndpoint)
	state.LogoURL = util.StringValueOrNull(&res.Data.LogoURL)
	state.Owner = util.StringValueOrNull(&res.Data.Owner)
	state.Domains = listStringOrNull(res.Data.Domains)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FederationProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config federationProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := prepareFederationProviderModel(ctx, plan, config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.CidaasClient.FederationProvider.Update(ctx, plan.ID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update federation provider", util.FormatErrorMessage(err))
		return
	}

	plan.ID = types.StringValue(res.Data.ID)
	if usingFederationWriteOnlySecret(config) {
		plan.ClientSecret = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FederationProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state federationProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.CidaasClient.FederationProvider.Delete(ctx, state.ID.ValueString())
	if err != nil && !util.IsResourceNotFound(err) {
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
	ClientSecretWO        types.String `tfsdk:"client_secret_wo"`
	ClientSecretWOVersion types.String `tfsdk:"client_secret_wo_version"`
	AuthorizationEndpoint types.String `tfsdk:"authorization_endpoint"`
	TokenEndpoint         types.String `tfsdk:"token_endpoint"`
	UserinfoEndpoint      types.String `tfsdk:"userinfo_endpoint"`
	LogoURL               types.String `tfsdk:"logo_url"`
	Domains               types.List   `tfsdk:"domains"`
	Owner                 types.String `tfsdk:"owner"`
}

func usingFederationWriteOnlySecret(config federationProviderModel) bool {
	return !config.ClientSecretWO.IsNull() && !config.ClientSecretWO.IsUnknown()
}

func prepareFederationProviderModel(ctx context.Context, plan, config federationProviderModel, diags *diag.Diagnostics) *cidaas.ProviderConfigModel {
	ownerVal := plan.Owner.ValueString()
	if ownerVal == "" {
		ownerVal = "client"
	}
	secretVal := plan.ClientSecret.ValueString()
	if usingFederationWriteOnlySecret(config) {
		secretVal = config.ClientSecretWO.ValueString()
	}
	pc := &cidaas.ProviderConfigModel{
		ID:                    plan.ID.ValueString(),
		ProviderName:          plan.ProviderName.ValueString(),
		DisplayName:           plan.DisplayName.ValueString(),
		StandardType:          plan.StandardType.ValueString(),
		ClientID:              plan.ClientID.ValueString(),
		ClientSecret:          secretVal,
		AuthorizationEndpoint: plan.AuthorizationEndpoint.ValueString(),
		TokenEndpoint:         plan.TokenEndpoint.ValueString(),
		UserinfoEndpoint:      plan.UserinfoEndpoint.ValueString(),
		LogoURL:               plan.LogoURL.ValueString(),
		Owner:                 ownerVal,
	}
	if !plan.Domains.IsNull() && !plan.Domains.IsUnknown() {
		pc.Domains = util.ListToStrings(plan.Domains)
	}
	return pc
}
