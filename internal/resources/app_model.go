package resources

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const (
	CREATE = 1
	READ   = 2
	IMPORT = 3
	UPDATE = 4
)

type AppConfig struct {
	ID                              types.String `tfsdk:"id"`
	ClientType                      types.String `tfsdk:"client_type"`
	AccentColor                     types.String `tfsdk:"accent_color"`
	PrimaryColor                    types.String `tfsdk:"primary_color"`
	MediaType                       types.String `tfsdk:"media_type"`
	ContentAlign                    types.String `tfsdk:"content_align"`
	HostedPageGroup                 types.String `tfsdk:"hosted_page_group"`
	ClientName                      types.String `tfsdk:"client_name"`
	ClientDisplayName               types.String `tfsdk:"client_display_name"`
	CompanyName                     types.String `tfsdk:"company_name"`
	CompanyAddress                  types.String `tfsdk:"company_address"`
	CompanyWebsite                  types.String `tfsdk:"company_website"`
	TemplateGroupID                 types.String `tfsdk:"template_group_id"`
	ClientID                        types.String `tfsdk:"client_id"`
	ClientSecret                    types.String `tfsdk:"client_secret"`
	ClientSecretWO                  types.String `tfsdk:"client_secret_wo"`
	ClientSecretWOVersion           types.String `tfsdk:"client_secret_wo_version"`
	PolicyURI                       types.String `tfsdk:"policy_uri"`
	TosURI                          types.String `tfsdk:"tos_uri"`
	ImprintURI                      types.String `tfsdk:"imprint_uri"`
	TokenEndpointAuthMethod         types.String `tfsdk:"token_endpoint_auth_method"`
	TokenEndpointAuthSigningAlg     types.String `tfsdk:"token_endpoint_auth_signing_alg"`
	CaptchaRef                      types.String `tfsdk:"captcha_ref"`
	CommunicationMediumVerification types.String `tfsdk:"communication_medium_verification"`
	BotProvider                     types.String `tfsdk:"bot_provider"`
	BackchannelLogoutURI            types.String `tfsdk:"backchannel_logout_uri"`
	LogoAlign                       types.String `tfsdk:"logo_align"`
	Webfinger                       types.String `tfsdk:"webfinger"`
	LogoURI                         types.String `tfsdk:"logo_uri"`
	InitiateLoginURI                types.String `tfsdk:"initiate_login_uri"`
	RegistrationClientURI           types.String `tfsdk:"registration_client_uri"`
	RegistrationAccessToken         types.String `tfsdk:"registration_access_token"`
	ClientURI                       types.String `tfsdk:"client_uri"`
	JwksURI                         types.String `tfsdk:"jwks_uri"`
	Jwks                            types.String `tfsdk:"jwks"`
	SectorIdentifierURI             types.String `tfsdk:"sector_identifier_uri"`
	SubjectType                     types.String `tfsdk:"subject_type"`
	IDTokenSignedResponseAlg        types.String `tfsdk:"id_token_signed_response_alg"`
	IDTokenEncryptedResponseAlg     types.String `tfsdk:"id_token_encrypted_response_alg"`
	IDTokenEncryptedResponseEnc     types.String `tfsdk:"id_token_encrypted_response_enc"`
	UserinfoSignedResponseAlg       types.String `tfsdk:"userinfo_signed_response_alg"`
	UserinfoEncryptedResponseAlg    types.String `tfsdk:"userinfo_encrypted_response_alg"`
	UserinfoEncryptedResponseEnc    types.String `tfsdk:"userinfo_encrypted_response_enc"`
	RequestObjectSigningAlg         types.String `tfsdk:"request_object_signing_alg"`
	RequestObjectEncryptionAlg      types.String `tfsdk:"request_object_encryption_alg"`
	RequestObjectEncryptionEnc      types.String `tfsdk:"request_object_encryption_enc"`
	Description                     types.String `tfsdk:"description"`
	ConsentPageGroup                types.String `tfsdk:"consent_page_group"`
	PasswordPolicyRef               types.String `tfsdk:"password_policy_ref"`
	BlockingMechanismRef            types.String `tfsdk:"blocking_mechanism_ref"`
	Sub                             types.String `tfsdk:"sub"`
	Role                            types.String `tfsdk:"role"`
	MfaConfiguration                types.String `tfsdk:"mfa_configuration"`
	BackgroundURI                   types.String `tfsdk:"background_uri"`
	VideoURL                        types.String `tfsdk:"video_url"`
	BotCaptchaRef                   types.String `tfsdk:"bot_captcha_ref"`
	OauthStandard                   types.String `tfsdk:"oauth_standard"`
	RequirePKCE                     types.Bool   `tfsdk:"require_pkce"`
	DisableInsecurePKCEMethod       types.Bool   `tfsdk:"disable_insecure_pkce_method"`

	EnableDeduplication              types.Bool `tfsdk:"enable_deduplication"`
	AutoLoginAfterRegister           types.Bool `tfsdk:"auto_login_after_register"`
	EnablePasswordlessAuth           types.Bool `tfsdk:"enable_passwordless_auth"`
	RegisterWithLoginInformation     types.Bool `tfsdk:"register_with_login_information"`
	AllowDisposableEmail             types.Bool `tfsdk:"allow_disposable_email"`
	ValidatePhoneNumber              types.Bool `tfsdk:"validate_phone_number"`
	IsHybridApp                      types.Bool `tfsdk:"is_hybrid_app"`
	JweEnabled                       types.Bool `tfsdk:"jwe_enabled"`
	UserConsent                      types.Bool `tfsdk:"user_consent"`
	Enabled                          types.Bool `tfsdk:"enabled"`
	SmartMfa                         types.Bool `tfsdk:"smart_mfa"`
	IsRememberMeSelected             types.Bool `tfsdk:"is_remember_me_selected"`
	EnableBotDetection               types.Bool `tfsdk:"enable_bot_detection"`
	IsLoginSuccessPageEnabled        types.Bool `tfsdk:"is_login_success_page_enabled"`
	IsRegisterSuccessPageEnabled     types.Bool `tfsdk:"is_register_success_page_enabled"`
	IsGroupLoginSelectionEnabled     types.Bool `tfsdk:"is_group_login_selection_enabled"`
	AllowGuestLogin                  types.Bool `tfsdk:"allow_guest_login"`
	RequireAuthTime                  types.Bool `tfsdk:"require_auth_time"`
	EnableLoginSpi                   types.Bool `tfsdk:"enable_login_spi"`
	BackchannelLogoutSessionRequired types.Bool `tfsdk:"backchannel_logout_session_required"`
	AcceptRolesInTheRegistration     types.Bool `tfsdk:"accept_roles_in_the_registration"`

	DefaultMaxAge                 types.Int64 `tfsdk:"default_max_age"`
	TokenLifetimeInSeconds        types.Int64 `tfsdk:"token_lifetime_in_seconds"`
	IDTokenLifetimeInSeconds      types.Int64 `tfsdk:"id_token_lifetime_in_seconds"`
	RefreshTokenLifetimeInSeconds types.Int64 `tfsdk:"refresh_token_lifetime_in_seconds"`

	AllowLoginWith               types.Set `tfsdk:"allow_login_with"`
	RedirectURIS                 types.Set `tfsdk:"redirect_uris"`
	AllowedLogoutUrls            types.Set `tfsdk:"allowed_logout_urls"`
	AllowedScopes                types.Set `tfsdk:"allowed_scopes"`
	ResponseTypes                types.Set `tfsdk:"response_types"`
	GrantTypes                   types.Set `tfsdk:"grant_types"`
	LoginProviders               types.Set `tfsdk:"login_providers"`
	AdditionalAccessTokenPayload types.Set `tfsdk:"additional_access_token_payload"`
	RequiredFields               types.Set `tfsdk:"required_fields"`
	AllowedWebOrigins            types.Set `tfsdk:"allowed_web_origins"`
	AllowedOrigins               types.Set `tfsdk:"allowed_origins"`
	AllowedNativeClients         types.Set `tfsdk:"allowed_native_clients"`
	Contacts                     types.Set `tfsdk:"contacts"`
	DefaultAcrValues             types.Set `tfsdk:"default_acr_values"`
	WebMessageUris               types.Set `tfsdk:"web_message_uris"`
	AllowedFields                types.Set `tfsdk:"allowed_fields"`
	AllowedMfa                   types.Set `tfsdk:"allowed_mfa"`
	CaptchaRefs                  types.Set `tfsdk:"captcha_refs"`
	ConsentRefs                  types.Set `tfsdk:"consent_refs"`
	AllowedRoles                 types.Set `tfsdk:"allowed_roles"`
	DefaultRoles                 types.Set `tfsdk:"default_roles"`
	GroupIDs                     types.Set `tfsdk:"group_ids"`
	GroupTypes                   types.Set `tfsdk:"group_types"`
	PostLogoutRedirectUris       types.Set `tfsdk:"post_logout_redirect_uris"`
	RequestUris                  types.Set `tfsdk:"request_uris"`
	DefaultScopes                types.Set `tfsdk:"default_scopes"`
	PendingScopes                types.Set `tfsdk:"pending_scopes"`
	SuggestMfa                   types.Set `tfsdk:"suggest_mfa"`

	ApplicationMetaData types.Map `tfsdk:"application_meta_data"`

	SocialProviders         types.List `tfsdk:"social_providers"`
	CustomProviders         types.List `tfsdk:"custom_providers"`
	SamlProviders           types.List `tfsdk:"saml_providers"`
	AdProviders             types.List `tfsdk:"ad_providers"`
	AllowedGroups           types.List `tfsdk:"allowed_groups"`
	OperationsAllowedGroups types.List `tfsdk:"operations_allowed_groups"`
	AllowGuestLoginGroups   types.List `tfsdk:"allow_guest_login_groups"`

	socialProviders         []*SocialProviderData
	customProviders         []*ProviderMetadData
	samlProviders           []*ProviderMetadData
	adProviders             []*ProviderMetadData
	allowedGroups           []*AllowedGroups
	operationsAllowedGroups []*AllowedGroups
	allowGuestLoginGroups   []*AllowedGroups

	LoginSpi                   types.Object `tfsdk:"login_spi"`
	GroupSelection             types.Object `tfsdk:"group_selection"`
	Mfa                        types.Object `tfsdk:"mfa"`
	MobileSettings             types.Object `tfsdk:"mobile_settings"`
	SuggestVerificationMethods types.Object `tfsdk:"suggest_verification_methods"`
	GroupRoleRestriction       types.Object `tfsdk:"group_role_restriction"`

	loginSpi                   *LoginSPI
	groupSelection             *GroupSelection
	mfa                        *MfaOption
	mobileSettings             *AppMobileSettings
	suggestVerificationMethods *SuggestVerificationMethods
	groupRoleRestriction       *GroupRoleRestriction
}

