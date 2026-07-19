package resources

import (
	"context"
	"fmt"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type AppResource struct {
	BaseResource
}

func NewAppResource() resource.Resource {
	return &AppResource{
		BaseResource: NewBaseResource(
			BaseResourceConfig{
				Name:   RESOURCE_APP,
				Schema: &resourceAppSchema,
			},
		),
	}
}

func (r *AppResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		// Not ExactlyOneOf: client_secret is Computed, so omitting both enables cidaas auto-generation.
		resourcevalidator.Conflicting(
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

func (r *AppResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config AppConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	validationRules := []struct {
		clientTypes   []string
		attribute     types.Set
		attributeName string
		errorMessage  string
	}{
		{
			clientTypes:   []string{"SINGLE_PAGE", "REGULAR_WEB", "THIRD_PARTY"},
			attribute:     config.RedirectURIS,
			attributeName: "redirect_uris",
			errorMessage:  `The argument "redirect_uris" is required when client_type is %s, but no definition was found.`,
		},
		{
			clientTypes:   []string{"SINGLE_PAGE", "REGULAR_WEB", "THIRD_PARTY"},
			attribute:     config.AllowedLogoutUrls,
			attributeName: "allowed_logout_urls",
			errorMessage:  `The argument "allowed_logout_urls" is required when client_type is %s, but no definition was found.`,
		},
		{
			clientTypes:   []string{"DEVICE"},
			attribute:     config.GrantTypes,
			attributeName: "grant_types",
			errorMessage:  `The argument "grant_types" is required when client_type is %s, but no definition was found.`,
		},
	}

	for _, rule := range validationRules {
		if util.Contains(rule.clientTypes, config.ClientType.ValueString()) && rule.attribute.IsNull() {
			resp.Diagnostics.AddError(
				"Missing required argument",
				fmt.Sprintf(rule.errorMessage, config.ClientType.ValueString()),
			)
		}
	}
}

func (r *AppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config AppConfig
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(plan.ExtractAppConfigs(ctx)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan data or extract app configurations", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Debug(ctx, "successfully extracted app configurations")

	appModel, diags := prepareAppModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to prepare app model", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Debug(ctx, "successfully prepared app model")
	usingWO := !config.ClientSecretWO.IsNull() && !config.ClientSecretWO.IsUnknown()
	if usingWO {
		appModel.ClientSecret = config.ClientSecretWO.ValueString()
	}
	res, err := r.cidaasClient.Apps.Create(ctx, *appModel)
	if err != nil {
		tflog.Error(ctx, "failed to create app via API", util.H{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("failed to create app", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully created app via API", util.H{
		"client_id": res.Data.ClientID,
	})

	plan.ID = util.StringValueOrNull(&res.Data.ID)
	plan.ClientID = util.StringValueOrNull(&res.Data.ClientID)
	if usingWO {
		plan.ClientSecret = types.StringNull()
	} else {
		plan.ClientSecret = util.StringValueOrNull(&res.Data.ClientSecret)
	}

	// Set the updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	tflog.Info(ctx, "resource app created successfully", util.H{
		"app_id":    res.Data.ID,
		"client_id": res.Data.ClientID,
	})
}

func (r *AppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(state.ExtractAppConfigs(ctx)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data or extract app configurations", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Debug(ctx, "successfully extracted app configurations")
	data, err := r.cidaasClient.Apps.Get(ctx, state.ClientID.ValueString())
	if err != nil {
		if readHandleNotFound(ctx, resp, err) {
			return
		}
		tflog.Error(ctx, "failed to read app via API", util.H{
			"client_id": state.ClientID.ValueString(),
			"error":     err.Error(),
		})
		resp.Diagnostics.AddError("failed to read app", util.FormatErrorMessage(err))
		return
	}

	isImport := !state.ClientID.IsNull() && state.ClientName.IsNull()
	updateAppState(&state, *data, isImport)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Debug(ctx, "resource app read successfully", util.H{
		"app_id":    state.ID.ValueString(),
		"client_id": state.ClientID.ValueString(),
		"is_import": isImport,
	})
}

func (r *AppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config AppConfig

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(plan.ExtractAppConfigs(ctx)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(config.ExtractAppConfigs(ctx)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get plan/state/config data or extract app configurations", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Debug(ctx, "successfully extracted app configurations", util.H{
		"app_id":    state.ID.ValueString(),
		"client_id": state.ClientID.ValueString(),
	})

	appModel, diags := prepareAppModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to prepare app model for update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Debug(ctx, "successfully prepared app model for update")
	usingWO := !config.ClientSecretWO.IsNull() && !config.ClientSecretWO.IsUnknown()
	if usingWO {
		appModel.ClientSecret = config.ClientSecretWO.ValueString()
	}
	appModel.ID = state.ID.ValueString()
	_, err := r.cidaasClient.Apps.Update(ctx, *appModel)
	if err != nil {
		tflog.Error(ctx, "failed to update app via API", util.H{
			"app_id": state.ID.ValueString(),
			"error":  err.Error(),
		})
		resp.Diagnostics.AddError("failed to update app", util.FormatErrorMessage(err))
		return
	}
	tflog.Info(ctx, "successfully updated app via API", util.H{
		"client_id": state.ClientID.ValueString(),
	})

	if usingWO {
		// Also clears any plaintext value left from a prior non-write-only apply.
		plan.ClientSecret = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state after update", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Debug(ctx, "resource app updated successfully")
}

func (r *AppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppConfig
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get state data for deletion", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	err := r.cidaasClient.Apps.Delete(ctx, state.ClientID.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to delete app via API", util.H{
			"client_id": state.ClientID.ValueString(),
			"error":     err.Error(),
		})
		resp.Diagnostics.AddError("failed to delete app", util.FormatErrorMessage(err))
		return
	}

	tflog.Info(ctx, "resource app deleted successfully", util.H{
		"client_id": state.ClientID.ValueString(),
	})
}

func (r *AppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("client_id"), req, resp)
}

// updateAppState updates the Terraform state with data from the API response.
// During import operations (isImport=true), all fields are updated.
// During normal reads (isImport=false), only configured fields from tf file are updated.
func updateAppState(state *AppConfig, resp cidaas.AppResponse, isImport bool) {
	data := resp.Data
	state.ID = util.StringValueOrNull(&data.ID)
	state.ClientID = util.StringValueOrNull(&data.ClientID)
	// A null state value signals client_secret_wo is in use — leave state null.
	if !state.ClientSecret.IsNull() || isImport {
		state.ClientSecret = util.StringValueOrNull(&data.ClientSecret)
	}
	state.ClientName = util.StringValueOrNull(&data.ClientName)
	state.ClientType = util.StringValueOrNull(&data.ClientType)
	state.CompanyName = util.StringValueOrNull(&data.CompanyName)
	state.CompanyAddress = util.StringValueOrNull(&data.CompanyAddress)
	state.CompanyWebsite = util.StringValueOrNull(&data.CompanyWebsite)
	state.AllowedScopes = util.SetValueOrNull(data.AllowedScopes)

	// RedirectURIS, AllowedLogoutUrls, GrantTypes are required to be set when import is done based on the client_type
	if !state.RedirectURIS.IsNull() || isImport {
		state.RedirectURIS = util.SetValueOrNull(data.RedirectURIS)
	}
	if !state.AllowedLogoutUrls.IsNull() || isImport {
		state.AllowedLogoutUrls = util.SetValueOrNull(data.AllowedLogoutUrls)
	}
	if !state.GrantTypes.IsNull() || isImport {
		state.GrantTypes = util.SetValueOrNull(data.GrantTypes)
	}
	// String attributes
	if !state.AccentColor.IsNull() || isImport {
		state.AccentColor = util.StringValueOrNull(&data.AccentColor)
	}
	if !state.PrimaryColor.IsNull() || isImport {
		state.PrimaryColor = util.StringValueOrNull(&data.PrimaryColor)
	}
	if !state.MediaType.IsNull() || isImport {
		state.MediaType = util.StringValueOrNull(&data.MediaType)
	}
	if !state.HostedPageGroup.IsNull() || isImport {
		state.HostedPageGroup = util.StringValueOrNull(&data.HostedPageGroup)
	}
	if !state.TemplateGroupID.IsNull() || isImport {
		state.TemplateGroupID = util.StringValueOrNull(&data.TemplateGroupID)
	}
	if !state.LogoAlign.IsNull() || isImport {
		state.LogoAlign = util.StringValueOrNull(&data.LogoAlign)
	}
	if !state.Webfinger.IsNull() || isImport {
		state.Webfinger = util.StringValueOrNull(&data.Webfinger)
	}
	if !state.ContentAlign.IsNull() || isImport {
		state.ContentAlign = util.StringValueOrNull(&data.ContentAlign)
	}
	if !state.ClientDisplayName.IsNull() || isImport {
		state.ClientDisplayName = util.StringValueOrNull(&data.ClientDisplayName)
	}
	if !state.PolicyURI.IsNull() || isImport {
		state.PolicyURI = util.StringValueOrNull(&data.PolicyURI)
	}
	if !state.TosURI.IsNull() || isImport {
		state.TosURI = util.StringValueOrNull(&data.TosURI)
	}
	if !state.ImprintURI.IsNull() || isImport {
		state.ImprintURI = util.StringValueOrNull(&data.ImprintURI)
	}
	if !state.TokenEndpointAuthMethod.IsNull() || isImport {
		state.TokenEndpointAuthMethod = util.StringValueOrNull(&data.TokenEndpointAuthMethod)
	}
	if !state.TokenEndpointAuthSigningAlg.IsNull() || isImport {
		state.TokenEndpointAuthSigningAlg = util.StringValueOrNull(&data.TokenEndpointAuthSigningAlg)
	}
	if !state.CaptchaRef.IsNull() || isImport {
		state.CaptchaRef = util.StringValueOrNull(&data.CaptchaRef)
	}
	if !state.CommunicationMediumVerification.IsNull() || isImport {
		state.CommunicationMediumVerification = util.StringValueOrNull(&data.CommunicationMediumVerification)
	}
	if !state.BackchannelLogoutURI.IsNull() || isImport {
		state.BackchannelLogoutURI = util.StringValueOrNull(&data.BackchannelLogoutURI)
	}
	if !state.LogoURI.IsNull() || isImport {
		state.LogoURI = util.StringValueOrNull(&data.LogoURI)
	}
	if !state.InitiateLoginURI.IsNull() || isImport {
		state.InitiateLoginURI = util.StringValueOrNull(&data.InitiateLoginURI)
	}
	if !state.RegistrationClientURI.IsNull() || isImport {
		state.RegistrationClientURI = util.StringValueOrNull(&data.RegistrationClientURI)
	}
	if !state.RegistrationAccessToken.IsNull() || isImport {
		state.RegistrationAccessToken = util.StringValueOrNull(&data.RegistrationAccessToken)
	}
	if !state.ClientURI.IsNull() || isImport {
		state.ClientURI = util.StringValueOrNull(&data.ClientURI)
	}
	if !state.JwksURI.IsNull() || isImport {
		state.JwksURI = util.StringValueOrNull(&data.JwksURI)
	}
	if !state.Jwks.IsNull() || isImport {
		state.Jwks = util.StringValueOrNull(&data.Jwks)
	}
	if !state.SectorIdentifierURI.IsNull() || isImport {
		state.SectorIdentifierURI = util.StringValueOrNull(&data.SectorIdentifierURI)
	}
	if !state.SubjectType.IsNull() || isImport {
		state.SubjectType = util.StringValueOrNull(&data.SubjectType)
	}
	if !state.IDTokenSignedResponseAlg.IsNull() || isImport {
		state.IDTokenSignedResponseAlg = util.StringValueOrNull(&data.IDTokenSignedResponseAlg)
	}
	if !state.IDTokenEncryptedResponseAlg.IsNull() || isImport {
		state.IDTokenEncryptedResponseAlg = util.StringValueOrNull(&data.IDTokenEncryptedResponseAlg)
	}
	if !state.IDTokenEncryptedResponseEnc.IsNull() || isImport {
		state.IDTokenEncryptedResponseEnc = util.StringValueOrNull(&data.IDTokenEncryptedResponseEnc)
	}
	if !state.UserinfoSignedResponseAlg.IsNull() || isImport {
		state.UserinfoSignedResponseAlg = util.StringValueOrNull(&data.UserinfoSignedResponseAlg)
	}
	if !state.UserinfoEncryptedResponseAlg.IsNull() || isImport {
		state.UserinfoEncryptedResponseAlg = util.StringValueOrNull(&data.UserinfoEncryptedResponseAlg)
	}
	if !state.UserinfoEncryptedResponseEnc.IsNull() || isImport {
		state.UserinfoEncryptedResponseEnc = util.StringValueOrNull(&data.UserinfoEncryptedResponseEnc)
	}
	if !state.RequestObjectSigningAlg.IsNull() || isImport {
		state.RequestObjectSigningAlg = util.StringValueOrNull(&data.RequestObjectSigningAlg)
	}
	if !state.RequestObjectEncryptionAlg.IsNull() || isImport {
		state.RequestObjectEncryptionAlg = util.StringValueOrNull(&data.RequestObjectEncryptionAlg)
	}
	if !state.RequestObjectEncryptionEnc.IsNull() || isImport {
		state.RequestObjectEncryptionEnc = util.StringValueOrNull(&data.RequestObjectEncryptionEnc)
	}
	if !state.Description.IsNull() || isImport {
		state.Description = util.StringValueOrNull(&data.Description)
	}
	if !state.ConsentPageGroup.IsNull() || isImport {
		state.ConsentPageGroup = util.StringValueOrNull(&data.ConsentPageGroup)
	}
	if !state.PasswordPolicyRef.IsNull() || isImport {
		state.PasswordPolicyRef = util.StringValueOrNull(&data.PasswordPolicyRef)
	}
	if !state.BlockingMechanismRef.IsNull() || isImport {
		state.BlockingMechanismRef = util.StringValueOrNull(&data.BlockingMechanismRef)
	}
	if !state.Sub.IsNull() || isImport {
		state.Sub = util.StringValueOrNull(&data.Sub)
	}
	if !state.Role.IsNull() || isImport {
		state.Role = util.StringValueOrNull(&data.Role)
	}
	if !state.MfaConfiguration.IsNull() || isImport {
		state.MfaConfiguration = util.StringValueOrNull(&data.MfaConfiguration)
	}
	if !state.BackgroundURI.IsNull() || isImport {
		state.BackgroundURI = util.StringValueOrNull(&data.BackgroundURI)
	}
	if !state.VideoURL.IsNull() || isImport {
		state.VideoURL = util.StringValueOrNull(&data.VideoURL)
	}
	if !state.BotCaptchaRef.IsNull() || isImport {
		state.BotCaptchaRef = util.StringValueOrNull(&data.BotCaptchaRef)
	}
	if !state.BotProvider.IsNull() || isImport {
		state.BotProvider = util.StringValueOrNull(&data.BotProvider)
	}
	if !state.OauthStandard.IsNull() || isImport {
		state.OauthStandard = util.StringValueOrNull(&data.OauthStandard)
	}

	// Boolean attributes
	if !state.AllowGuestLogin.IsNull() || isImport {
		state.AllowGuestLogin = util.BoolValueOrNull(data.AllowGuestLogin)
	}
	if !state.EnableDeduplication.IsNull() || isImport {
		state.EnableDeduplication = util.BoolValueOrNull(data.EnableDeduplication)
	}
	if !state.AutoLoginAfterRegister.IsNull() || isImport {
		state.AutoLoginAfterRegister = util.BoolValueOrNull(data.AutoLoginAfterRegister)
	}
	if !state.RegisterWithLoginInformation.IsNull() || isImport {
		state.RegisterWithLoginInformation = util.BoolValueOrNull(data.RegisterWithLoginInformation)
	}
	if !state.IsHybridApp.IsNull() || isImport {
		state.IsHybridApp = util.BoolValueOrNull(data.IsHybridApp)
	}
	if !state.Enabled.IsNull() || isImport {
		state.Enabled = util.BoolValueOrNull(data.Enabled)
	}
	if !state.IsRememberMeSelected.IsNull() || isImport {
		state.IsRememberMeSelected = util.BoolValueOrNull(data.IsRememberMeSelected)
	}
	if !state.AllowDisposableEmail.IsNull() || isImport {
		state.AllowDisposableEmail = util.BoolValueOrNull(data.AllowDisposableEmail)
	}
	if !state.ValidatePhoneNumber.IsNull() || isImport {
		state.ValidatePhoneNumber = util.BoolValueOrNull(data.ValidatePhoneNumber)
	}
	if !state.SmartMfa.IsNull() || isImport {
		state.SmartMfa = util.BoolValueOrNull(data.SmartMfa)
	}
	if !state.EnableBotDetection.IsNull() || isImport {
		state.EnableBotDetection = util.BoolValueOrNull(data.EnableBotDetection)
	}
	if !state.IsLoginSuccessPageEnabled.IsNull() || isImport {
		state.IsLoginSuccessPageEnabled = util.BoolValueOrNull(data.IsLoginSuccessPageEnabled)
	}
	if !state.IsRegisterSuccessPageEnabled.IsNull() || isImport {
		state.IsRegisterSuccessPageEnabled = util.BoolValueOrNull(data.IsRegisterSuccessPageEnabled)
	}
	if !state.IsGroupLoginSelectionEnabled.IsNull() || isImport {
		state.IsGroupLoginSelectionEnabled = util.BoolValueOrNull(data.IsGroupLoginSelectionEnabled)
	}

	if !state.JweEnabled.IsNull() || isImport {
		state.JweEnabled = util.BoolValueOrNull(data.JweEnabled)
	}
	if !state.UserConsent.IsNull() || isImport {
		state.UserConsent = util.BoolValueOrNull(data.UserConsent)
	}
	if !state.EnablePasswordlessAuth.IsNull() || isImport {
		state.EnablePasswordlessAuth = util.BoolValueOrNull(data.EnablePasswordlessAuth)
	}
	if !state.RequireAuthTime.IsNull() || isImport {
		state.RequireAuthTime = util.BoolValueOrNull(data.RequireAuthTime)
	}
	if !state.EnableLoginSpi.IsNull() || isImport {
		state.EnableLoginSpi = util.BoolValueOrNull(data.EnableLoginSpi)
	}
	if !state.BackchannelLogoutSessionRequired.IsNull() || isImport {
		state.BackchannelLogoutSessionRequired = util.BoolValueOrNull(data.BackchannelLogoutSessionRequired)
	}
	if !state.AcceptRolesInTheRegistration.IsNull() || isImport {
		state.AcceptRolesInTheRegistration = util.BoolValueOrNull(data.AcceptRolesInTheRegistration)
	}
	if !state.RequirePKCE.IsNull() || isImport {
		state.RequirePKCE = util.BoolValueOrNull(data.RequirePKCE)
	}
	if !state.DisableInsecurePKCEMethod.IsNull() || isImport {
		state.DisableInsecurePKCEMethod = util.BoolValueOrNull(data.DisableInsecurePKCEMethod)
	}

	// Integer attributes
	if !state.DefaultMaxAge.IsNull() || isImport {
		state.DefaultMaxAge = util.Int64ValueOrNull(data.DefaultMaxAge)
	}
	if !state.TokenLifetimeInSeconds.IsNull() || isImport {
		state.TokenLifetimeInSeconds = util.Int64ValueOrNull(data.TokenLifetimeInSeconds)
	}
	if !state.IDTokenLifetimeInSeconds.IsNull() || isImport {
		state.IDTokenLifetimeInSeconds = util.Int64ValueOrNull(data.IDTokenLifetimeInSeconds)
	}
	if !state.RefreshTokenLifetimeInSeconds.IsNull() || isImport {
		state.RefreshTokenLifetimeInSeconds = util.Int64ValueOrNull(data.RefreshTokenLifetimeInSeconds)
	}

	// Set attributes
	if !state.AllowedWebOrigins.IsNull() || isImport {
		state.AllowedWebOrigins = util.SetValueOrNull(data.AllowedWebOrigins)
	}
	if !state.DefaultScopes.IsNull() || isImport {
		state.DefaultScopes = util.SetValueOrNull(data.DefaultScopes)
	}
	if !state.AllowedRoles.IsNull() || isImport {
		state.AllowedRoles = util.SetValueOrNull(data.AllowedRoles)
	}
	if !state.ResponseTypes.IsNull() || isImport {
		state.ResponseTypes = util.SetValueOrNull(data.ResponseTypes)
	}
	if !state.AllowLoginWith.IsNull() || isImport {
		state.AllowLoginWith = util.SetValueOrNull(data.AllowLoginWith)
	}
	if !state.PostLogoutRedirectUris.IsNull() || isImport {
		state.PostLogoutRedirectUris = util.SetValueOrNull(data.PostLogoutRedirectUris)
	}
	if !state.GroupIDs.IsNull() || isImport {
		state.GroupIDs = util.SetValueOrNull(data.GroupIDs)
	}
	if !state.AllowedFields.IsNull() || isImport {
		state.AllowedFields = util.SetValueOrNull(data.AllowedFields)
	}
	if !state.GroupTypes.IsNull() || isImport {
		state.GroupTypes = util.SetValueOrNull(data.GroupTypes)
	}
	if !state.AdditionalAccessTokenPayload.IsNull() || isImport {
		state.AdditionalAccessTokenPayload = util.SetValueOrNull(data.AdditionalAccessTokenPayload)
	}
	if !state.RequiredFields.IsNull() || isImport {
		state.RequiredFields = util.SetValueOrNull(data.RequiredFields)
	}
	if !state.Contacts.IsNull() || isImport {
		state.Contacts = util.SetValueOrNull(data.Contacts)
	}
	if !state.WebMessageUris.IsNull() || isImport {
		state.WebMessageUris = util.SetValueOrNull(data.WebMessageUris)
	}
	if !state.AllowedOrigins.IsNull() || isImport {
		state.AllowedOrigins = util.SetValueOrNull(data.AllowedOrigins)
	}
	if !state.AllowedNativeClients.IsNull() || isImport {
		state.AllowedNativeClients = util.SetValueOrNull(data.AllowedNativeClients)
	}
	if !state.LoginProviders.IsNull() || isImport {
		state.LoginProviders = util.SetValueOrNull(data.LoginProviders)
	}
	if !state.PendingScopes.IsNull() || isImport {
		state.PendingScopes = util.SetValueOrNull(data.PendingScopes)
	}
	if !state.AllowedMfa.IsNull() || isImport {
		state.AllowedMfa = util.SetValueOrNull(data.AllowedMfa)
	}
	if !state.DefaultRoles.IsNull() || isImport {
		state.DefaultRoles = util.SetValueOrNull(data.DefaultRoles)
	}
	if !state.DefaultAcrValues.IsNull() || isImport {
		state.DefaultAcrValues = util.SetValueOrNull(data.DefaultAcrValues)
	}
	if !state.SuggestMfa.IsNull() || isImport {
		state.SuggestMfa = util.SetValueOrNull(data.SuggestMfa)
	}
	if !state.CaptchaRefs.IsNull() || isImport {
		state.CaptchaRefs = util.SetValueOrNull(data.CaptchaRefs)
	}
	if !state.ConsentRefs.IsNull() || isImport {
		state.ConsentRefs = util.SetValueOrNull(data.ConsentRefs)
	}
	if !state.RequestUris.IsNull() || isImport {
		state.RequestUris = util.SetValueOrNull(data.RequestUris)
	}

	// Map & List attributes
	if (!state.ApplicationMetaData.IsNull() || isImport) && len(data.ApplicationMetaData) > 0 {
		metaData := map[string]attr.Value{}
		for key, value := range data.ApplicationMetaData {
			val := value
			metaData[key] = util.StringValueOrNull(&val)
		}
		state.ApplicationMetaData = types.MapValueMust(types.StringType, metaData)
	}

	if ((!state.SocialProviders.IsNull() && len(state.SocialProviders.Elements()) > 0) || isImport) && len(data.SocialProviders) > 0 {
		var spObjectValues []attr.Value
		spObjectType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"provider_name": types.StringType,
				"social_id":     types.StringType,
			},
		}

		for _, sp := range data.SocialProviders {
			providerName := sp.ProviderName
			socialID := sp.SocialID
			objValue := types.ObjectValueMust(
				spObjectType.AttrTypes,
				map[string]attr.Value{
					"provider_name": util.StringValueOrNull(&providerName),
					"social_id":     util.StringValueOrNull(&socialID),
				})
			spObjectValues = append(spObjectValues, objValue)
		}
		state.SocialProviders = types.ListValueMust(spObjectType, spObjectValues)
	}

	providerObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"provider_name":       types.StringType,
			"display_name":        types.StringType,
			"logo_url":            types.StringType,
			"type":                types.StringType,
			"is_provider_visible": types.BoolType,
			"domains":             types.SetType{ElemType: types.StringType},
		},
	}

	createProviderMetaObjectValues := func(providers []cidaas.IProviderMetadData) []attr.Value {
		var providerMetaObjectValues []attr.Value
		if len(providers) > 0 {
			for _, provider := range providers {
				providerName := provider.ProviderName
				displayName := provider.DisplayName
				logoURL := provider.LogoURL
				providerType := provider.Type
				objValue := types.ObjectValueMust(
					providerObjectType.AttrTypes,
					map[string]attr.Value{
						"provider_name":       util.StringValueOrNull(&providerName),
						"display_name":        util.StringValueOrNull(&displayName),
						"logo_url":            util.StringValueOrNull(&logoURL),
						"type":                util.StringValueOrNull(&providerType),
						"is_provider_visible": util.BoolValueOrNull(provider.IsProviderVisible),
						"domains":             util.SetValueOrNull(provider.Domains),
					})
				providerMetaObjectValues = append(providerMetaObjectValues, objValue)
			}
		}
		return providerMetaObjectValues
	}

	if ((!state.CustomProviders.IsNull() && len(state.CustomProviders.Elements()) > 0) || isImport) && len(data.CustomProviders) > 0 {
		state.CustomProviders = types.ListValueMust(providerObjectType, createProviderMetaObjectValues(data.CustomProviders))
	}
	if ((!state.SamlProviders.IsNull() && len(state.SamlProviders.Elements()) > 0) || isImport) && len(data.SamlProviders) > 0 {
		state.SamlProviders = types.ListValueMust(providerObjectType, createProviderMetaObjectValues(data.SamlProviders))
	}
	if ((!state.AdProviders.IsNull() && len(state.AdProviders.Elements()) > 0) || isImport) && len(data.AdProviders) > 0 {
		state.AdProviders = types.ListValueMust(providerObjectType, createProviderMetaObjectValues(data.AdProviders))
	}

	allowedGroupsObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"group_id":      types.StringType,
			"group_type":    types.StringType,
			"roles":         types.SetType{ElemType: types.StringType},
			"default_roles": types.SetType{ElemType: types.StringType},
		},
	}

	createAllowedGroupsObjectValues := func(groups []cidaas.IAllowedGroups) []attr.Value {
		var allowedGroupObjectValues []attr.Value
		if len(groups) > 0 {
			for _, group := range groups {
				groupID := group.GroupID
				groupType := group.GroupType
				objValue := types.ObjectValueMust(
					allowedGroupsObjectType.AttrTypes,
					map[string]attr.Value{
						"group_id":      util.StringValueOrNull(&groupID),
						"group_type":    util.StringValueOrNull(&groupType),
						"roles":         util.SetValueOrNull(group.Roles),
						"default_roles": util.SetValueOrNull(group.DefaultRoles),
					})
				allowedGroupObjectValues = append(allowedGroupObjectValues, objValue)
			}
		}
		return allowedGroupObjectValues
	}

	if ((!state.AllowedGroups.IsNull() && len(state.AllowedGroups.Elements()) > 0) || isImport) && len(data.AllowedGroups) > 0 {
		state.AllowedGroups = types.ListValueMust(allowedGroupsObjectType, createAllowedGroupsObjectValues(data.AllowedGroups))
	}
	if ((!state.OperationsAllowedGroups.IsNull() && len(state.OperationsAllowedGroups.Elements()) > 0) || isImport) && len(data.OperationsAllowedGroups) > 0 {
		state.OperationsAllowedGroups = types.ListValueMust(allowedGroupsObjectType, createAllowedGroupsObjectValues(data.OperationsAllowedGroups))
	}

	if ((!state.AllowGuestLoginGroups.IsNull() && len(state.AllowGuestLoginGroups.Elements()) > 0) || isImport) && len(data.AllowGuestLoginGroups) > 0 {
		var allowedGroupObjectValues []attr.Value
		for _, group := range data.AllowGuestLoginGroups {
			groupID := group.GroupID
			groupType := group.GroupType
			objValue := types.ObjectValueMust(
				allowedGroupsObjectType.AttrTypes,
				map[string]attr.Value{
					"group_id":      util.StringValueOrNull(&groupID),
					"group_type":    util.StringValueOrNull(&groupType),
					"roles":         util.SetValueOrNull(group.Roles),
					"default_roles": util.SetValueOrNull(group.DefaultRoles),
				})
			allowedGroupObjectValues = append(allowedGroupObjectValues, objValue)
		}
		state.AllowGuestLoginGroups = types.ListValueMust(allowedGroupsObjectType, allowedGroupObjectValues)
	}

	if (!state.LoginSpi.IsNull() || isImport) && data.LoginSpi != nil {
		oauthClientID := util.StringValueOrNull(&data.LoginSpi.OauthClientID)
		spiURL := util.StringValueOrNull(&data.LoginSpi.SpiURL)

		loginSpi := types.ObjectValueMust(
			map[string]attr.Type{
				"oauth_client_id": types.StringType,
				"spi_url":         types.StringType,
			},
			map[string]attr.Value{
				"oauth_client_id": oauthClientID,
				"spi_url":         spiURL,
			})
		state.LoginSpi = loginSpi
	}

	if (!state.GroupSelection.IsNull() || isImport) && data.GroupSelection != nil {
		alwaysShowGroupSelection := util.BoolValueOrNull(data.GroupSelection.AlwaysShowGroupSelection)
		selectableGroups := util.SetValueOrNull(data.GroupSelection.SelectableGroups)
		selectableGroupTypes := util.SetValueOrNull(data.GroupSelection.SelectableGroupTypes)

		groupSelection := types.ObjectValueMust(
			map[string]attr.Type{
				"always_show_group_selection": types.BoolType,
				"selectable_groups":           types.SetType{ElemType: types.StringType},
				"selectable_group_types":      types.SetType{ElemType: types.StringType},
			},
			map[string]attr.Value{
				"always_show_group_selection": alwaysShowGroupSelection,
				"selectable_groups":           selectableGroups,
				"selectable_group_types":      selectableGroupTypes,
			})
		state.GroupSelection = groupSelection
	}

	if (!state.Mfa.IsNull() || isImport) && data.Mfa != nil {
		setting := util.StringValueOrNull(&data.Mfa.Setting)
		// Don't persist 0 or empty: treat as null to avoid drift when API returns no value
		timeInterval := types.Int64Null()
		if data.Mfa.TimeIntervalInSeconds != nil && *data.Mfa.TimeIntervalInSeconds != 0 {
			timeInterval = types.Int64Value(*data.Mfa.TimeIntervalInSeconds)
		}
		allowedMethods := util.SetValueOrNull(data.Mfa.AllowedMethods)

		mfa := types.ObjectValueMust(
			map[string]attr.Type{
				"setting":                  types.StringType,
				"time_interval_in_seconds": types.Int64Type,
				"allowed_methods":          types.SetType{ElemType: types.StringType},
			},
			map[string]attr.Value{
				"setting":                  setting,
				"time_interval_in_seconds": timeInterval,
				"allowed_methods":          allowedMethods,
			},
		)
		state.Mfa = mfa
	}

	if (!state.MobileSettings.IsNull() || isImport) && data.MobileSettings != nil {
		mobileSettings := types.ObjectValueMust(
			map[string]attr.Type{
				"team_id":      types.StringType,
				"bundle_id":    types.StringType,
				"package_name": types.StringType,
				"key_hash":     types.StringType,
			},
			map[string]attr.Value{
				"team_id":      util.StringValueOrNull(&data.MobileSettings.TeamID),
				"bundle_id":    util.StringValueOrNull(&data.MobileSettings.BundleID),
				"package_name": util.StringValueOrNull(&data.MobileSettings.PackageName),
				"key_hash":     util.StringValueOrNull(&data.MobileSettings.KeyHash),
			})
		state.MobileSettings = mobileSettings
	}

	if (!state.SuggestVerificationMethods.IsNull() || isImport) && data.SuggestVerificationMethods != nil {
		skipDurationInDays := types.Int32Value(data.SuggestVerificationMethods.SkipDurationInDays)
		skipUntil := util.StringValueOrNull(&data.SuggestVerificationMethods.MandatoryConfig.SkipUntil)
		mandatatoryRange := util.StringValueOrNull(&data.SuggestVerificationMethods.MandatoryConfig.Range)
		mandatoryMethods := util.SetValueOrNull(data.SuggestVerificationMethods.MandatoryConfig.Methods)
		optionalMethods := util.SetValueOrNull(data.SuggestVerificationMethods.OptionalConfig.Methods)

		mandateConfigType := map[string]attr.Type{
			"methods":    types.SetType{ElemType: types.StringType},
			"range":      types.StringType,
			"skip_until": types.StringType,
		}
		mandatoryConfig := types.ObjectValueMust(
			mandateConfigType,
			map[string]attr.Value{
				"methods":    mandatoryMethods,
				"range":      mandatatoryRange,
				"skip_until": skipUntil,
			},
		)

		optionalConfigType := map[string]attr.Type{
			"methods": types.SetType{ElemType: types.StringType},
		}
		optionalConfig := types.ObjectValueMust(
			optionalConfigType,
			map[string]attr.Value{
				"methods": optionalMethods,
			},
		)

		obj := types.ObjectValueMust(
			map[string]attr.Type{
				"mandatory_config":      types.ObjectType{AttrTypes: mandateConfigType},
				"optional_config":       types.ObjectType{AttrTypes: optionalConfigType},
				"skip_duration_in_days": types.Int32Type,
			},
			map[string]attr.Value{
				"mandatory_config":      mandatoryConfig,
				"optional_config":       optionalConfig,
				"skip_duration_in_days": skipDurationInDays,
			},
		)
		state.SuggestVerificationMethods = obj
	}

	if (!state.GroupRoleRestriction.IsNull() || isImport) && data.GroupRoleRestriction != nil &&
		(len(data.GroupRoleRestriction.Filters) > 0 || len(data.GroupRoleRestriction.ResponseHints) > 0) {
		roleFilterType := map[string]attr.Type{
			"match_condition": types.StringType,
			"roles":           types.SetType{ElemType: types.StringType},
		}
		filterType := map[string]attr.Type{
			"group_id":    types.StringType,
			"group_type":  types.StringType,
			"role_filter": types.ObjectType{AttrTypes: roleFilterType},
		}

		filterObjectType := types.ObjectType{
			AttrTypes: filterType,
		}
		var filters basetypes.ListValue
		var filterObjectValues []attr.Value

		parentMatchCondition := util.StringValueOrNull(&data.GroupRoleRestriction.MatchCondition)
		for _, grr := range data.GroupRoleRestriction.Filters {
			groupID := grr.GroupID
			groupType := grr.GroupType
			var roleFilterValue attr.Value
			if grr.RoleFilter.IsEmpty() {
				// API often returns roleFilter: {} when unset; treat as null to match omitted config.
				roleFilterValue = types.ObjectNull(roleFilterType)
			} else {
				matchCondition := grr.RoleFilter.MatchCondition
				roles := grr.RoleFilter.Roles
				roleFilterValue = types.ObjectValueMust(
					roleFilterType,
					map[string]attr.Value{
						"match_condition": util.StringValueOrNull(&matchCondition),
						"roles":           util.SetValueOrNull(roles),
					},
				)
			}
			objValue := types.ObjectValueMust(
				filterType,
				map[string]attr.Value{
					"group_id":    util.StringValueOrNull(&groupID),
					"group_type":  util.StringValueOrNull(&groupType),
					"role_filter": roleFilterValue,
				})
			filterObjectValues = append(filterObjectValues, objValue)
		}
		filters = types.ListValueMust(filterObjectType, filterObjectValues)

		hints := util.SetValueOrNull(data.GroupRoleRestriction.ResponseHints)

		obj := types.ObjectValueMust(
			map[string]attr.Type{
				"match_condition": types.StringType,
				"filters":         types.ListType{ElemType: types.ObjectType{AttrTypes: filterType}},
				"hints":           types.SetType{ElemType: types.StringType},
			},
			map[string]attr.Value{
				"match_condition": parentMatchCondition,
				"filters":         filters,
				"hints":           hints,
			},
		)
		state.GroupRoleRestriction = obj
	}
}
