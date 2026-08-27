package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type UserSetupService struct {
	cfg Config
}

func NewUserSetupService(cfg Config) *UserSetupService {
	return &UserSetupService{cfg: cfg}
}

// UserAppSetupModel matches user-srv UserAppSetup JSON.
type UserAppSetupModel struct {
	ID          string          `json:"id,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	UserSetup   UserSetupDetail `json:"user_setup"`
	CreatedTime string          `json:"createdTime,omitempty"`
	UpdatedTime string          `json:"updatedTime,omitempty"`
}

// UserSetupDetail maps user-srv UserSetup JSON (consent_refs nested here).
type UserSetupDetail struct {
	AllowedFields                   []string       `json:"allowed_fields,omitempty"`
	RequiredFields                  []string       `json:"required_fields,omitempty"`
	AllowLoginWith                  []string       `json:"allow_login_with,omitempty"`
	ConsentRefs                     []string       `json:"consent_refs,omitempty"`
	EnableDeduplication             *bool          `json:"enable_deduplication,omitempty"`
	ValidatePhoneNumber             *bool          `json:"validate_phone_number,omitempty"`
	ValidateEmail                   *bool          `json:"validate_email,omitempty"`
	AutoActivateUser                *bool          `json:"auto_activate_user,omitempty"`
	AllowDisposableEmail            *bool          `json:"allow_disposable_email,omitempty"`
	AcceptRolesInRegistration       *bool          `json:"accept_roles_in_the_registration,omitempty"`
	SendWelcomeNotification         *bool          `json:"send_welcome_notification,omitempty"`
	BirthdateAsDate                 *bool          `json:"birthdate_as_date,omitempty"`
	CommunicationMediumVerification string         `json:"communication_medium_verification,omitempty"`
	AutoConfirmCommunicationMethod  []string       `json:"auto_confirm_communication_method,omitempty"`
	VerificationForMedium           []string       `json:"verification_for_medium,omitempty"`
	OperationsAllowedGroups         []AllowedGroup `json:"operations_allowed_groups,omitempty"`
}

// AllowedGroup maps user-srv groupId / roles / default_roles.
type AllowedGroup struct {
	GroupID      string   `json:"groupId"`
	Roles        []string `json:"roles,omitempty"`
	DefaultRoles []string `json:"default_roles,omitempty"`
}

type UserAppSetupResponse struct {
	Success bool              `json:"success"`
	Status  int               `json:"status"`
	Data    UserAppSetupModel `json:"data"`
}

func (s *UserSetupService) Create(ctx context.Context, model UserAppSetupModel) (*UserAppSetupResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "user-srv/usersetup")
	if err != nil {
		return nil, err
	}
	var out UserAppSetupResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPost, endpoint, model, &out, http.StatusCreated, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *UserSetupService) Get(ctx context.Context, id string) (*UserAppSetupResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "user-srv/usersetup", id)
	if err != nil {
		return nil, err
	}
	var out UserAppSetupResponse
	if err := requestJSON(ctx, s.cfg, http.MethodGet, endpoint, nil, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update PATCHes /user-srv/usersetup/{id} (user-srv exposes PATCH, not PUT).
func (s *UserSetupService) Update(ctx context.Context, id string, model UserAppSetupModel) (*UserAppSetupResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "user-srv/usersetup", id)
	if err != nil {
		return nil, err
	}
	var out UserAppSetupResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPatch, endpoint, model, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *UserSetupService) Delete(ctx context.Context, id string) error {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "user-srv/usersetup", id)
	if err != nil {
		return err
	}
	if err := requestJSON(ctx, s.cfg, http.MethodDelete, endpoint, nil, nil, http.StatusNoContent, http.StatusOK); err != nil {
		return fmt.Errorf("delete user setup: %w", err)
	}
	return nil
}