type AllowedGroups struct {
	GroupType    types.String `tfsdk:"group_type"`
	GroupID      types.String `tfsdk:"group_id"`
	Roles        types.Set    `tfsdk:"roles"`
	DefaultRoles types.Set    `tfsdk:"default_roles"`
}

type LoginSPI struct {
	OauthClientID types.String `tfsdk:"oauth_client_id"`
	SpiURL        types.String `tfsdk:"spi_url"`
}

type AppMobileSettings struct {
	TeamID      types.String `tfsdk:"team_id"`
	BundleID    types.String `tfsdk:"bundle_id"`
	PackageName types.String `tfsdk:"package_name"`
	KeyHash     types.String `tfsdk:"key_hash"`
}

type GroupSelection struct {
	AlwaysShowGroupSelection types.Bool `tfsdk:"always_show_group_selection"`
	SelectableGroups         types.Set  `tfsdk:"selectable_groups"`
	SelectableGroupTypes     types.Set  `tfsdk:"selectable_group_types"`
}

type MfaOption struct {
	Setting               types.String `tfsdk:"setting"`
	TimeIntervalInSeconds types.Int64  `tfsdk:"time_interval_in_seconds"`
	AllowedMethods        types.Set    `tfsdk:"allowed_methods"`
}

type SocialProviderData struct {
	ProviderName types.String `tfsdk:"provider_name"`
	SocialID     types.String `tfsdk:"social_id"`
}

