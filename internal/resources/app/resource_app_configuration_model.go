// Plan/state model for cidaas_app_configuration (appv3 / Trustdesk).
package app

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// appConfigurationConfig is the Terraform state/plan model for cidaas_app_configuration.
type appConfigurationConfig struct {
	ClientID            types.String `tfsdk:"client_id"`
	ClientName          types.String `tfsdk:"client_name"`
	ClientType          types.String `tfsdk:"client_type"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	GrantTypes          types.List   `tfsdk:"grant_types"`
	ResponseTypes       types.List   `tfsdk:"response_types"`
	RedirectURIs        types.Object `tfsdk:"redirect_uris"`
	Scopes              types.Object `tfsdk:"scopes"`
	TokenLifetimes      types.Object `tfsdk:"token_lifetimes"`
	AuthenticationSetup types.Object `tfsdk:"authentication_setup"`
	HostedPagesLayoutID types.String `tfsdk:"hosted_pages_layout_id"`
	UserSetupID         types.String `tfsdk:"user_setup_id"`
	OwnershipDetails    types.Object `tfsdk:"ownership_details"`
	ClientAuthConfig    types.Object `tfsdk:"client_auth_config"`
	SigningKeyConfig    types.Object `tfsdk:"signing_key_config"`
	CreatedTime         types.String `tfsdk:"created_time"`
	UpdatedTime         types.String `tfsdk:"updated_time"`
	Owner               types.String `tfsdk:"owner"`

	redirectURIs        *redirectURIsConfig
	scopes              *scopesConfig
	tokenLifetimes      *tokenLifetimesConfig
	authenticationSetup *authenticationSetupConfig
	ownershipDetails    *ownershipDetailsConfig
	clientAuthConfig    *clientAuthConfigBlock
}

// clientAuthConfigBlock is the Terraform nested block for client_auth_config.
type clientAuthConfigBlock struct {
	TokenEndpointAuthMethod types.String `tfsdk:"token_endpoint_auth_method"`
}

// redirectURIsConfig is the Terraform nested block for redirect_uris.
type redirectURIsConfig struct {
	RedirectURIs           types.List `tfsdk:"redirect_uris"`
	AllowedLogoutUrls      types.List `tfsdk:"allowed_logout_urls"`
	PostLogoutRedirectURIs types.List `tfsdk:"post_logout_redirect_uris"`
	AllowedWebOrigins      types.List `tfsdk:"allowed_web_origins"`
}

// scopesConfig is the Terraform nested block for scopes.
type scopesConfig struct {
	AllowedScopes types.List `tfsdk:"allowed_scopes"`
	DefaultScopes types.List `tfsdk:"default_scopes"`
}

// tokenLifetimesConfig is the Terraform nested block for token_lifetimes.
type tokenLifetimesConfig struct {
	TokenLifetimeInSeconds        types.Int64 `tfsdk:"token_lifetime_in_seconds"`
	RefreshTokenLifetimeInSeconds types.Int64 `tfsdk:"refresh_token_lifetime_in_seconds"`
	IDTokenLifetimeInSeconds      types.Int64 `tfsdk:"id_token_lifetime_in_seconds"`
	CodeLifetimeInSeconds         types.Int64 `tfsdk:"code_lifetime_in_seconds"`
	DefaultMaxAge                 types.Int64 `tfsdk:"default_max_age"`
}

// authenticationSetupConfig is the Terraform nested block for authentication_setup (extdep IDs + flags).
type authenticationSetupConfig struct {
	VerificationOptionsID        types.String `tfsdk:"verification_options_id"`
	GroupSelectionID             types.String `tfsdk:"group_selection_id"`
	GroupVerificationRequestID   types.String `tfsdk:"group_verification_request_id"`
	TemplateGroupID              types.String `tfsdk:"template_group_id"`
	AllowGuestLogin              types.Bool   `tfsdk:"allow_guest_login"`
	IsRememberMeSelected         types.Bool   `tfsdk:"is_remember_me_selected"`
	AdminClient                  types.Bool   `tfsdk:"admin_client"`
	IsLoginSuccessPageEnabled    types.Bool   `tfsdk:"is_login_success_page_enabled"`
	IsRegisterSuccessPageEnabled types.Bool   `tfsdk:"is_register_success_page_enabled"`
}

// ownershipDetailsConfig is the Terraform nested block for ownership_details (required on create).
type ownershipDetailsConfig struct {
	CompanyName    types.String `tfsdk:"company_name"`
	CompanyAddress types.String `tfsdk:"company_address"`
	CompanyWebsite types.String `tfsdk:"company_website"`
}

// extract decodes nested Terraform object attributes into typed helper structs on c.
func (c *appConfigurationConfig) extract(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	if !c.RedirectURIs.IsNull() && !c.RedirectURIs.IsUnknown() {
		c.redirectURIs = &redirectURIsConfig{}
		diags.Append(c.RedirectURIs.As(ctx, c.redirectURIs, basetypes.ObjectAsOptions{})...)
	}
	if !c.Scopes.IsNull() && !c.Scopes.IsUnknown() {
		c.scopes = &scopesConfig{}
		diags.Append(c.Scopes.As(ctx, c.scopes, basetypes.ObjectAsOptions{})...)
	}
	if !c.TokenLifetimes.IsNull() && !c.TokenLifetimes.IsUnknown() {
		c.tokenLifetimes = &tokenLifetimesConfig{}
		diags.Append(c.TokenLifetimes.As(ctx, c.tokenLifetimes, basetypes.ObjectAsOptions{})...)
	}
	if !c.AuthenticationSetup.IsNull() && !c.AuthenticationSetup.IsUnknown() {
		c.authenticationSetup = &authenticationSetupConfig{}
		diags.Append(c.AuthenticationSetup.As(ctx, c.authenticationSetup, basetypes.ObjectAsOptions{})...)
	}
	if !c.OwnershipDetails.IsNull() && !c.OwnershipDetails.IsUnknown() {
		c.ownershipDetails = &ownershipDetailsConfig{}
		diags.Append(c.OwnershipDetails.As(ctx, c.ownershipDetails, basetypes.ObjectAsOptions{})...)
	}
	if !c.ClientAuthConfig.IsNull() && !c.ClientAuthConfig.IsUnknown() {
		c.clientAuthConfig = &clientAuthConfigBlock{}
		diags.Append(c.ClientAuthConfig.As(ctx, c.clientAuthConfig, basetypes.ObjectAsOptions{})...)
	}
	return diags
}

// toModel builds the app-srv appv3 JSON payload from Terraform plan/state.
func (c *appConfigurationConfig) toModel(ctx context.Context) (client.AppConfigurationModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	model := client.AppConfigurationModel{
		ClientID:   c.ClientID.ValueString(),
		ClientName: c.ClientName.ValueString(),
		ClientType: c.ClientType.ValueString(),
		// ponytail: always owner=client — app-srv checkAppAccess and Trust Desk search require it; omitting yields 403 APP10023 and hides the app from UI.
		Owner:               client.OwnerClient,
		Enabled:             boolPtr(c.Enabled),
		HostedPagesLayoutID: c.HostedPagesLayoutID.ValueString(),
		UserSetupID:         c.UserSetupID.ValueString(),
	}
	var d diag.Diagnostics
	model.GrantTypes, d = listToStrings(ctx, c.GrantTypes)
	diags.Append(d...)
	model.ResponseTypes, d = listToStrings(ctx, c.ResponseTypes)
	diags.Append(d...)

	if c.redirectURIs != nil {
		redirectURIs, d := redirectURIsToClient(ctx, c.redirectURIs)
		diags.Append(d...)
		model.RedirectURIs = redirectURIs
	}
	if c.scopes != nil {
		scopes, d := scopesToClient(ctx, c.scopes)
		diags.Append(d...)
		model.Scopes = scopes
	}
	if c.tokenLifetimes != nil {
		model.TokenLifetimes = tokenLifetimesToClient(c.tokenLifetimes)
	}
	if c.authenticationSetup != nil {
		model.AuthenticationSetup = authenticationSetupToClient(c.authenticationSetup)
	}
	if c.ownershipDetails != nil {
		model.OwnershipDetails = &client.OwnershipDetailsConfig{
			CompanyName:    c.ownershipDetails.CompanyName.ValueString(),
			CompanyAddress: c.ownershipDetails.CompanyAddress.ValueString(),
			CompanyWebsite: c.ownershipDetails.CompanyWebsite.ValueString(),
		}
	}
	if c.clientAuthConfig != nil && !c.clientAuthConfig.TokenEndpointAuthMethod.IsNull() && !c.clientAuthConfig.TokenEndpointAuthMethod.IsUnknown() {
		model.ClientAuthConfig = &client.AuthConfig{
			TokenEndpointAuthMethod: c.clientAuthConfig.TokenEndpointAuthMethod.ValueString(),
		}
	}
	return model, diags
}

// redirectURIsToClient maps the redirect_uris nested block to the API struct.
func redirectURIsToClient(ctx context.Context, cfg *redirectURIsConfig) (*client.RedirectURIsConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := &client.RedirectURIsConfig{}
	var d diag.Diagnostics
	out.RedirectURIs, d = listToStrings(ctx, cfg.RedirectURIs)
	diags.Append(d...)
	out.AllowedLogoutUrls, d = listToStrings(ctx, cfg.AllowedLogoutUrls)
	diags.Append(d...)
	out.PostLogoutRedirectURIs, d = listToStrings(ctx, cfg.PostLogoutRedirectURIs)
	diags.Append(d...)
	out.AllowedWebOrigins, d = listToStrings(ctx, cfg.AllowedWebOrigins)
	diags.Append(d...)
	return out, diags
}

// scopesToClient maps the scopes nested block to the API struct.
func scopesToClient(ctx context.Context, cfg *scopesConfig) (*client.ScopesConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := &client.ScopesConfig{}
	var d diag.Diagnostics
	out.AllowedScopes, d = listToStrings(ctx, cfg.AllowedScopes)
	diags.Append(d...)
	out.DefaultScopes, d = listToStrings(ctx, cfg.DefaultScopes)
	diags.Append(d...)
	return out, diags
}

// tokenLifetimesToClient maps the token_lifetimes nested block to the API struct.
func tokenLifetimesToClient(cfg *tokenLifetimesConfig) *client.TokenLifetimesConfig {
	return &client.TokenLifetimesConfig{
		TokenLifetimeInSeconds:        int64Ptr(cfg.TokenLifetimeInSeconds),
		RefreshTokenLifetimeInSeconds: int64Ptr(cfg.RefreshTokenLifetimeInSeconds),
		IDTokenLifetimeInSeconds:      int64Ptr(cfg.IDTokenLifetimeInSeconds),
		CodeLifetimeInSeconds:         int64Ptr(cfg.CodeLifetimeInSeconds),
		DefaultMaxAge:                 int64Ptr(cfg.DefaultMaxAge),
	}
}

// authenticationSetupToClient maps the authentication_setup nested block to the API struct.
func authenticationSetupToClient(cfg *authenticationSetupConfig) *client.AuthenticationSetupConfig {
	return &client.AuthenticationSetupConfig{
		VerificationOptionsID:        cfg.VerificationOptionsID.ValueString(),
		GroupSelectionID:             cfg.GroupSelectionID.ValueString(),
		GroupVerificationRequestID:   cfg.GroupVerificationRequestID.ValueString(),
		TemplateGroupID:              cfg.TemplateGroupID.ValueString(),
		AllowGuestLogin:              boolPtr(cfg.AllowGuestLogin),
		IsRememberMeSelected:         boolPtr(cfg.IsRememberMeSelected),
		AdminClient:                  boolPtr(cfg.AdminClient),
		IsLoginSuccessPageEnabled:    boolPtr(cfg.IsLoginSuccessPageEnabled),
		IsRegisterSuccessPageEnabled: boolPtr(cfg.IsRegisterSuccessPageEnabled),
	}
}

// flattenAppConfiguration maps an app-srv appv3 response into Terraform state.
func flattenAppConfiguration(model client.AppConfigurationModel) (appConfigurationConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := appConfigurationConfig{
		ClientID:            types.StringValue(model.ClientID),
		ClientName:          types.StringValue(model.ClientName),
		ClientType:          types.StringValue(model.ClientType),
		Owner:               stringOrNull(model.Owner),
		Enabled:             boolValueOrNull(model.Enabled),
		GrantTypes:          stringList(model.GrantTypes),
		ResponseTypes:       stringList(model.ResponseTypes),
		HostedPagesLayoutID: stringOrNull(model.HostedPagesLayoutID),
		UserSetupID:         stringOrNull(model.UserSetupID),
		CreatedTime:         stringOrNull(model.CreatedTime),
		UpdatedTime:         stringOrNull(model.UpdatedTime),
	}
	if model.RedirectURIs != nil {
		obj, d := types.ObjectValue(redirectURIsAttrTypes(), map[string]attr.Value{
			"redirect_uris":             stringList(model.RedirectURIs.RedirectURIs),
			"allowed_logout_urls":       stringList(model.RedirectURIs.AllowedLogoutUrls),
			"post_logout_redirect_uris": stringList(model.RedirectURIs.PostLogoutRedirectURIs),
			"allowed_web_origins":       stringList(model.RedirectURIs.AllowedWebOrigins),
		})
		diags.Append(d...)
		cfg.RedirectURIs = obj
	} else {
		cfg.RedirectURIs = types.ObjectNull(redirectURIsAttrTypes())
	}
	if model.Scopes != nil {
		obj, d := types.ObjectValue(scopesAttrTypes(), map[string]attr.Value{
			"allowed_scopes": stringList(model.Scopes.AllowedScopes),
			"default_scopes": stringList(model.Scopes.DefaultScopes),
		})
		diags.Append(d...)
		cfg.Scopes = obj
	} else {
		cfg.Scopes = types.ObjectNull(scopesAttrTypes())
	}
	if model.TokenLifetimes != nil {
		tl := model.TokenLifetimes
		obj, d := types.ObjectValue(tokenLifetimesAttrTypes(), map[string]attr.Value{
			"token_lifetime_in_seconds":         int64ValueOrNull(tl.TokenLifetimeInSeconds),
			"refresh_token_lifetime_in_seconds": int64ValueOrNull(tl.RefreshTokenLifetimeInSeconds),
			"id_token_lifetime_in_seconds":      int64ValueOrNull(tl.IDTokenLifetimeInSeconds),
			"code_lifetime_in_seconds":          int64ValueOrNull(tl.CodeLifetimeInSeconds),
			"default_max_age":                   int64ValueOrNull(tl.DefaultMaxAge),
		})
		diags.Append(d...)
		cfg.TokenLifetimes = obj
	} else {
		cfg.TokenLifetimes = types.ObjectNull(tokenLifetimesAttrTypes())
	}
	if model.AuthenticationSetup != nil {
		a := model.AuthenticationSetup
		obj, d := types.ObjectValue(authenticationSetupAttrTypes(), map[string]attr.Value{
			"verification_options_id":          stringOrNull(a.VerificationOptionsID),
			"group_selection_id":               stringOrNull(a.GroupSelectionID),
			"group_verification_request_id":    stringOrNull(a.GroupVerificationRequestID),
			"template_group_id":                stringOrNull(a.TemplateGroupID),
			"allow_guest_login":                boolValueOrNull(a.AllowGuestLogin),
			"is_remember_me_selected":          boolValueOrNull(a.IsRememberMeSelected),
			"admin_client":                     boolValueOrNull(a.AdminClient),
			"is_login_success_page_enabled":    boolValueOrNull(a.IsLoginSuccessPageEnabled),
			"is_register_success_page_enabled": boolValueOrNull(a.IsRegisterSuccessPageEnabled),
		})
		diags.Append(d...)
		cfg.AuthenticationSetup = obj
	} else {
		cfg.AuthenticationSetup = types.ObjectNull(authenticationSetupAttrTypes())
	}
	if model.OwnershipDetails != nil {
		o := model.OwnershipDetails
		obj, d := types.ObjectValue(ownershipDetailsAttrTypes(), map[string]attr.Value{
			"company_name":    types.StringValue(o.CompanyName),
			"company_address": types.StringValue(o.CompanyAddress),
			"company_website": types.StringValue(o.CompanyWebsite),
		})
		diags.Append(d...)
		cfg.OwnershipDetails = obj
	}
	if model.ClientAuthConfig != nil {
		obj, d := types.ObjectValue(clientAuthConfigAttrTypes(), map[string]attr.Value{
			"token_endpoint_auth_method": stringOrNull(model.ClientAuthConfig.TokenEndpointAuthMethod),
		})
		diags.Append(d...)
		cfg.ClientAuthConfig = obj
	} else {
		cfg.ClientAuthConfig = types.ObjectNull(clientAuthConfigAttrTypes())
	}
	if model.SigningKeyConfig != nil {
		obj, d := types.ObjectValue(signingKeyConfigAttrTypes(), map[string]attr.Value{
			"active_kid": stringOrNull(model.SigningKeyConfig.ActiveKID),
			"next_kid":   stringOrNull(model.SigningKeyConfig.NextKID),
		})
		diags.Append(d...)
		cfg.SigningKeyConfig = obj
	} else {
		cfg.SigningKeyConfig = types.ObjectNull(signingKeyConfigAttrTypes())
	}
	return cfg, diags
}
