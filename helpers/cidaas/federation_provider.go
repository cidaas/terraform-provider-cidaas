package cidaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

// Standard types accepted by /federation/providers.
const (
	FederationStandardTypeOAuth2 = "OAUTH2"
	FederationStandardTypeOIDC   = "OPENID_CONNECT"
	FederationStandardTypeSAML   = "SAML"
	FederationStandardTypeLDAP   = "LDAP"
)

// FederationSocialProviders is the catalog of social providers whose endpoints are owned by cidaas
// (provider-model knownSocialProviderNames). For these, OAuth2 endpoints and userinfoSource are optional.
var FederationSocialProviders = []string{
	"amazon", "apple", "discord", "facebook", "foursquare", "github", "gitlab", "google", "instagram",
	"linkedin", "microsoft", "netid", "netid_qa", "officeenterprise", "paypal", "paypal_sandbox",
	"salesforce", "slack", "tiktok", "twitter", "wordpress", "x", "yahoo", "yammer",
}

// IsFederationSocialProvider reports whether provider is a catalog social provider.
func IsFederationSocialProvider(provider string) bool {
	return util.Contains(FederationSocialProviders, provider)
}

// FederationProviderModel mirrors provider-model ProviderConfig (base fields + exactly one protocol nest).
type FederationProviderModel struct {
	ID                    string                                `json:"id,omitempty"`
	MongoID               string                                `json:"_id,omitempty"`
	ProviderName          string                                `json:"provider_name,omitempty"`
	Provider              string                                `json:"provider,omitempty"`
	DisplayName           string                                `json:"display_name,omitempty"`
	StandardType          string                                `json:"standard_type,omitempty"`
	LogoURL               string                                `json:"logo_url,omitempty"`
	Owner                 string                                `json:"owner,omitempty"`
	Description           string                                `json:"description,omitempty"`
	Domains               []string                              `json:"domains,omitempty"`
	EnabledForAdminPortal *bool                                 `json:"enabled_for_admin_portal,omitempty"`
	AttributeMapping      map[string]FederationAttributeMapping `json:"attribute_mapping,omitempty"`
	CreatedTime           string                                `json:"createdTime,omitempty"`
	UpdatedTime           string                                `json:"updatedTime,omitempty"`
	OAuth2                *FederationOAuth2Config               `json:"oauth2,omitempty"`
	SAML                  *FederationSAMLConfig                 `json:"saml,omitempty"`
	LDAP                  *FederationLDAPConfig                 `json:"ldap,omitempty"`
}

// FederationAttributeMapping maps one external attribute to a cidaas field (map key = FieldKey).
type FederationAttributeMapping struct {
	FieldKey         string                    `json:"field_key"`
	ExternalKey      string                    `json:"external_key"`
	DataType         string                    `json:"data_type,omitempty"`
	Default          interface{}               `json:"default,omitempty"`
	Ignore           bool                      `json:"ignore"`
	Transform        *FederationValueTransform `json:"transform,omitempty"`
	EmailVerifiedKey string                    `json:"email_verified_key,omitempty"`
}

// FederationValueTransform is an optional transformation applied to a mapped value.
type FederationValueTransform struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// FederationOAuth2Config is the oauth2 nest, used for OAUTH2 and OPENID_CONNECT providers.
type FederationOAuth2Config struct {
	ClientID                string                        `json:"client_id"`
	ClientSecret            string                        `json:"client_secret,omitempty"`
	Pkce                    bool                          `json:"pkce,omitempty"`
	Scopes                  []string                      `json:"scopes,omitempty"`
	ScopesConfig            *FederationOAuth2ScopesConfig `json:"scopes_config,omitempty"`
	AuthorizationParameters map[string]interface{}        `json:"authorization_parameters,omitempty"`
	AuthorizationEndpoint   string                        `json:"authorization_endpoint,omitempty"`
	TokenEndpoint           string                        `json:"token_endpoint,omitempty"`
	WellknownEndpoint       string                        `json:"wellknown_endpoint,omitempty"`
	UserinfoEndpoint        string                        `json:"userinfo_endpoint,omitempty"`
	TokenEndpointAuth       json.RawMessage               `json:"token_endpoint_auth,omitempty"`
	UserinfoSource          string                        `json:"userinfoSource,omitempty"`
	AmrConfig               []FederationAmrConfig         `json:"amrConfig,omitempty"`
	Groups                  string                        `json:"groups,omitempty"`
	Claims                  *FederationOAuth2Claims       `json:"claims,omitempty"`
}

// FederationOAuth2ScopesConfig is the full scope setup (display label and per-scope flags).
type FederationOAuth2ScopesConfig struct {
	DisplayLabel string                       `json:"display_label,omitempty"`
	Scopes       []FederationOAuth2ScopeSetup `json:"scopes,omitempty"`
}