type ProviderMetadData struct {
	LogoURL           types.String `tfsdk:"logo_url"`
	ProviderName      types.String `tfsdk:"provider_name"`
	DisplayName       types.String `tfsdk:"display_name"`
	Type              types.String `tfsdk:"type"`
	IsProviderVisible types.Bool   `tfsdk:"is_provider_visible"`
	Domains           types.Set    `tfsdk:"domains"`
}
type SuggestVerificationMethods struct {
	MandatoryConfig    types.Object `tfsdk:"mandatory_config"`
	OptionalConfig     types.Object `tfsdk:"optional_config"`
	SkipDurationInDays types.Int32  `tfsdk:"skip_duration_in_days"`
}

type MandatoryConfig struct {
	Methods   types.Set    `tfsdk:"methods"`
	SkipUntil types.String `tfsdk:"skip_until"`
	Range     types.String `tfsdk:"range"`
}

type OptionalConfig struct {
	Methods types.Set `tfsdk:"methods"`
}

type GroupRoleRestriction struct {
	MatchCondition types.String `tfsdk:"match_condition"`
	Filters        types.List   `tfsdk:"filters"`
	Hints          types.Set    `tfsdk:"hints"`
}
type GroupRoleFilters struct {
	GroupID    types.String `tfsdk:"group_id"`
	GroupType  types.String `tfsdk:"group_type"`
	RoleFilter types.Object `tfsdk:"role_filter"`
}

type RoleFilter struct {
	MatchCondition types.String `tfsdk:"match_condition"`
	Roles          types.Set    `tfsdk:"roles"`
}

func (w *AppConfig) ExtractAppConfigs(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	if !w.LoginSpi.IsNull() && !w.LoginSpi.IsUnknown() {
		w.loginSpi = &LoginSPI{}
		diags = w.LoginSpi.As(ctx, w.loginSpi, basetypes.ObjectAsOptions{})
	}
	if !w.GroupSelection.IsNull() && !w.GroupSelection.IsUnknown() {
		w.groupSelection = &GroupSelection{}
		diags = w.GroupSelection.As(ctx, w.groupSelection, basetypes.ObjectAsOptions{})
	}
	if !w.Mfa.IsNull() && !w.Mfa.IsUnknown() {
		w.mfa = &MfaOption{}
		diags = w.Mfa.As(ctx, w.mfa, basetypes.ObjectAsOptions{})
	}
	if !w.MobileSettings.IsNull() && !w.MobileSettings.IsUnknown() {
		w.mobileSettings = &AppMobileSettings{}
		diags = w.MobileSettings.As(ctx, w.mobileSettings, basetypes.ObjectAsOptions{})
	}
	if !w.SocialProviders.IsNull() && !w.SocialProviders.IsUnknown() {
		w.socialProviders = make([]*SocialProviderData, 0, len(w.SocialProviders.Elements()))
		diags = w.SocialProviders.ElementsAs(ctx, &w.socialProviders, false)
	}
	if !w.CustomProviders.IsNull() && !w.CustomProviders.IsUnknown() {
		w.customProviders = make([]*ProviderMetadData, 0, len(w.CustomProviders.Elements()))
		diags = w.CustomProviders.ElementsAs(ctx, &w.customProviders, false)
	}
	if !w.SamlProviders.IsNull() && !w.SamlProviders.IsUnknown() {
		w.samlProviders = make([]*ProviderMetadData, 0, len(w.SamlProviders.Elements()))
		diags = w.SamlProviders.ElementsAs(ctx, &w.samlProviders, false)
	}
	if !w.AdProviders.IsNull() && !w.AdProviders.IsUnknown() {
		w.adProviders = make([]*ProviderMetadData, 0, len(w.AdProviders.Elements()))
		diags = w.AdProviders.ElementsAs(ctx, &w.adProviders, false)
	}
	if !w.AllowedGroups.IsNull() && !w.AllowedGroups.IsUnknown() {
		w.allowedGroups = make([]*AllowedGroups, 0, len(w.AllowedGroups.Elements()))
		diags = w.AllowedGroups.ElementsAs(ctx, &w.allowedGroups, false)
	}
	if !w.OperationsAllowedGroups.IsNull() && !w.OperationsAllowedGroups.IsUnknown() {
		w.operationsAllowedGroups = make([]*AllowedGroups, 0, len(w.OperationsAllowedGroups.Elements()))
		diags = w.OperationsAllowedGroups.ElementsAs(ctx, &w.operationsAllowedGroups, false)
	}
	if !w.AllowGuestLoginGroups.IsNull() && !w.AllowGuestLoginGroups.IsUnknown() {
		w.allowGuestLoginGroups = make([]*AllowedGroups, 0, len(w.AllowGuestLoginGroups.Elements()))
		diags = w.AllowGuestLoginGroups.ElementsAs(ctx, &w.allowGuestLoginGroups, false)
	}
	if !w.SuggestVerificationMethods.IsNull() && !w.SuggestVerificationMethods.IsUnknown() {
		w.suggestVerificationMethods = &SuggestVerificationMethods{}
		diags = w.SuggestVerificationMethods.As(ctx, w.suggestVerificationMethods, basetypes.ObjectAsOptions{})
	}
	if !w.GroupRoleRestriction.IsNull() && !w.GroupRoleRestriction.IsUnknown() {
		w.groupRoleRestriction = &GroupRoleRestriction{}
		diags = w.GroupRoleRestriction.As(ctx, w.groupRoleRestriction, basetypes.ObjectAsOptions{})
	}
	return diags
}

