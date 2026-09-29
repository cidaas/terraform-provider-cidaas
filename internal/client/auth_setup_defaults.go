package client

import (
	"context"
	"net/http"
	"net/url"
)

// AuthSetupDefaultsService calls app-srv tenant auth-setup-defaults APIs.
type AuthSetupDefaultsService struct {
	cfg Config
}

// NewAuthSetupDefaultsService returns an AuthSetupDefaultsService bound to cfg.
func NewAuthSetupDefaultsService(cfg Config) *AuthSetupDefaultsService {
	return &AuthSetupDefaultsService{cfg: cfg}
}

// AuthSetupDefaultsEntity matches Auth_AuthSetupDefaults `_id="default"` JSON.
type AuthSetupDefaultsEntity struct {
	ID                string             `json:"_id,omitempty"`
	Name              string             `json:"name,omitempty"`
	Description       string             `json:"description,omitempty"`
	Owner             string             `json:"owner,omitempty"`
	AuthSetupDefaults *AuthSetupDefaults `json:"auth_setup_defaults,omitempty"`
	CreatedTime       string             `json:"createdTime,omitempty"`
	UpdatedTime       string             `json:"updatedTime,omitempty"`
}

// AuthSetupDefaults holds the six globally-defaultable AuthenticationSetup fields.
type AuthSetupDefaults struct {
	AutoLoginAfterRegister       *bool `json:"auto_login_after_register,omitempty"`
	RegisterWithLoginInformation *bool `json:"register_with_login_information,omitempty"`
	EnablePasswordLessAuth       *bool `json:"enable_password_less_auth,omitempty"`
	AllowUserLevelMultiProvider  *bool `json:"allow_user_level_multi_provider,omitempty"`
	SocialBusinessIDs            *bool `json:"social_business_ids,omitempty"`
	NetID                        *bool `json:"net_id,omitempty"`
}

// AuthSetupDefaultsResponse is the standard {success,status,data} envelope.
type AuthSetupDefaultsResponse struct {
	Success bool                    `json:"success"`
	Status  int                     `json:"status"`
	Data    AuthSetupDefaultsEntity `json:"data"`
}

// Get GETs /app-srv/apps/auth-setup-defaults (admin-only).
func (s *AuthSetupDefaultsService) Get(ctx context.Context) (*AuthSetupDefaultsResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "app-srv/apps/auth-setup-defaults")
	if err != nil {
		return nil, err
	}
	var out AuthSetupDefaultsResponse
	if err := requestJSON(ctx, s.cfg, http.MethodGet, endpoint, nil, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update PUTs /app-srv/apps/auth-setup-defaults (full-object replace of editable fields).
func (s *AuthSetupDefaultsService) Update(ctx context.Context, entity AuthSetupDefaultsEntity) (*AuthSetupDefaultsResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "app-srv/apps/auth-setup-defaults")
	if err != nil {
		return nil, err
	}
	var out AuthSetupDefaultsResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPut, endpoint, entity, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}
