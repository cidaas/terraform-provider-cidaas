//nolint:revive
package identity_provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// federationProviderModel is shared by the resource and (after stripping write-only attributes) the data source.
type federationProviderModel struct {
	ID                    types.String                               `tfsdk:"id"`
	ProviderName          types.String                               `tfsdk:"provider_name"`
	Provider              types.String                               `tfsdk:"provider_type"`
	StandardType          types.String                               `tfsdk:"standard_type"`
	DisplayName           types.String                               `tfsdk:"display_name"`
	LogoURL               types.String                               `tfsdk:"logo_url"`
	Owner                 types.String                               `tfsdk:"owner"`
	Description           types.String                               `tfsdk:"description"`
	Domains               types.List                                 `tfsdk:"domains"`
	EnabledForAdminPortal types.Bool                                 `tfsdk:"enabled_for_admin_portal"`
	AttributeMapping      map[string]federationAttributeMappingModel `tfsdk:"attribute_mapping"`
	OAuth2                *federationOAuth2Model                     `tfsdk:"oauth2"`
	SAML                  *federationSAMLModel                       `tfsdk:"saml"`
	LDAP                  *federationLDAPModel                       `tfsdk:"ldap"`
	CreatedTime           types.String                               `tfsdk:"created_time"`
	UpdatedTime           types.String                               `tfsdk:"updated_time"`
}

// federationAttributeMappingModel is one entry of the attribute_mapping map.
type federationAttributeMappingModel struct {
	FieldKey         types.String              `tfsdk:"field_key"`
	ExternalKey      types.String              `tfsdk:"external_key"`
	DataType         types.String              `tfsdk:"data_type"`
	Default          types.String              `tfsdk:"default"`
	Ignore           types.Bool                `tfsdk:"ignore"`
	Transform        *federationTransformModel `tfsdk:"transform"`
	EmailVerifiedKey types.String              `tfsdk:"email_verified_key"`
}

// federationTransformModel is attribute_mapping.transform.
type federationTransformModel struct {
	Type   types.String `tfsdk:"type"`
	Config types.Map    `tfsdk:"config"`
}

// federationOAuth2Model is the Terraform side of the oauth2 block, including the write-only secret attributes.
type federationOAuth2Model struct {
	ClientID                types.String                 `tfsdk:"client_id"`
	ClientSecret            types.String                 `tfsdk:"client_secret"`
	ClientSecretWO          types.String                 `tfsdk:"client_secret_wo"`
	ClientSecretWOVersion   types.String                 `tfsdk:"client_secret_wo_version"`
	Pkce                    types.Bool                   `tfsdk:"pkce"`
	UserinfoSource          types.String                 `tfsdk:"userinfo_source"`
	AuthorizationEndpoint   types.String                 `tfsdk:"authorization_endpoint"`
	TokenEndpoint           types.String                 `tfsdk:"token_endpoint"`
	UserinfoEndpoint        types.String                 `tfsdk:"userinfo_endpoint"`
	WellknownEndpoint       types.String                 `tfsdk:"wellknown_endpoint"`
	AuthorizationParameters types.Map                    `tfsdk:"authorization_parameters"`
	TokenEndpointAuth       types.String                 `tfsdk:"token_endpoint_auth"`
	Scopes                  types.List                   `tfsdk:"scopes"`
	ScopesConfig            *federationScopesConfigModel `tfsdk:"scopes_config"`
	Claims                  *federationClaimsModel       `tfsdk:"claims"`
	AmrConfig               []federationAmrConfigModel   `tfsdk:"amr_config"`
	Groups                  types.String                 `tfsdk:"groups"`
}

// federationScopesConfigModel is oauth2.scopes_config.
type federationScopesConfigModel struct {
	DisplayLabel types.String                `tfsdk:"display_label"`
	Scopes       []federationScopeSetupModel `tfsdk:"scopes"`
}

// federationScopeSetupModel is one entry of oauth2.scopes_config.scopes.
type federationScopeSetupModel struct {
	ScopeName   types.String `tfsdk:"scope_name"`
	Required    types.Bool   `tfsdk:"required"`
	Recommended types.Bool   `tfsdk:"recommended"`
}

// federationClaimsModel is oauth2.claims.
type federationClaimsModel struct {
	RequiredClaims *federationClaimNamesModel `tfsdk:"required_claims"`
	OptionalClaims *federationClaimNamesModel `tfsdk:"optional_claims"`
}

// federationClaimNamesModel is oauth2.claims.required_claims / optional_claims.
type federationClaimNamesModel struct {
	UserInfo types.List `tfsdk:"user_info"`
	IDToken  types.List `tfsdk:"id_token"`
}