func prepareAppModel(ctx context.Context, plan AppConfig) (*cidaas.AppModel, diag.Diagnostics) { //nolint:gocognit,gocyclo

	assignSetValues := func(ctx context.Context, planValue basetypes.SetValue, target *[]string) diag.Diagnostics {
		if !planValue.IsNull() && len(planValue.Elements()) > 0 {
			return planValue.ElementsAs(ctx, target, false)
		}
		return nil
	}

	app := cidaas.AppModel{
		AppOwner:                      "CLIENT",
		CompanyName:                   plan.CompanyName.ValueString(),
		CompanyAddress:                plan.CompanyAddress.ValueString(),
		CompanyWebsite:                plan.CompanyWebsite.ValueString(),
		ClientType:                    plan.ClientType.ValueString(),
		AccentColor:                   plan.AccentColor.ValueString(),
		PrimaryColor:                  plan.PrimaryColor.ValueString(),
		MediaType:                     plan.MediaType.ValueString(),
		ContentAlign:                  plan.ContentAlign.ValueString(),
		EnableDeduplication:           plan.EnableDeduplication.ValueBoolPointer(),
		AutoLoginAfterRegister:        plan.AutoLoginAfterRegister.ValueBoolPointer(),
		EnablePasswordlessAuth:        plan.EnablePasswordlessAuth.ValueBoolPointer(),
		RegisterWithLoginInformation:  plan.RegisterWithLoginInformation.ValueBoolPointer(),
		AllowDisposableEmail:          plan.AllowDisposableEmail.ValueBoolPointer(),
		ValidatePhoneNumber:           plan.ValidatePhoneNumber.ValueBoolPointer(),
		HostedPageGroup:               plan.HostedPageGroup.ValueString(),
		ClientName:                    plan.ClientName.ValueString(),
		ClientDisplayName:             plan.ClientDisplayName.ValueString(),
		IsHybridApp:                   plan.IsHybridApp.ValueBoolPointer(),
		DefaultMaxAge:                 plan.DefaultMaxAge.ValueInt64Pointer(),
		TokenLifetimeInSeconds:        plan.TokenLifetimeInSeconds.ValueInt64Pointer(),
		IDTokenLifetimeInSeconds:      plan.IDTokenLifetimeInSeconds.ValueInt64Pointer(),
		RefreshTokenLifetimeInSeconds: plan.RefreshTokenLifetimeInSeconds.ValueInt64Pointer(),
		TemplateGroupID:               plan.TemplateGroupID.ValueString(),
		ClientID:                      plan.ClientID.ValueString(),
		ClientSecret:                  plan.ClientSecret.ValueString(),
		PolicyURI:                     plan.PolicyURI.ValueString(),
		TosURI:                        plan.TosURI.ValueString(),
		ImprintURI:                    plan.ImprintURI.ValueString(),
		TokenEndpointAuthMethod:       plan.TokenEndpointAuthMethod.ValueString(),
		TokenEndpointAuthSigningAlg:   plan.TokenEndpointAuthSigningAlg.ValueString(),
		JweEnabled:                    plan.JweEnabled.ValueBoolPointer(),
		UserConsent:                   plan.UserConsent.ValueBoolPointer(),
		Enabled:                       plan.Enabled.ValueBoolPointer(),
		// AlwaysAskMfa:                     plan.AlwaysAskMfa.ValueBoolPointer(),
		SmartMfa:                         plan.SmartMfa.ValueBoolPointer(),
		CaptchaRef:                       plan.CaptchaRef.ValueString(),
		CommunicationMediumVerification:  plan.CommunicationMediumVerification.ValueString(),
		IsRememberMeSelected:             plan.IsRememberMeSelected.ValueBoolPointer(),
		EnableBotDetection:               plan.EnableBotDetection.ValueBoolPointer(),
		BotProvider:                      plan.BotProvider.ValueString(),
		IsLoginSuccessPageEnabled:        plan.IsLoginSuccessPageEnabled.ValueBoolPointer(),
		IsRegisterSuccessPageEnabled:     plan.IsRegisterSuccessPageEnabled.ValueBoolPointer(),
		IsGroupLoginSelectionEnabled:     plan.IsGroupLoginSelectionEnabled.ValueBoolPointer(),
		BackchannelLogoutURI:             plan.BackchannelLogoutURI.ValueString(),
		LogoAlign:                        plan.LogoAlign.ValueString(),
		Webfinger:                        plan.Webfinger.ValueString(),
		LogoURI:                          plan.LogoURI.ValueString(),
		InitiateLoginURI:                 plan.InitiateLoginURI.ValueString(),
		RegistrationClientURI:            plan.RegistrationClientURI.ValueString(),
		RegistrationAccessToken:          plan.RegistrationAccessToken.ValueString(),
		ClientURI:                        plan.ClientURI.ValueString(),
		JwksURI:                          plan.JwksURI.ValueString(),
		Jwks:                             plan.Jwks.ValueString(),
		SectorIdentifierURI:              plan.SectorIdentifierURI.ValueString(),
		SubjectType:                      plan.SubjectType.ValueString(),
		IDTokenSignedResponseAlg:         plan.IDTokenSignedResponseAlg.ValueString(),
		IDTokenEncryptedResponseAlg:      plan.IDTokenEncryptedResponseAlg.ValueString(),
		IDTokenEncryptedResponseEnc:      plan.IDTokenEncryptedResponseEnc.ValueString(),
		UserinfoSignedResponseAlg:        plan.UserinfoSignedResponseAlg.ValueString(),
		UserinfoEncryptedResponseAlg:     plan.UserinfoEncryptedResponseAlg.ValueString(),
		UserinfoEncryptedResponseEnc:     plan.UserinfoEncryptedResponseEnc.ValueString(),
		RequestObjectSigningAlg:          plan.RequestObjectSigningAlg.ValueString(),
		RequestObjectEncryptionAlg:       plan.RequestObjectEncryptionAlg.ValueString(),
		RequestObjectEncryptionEnc:       plan.RequestObjectEncryptionEnc.ValueString(),
		Description:                      plan.Description.ValueString(),
		ConsentPageGroup:                 plan.ConsentPageGroup.ValueString(),
		BlockingMechanismRef:             plan.BlockingMechanismRef.ValueString(),
		Sub:                              plan.Sub.ValueString(),
		Role:                             plan.Role.ValueString(),
		MfaConfiguration:                 plan.MfaConfiguration.ValueString(),
		AllowGuestLogin:                  plan.AllowGuestLogin.ValueBoolPointer(),
		BackgroundURI:                    plan.BackgroundURI.ValueString(),
		VideoURL:                         plan.VideoURL.ValueString(),
		BotCaptchaRef:                    plan.BotCaptchaRef.ValueString(),
		RequireAuthTime:                  plan.RequireAuthTime.ValueBoolPointer(),
		EnableLoginSpi:                   plan.EnableLoginSpi.ValueBoolPointer(),
		BackchannelLogoutSessionRequired: plan.BackchannelLogoutSessionRequired.ValueBoolPointer(),
		AcceptRolesInTheRegistration:     plan.AcceptRolesInTheRegistration.ValueBoolPointer(),
		PasswordPolicyRef:                plan.PasswordPolicyRef.ValueString(),
		RequirePKCE:                      plan.RequirePKCE.ValueBoolPointer(),
		DisableInsecurePKCEMethod:        plan.DisableInsecurePKCEMethod.ValueBoolPointer(),
		OauthStandard:                    plan.OauthStandard.ValueString(),
	}

	var diags diag.Diagnostics

	diags.Append(assignSetValues(ctx, plan.AllowedScopes, &app.AllowedScopes)...)
	diags.Append(assignSetValues(ctx, plan.RedirectURIS, &app.RedirectURIS)...)
	diags.Append(assignSetValues(ctx, plan.AllowedLogoutUrls, &app.AllowedLogoutUrls)...)
	diags.Append(assignSetValues(ctx, plan.AllowedWebOrigins, &app.AllowedWebOrigins)...)
	diags.Append(assignSetValues(ctx, plan.AllowedOrigins, &app.AllowedOrigins)...)
	diags.Append(assignSetValues(ctx, plan.AllowedNativeClients, &app.AllowedNativeClients)...)
	diags.Append(assignSetValues(ctx, plan.LoginProviders, &app.LoginProviders)...)
	diags.Append(assignSetValues(ctx, plan.DefaultScopes, &app.DefaultScopes)...)
	diags.Append(assignSetValues(ctx, plan.PendingScopes, &app.PendingScopes)...)
	diags.Append(assignSetValues(ctx, plan.AllowedMfa, &app.AllowedMfa)...)
	diags.Append(assignSetValues(ctx, plan.AllowedRoles, &app.AllowedRoles)...)
	diags.Append(assignSetValues(ctx, plan.DefaultRoles, &app.DefaultRoles)...)
	diags.Append(assignSetValues(ctx, plan.PostLogoutRedirectUris, &app.PostLogoutRedirectUris)...)
	diags.Append(assignSetValues(ctx, plan.GroupIDs, &app.GroupIDs)...)
	diags.Append(assignSetValues(ctx, plan.AllowedFields, &app.AllowedFields)...)
	diags.Append(assignSetValues(ctx, plan.WebMessageUris, &app.WebMessageUris)...)
	diags.Append(assignSetValues(ctx, plan.DefaultAcrValues, &app.DefaultAcrValues)...)
	diags.Append(assignSetValues(ctx, plan.GroupTypes, &app.GroupTypes)...)
	diags.Append(assignSetValues(ctx, plan.SuggestMfa, &app.SuggestMfa)...)
	diags.Append(assignSetValues(ctx, plan.CaptchaRefs, &app.CaptchaRefs)...)
	diags.Append(assignSetValues(ctx, plan.ConsentRefs, &app.ConsentRefs)...)
	diags.Append(assignSetValues(ctx, plan.AllowLoginWith, &app.AllowLoginWith)...)
	diags.Append(assignSetValues(ctx, plan.AllowedLogoutUrls, &app.AllowedLogoutUrls)...)
	diags.Append(assignSetValues(ctx, plan.ResponseTypes, &app.ResponseTypes)...)
	diags.Append(assignSetValues(ctx, plan.GrantTypes, &app.GrantTypes)...)
	diags.Append(assignSetValues(ctx, plan.AdditionalAccessTokenPayload, &app.AdditionalAccessTokenPayload)...)
	diags.Append(assignSetValues(ctx, plan.RequiredFields, &app.RequiredFields)...)
	diags.Append(assignSetValues(ctx, plan.Contacts, &app.Contacts)...)
	diags.Append(assignSetValues(ctx, plan.RequestUris, &app.RequestUris)...)

	// SocialProviders
	if !plan.SocialProviders.IsNull() && len(plan.socialProviders) > 0 {
		target := []cidaas.ISocialProviderData{}
		var objectValues []attr.Value
		spObjectType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"provider_name": types.StringType,
				"social_id":     types.StringType,
			},
		}
		objValue := types.ObjectValueMust(
			spObjectType.AttrTypes,
			map[string]attr.Value{
				"provider_name": types.StringNull(),
				"social_id":     types.StringNull(),
			})
		objectValues = append(objectValues, objValue)
		emptyProvider := types.ListValueMust(spObjectType, objectValues)
		if plan.SocialProviders.Equal(emptyProvider) {
			app.SocialProviders = target
		} else {
			for _, sp := range plan.socialProviders {
				target = append(target, cidaas.ISocialProviderData{
					ProviderName: sp.ProviderName.ValueString(),
					SocialID:     sp.SocialID.ValueString(),
				})
				app.SocialProviders = target
			}
		}
	}

	// helper function to get the []strings from SetValue
	getSetAsStrings := func(ctx context.Context, planValue basetypes.SetValue) ([]string, diag.Diagnostics) {
		var result []string
		diags := assignSetValues(ctx, planValue, &result)
		return result, diags
	}

	setProviders := func(ctx context.Context, source []*ProviderMetadData, target *[]cidaas.IProviderMetadData) diag.Diagnostics {
		for _, cp := range source {
			temp := cidaas.IProviderMetadData{
				ProviderName:      cp.ProviderName.ValueString(),
				Type:              cp.Type.ValueString(),
				DisplayName:       cp.DisplayName.ValueString(),
				LogoURL:           cp.LogoURL.ValueString(),
				IsProviderVisible: cp.IsProviderVisible.ValueBoolPointer(),
			}
			diags := cp.Domains.ElementsAs(ctx, &temp.Domains, false)
			if diags.HasError() {
				return diags
			}
			*target = append(*target, temp)
		}
		return nil
	}

	var objectValues []attr.Value
	spObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"provider_name":       types.StringType,
			"display_name":        types.StringType,
			"logo_url":            types.StringType,
			"type":                types.StringType,
			"is_provider_visible": types.BoolType,
			"domains":             types.SetType{ElemType: types.StringType},
		},
	}
	objValue := types.ObjectValueMust(
		spObjectType.AttrTypes,
		map[string]attr.Value{
			"provider_name":       types.StringNull(),
			"display_name":        types.StringNull(),
			"logo_url":            types.StringNull(),
			"type":                types.StringNull(),
			"is_provider_visible": types.BoolNull(),
			"domains":             types.SetNull(types.StringType),
		})
	objectValues = append(objectValues, objValue)
	emptyProvider := types.ListValueMust(spObjectType, objectValues)

	// CustomProviders
	if !plan.CustomProviders.IsNull() && len(plan.customProviders) > 0 {
		if plan.CustomProviders.Equal(emptyProvider) {
			app.CustomProviders = []cidaas.IProviderMetadData{}
		} else {
			diags.Append(setProviders(ctx, plan.customProviders, &app.CustomProviders)...)
			if diags.HasError() {
				return nil, diags
			}
		}
	}

	// SamlProviders
	if !plan.SamlProviders.IsNull() && len(plan.samlProviders) > 0 {
		if plan.SamlProviders.Equal(emptyProvider) {
			app.SamlProviders = []cidaas.IProviderMetadData{}
		} else {
			diags.Append(setProviders(ctx, plan.samlProviders, &app.SamlProviders)...)
			if diags.HasError() {
				return nil, diags
			}
		}
	}
	// AdProviders
	if !plan.AdProviders.IsNull() && len(plan.adProviders) > 0 {
		if plan.AdProviders.Equal(emptyProvider) {
			app.AdProviders = []cidaas.IProviderMetadData{}
		} else {
			diags.Append(setProviders(ctx, plan.adProviders, &app.AdProviders)...)
			if diags.HasError() {
				return nil, diags
			}
		}
	}

	var allowedGroupObjectValues []attr.Value
	agObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"group_id":      types.StringType,
			"roles":         types.SetType{ElemType: types.StringType},
			"default_roles": types.SetType{ElemType: types.StringType},
		},
	}
	value := types.ObjectValueMust(
		agObjectType.AttrTypes,
		map[string]attr.Value{
			"group_id":      types.StringNull(),
			"roles":         types.SetNull(types.StringType),
			"default_roles": types.SetNull(types.StringType),
		})
	allowedGroupObjectValues = append(allowedGroupObjectValues, value)
	emptyAllowedGroups := types.ListValueMust(agObjectType, allowedGroupObjectValues)

	// AllowedGroups
	if !plan.AllowedGroups.IsNull() && len(plan.allowedGroups) > 0 {
		if plan.AllowedGroups.Equal(emptyAllowedGroups) {
			app.AllowedGroups = []cidaas.IAllowedGroups{}
		} else {
			for _, ag := range plan.allowedGroups {
				temp := cidaas.IAllowedGroups{
					GroupID:   ag.GroupID.ValueString(),
					GroupType: ag.GroupType.ValueString(),
				}
				diags := ag.Roles.ElementsAs(ctx, &temp.Roles, false)
				if diags.HasError() {
					return nil, diags
				}
				diags = ag.DefaultRoles.ElementsAs(ctx, &temp.DefaultRoles, false)
				if diags.HasError() {
					return nil, diags
				}
				app.AllowedGroups = append(app.AllowedGroups, temp)
			}
		}
	}

	// OperationsAllowedGroups
	if !plan.OperationsAllowedGroups.IsNull() && len(plan.operationsAllowedGroups) > 0 {
		if plan.OperationsAllowedGroups.Equal(emptyAllowedGroups) {
			app.OperationsAllowedGroups = []cidaas.IAllowedGroups{}
		} else {
			for _, oag := range plan.operationsAllowedGroups {
				temp := cidaas.IAllowedGroups{
					GroupID: oag.GroupID.ValueString(),
				}
				diags := oag.Roles.ElementsAs(ctx, &temp.Roles, false)
				if diags.HasError() {
					return nil, diags
				}
				diags = oag.DefaultRoles.ElementsAs(ctx, &temp.DefaultRoles, false)
				if diags.HasError() {
					return nil, diags
				}
				app.OperationsAllowedGroups = append(app.OperationsAllowedGroups, temp)
			}
		}
	}

	// AllowGuestLoginGroups
	if !plan.AllowGuestLoginGroups.IsNull() && len(plan.allowGuestLoginGroups) > 0 {
		for _, aglg := range plan.allowGuestLoginGroups {
			temp := cidaas.IAllowedGroups{
				GroupID: aglg.GroupID.ValueString(),
			}
			diags := aglg.Roles.ElementsAs(ctx, &temp.Roles, false)
			if diags.HasError() {
				return nil, diags
			}
			diags = aglg.DefaultRoles.ElementsAs(ctx, &temp.DefaultRoles, false)
			if diags.HasError() {
				return nil, diags
			}
			app.AllowGuestLoginGroups = append(app.AllowGuestLoginGroups, temp)
		}
	}

	// LoginSpi
	if plan.loginSpi != nil {
		loginSpi := &cidaas.ILoginSPI{}
		if !plan.loginSpi.OauthClientID.IsNull() {
			loginSpi.OauthClientID = plan.loginSpi.OauthClientID.ValueString()
		}
		if !plan.loginSpi.SpiURL.IsNull() {
			loginSpi.SpiURL = plan.loginSpi.SpiURL.ValueString()
		}
		app.LoginSpi = loginSpi
	}
	// GroupSelection
	if plan.groupSelection != nil {
		groupSelection := &cidaas.IGroupSelection{}
		if !plan.groupSelection.AlwaysShowGroupSelection.IsNull() && !plan.groupSelection.AlwaysShowGroupSelection.IsUnknown() {
			groupSelection.AlwaysShowGroupSelection = plan.groupSelection.AlwaysShowGroupSelection.ValueBoolPointer()
		}
		if !plan.groupSelection.SelectableGroups.IsNull() && !plan.groupSelection.SelectableGroups.IsUnknown() {
			selectableGroups, diags := getSetAsStrings(ctx, plan.groupSelection.SelectableGroups)
			if diags.HasError() {
				return nil, diags
			}
			groupSelection.SelectableGroups = selectableGroups
		}

		if !plan.groupSelection.SelectableGroupTypes.IsNull() && !plan.groupSelection.SelectableGroupTypes.IsUnknown() {
			selectableGroupTypes, diags := getSetAsStrings(ctx, plan.groupSelection.SelectableGroupTypes)
			if diags.HasError() {
				return nil, diags
			}
			groupSelection.SelectableGroupTypes = selectableGroupTypes
		}
		app.GroupSelection = groupSelection
	}
	// Mfa
	if plan.mfa != nil {
		mfa := &cidaas.IMfaOption{}
		if !plan.mfa.Setting.IsNull() && !plan.mfa.Setting.IsUnknown() {
			mfa.Setting = plan.mfa.Setting.ValueString()
		}
		if !plan.mfa.TimeIntervalInSeconds.IsNull() && !plan.mfa.TimeIntervalInSeconds.IsUnknown() {
			mfa.TimeIntervalInSeconds = plan.mfa.TimeIntervalInSeconds.ValueInt64Pointer()
		}
		if !plan.mfa.AllowedMethods.IsNull() && !plan.mfa.AllowedMethods.IsUnknown() {
			allowedMethods, diags := getSetAsStrings(ctx, plan.mfa.AllowedMethods)
			if diags.HasError() {
				return nil, diags
			}
			mfa.AllowedMethods = allowedMethods
		}
		app.Mfa = mfa
	}
	// MobileSettings
	if plan.mobileSettings != nil {
		mobileSettings := &cidaas.IAppMobileSettings{}
		if !plan.mobileSettings.TeamID.IsNull() && !plan.mobileSettings.TeamID.IsUnknown() {
			mobileSettings.TeamID = plan.mobileSettings.TeamID.ValueString()
		}
		if !plan.mobileSettings.BundleID.IsNull() && !plan.mobileSettings.BundleID.IsUnknown() {
			mobileSettings.BundleID = plan.mobileSettings.BundleID.ValueString()
		}
		if !plan.mobileSettings.PackageName.IsNull() && !plan.mobileSettings.PackageName.IsUnknown() {
			mobileSettings.PackageName = plan.mobileSettings.PackageName.ValueString()
		}
		if !plan.mobileSettings.KeyHash.IsNull() && !plan.mobileSettings.KeyHash.IsUnknown() {
			mobileSettings.KeyHash = plan.mobileSettings.KeyHash.ValueString()
		}
		app.MobileSettings = mobileSettings
	}
	// ApplicationMetaData
	if len(plan.ApplicationMetaData.Elements()) > 0 {
		diags.Append(plan.ApplicationMetaData.ElementsAs(ctx, &app.ApplicationMetaData, false)...)
	}
	// SuggestVerificationMethods
	if plan.suggestVerificationMethods != nil {
		svm := &cidaas.SuggestVerificationMethods{}
		if !plan.suggestVerificationMethods.MandatoryConfig.IsNull() && !plan.suggestVerificationMethods.MandatoryConfig.IsUnknown() {
			mf := &MandatoryConfig{}
			diags = plan.suggestVerificationMethods.MandatoryConfig.As(ctx, mf, basetypes.ObjectAsOptions{})
			diags.Append(assignSetValues(ctx, mf.Methods, &svm.MandatoryConfig.Methods)...)
			svm.MandatoryConfig.Range = mf.Range.ValueString()
			svm.MandatoryConfig.SkipUntil = mf.SkipUntil.ValueString()
		}
		if !plan.suggestVerificationMethods.OptionalConfig.IsNull() && !plan.suggestVerificationMethods.OptionalConfig.IsUnknown() {
			of := &OptionalConfig{}
			diags = plan.suggestVerificationMethods.OptionalConfig.As(ctx, of, basetypes.ObjectAsOptions{})
			diags.Append(assignSetValues(ctx, of.Methods, &svm.OptionalConfig.Methods)...)
		}
		svm.SkipDurationInDays = plan.suggestVerificationMethods.SkipDurationInDays.ValueInt32()
		app.SuggestVerificationMethods = svm
	}
	// GroupRoleRestriction
	if plan.groupRoleRestriction != nil {
		grr := &cidaas.GroupRoleRestriction{}
		if !plan.groupRoleRestriction.Filters.IsNull() && !plan.groupRoleRestriction.Filters.IsUnknown() {
			filters := make([]GroupRoleFilters, 0, len(plan.groupRoleRestriction.Filters.Elements()))
			diags = plan.groupRoleRestriction.Filters.ElementsAs(ctx, &filters, false)
			target := []cidaas.GroupRoleFilters{}
			for _, f := range filters {
				var rf *cidaas.RoleFilter
				if !f.RoleFilter.IsNull() && !f.RoleFilter.IsUnknown() {
					tfrf := &RoleFilter{}
					diags = f.RoleFilter.As(ctx, tfrf, basetypes.ObjectAsOptions{})
					candidate := &cidaas.RoleFilter{}
					diags.Append(assignSetValues(ctx, tfrf.Roles, &candidate.Roles)...)
					candidate.MatchCondition = tfrf.MatchCondition.ValueString()
					// Omit empty roleFilter from the API payload so Terraform null and
					// API-normalized {} do not drift (Go omitempty never skips empty structs).
					if !candidate.IsEmpty() {
						rf = candidate
					}
				}
				target = append(target, cidaas.GroupRoleFilters{
					GroupID:    f.GroupID.ValueString(),
					GroupType:  f.GroupType.ValueString(),
					RoleFilter: rf,
				})
				grr.Filters = target
			}
		}
		grr.MatchCondition = plan.groupRoleRestriction.MatchCondition.ValueString()
		if !plan.groupRoleRestriction.Hints.IsNull() && !plan.groupRoleRestriction.Hints.IsUnknown() {
			diags.Append(assignSetValues(ctx, plan.groupRoleRestriction.Hints, &grr.ResponseHints)...)
		}
		app.GroupRoleRestriction = grr
	}
	return &app, diags
}
