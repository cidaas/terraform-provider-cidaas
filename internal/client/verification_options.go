package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// VerificationOptionsService calls verification-actions-srv verification-options APIs.
type VerificationOptionsService struct {
	cfg Config
}

// NewVerificationOptionsService returns a VerificationOptionsService bound to cfg.
func NewVerificationOptionsService(cfg Config) *VerificationOptionsService {
	return &VerificationOptionsService{cfg: cfg}
}

// VerificationOptionsModel matches VerificationOptionsEntity JSON.
type VerificationOptionsModel struct {
	ID                  string               `json:"id,omitempty"`
	Name                string               `json:"name"`
	Description         string               `json:"description,omitempty"`
	Owner               string               `json:"owner,omitempty"`
	VerificationOptions *VerificationOptions `json:"verification_options,omitempty"`
	CreatedTime         string               `json:"createdTime,omitempty"`
	UpdatedTime         string               `json:"updatedTime,omitempty"`
}

// VerificationOptions is the nested verification_options payload.
type VerificationOptions struct {
	Setting                     string          `json:"setting,omitempty"`
	TimeIntervalInSeconds       *int            `json:"time_interval_in_seconds,omitempty"`
	AllowedMethods              []string        `json:"allowed_methods"`
	PasswordPolicyRef           string          `json:"password_policy_ref,omitempty"`
	UseDefaultPasswordPolicy    bool            `json:"use_default_password_policy"`
	SuggestVerificationMethodID string          `json:"suggest_verification_method_id,omitempty"`
	AppAttest                   *AppAttestEntry `json:"app_attest,omitempty"`
}

// AppAttestEntry holds optional Android/iOS app attestation config.
type AppAttestEntry struct {
	Android *AppAttestAndroidEntry `json:"android,omitempty"`
	IOS     *AppAttestIOSEntry     `json:"ios,omitempty"`
}

// AppAttestAndroidEntry is Android app_attest settings.
type AppAttestAndroidEntry struct {
	Provider            string   `json:"provider"`
	RelaxAppRecognition bool     `json:"relax_app_recognition,omitempty"`
	CertDigests         []string `json:"cert_digests,omitempty"`
	GCPServiceAccount   string   `json:"gcp_service_account,omitempty"`
	ProjectID           string   `json:"project_id,omitempty"`
	IOSAppID            string   `json:"ios_app_id,omitempty"`
	AndroidAppID        string   `json:"android_app_id,omitempty"`
}

// AppAttestIOSEntry is iOS app_attest settings.
type AppAttestIOSEntry struct {
	Provider          string `json:"provider"`
	AppleRootCert     string `json:"apple_root_cert,omitempty"`
	GCPServiceAccount string `json:"gcp_service_account,omitempty"`
	ProjectID         string `json:"project_id,omitempty"`
	IOSAppID          string `json:"ios_app_id,omitempty"`
	AndroidAppID      string `json:"android_app_id,omitempty"`
}

// VerificationOptionsResponse is the standard {success,status,data} envelope.
type VerificationOptionsResponse struct {
	Success bool                     `json:"success"`
	Status  int                      `json:"status"`
	Data    VerificationOptionsModel `json:"data"`
}

// Create POSTs /verification-actions-srv/verification-options/.
func (s *VerificationOptionsService) Create(ctx context.Context, model VerificationOptionsModel) (*VerificationOptionsResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "verification-actions-srv/verification-options")
	if err != nil {
		return nil, err
	}
	endpoint += "/"
	var out VerificationOptionsResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPost, endpoint, model, &out, http.StatusCreated, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get GETs /verification-actions-srv/verification-options/{id}.
func (s *VerificationOptionsService) Get(ctx context.Context, id string) (*VerificationOptionsResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "verification-actions-srv/verification-options", id)
	if err != nil {
		return nil, err
	}
	var out VerificationOptionsResponse
	if err := requestJSON(ctx, s.cfg, http.MethodGet, endpoint, nil, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update uses PUT — what verification-actions-srv exposes.
func (s *VerificationOptionsService) Update(ctx context.Context, id string, model VerificationOptionsModel) (*VerificationOptionsResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "verification-actions-srv/verification-options", id)
	if err != nil {
		return nil, err
	}
	var out VerificationOptionsResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPut, endpoint, model, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete DELETEs /verification-actions-srv/verification-options/{id}.
func (s *VerificationOptionsService) Delete(ctx context.Context, id string) error {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "verification-actions-srv/verification-options", id)
	if err != nil {
		return err
	}
	if err := requestJSON(ctx, s.cfg, http.MethodDelete, endpoint, nil, nil, http.StatusNoContent, http.StatusOK); err != nil {
		return fmt.Errorf("delete verification options: %w", err)
	}
	return nil
}