// FederationOAuth2ScopeSetup is a single scope entry of FederationOAuth2ScopesConfig.
type FederationOAuth2ScopeSetup struct {
	ScopeName   string `json:"scope_name,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Recommended bool   `json:"recommended,omitempty"`
}

// FederationOAuth2Claims lists the claims that are required or optionally read from the provider.
type FederationOAuth2Claims struct {
	RequiredClaims *FederationOAuth2ClaimNames `json:"required_claims,omitempty"`
	OptionalClaims *FederationOAuth2ClaimNames `json:"optional_claims,omitempty"`
}

// FederationOAuth2ClaimNames holds claim names per source (userinfo response, ID token).
type FederationOAuth2ClaimNames struct {
	UserInfo []string `json:"user_info,omitempty"`
	IDToken  []string `json:"id_token,omitempty"`
}

// FederationAmrConfig maps an external AMR value to a cidaas AMR value.
type FederationAmrConfig struct {
	AmrValue    string `json:"amrValue"`
	ExtAmrValue string `json:"extAmrValue"`
}

// FederationSAMLConfig is the saml nest. SP certificates are generated by the API and never sent.
type FederationSAMLConfig struct {
	MetaData                 string                            `json:"meta_data,omitempty"`
	SignAuthRequest          bool                              `json:"sign_auth_request"`
	UserIDAttribute          string                            `json:"user_id_attribute,omitempty"`
	ProtocolBindingMethod    string                            `json:"protocol_binding_method,omitempty"`
	IDPInitiatedConfig       *FederationSAMLIDPInitiatedConfig `json:"idp_initiated_config,omitempty"`
	SignatureConfig          *FederationSAMLSignatureConfig    `json:"signature_config,omitempty"`
	EncryptionConfig         *FederationSAMLEncryptionConfig   `json:"encryption_config,omitempty"`
	SPSigningCertificates    []FederationSAMLCertificate       `json:"sp_signing_certificates,omitempty"`
	SPEncryptionCertificates []FederationSAMLCertificate       `json:"sp_encryption_certificates,omitempty"`
}

// FederationSAMLIDPInitiatedConfig configures IdP-initiated login. ID is assigned by the API.
type FederationSAMLIDPInitiatedConfig struct {
	ID          string `json:"_id,omitempty"`
	ClientID    string `json:"client_id"`
	Protocol    string `json:"protocol"`
	QueryString string `json:"query_string,omitempty"`
	Enabled     bool   `json:"enabled"`
}

// FederationSAMLSignatureConfig configures signing of outgoing SAML messages.
type FederationSAMLSignatureConfig struct {
	SigningAlgorithm string   `json:"signing_algorithm,omitempty"`
	SignTargets      []string `json:"sign_targets"`
}

// FederationSAMLEncryptionConfig configures encryption of SAML messages.
type FederationSAMLEncryptionConfig struct {
	EncryptTargets             []string `json:"encrypt_targets"`
	RequireEncryptedFor        []string `json:"require_encrypted_for"`
	ContentEncryptionAlgorithm string   `json:"content_encryption_algorithm,omitempty"`
	KeyTransportAlgorithm      string   `json:"key_transport_algorithm,omitempty"`
}

// FederationSAMLCertificate is API-generated (read-only).
type FederationSAMLCertificate struct {
	Certificate       string `json:"certificate,omitempty"`
	Sha256Fingerprint string `json:"sha256_fingerprint,omitempty"`
}

// FederationLDAPConfig is the ldap nest. BindCredentials and BindPassword are aliases on the API.
type FederationLDAPConfig struct {
	ConnectionString string                `json:"connection_string,omitempty"`
	Host             string                `json:"host,omitempty"`
	Port             int64                 `json:"port,omitempty"`
	UseSSL           *bool                 `json:"use_ssl,omitempty"`
	BindDN           string                `json:"bind_dn,omitempty"`
	BindCredentials  string                `json:"bind_credentials,omitempty"`
	BindPassword     string                `json:"bind_password,omitempty"`
	BaseDN           string                `json:"base_dn,omitempty"`
	SearchBase       string                `json:"search_base,omitempty"`
	Filter           string                `json:"filter,omitempty"`
	UserFilter       string                `json:"user_filter,omitempty"`
	UniqueField      string                `json:"unique_field,omitempty"`
	UserSync         *bool                 `json:"user_sync,omitempty"`
	DefaultGroups    []FederationLDAPGroup `json:"default_groups,omitempty"`
}

// FederationLDAPGroup is a default group (and its roles) assigned to LDAP users.
type FederationLDAPGroup struct {
	GroupID string   `json:"groupId,omitempty"`
	Roles   []string `json:"roles,omitempty"`
}

// FederationProviderResponse wraps a single provider returned by the API.
type FederationProviderResponse struct {
	Success bool                    `json:"success,omitempty"`
	Status  int                     `json:"status,omitempty"`
	Data    FederationProviderModel `json:"data,omitempty"`
}

// FederationProviderListResponse wraps the provider list returned by GET /federation/providers.
type FederationProviderListResponse struct {
	Success bool                      `json:"success,omitempty"`
	Status  int                       `json:"status,omitempty"`
	Data    []FederationProviderModel `json:"data,omitempty"`
}

// FederationProviderListParams are the query filters supported by GET /federation/providers.
type FederationProviderListParams struct {
	ProviderName string
	StandardType string
	Size         int
	From         int
}

// FederationProvider is the client for /federation/providers.
type FederationProvider struct {
	ClientConfig
}

// NewFederationProvider returns a FederationProvider client for the given configuration.
func NewFederationProvider(clientConfig ClientConfig) *FederationProvider {
	return &FederationProvider{clientConfig}
}

// buildFederationProviderPayload returns the request body: read-only fields (timestamps, SP certificates)
// are dropped, only the nest matching standard_type is sent, and id is set only for updates.
func buildFederationProviderPayload(pc *FederationProviderModel, id string) *FederationProviderModel {
	payload := *pc
	payload.ID = id
	payload.MongoID = ""
	payload.CreatedTime = ""
	payload.UpdatedTime = ""
	if payload.Owner == "" {
		payload.Owner = "client"
	}

	oauth2, saml, ldap := payload.OAuth2, payload.SAML, payload.LDAP
	payload.OAuth2, payload.SAML, payload.LDAP = nil, nil, nil
	switch payload.StandardType {
	case FederationStandardTypeOAuth2, FederationStandardTypeOIDC:
		payload.OAuth2 = oauth2
	case FederationStandardTypeSAML:
		if saml != nil {
			s := *saml
			s.SPSigningCertificates = nil
			s.SPEncryptionCertificates = nil
			payload.SAML = &s
		}
	case FederationStandardTypeLDAP:
		payload.LDAP = ldap
	}
	return &payload
}

// normalizeFederationProvider fills ID from the legacy `_id` key when the API returns it.
func normalizeFederationProvider(data *FederationProviderModel) {
	if data.ID == "" {
		data.ID = data.MongoID
	}
}

// do sends a request with an optional body and decodes a single-provider response.
func (f *FederationProvider) do(ctx context.Context, method, url string, body interface{}) (*FederationProviderResponse, error) {
	var response FederationProviderResponse
	client, err := util.NewHTTPClient(url, method, f.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, body)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	normalizeFederationProvider(&response.Data)
	return &response, nil
}

// Create sends POST /federation/providers. The payload never contains an id or read-only fields.
func (f *FederationProvider) Create(ctx context.Context, pc *FederationProviderModel) (*FederationProviderResponse, error) {
	return f.do(ctx, http.MethodPost, fmt.Sprintf("%s/%s", f.BaseURL, "federation/providers"), buildFederationProviderPayload(pc, ""))
}

// Get sends GET /federation/providers/:id.
func (f *FederationProvider) Get(ctx context.Context, id string) (*FederationProviderResponse, error) {
	return f.do(ctx, http.MethodGet, fmt.Sprintf("%s/%s/%s", f.BaseURL, "federation/providers", url.PathEscape(id)), nil)
}

// Update sends PUT /federation/providers/:id. The API replaces the whole document, so the full model is sent.
func (f *FederationProvider) Update(ctx context.Context, id string, pc *FederationProviderModel) (*FederationProviderResponse, error) {
	return f.do(ctx, http.MethodPut, fmt.Sprintf("%s/%s/%s", f.BaseURL, "federation/providers", url.PathEscape(id)), buildFederationProviderPayload(pc, id))
}

// Delete sends DELETE /federation/providers/:id.
func (f *FederationProvider) Delete(ctx context.Context, id string) error {
	reqURL := fmt.Sprintf("%s/%s/%s", f.BaseURL, "federation/providers", url.PathEscape(id))
	client, err := util.NewHTTPClient(reqURL, http.MethodDelete, f.AccessToken)
	if err != nil {
		return err
	}
	res, err := client.MakeRequest(ctx, nil)
	if err := util.HandleResponseError(res, err); err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	return nil
}

// List sends GET /federation/providers with optional filters. The API caps size at 100.
func (f *FederationProvider) List(ctx context.Context, params FederationProviderListParams) ([]FederationProviderModel, error) {
	query := url.Values{}
	if params.ProviderName != "" {
		query.Set("provider_name", params.ProviderName)
	}
	if params.StandardType != "" {
		query.Set("standard_type", params.StandardType)
	}
	if params.Size > 0 {
		query.Set("size", strconv.Itoa(params.Size))
	}
	if params.From > 0 {
		query.Set("from", strconv.Itoa(params.From))
	}
	reqURL := fmt.Sprintf("%s/%s", f.BaseURL, "federation/providers")
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	var response FederationProviderListResponse
	client, err := util.NewHTTPClient(reqURL, http.MethodGet, f.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, nil)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	for i := range response.Data {
		normalizeFederationProvider(&response.Data[i])
	}
	return response.Data, nil
}
