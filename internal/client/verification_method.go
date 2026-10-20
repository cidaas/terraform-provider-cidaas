package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// VerificationMethodOwnerClient marks pluggable (customer-defined) verification methods. Every other owner
// (e.g. "system") is a built-in method that can only be enabled/disabled, never created or deleted.
const VerificationMethodOwnerClient = "CLIENT"

// VerificationMethodService calls verification-actions-srv /config (Admin → Multifactor Settings → MFA Settings).
type VerificationMethodService struct {
	cfg Config
}

// NewVerificationMethodService returns a VerificationMethodService bound to cfg.
func NewVerificationMethodService(cfg Config) *VerificationMethodService {
	return &VerificationMethodService{cfg: cfg}
}

// VerificationMethodModel matches verification-go-lib VerificationSettings (plus name/description on reads).
type VerificationMethodModel struct {
	ID                       string                        `json:"_id,omitempty"`
	VerificationType         string                        `json:"verificationType,omitempty"`
	Active                   *bool                         `json:"active,omitempty"`
	Owner                    string                        `json:"owner,omitempty"`
	LocaleText               []VerificationMethodLocale    `json:"localeText,omitempty"`
	Icon                     string                        `json:"icon,omitempty"`
	URL                      string                        `json:"url,omitempty"`
	AuthType                 string                        `json:"auth_type,omitempty"`
	Prerequisite             *string                       `json:"prerequisite,omitempty"`
	CidaasAuthDetails        *VerificationMethodAuth       `json:"cidaasAuthDetails,omitempty"`
	AlgorithmTypeID          string                        `json:"algorithmTypeId,omitempty"`
	Category                 string                        `json:"category,omitempty"`
	Code                     *VerificationMethodCode       `json:"code,omitempty"`
	Factor                   string                        `json:"factor,omitempty"`
	AMR                      []string                      `json:"amr_values,omitempty"`
	ACR                      string                        `json:"acr_pin,omitempty"`
	RequiresAuthenticatorApp *bool                         `json:"requiresAuthenticatorApp,omitempty"`
	EncryptionConfig         *VerificationMethodEncryption `json:"encryptionConfig,omitempty"`
	Name                     string                        `json:"name,omitempty"`
	Description              string                        `json:"description,omitempty"`
	CreatedTime              string                        `json:"createdTime,omitempty"`
	UpdatedTime              string                        `json:"updatedTime,omitempty"`
}

// VerificationMethodLocale is one localeText entry (display name and description per locale).
type VerificationMethodLocale struct {
	Locale      string `json:"locale,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// VerificationMethodAuth is cidaasAuthDetails: the cidaas client used to call a pluggable method's URL.
type VerificationMethodAuth struct {
	ClientID string `json:"client_id,omitempty"`
}

// VerificationMethodCode holds code settings of code-based methods (read-only).
type VerificationMethodCode struct {
	Length   int64 `json:"length,omitempty"`
	Validity int64 `json:"validity,omitempty"`
}

// VerificationMethodEncryption is encryptionConfig. Only Enabled and Mode can be updated.
type VerificationMethodEncryption struct {
	Enabled         bool     `json:"enabled"`
	Mode            string   `json:"mode"`
	EncryptedFields []string `json:"encryptedFields"`
}

// IsPluggable reports whether the method is customer-defined (owner CLIENT), mirroring verification-actions-srv.
func (m *VerificationMethodModel) IsPluggable() bool {
	return strings.EqualFold(m.Owner, VerificationMethodOwnerClient)
}

type verificationMethodListResponse struct {
	Success bool                      `json:"success"`
	Status  int                       `json:"status"`
	Data    []VerificationMethodModel `json:"data"`
}

type verificationMethodResponse struct {
	Success bool                    `json:"success"`
	Status  int                     `json:"status"`
	Data    VerificationMethodModel `json:"data"`
}

func (s *VerificationMethodService) endpoint(elem ...string) (string, error) {
	return url.JoinPath(s.cfg.BaseURL, append([]string{"verification-actions-srv/config"}, elem...)...)
}

// List GETs /verification-actions-srv/config (all system and pluggable methods).
// The service answers 204 without a body when it fails to read the settings, so that is an error here.
func (s *VerificationMethodService) List(ctx context.Context) ([]VerificationMethodModel, error) {
	endpoint, err := s.endpoint()
	if err != nil {
		return nil, err
	}
	endpoint += "/"
	httpClient := NewHTTPClient(endpoint, http.MethodGet, s.cfg.AccessToken)
	resp, err := httpClient.DoJSON(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNoContent {
		return nil, errors.New("list verification methods: verification-actions-srv returned 204 No Content (settings could not be read)")
	}
	if err := ExpectStatus(resp, http.StatusOK); err != nil {
		return nil, fmt.Errorf("list verification methods: %w", err)
	}
	var out verificationMethodListResponse
	if err := DecodeJSON(resp, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// Get returns one method by verificationType. There is no single-item endpoint, so it filters List.
func (s *VerificationMethodService) Get(ctx context.Context, verificationType string) (*VerificationMethodModel, error) {
	methods, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range methods {
		if methods[i].VerificationType == verificationType {
			return &methods[i], nil
		}
	}
	return nil, fmt.Errorf("%w: verification method %q", ErrNotFound, verificationType)
}

// Create POSTs /verification-actions-srv/config. The service only creates pluggable methods (owner CLIENT).
func (s *VerificationMethodService) Create(ctx context.Context, model VerificationMethodModel) (*VerificationMethodModel, error) {
	endpoint, err := s.endpoint()
	if err != nil {
		return nil, err
	}
	endpoint += "/"
	var out verificationMethodResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPost, endpoint, model, &out, http.StatusCreated, http.StatusOK); err != nil {
		return nil, fmt.Errorf("create verification method: %w", err)
	}
	return &out.Data, nil
}

// Update PUTs /verification-actions-srv/config. localeText is always replaced by the service, so callers
// must send the full desired list.
func (s *VerificationMethodService) Update(ctx context.Context, model VerificationMethodModel) (*VerificationMethodModel, error) {
	endpoint, err := s.endpoint()
	if err != nil {
		return nil, err
	}
	endpoint += "/"
	var out verificationMethodResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPut, endpoint, model, &out, http.StatusOK); err != nil {
		if isVerificationMethodMissing(err) {
			return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
		}
		return nil, fmt.Errorf("update verification method: %w", err)
	}
	return &out.Data, nil
}

// Delete DELETEs /verification-actions-srv/config/{verificationType}. Only pluggable methods can be deleted.
func (s *VerificationMethodService) Delete(ctx context.Context, verificationType string) error {
	endpoint, err := s.endpoint(verificationType)
	if err != nil {
		return err
	}
	if err := requestJSON(ctx, s.cfg, http.MethodDelete, endpoint, nil, nil, http.StatusOK, http.StatusNoContent); err != nil {
		if isVerificationMethodMissing(err) {
			return fmt.Errorf("%w: %v", ErrNotFound, err)
		}
		return fmt.Errorf("delete verification method: %w", err)
	}
	return nil
}

// isVerificationMethodMissing maps the service's 400 "verification method not found" to not-found.
func isVerificationMethodMissing(err error) bool {
	if err == nil || errors.Is(err, ErrNotFound) {
		return false
	}
	return strings.Contains(err.Error(), "verification method not found")
}
