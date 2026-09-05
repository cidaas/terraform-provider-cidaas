package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// AppConfigurationService calls app-srv app configuration APIs (/app-srv/apps).
type AppConfigurationService struct {
	cfg Config
}

// NewAppConfigurationService returns an AppConfigurationService bound to cfg.
func NewAppConfigurationService(cfg Config) *AppConfigurationService {
	return &AppConfigurationService{cfg: cfg}
}

// OwnerClient is the appv3 owner value for customer (Trustdesk) applications.
const OwnerClient = "client"

// AppConfigurationModel matches appv3.App JSON (MVP subset for Terraform).
type AppConfigurationModel struct {
	ID                        string                     `json:"_id,omitempty"`
	ClientID                  string                     `json:"client_id,omitempty"`
	ClientName                string                     `json:"client_name"`
	ClientType                string                     `json:"client_type"`
	Owner                     string                     `json:"owner,omitempty"`
	Enabled                   *bool                      `json:"enabled,omitempty"`
	RequirePKCE               *bool                      `json:"require_pkce,omitempty"`
	DisableInsecurePKCEMethod *bool                      `json:"disable_insecure_pkce_method,omitempty"`
	PKCE                      *PKCEConfig                `json:"pkce,omitempty"`
	GrantTypes                []string                   `json:"grant_types,omitempty"`
	ResponseTypes             []string                   `json:"response_types,omitempty"`
	RedirectURIs              *RedirectURIsConfig        `json:"redirect_uris,omitempty"`
	Scopes                    *ScopesConfig              `json:"scopes,omitempty"`
	TokenLifetimes            *TokenLifetimesConfig      `json:"token_lifetimes,omitempty"`
	AuthenticationSetup       *AuthenticationSetupConfig `json:"authentication_setup,omitempty"`
	HostedPagesLayoutID       string                     `json:"hosted_pages_layout_id,omitempty"`
	UserSetupID               string                     `json:"user_setup_id,omitempty"`
	OwnershipDetails          *OwnershipDetailsConfig    `json:"owner_ship_details,omitempty"`
	ClientAuthConfig          *AuthConfig                `json:"client_auth_config,omitempty"`
	SigningKeyConfig          *SigningKeyConfig          `json:"signing_key_config,omitempty"`
	CreatedTime               string                     `json:"created_time,omitempty"`
	UpdatedTime               string                     `json:"updated_time,omitempty"`
}

// PKCEConfig maps the nested pkce object required by app-srv validation.
type PKCEConfig struct {
	RequirePKCE               *bool `json:"require_pkce,omitempty"`
	DisableInsecurePKCEMethod *bool `json:"disable_insecure_pkce_method,omitempty"`
}

// RedirectURIsConfig maps redirect_uris nested object.
type RedirectURIsConfig struct {
	RedirectURIs           []string `json:"redirect_uris,omitempty"`
	AllowedLogoutUrls      []string `json:"allowed_logout_urls,omitempty"`
	PostLogoutRedirectURIs []string `json:"post_logout_redirect_uris,omitempty"`
	AllowedWebOrigins      []string `json:"allowed_web_origins,omitempty"`
}

// ScopesConfig maps scopes nested object.
type ScopesConfig struct {
	AllowedScopes []string `json:"allowed_scopes,omitempty"`
	DefaultScopes []string `json:"default_scopes,omitempty"`
}

// TokenLifetimesConfig maps token_lifetimes nested object.
type TokenLifetimesConfig struct {
	TokenLifetimeInSeconds        *int64 `json:"token_lifetime_in_seconds,omitempty"`
	RefreshTokenLifetimeInSeconds *int64 `json:"refresh_token_lifetime_in_seconds,omitempty"`
	IDTokenLifetimeInSeconds      *int64 `json:"id_token_lifetime_in_seconds,omitempty"`
	CodeLifetimeInSeconds         *int64 `json:"code_lifetime_in_seconds,omitempty"`
	DefaultMaxAge                 *int64 `json:"default_max_age,omitempty"`
}

// AuthenticationSetupConfig maps authentication_setup nested object (extdep IDs + flags).
type AuthenticationSetupConfig struct {
	VerificationOptionsID        string `json:"verification_options_id,omitempty"`
	GroupSelectionID             string `json:"group_selection_id,omitempty"`
	GroupVerificationRequestID   string `json:"group_verification_request_id,omitempty"`
	TemplateGroupID              string `json:"template_group_id,omitempty"`
	AllowGuestLogin              *bool  `json:"allow_guest_login,omitempty"`
	IsRememberMeSelected         *bool  `json:"is_remember_me_selected,omitempty"`
	AdminClient                  *bool  `json:"admin_client,omitempty"`
	IsLoginSuccessPageEnabled    *bool  `json:"is_login_success_page_enabled,omitempty"`
	IsRegisterSuccessPageEnabled *bool  `json:"is_register_success_page_enabled,omitempty"`
}

// OwnershipDetailsConfig maps owner_ship_details (required on create).
type OwnershipDetailsConfig struct {
	CompanyName    string   `json:"company_name"`
	CompanyAddress string   `json:"company_address"`
	CompanyWebsite string   `json:"company_website"`
	GroupIDs       []string `json:"groupIds,omitempty"`
}

// AuthConfig maps client_auth_config (OAuth client authentication).
type AuthConfig struct {
	TokenEndpointAuthMethod string `json:"token_endpoint_auth_method,omitempty"`
}

// SigningKeyConfig is server-assigned signing key metadata.
type SigningKeyConfig struct {
	ActiveKID string `json:"active_kid,omitempty"`
	NextKID   string `json:"next_kid,omitempty"`
}

// AppConfigurationResponse is the standard {success,status,data} envelope.
type AppConfigurationResponse struct {
	Success bool                  `json:"success"`
	Status  int                   `json:"status"`
	Data    AppConfigurationModel `json:"data"`
}

// Create POSTs /app-srv/apps/.
func (s *AppConfigurationService) Create(ctx context.Context, model AppConfigurationModel) (*AppConfigurationResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "app-srv/apps")
	if err != nil {
		return nil, err
	}
	endpoint += "/"
	var out AppConfigurationResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPost, endpoint, model, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get GETs /app-srv/apps/{clientID}.
func (s *AppConfigurationService) Get(ctx context.Context, clientID string) (*AppConfigurationResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "app-srv/apps", clientID)
	if err != nil {
		return nil, err
	}
	var out AppConfigurationResponse
	if err := requestJSON(ctx, s.cfg, http.MethodGet, endpoint, nil, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update PUTs /app-srv/apps/{clientID} (full replace).
func (s *AppConfigurationService) Update(ctx context.Context, clientID string, model AppConfigurationModel) (*AppConfigurationResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "app-srv/apps", clientID)
	if err != nil {
		return nil, err
	}
	var out AppConfigurationResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPut, endpoint, model, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete DELETEs /app-srv/apps/{clientID}.
func (s *AppConfigurationService) Delete(ctx context.Context, clientID string) error {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "app-srv/apps", clientID)
	if err != nil {
		return err
	}
	if err := requestJSON(ctx, s.cfg, http.MethodDelete, endpoint, nil, nil, http.StatusNoContent, http.StatusOK); err != nil {
		return fmt.Errorf("delete app configuration: %w", err)
	}
	return nil
}
