package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// TranslationsService calls hostedpages-srv translations APIs.
type TranslationsService struct {
	cfg Config
}

// NewTranslationsService returns a TranslationsService bound to cfg.
func NewTranslationsService(cfg Config) *TranslationsService {
	return &TranslationsService{cfg: cfg}
}

// TranslationModel matches hostedpages-srv TranslationDetails JSON.
type TranslationModel struct {
	Locale      string         `json:"locale"`
	Enabled     bool           `json:"enabled"`
	Bundled     bool           `json:"bundled,omitempty"`
	Translation map[string]any `json:"translation"`
}

// TranslationResponse is the standard {success,status,data} envelope.
type TranslationResponse struct {
	Success bool             `json:"success"`
	Status  int              `json:"status"`
	Data    TranslationModel `json:"data"`
}

// Create POSTs /hostedpages-srv/translations.
func (s *TranslationsService) Create(ctx context.Context, model TranslationModel) (*TranslationResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/translations")
	if err != nil {
		return nil, err
	}
	var out TranslationResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPost, endpoint, model, &out, http.StatusCreated, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get GETs /hostedpages-srv/translations/{locale}.
func (s *TranslationsService) Get(ctx context.Context, localeID string) (*TranslationResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/translations", localeID)
	if err != nil {
		return nil, err
	}
	var out TranslationResponse
	if err := requestJSON(ctx, s.cfg, http.MethodGet, endpoint, nil, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update PUTs /hostedpages-srv/translations/{locale}.
func (s *TranslationsService) Update(ctx context.Context, localeID string, model TranslationModel) (*TranslationResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/translations", localeID)
	if err != nil {
		return nil, err
	}
	if model.Locale == "" {
		model.Locale = localeID
	}
	var out TranslationResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPut, endpoint, model, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete DELETEs /hostedpages-srv/translations/{locale}.
func (s *TranslationsService) Delete(ctx context.Context, localeID string) error {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/translations", localeID)
	if err != nil {
		return err
	}
	if err := requestJSON(ctx, s.cfg, http.MethodDelete, endpoint, nil, nil, http.StatusOK, http.StatusNoContent); err != nil {
		return fmt.Errorf("delete translation: %w", err)
	}
	return nil
}