// federationAmrConfigModel is one entry of oauth2.amr_config.
type federationAmrConfigModel struct {
	AmrValue    types.String `tfsdk:"amr_value"`
	ExtAmrValue types.String `tfsdk:"ext_amr_value"`
}

// federationSAMLModel is the Terraform side of the saml block.
type federationSAMLModel struct {
	MetaData                 types.String                     `tfsdk:"meta_data"`
	SignAuthRequest          types.Bool                       `tfsdk:"sign_auth_request"`
	UserIDAttribute          types.String                     `tfsdk:"user_id_attribute"`
	ProtocolBindingMethod    types.String                     `tfsdk:"protocol_binding_method"`
	IDPInitiatedConfig       *federationIDPInitiatedModel     `tfsdk:"idp_initiated_config"`
	SignatureConfig          *federationSignatureConfigModel  `tfsdk:"signature_config"`
	EncryptionConfig         *federationEncryptionConfigModel `tfsdk:"encryption_config"`
	SPSigningCertificates    types.List                       `tfsdk:"sp_signing_certificates"`
	SPEncryptionCertificates types.List                       `tfsdk:"sp_encryption_certificates"`
}

// federationIDPInitiatedModel is saml.idp_initiated_config.
type federationIDPInitiatedModel struct {
	ID          types.String `tfsdk:"id"`
	ClientID    types.String `tfsdk:"client_id"`
	Protocol    types.String `tfsdk:"protocol"`
	QueryString types.String `tfsdk:"query_string"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

// federationSignatureConfigModel is saml.signature_config.
type federationSignatureConfigModel struct {
	SigningAlgorithm types.String `tfsdk:"signing_algorithm"`
	SignTargets      types.List   `tfsdk:"sign_targets"`
}

// federationEncryptionConfigModel is saml.encryption_config.
type federationEncryptionConfigModel struct {
	EncryptTargets             types.List   `tfsdk:"encrypt_targets"`
	RequireEncryptedFor        types.List   `tfsdk:"require_encrypted_for"`
	ContentEncryptionAlgorithm types.String `tfsdk:"content_encryption_algorithm"`
	KeyTransportAlgorithm      types.String `tfsdk:"key_transport_algorithm"`
}

var federationCertificateType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"certificate":        types.StringType,
	"sha256_fingerprint": types.StringType,
}}

// federationLDAPModel is the Terraform side of the ldap block, including the write-only bind credentials.
type federationLDAPModel struct {
	ConnectionString         types.String               `tfsdk:"connection_string"`
	Host                     types.String               `tfsdk:"host"`
	Port                     types.Int64                `tfsdk:"port"`
	UseSSL                   types.Bool                 `tfsdk:"use_ssl"`
	BindDN                   types.String               `tfsdk:"bind_dn"`
	BindCredentials          types.String               `tfsdk:"bind_credentials"`
	BindCredentialsWO        types.String               `tfsdk:"bind_credentials_wo"`
	BindCredentialsWOVersion types.String               `tfsdk:"bind_credentials_wo_version"`
	BaseDN                   types.String               `tfsdk:"base_dn"`
	SearchBase               types.String               `tfsdk:"search_base"`
	Filter                   types.String               `tfsdk:"filter"`
	UserFilter               types.String               `tfsdk:"user_filter"`
	UniqueField              types.String               `tfsdk:"unique_field"`
	UserSync                 types.Bool                 `tfsdk:"user_sync"`
	DefaultGroups            []federationLDAPGroupModel `tfsdk:"default_groups"`
}

// federationLDAPGroupModel is one entry of ldap.default_groups.
type federationLDAPGroupModel struct {
	GroupID types.String `tfsdk:"group_id"`
	Roles   types.List   `tfsdk:"roles"`
}

// known reports whether v is neither null nor unknown.
func known(v attr.Value) bool {
	return !v.IsNull() && !v.IsUnknown()
}

// ---------------------------------------------------------------------------
// Terraform -> API
// ---------------------------------------------------------------------------

// expandFederationProvider builds the API model from the plan. Write-only secrets are read from config,
// because Terraform never puts them in the plan.
func expandFederationProvider(plan, config *federationProviderModel) (*cidaas.FederationProviderModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	pc := &cidaas.FederationProviderModel{
		ID:           plan.ID.ValueString(),
		ProviderName: plan.ProviderName.ValueString(),
		Provider:     plan.Provider.ValueString(),
		StandardType: plan.StandardType.ValueString(),
		DisplayName:  plan.DisplayName.ValueString(),
		LogoURL:      plan.LogoURL.ValueString(),
		Owner:        plan.Owner.ValueString(),
		Description:  plan.Description.ValueString(),
	}
	if known(plan.Domains) {
		pc.Domains = util.ListToStrings(plan.Domains)
	}
	if known(plan.EnabledForAdminPortal) {
		v := plan.EnabledForAdminPortal.ValueBool()
		pc.EnabledForAdminPortal = &v
	}
	if len(plan.AttributeMapping) > 0 {
		pc.AttributeMapping = make(map[string]cidaas.FederationAttributeMapping, len(plan.AttributeMapping))
		for key, m := range plan.AttributeMapping {
			am := cidaas.FederationAttributeMapping{
				FieldKey:         m.FieldKey.ValueString(),
				ExternalKey:      m.ExternalKey.ValueString(),
				DataType:         m.DataType.ValueString(),
				Ignore:           m.Ignore.ValueBool(),
				EmailVerifiedKey: m.EmailVerifiedKey.ValueString(),
			}
			if known(m.Default) {
				am.Default = m.Default.ValueString()
			}
			if m.Transform != nil {
				am.Transform = &cidaas.FederationValueTransform{
					Type:   m.Transform.Type.ValueString(),
					Config: mapToInterfaceMap(m.Transform.Config),
				}
			}
			pc.AttributeMapping[key] = am
		}
	}

	var oauth2Config *federationOAuth2Model
	var ldapConfig *federationLDAPModel
	if config != nil {
		oauth2Config, ldapConfig = config.OAuth2, config.LDAP
	}
	if plan.OAuth2 != nil {
		pc.OAuth2, diags = expandFederationOAuth2(plan.OAuth2, oauth2Config)
	}
	if plan.SAML != nil {
		pc.SAML = expandFederationSAML(plan.SAML)
	}
	if plan.LDAP != nil {
		pc.LDAP = expandFederationLDAP(plan.LDAP, ldapConfig)
	}
	return pc, diags
}

// expandFederationOAuth2 maps the oauth2 block to the API. A configured client_secret_wo replaces client_secret.
func expandFederationOAuth2(plan, config *federationOAuth2Model) (*cidaas.FederationOAuth2Config, diag.Diagnostics) {
	var diags diag.Diagnostics
	o := &cidaas.FederationOAuth2Config{
		ClientID:                plan.ClientID.ValueString(),
		ClientSecret:            plan.ClientSecret.ValueString(),
		Pkce:                    plan.Pkce.ValueBool(),
		UserinfoSource:          plan.UserinfoSource.ValueString(),
		AuthorizationEndpoint:   plan.AuthorizationEndpoint.ValueString(),
		TokenEndpoint:           plan.TokenEndpoint.ValueString(),
		UserinfoEndpoint:        plan.UserinfoEndpoint.ValueString(),
		WellknownEndpoint:       plan.WellknownEndpoint.ValueString(),
		AuthorizationParameters: mapToInterfaceMap(plan.AuthorizationParameters),
		Groups:                  plan.Groups.ValueString(),
	}
	if config != nil && known(config.ClientSecretWO) {
		o.ClientSecret = config.ClientSecretWO.ValueString()
	}
	if known(plan.Scopes) {
		o.Scopes = util.ListToStrings(plan.Scopes)
	}
	if known(plan.TokenEndpointAuth) && plan.TokenEndpointAuth.ValueString() != "" {
		raw := json.RawMessage(plan.TokenEndpointAuth.ValueString())
		if !json.Valid(raw) {
			diags.AddError("Invalid oauth2.token_endpoint_auth", "token_endpoint_auth must be a JSON object; use jsonencode({...}).")
		} else {
			o.TokenEndpointAuth = raw
		}
	}
	if plan.ScopesConfig != nil {
		sc := &cidaas.FederationOAuth2ScopesConfig{DisplayLabel: plan.ScopesConfig.DisplayLabel.ValueString()}
		for _, s := range plan.ScopesConfig.Scopes {
			sc.Scopes = append(sc.Scopes, cidaas.FederationOAuth2ScopeSetup{
				ScopeName:   s.ScopeName.ValueString(),
				Required:    s.Required.ValueBool(),
				Recommended: s.Recommended.ValueBool(),
			})
		}
		o.ScopesConfig = sc
	}
	if plan.Claims != nil {
		o.Claims = &cidaas.FederationOAuth2Claims{
			RequiredClaims: expandClaimNames(plan.Claims.RequiredClaims),
			OptionalClaims: expandClaimNames(plan.Claims.OptionalClaims),
		}
	}
	for _, a := range plan.AmrConfig {
		o.AmrConfig = append(o.AmrConfig, cidaas.FederationAmrConfig{
			AmrValue:    a.AmrValue.ValueString(),
			ExtAmrValue: a.ExtAmrValue.ValueString(),
		})
	}
	return o, diags
}

// expandClaimNames maps claim names to the API, or nil when the block is not set.
func expandClaimNames(m *federationClaimNamesModel) *cidaas.FederationOAuth2ClaimNames {
	if m == nil {
		return nil
	}
	return &cidaas.FederationOAuth2ClaimNames{
		UserInfo: listToStringsOrNil(m.UserInfo),
		IDToken:  listToStringsOrNil(m.IDToken),
	}
}

// expandFederationSAML maps the saml block to the API. idp_initiated_config.id is sent back so the API keeps it.
func expandFederationSAML(plan *federationSAMLModel) *cidaas.FederationSAMLConfig {
	s := &cidaas.FederationSAMLConfig{
		MetaData:              plan.MetaData.ValueString(),
		SignAuthRequest:       plan.SignAuthRequest.ValueBool(),
		UserIDAttribute:       plan.UserIDAttribute.ValueString(),
		ProtocolBindingMethod: plan.ProtocolBindingMethod.ValueString(),
	}
	if c := plan.IDPInitiatedConfig; c != nil {
		s.IDPInitiatedConfig = &cidaas.FederationSAMLIDPInitiatedConfig{
			ID:          valueStringIfKnown(c.ID),
			ClientID:    c.ClientID.ValueString(),
			Protocol:    c.Protocol.ValueString(),
			QueryString: c.QueryString.ValueString(),
			Enabled:     c.Enabled.ValueBool(),
		}
	}
	if c := plan.SignatureConfig; c != nil {
		s.SignatureConfig = &cidaas.FederationSAMLSignatureConfig{
			SigningAlgorithm: c.SigningAlgorithm.ValueString(),
			SignTargets:      listToStringsOrEmpty(c.SignTargets),
		}
	}
	if c := plan.EncryptionConfig; c != nil {
		s.EncryptionConfig = &cidaas.FederationSAMLEncryptionConfig{
			EncryptTargets:             listToStringsOrEmpty(c.EncryptTargets),
			RequireEncryptedFor:        listToStringsOrEmpty(c.RequireEncryptedFor),
			ContentEncryptionAlgorithm: c.ContentEncryptionAlgorithm.ValueString(),
			KeyTransportAlgorithm:      c.KeyTransportAlgorithm.ValueString(),
		}
	}
	return s
}

// expandFederationLDAP maps the ldap block to the API. A configured bind_credentials_wo replaces bind_credentials.
func expandFederationLDAP(plan, config *federationLDAPModel) *cidaas.FederationLDAPConfig {
	l := &cidaas.FederationLDAPConfig{
		ConnectionString: plan.ConnectionString.ValueString(),
		Host:             plan.Host.ValueString(),
		Port:             plan.Port.ValueInt64(),
		BindDN:           plan.BindDN.ValueString(),
		BindCredentials:  plan.BindCredentials.ValueString(),
		BaseDN:           plan.BaseDN.ValueString(),
		SearchBase:       plan.SearchBase.ValueString(),
		Filter:           plan.Filter.ValueString(),
		UserFilter:       plan.UserFilter.ValueString(),
		UniqueField:      plan.UniqueField.ValueString(),
	}
	if config != nil && known(config.BindCredentialsWO) {
		l.BindCredentials = config.BindCredentialsWO.ValueString()
	}
	if known(plan.UseSSL) {
		v := plan.UseSSL.ValueBool()
		l.UseSSL = &v
	}
	if known(plan.UserSync) {
		v := plan.UserSync.ValueBool()
		l.UserSync = &v
	}
	for _, g := range plan.DefaultGroups {
		l.DefaultGroups = append(l.DefaultGroups, cidaas.FederationLDAPGroup{
			GroupID: g.GroupID.ValueString(),
			Roles:   listToStringsOrNil(g.Roles),
		})
	}
	return l
}

// ---------------------------------------------------------------------------
// API -> Terraform
// ---------------------------------------------------------------------------

// flattenFederationProvider maps the API response to state. prior is the plan (after create/update) or the
// current state (read); it is used for values the API never returns (write-only secrets, wellknown_endpoint)
// and to keep secrets out of state when the practitioner uses the write-only variants.
// A nil prior means import / data source: everything the API returns is stored.
func flattenFederationProvider(api *cidaas.FederationProviderModel, prior *federationProviderModel) federationProviderModel {
	if prior == nil {
		prior = &federationProviderModel{}
	}
	out := federationProviderModel{
		ID:                    types.StringValue(api.ID),
		ProviderName:          types.StringValue(api.ProviderName),
		Provider:              util.StringValueOrNull(&api.Provider),
		StandardType:          util.StringValueOrNull(&api.StandardType),
		DisplayName:           util.StringValueOrNull(&api.DisplayName),
		LogoURL:               util.StringValueOrNull(&api.LogoURL),
		Owner:                 util.StringValueOrNull(&api.Owner),
		Description:           util.StringValueOrNull(&api.Description),
		Domains:               stringListOrPrior(api.Domains, prior.Domains),
		EnabledForAdminPortal: util.BoolValueOrNull(api.EnabledForAdminPortal),
		CreatedTime:           federationTimestamp(api.CreatedTime),
		UpdatedTime:           federationTimestamp(api.UpdatedTime),
	}
	if len(api.AttributeMapping) > 0 {
		out.AttributeMapping = make(map[string]federationAttributeMappingModel, len(api.AttributeMapping))
		for key, m := range api.AttributeMapping {
			var priorMapping federationAttributeMappingModel
			if prior.AttributeMapping != nil {
				priorMapping = prior.AttributeMapping[key]
			}
			out.AttributeMapping[key] = flattenAttributeMapping(m, priorMapping)
		}
	} else if prior.AttributeMapping != nil && len(prior.AttributeMapping) == 0 {
		out.AttributeMapping = map[string]federationAttributeMappingModel{}
	}

	// Import (state holds only id) and data sources have no configured provider_name.
	isImport := prior.ProviderName.IsNull()
	if api.OAuth2 != nil {
		out.OAuth2 = flattenFederationOAuth2(api.OAuth2, prior.OAuth2, isImport)
	}
	if api.SAML != nil {
		out.SAML = flattenFederationSAML(api.SAML, prior.SAML)
	}
	if api.LDAP != nil {
		out.LDAP = flattenFederationLDAP(api.LDAP, prior.LDAP, isImport)
	}
	return out
}

// flattenAttributeMapping maps one API attribute mapping to state. Non-string defaults are stored as JSON.
func flattenAttributeMapping(m cidaas.FederationAttributeMapping, prior federationAttributeMappingModel) federationAttributeMappingModel {
	out := federationAttributeMappingModel{
		FieldKey:         types.StringValue(m.FieldKey),
		ExternalKey:      util.StringValueOrNull(&m.ExternalKey),
		DataType:         util.StringValueOrNull(&m.DataType),
		Default:          types.StringNull(),
		Ignore:           types.BoolValue(m.Ignore),
		EmailVerifiedKey: util.StringValueOrNull(&m.EmailVerifiedKey),
	}
	if m.Default != nil {
		out.Default = types.StringValue(interfaceToString(m.Default))
	}
	if m.Transform != nil {
		var priorConfig types.Map
		if prior.Transform != nil {
			priorConfig = prior.Transform.Config
		}
		out.Transform = &federationTransformModel{
			Type:   util.StringValueOrNull(&m.Transform.Type),
			Config: stringMapOrPrior(m.Transform.Config, priorConfig),
		}
	}
	return out
}

// flattenFederationOAuth2 maps the API oauth2 nest to state. The secret is stored only when managed via client_secret or on import.
func flattenFederationOAuth2(api *cidaas.FederationOAuth2Config, prior *federationOAuth2Model, isImport bool) *federationOAuth2Model {
	if prior == nil {
		prior = &federationOAuth2Model{}
	}
	out := &federationOAuth2Model{
		ClientID:                types.StringValue(api.ClientID),
		ClientSecret:            types.StringNull(),
		ClientSecretWO:          types.StringNull(),
		ClientSecretWOVersion:   prior.ClientSecretWOVersion,
		Pkce:                    types.BoolValue(api.Pkce),
		UserinfoSource:          util.StringValueOrNull(&api.UserinfoSource),
		AuthorizationEndpoint:   util.StringValueOrNull(&api.AuthorizationEndpoint),
		TokenEndpoint:           util.StringValueOrNull(&api.TokenEndpoint),
		UserinfoEndpoint:        util.StringValueOrNull(&api.UserinfoEndpoint),
		WellknownEndpoint:       prior.WellknownEndpoint, // not persisted by the API (bson:"-"); keep what was configured
		AuthorizationParameters: stringMapOrPrior(api.AuthorizationParameters, prior.AuthorizationParameters),
		TokenEndpointAuth:       jsonOrPrior(api.TokenEndpointAuth, prior.TokenEndpointAuth),
		Scopes:                  stringListOrPrior(api.Scopes, prior.Scopes),
		Groups:                  util.StringValueOrNull(&api.Groups),
	}
	if out.ClientSecretWOVersion.IsUnknown() {
		out.ClientSecretWOVersion = types.StringNull()
	}
	if out.WellknownEndpoint.IsUnknown() {
		out.WellknownEndpoint = types.StringNull()
	}
	// Only store the secret when the practitioner manages it via client_secret (or on import).
	if known(prior.ClientSecret) || isImport {
		out.ClientSecret = util.StringValueOrNull(&api.ClientSecret)
	}
	if api.ScopesConfig != nil {
		sc := &federationScopesConfigModel{DisplayLabel: util.StringValueOrNull(&api.ScopesConfig.DisplayLabel)}
		for _, s := range api.ScopesConfig.Scopes {
			sc.Scopes = append(sc.Scopes, federationScopeSetupModel{
				ScopeName:   util.StringValueOrNull(&s.ScopeName),
				Required:    types.BoolValue(s.Required),
				Recommended: types.BoolValue(s.Recommended),
			})
		}
		if sc.Scopes == nil && prior.ScopesConfig != nil && prior.ScopesConfig.Scopes != nil {
			sc.Scopes = []federationScopeSetupModel{}
		}
		out.ScopesConfig = sc
	}
	if api.Claims != nil {
		var priorReq, priorOpt *federationClaimNamesModel
		if prior.Claims != nil {
			priorReq, priorOpt = prior.Claims.RequiredClaims, prior.Claims.OptionalClaims
		}
		out.Claims = &federationClaimsModel{
			RequiredClaims: flattenClaimNames(api.Claims.RequiredClaims, priorReq),
			OptionalClaims: flattenClaimNames(api.Claims.OptionalClaims, priorOpt),
		}
	}
	for _, a := range api.AmrConfig {
		out.AmrConfig = append(out.AmrConfig, federationAmrConfigModel{
			AmrValue:    types.StringValue(a.AmrValue),
			ExtAmrValue: types.StringValue(a.ExtAmrValue),
		})
	}
	if out.AmrConfig == nil && prior.AmrConfig != nil {
		out.AmrConfig = []federationAmrConfigModel{}
	}
	return out
}

// flattenClaimNames maps API claim names to state, keeping explicitly configured empty lists.
func flattenClaimNames(api *cidaas.FederationOAuth2ClaimNames, prior *federationClaimNamesModel) *federationClaimNamesModel {
	if api == nil {
		return nil
	}
	if prior == nil {
		prior = &federationClaimNamesModel{}
	}
	return &federationClaimNamesModel{
		UserInfo: stringListOrPrior(api.UserInfo, prior.UserInfo),
		IDToken:  stringListOrPrior(api.IDToken, prior.IDToken),
	}
}

// flattenFederationSAML maps the API saml nest to state, including the generated SP certificates.
func flattenFederationSAML(api *cidaas.FederationSAMLConfig, prior *federationSAMLModel) *federationSAMLModel {
	if prior == nil {
		prior = &federationSAMLModel{}
	}
	out := &federationSAMLModel{
		MetaData:                 util.StringValueOrNull(&api.MetaData),
		SignAuthRequest:          types.BoolValue(api.SignAuthRequest),
		UserIDAttribute:          util.StringValueOrNull(&api.UserIDAttribute),
		ProtocolBindingMethod:    util.StringValueOrNull(&api.ProtocolBindingMethod),
		SPSigningCertificates:    flattenCertificates(api.SPSigningCertificates),
		SPEncryptionCertificates: flattenCertificates(api.SPEncryptionCertificates),
	}
	if c := api.IDPInitiatedConfig; c != nil {
		out.IDPInitiatedConfig = &federationIDPInitiatedModel{
			ID:          util.StringValueOrNull(&c.ID),
			ClientID:    types.StringValue(c.ClientID),
			Protocol:    types.StringValue(c.Protocol),
			QueryString: util.StringValueOrNull(&c.QueryString),
			Enabled:     types.BoolValue(c.Enabled),
		}
	}
	if c := api.SignatureConfig; c != nil {
		var priorTargets types.List
		if prior.SignatureConfig != nil {
			priorTargets = prior.SignatureConfig.SignTargets
		}
		out.SignatureConfig = &federationSignatureConfigModel{
			SigningAlgorithm: util.StringValueOrNull(&c.SigningAlgorithm),
			SignTargets:      stringListOrPrior(c.SignTargets, priorTargets),
		}
	}
	if c := api.EncryptionConfig; c != nil {
		var priorEncrypt, priorRequire types.List
		if prior.EncryptionConfig != nil {
			priorEncrypt, priorRequire = prior.EncryptionConfig.EncryptTargets, prior.EncryptionConfig.RequireEncryptedFor
		}
		out.EncryptionConfig = &federationEncryptionConfigModel{
			EncryptTargets:             stringListOrPrior(c.EncryptTargets, priorEncrypt),
			RequireEncryptedFor:        stringListOrPrior(c.RequireEncryptedFor, priorRequire),
			ContentEncryptionAlgorithm: util.StringValueOrNull(&c.ContentEncryptionAlgorithm),
			KeyTransportAlgorithm:      util.StringValueOrNull(&c.KeyTransportAlgorithm),
		}
	}
	return out
}

// flattenCertificates maps API certificates to a list (empty, not null, so the computed value is always known).
func flattenCertificates(certs []cidaas.FederationSAMLCertificate) types.List {
	elems := make([]attr.Value, 0, len(certs))
	for _, c := range certs {
		elems = append(elems, types.ObjectValueMust(federationCertificateType.AttrTypes, map[string]attr.Value{
			"certificate":        util.StringValueOrNull(&c.Certificate),
			"sha256_fingerprint": util.StringValueOrNull(&c.Sha256Fingerprint),
		}))
	}
	return types.ListValueMust(federationCertificateType, elems)
}

// flattenFederationLDAP maps the API ldap nest to state. The bind password is stored only when managed via bind_credentials or on import.
func flattenFederationLDAP(api *cidaas.FederationLDAPConfig, prior *federationLDAPModel, isImport bool) *federationLDAPModel {
	if prior == nil {
		prior = &federationLDAPModel{}
	}
	out := &federationLDAPModel{
		ConnectionString:         util.StringValueOrNull(&api.ConnectionString),
		Host:                     util.StringValueOrNull(&api.Host),
		Port:                     types.Int64Null(),
		UseSSL:                   util.BoolValueOrNull(api.UseSSL),
		BindDN:                   util.StringValueOrNull(&api.BindDN),
		BindCredentials:          types.StringNull(),
		BindCredentialsWO:        types.StringNull(),
		BindCredentialsWOVersion: prior.BindCredentialsWOVersion,
		BaseDN:                   util.StringValueOrNull(&api.BaseDN),
		SearchBase:               util.StringValueOrNull(&api.SearchBase),
		Filter:                   util.StringValueOrNull(&api.Filter),
		UserFilter:               util.StringValueOrNull(&api.UserFilter),
		UniqueField:              util.StringValueOrNull(&api.UniqueField),
		UserSync:                 util.BoolValueOrNull(api.UserSync),
	}
	if out.BindCredentialsWOVersion.IsUnknown() {
		out.BindCredentialsWOVersion = types.StringNull()
	}
	if api.Port != 0 {
		out.Port = types.Int64Value(api.Port)
	}
	if known(prior.BindCredentials) || isImport {
		secret := api.BindCredentials
		if secret == "" {
			secret = api.BindPassword
		}
		out.BindCredentials = util.StringValueOrNull(&secret)
	}
	for _, g := range api.DefaultGroups {
		out.DefaultGroups = append(out.DefaultGroups, federationLDAPGroupModel{
			GroupID: util.StringValueOrNull(&g.GroupID),
			Roles:   listStringOrNull(g.Roles),
		})
	}
	if out.DefaultGroups == nil && prior.DefaultGroups != nil {
		out.DefaultGroups = []federationLDAPGroupModel{}
	}
	return out
}

// ---------------------------------------------------------------------------
// value helpers
// ---------------------------------------------------------------------------

// federationTimestamp normalizes an API timestamp to millisecond precision. Create/update responses carry
// nanoseconds while GET returns what MongoDB stored (milliseconds), so without this the value would change
// on the first refresh and break import verification.
func federationTimestamp(v string) types.String {
	if v == "" {
		return types.StringNull()
	}
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return types.StringValue(v)
	}
	return types.StringValue(t.UTC().Truncate(time.Millisecond).Format(time.RFC3339Nano))
}

// valueStringIfKnown returns the string value, or "" when v is null or unknown.
func valueStringIfKnown(v types.String) string {
	if !known(v) {
		return ""
	}
	return v.ValueString()
}

// listToStringsOrNil converts a known list to strings, or returns nil.
func listToStringsOrNil(l types.List) []string {
	if !known(l) {
		return nil
	}
	return util.ListToStrings(l)
}

// listToStringsOrEmpty is used for SAML target lists, which the API serializes without omitempty.
func listToStringsOrEmpty(l types.List) []string {
	if v := listToStringsOrNil(l); v != nil {
		return v
	}
	return []string{}
}

// stringListOrPrior returns null for an empty API list, unless the practitioner configured an explicit empty list.
func stringListOrPrior(values []string, prior types.List) types.List {
	if len(values) == 0 && known(prior) && len(prior.Elements()) == 0 {
		return types.ListValueMust(types.StringType, []attr.Value{})
	}
	return listStringOrNull(values)
}

// stringMapOrPrior converts an API map to a Terraform string map, keeping an explicitly configured empty map.
func stringMapOrPrior(values map[string]interface{}, prior types.Map) types.Map {
	if len(values) == 0 {
		if known(prior) && len(prior.Elements()) == 0 {
			return types.MapValueMust(types.StringType, map[string]attr.Value{})
		}
		return types.MapNull(types.StringType)
	}
	elems := make(map[string]attr.Value, len(values))
	for k, v := range values {
		elems[k] = types.StringValue(interfaceToString(v))
	}
	return types.MapValueMust(types.StringType, elems)
}

// mapToInterfaceMap converts a known Terraform string map for the API, or returns nil.
func mapToInterfaceMap(m types.Map) map[string]interface{} {
	if !known(m) {
		return nil
	}
	out := make(map[string]interface{}, len(m.Elements()))
	for k, v := range m.Elements() {
		if s, ok := v.(types.String); ok && known(s) {
			out[k] = s.ValueString()
		}
	}
	return out
}

// interfaceToString returns strings as-is and JSON-encodes any other value.
func interfaceToString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return ""
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprint(t)
		}
		return string(b)
	}
}

// jsonOrPrior keeps the configured JSON string when it is semantically equal to the API value,
// so formatting differences do not show up as drift.
func jsonOrPrior(raw json.RawMessage, prior types.String) types.String {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return types.StringNull()
	}
	if known(prior) && jsonEqual([]byte(prior.ValueString()), trimmed) {
		return prior
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, trimmed); err != nil {
		return types.StringValue(string(trimmed))
	}
	return types.StringValue(compact.String())
}

// jsonEqual reports whether a and b are semantically equal JSON documents.
func jsonEqual(a, b []byte) bool {
	var va, vb interface{}
	if json.Unmarshal(a, &va) != nil || json.Unmarshal(b, &vb) != nil {
		return false
	}
	ja, _ := json.Marshal(va)
	jb, _ := json.Marshal(vb)
	return bytes.Equal(ja, jb)
}

// conformValue drops object attributes that are not part of target (used to turn the resource
// representation into the data source representation, which has no write-only attributes).
func conformValue(v attr.Value, target attr.Type) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch tt := target.(type) {
	case basetypes.ObjectType:
		obj, ok := v.(basetypes.ObjectValue)
		if !ok {
			return v, diags
		}
		if obj.IsNull() {
			return types.ObjectNull(tt.AttrTypes), diags
		}
		if obj.IsUnknown() {
			return types.ObjectUnknown(tt.AttrTypes), diags
		}
		src := obj.Attributes()
		attrs := make(map[string]attr.Value, len(tt.AttrTypes))
		for name, at := range tt.AttrTypes {
			cv, d := conformValue(src[name], at)
			diags.Append(d...)
			attrs[name] = cv
		}
		out, d := types.ObjectValue(tt.AttrTypes, attrs)
		diags.Append(d...)
		return out, diags
	case basetypes.ListType:
		l, ok := v.(basetypes.ListValue)
		if !ok || l.IsNull() || l.IsUnknown() {
			if !ok {
				return v, diags
			}
			if l.IsNull() {
				return types.ListNull(tt.ElemType), diags
			}
			return types.ListUnknown(tt.ElemType), diags
		}
		elems := make([]attr.Value, 0, len(l.Elements()))
		for _, e := range l.Elements() {
			ce, d := conformValue(e, tt.ElemType)
			diags.Append(d...)
			elems = append(elems, ce)
		}
		out, d := types.ListValue(tt.ElemType, elems)
		diags.Append(d...)
		return out, diags
	case basetypes.MapType:
		m, ok := v.(basetypes.MapValue)
		if !ok || m.IsNull() || m.IsUnknown() {
			if !ok {
				return v, diags
			}
			if m.IsNull() {
				return types.MapNull(tt.ElemType), diags
			}
			return types.MapUnknown(tt.ElemType), diags
		}
		elems := make(map[string]attr.Value, len(m.Elements()))
		keys := make([]string, 0, len(m.Elements()))
		for k := range m.Elements() {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			ce, d := conformValue(m.Elements()[k], tt.ElemType)
			diags.Append(d...)
			elems[k] = ce
		}
		out, d := types.MapValue(tt.ElemType, elems)
		diags.Append(d...)
		return out, diags
	}
	return v, diags
}
