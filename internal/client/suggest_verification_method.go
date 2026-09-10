package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// SuggestVerificationMethodService calls verification-actions-srv suggest-verification-configs APIs.
type SuggestVerificationMethodService struct {
	cfg Config
}

// NewSuggestVerificationMethodService returns a SuggestVerificationMethodService bound to cfg.
func NewSuggestVerificationMethodService(cfg Config) *SuggestVerificationMethodService {
	return &SuggestVerificationMethodService{cfg: cfg}
}

// SuggestVerificationMethodModel matches SuggestVerificationMethodEntity JSON.
type SuggestVerificationMethodModel struct {
	ID                        string                          `json:"id,omitempty"`
	Name                      string                          `json:"name"`
	Description               string                          `json:"description,omitempty"`
	Owner                     string                          `json:"owner,omitempty"`
	SuggestVerificationMethod *SuggestVerificationMethodSetup `json:"suggest_verification_method,omitempty"`
	CreatedTime               string                          `json:"createdTime,omitempty"`
	UpdatedTime               string                          `json:"updatedTime,omitempty"`
}

// SuggestVerificationMethodSetup is the nested suggest_verification_method payload.
type SuggestVerificationMethodSetup struct {
	MandatoryConfig    *SuggestVerificationMandatoryConfig `json:"mandatoryConfig,omitempty"`
	OptionalConfig     *SuggestVerificationOptionalConfig  `json:"optionalConfig,omitempty"`
	SkipDurationInDays int                                 `json:"skipDurationInDays"`
}

// SuggestVerificationMandatoryConfig is mandatoryConfig (methods + range).
type SuggestVerificationMandatoryConfig struct {
	Methods   []string `json:"methods"`
	Range     string   `json:"range"`
	SkipUntil string   `json:"skipUntil,omitempty"`
}

// SuggestVerificationOptionalConfig is optionalConfig (methods only).
type SuggestVerificationOptionalConfig struct {
	Methods []string `json:"methods"`
}

// SuggestVerificationMethodResponse is the standard {success,status,data} envelope.
type SuggestVerificationMethodResponse struct {
	Success bool                           `json:"success"`
	Status  int                            `json:"status"`
	Data    SuggestVerificationMethodModel `json:"data"`
}

// Create POSTs /verification-actions-srv/suggest-verification-configs/.
func (s *SuggestVerificationMethodService) Create(ctx context.Context, model SuggestVerificationMethodModel) (*SuggestVerificationMethodResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "verification-actions-srv/suggest-verification-configs")
	if err != nil {
		return nil, err
	}
	endpoint += "/"
	var out SuggestVerificationMethodResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPost, endpoint, model, &out, http.StatusCreated, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get GETs /verification-actions-srv/suggest-verification-configs/{id}.
func (s *SuggestVerificationMethodService) Get(ctx context.Context, id string) (*SuggestVerificationMethodResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "verification-actions-srv/suggest-verification-configs", id)
	if err != nil {
		return nil, err
	}
	var out SuggestVerificationMethodResponse
	if err := requestJSON(ctx, s.cfg, http.MethodGet, endpoint, nil, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update uses PUT — what verification-actions-srv exposes.
func (s *SuggestVerificationMethodService) Update(ctx context.Context, id string, model SuggestVerificationMethodModel) (*SuggestVerificationMethodResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "verification-actions-srv/suggest-verification-configs", id)
	if err != nil {
		return nil, err
	}
	var out SuggestVerificationMethodResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPut, endpoint, model, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete DELETEs /verification-actions-srv/suggest-verification-configs/{id}.
func (s *SuggestVerificationMethodService) Delete(ctx context.Context, id string) error {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "verification-actions-srv/suggest-verification-configs", id)
	if err != nil {
		return err
	}
	if err := requestJSON(ctx, s.cfg, http.MethodDelete, endpoint, nil, nil, http.StatusNoContent, http.StatusOK); err != nil {
		return fmt.Errorf("delete suggest verification method: %w", err)
	}
	return nil
}
