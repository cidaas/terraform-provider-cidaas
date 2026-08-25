package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type TranslationsService struct {
	cfg Config
}

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

type TranslationResponse struct {
	Success bool             `json:"success"`
	Status  int              `json:"status"`
	Data    TranslationModel `json:"data"`
}

func (s *TranslationsService) Create(ctx context.Context, model TranslationModel) (*TranslationResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/translations")
	if err != nil {
		return nil, err
	}
	httpClient := NewHTTPClient(endpoint, http.MethodPost, s.cfg.AccessToken)
	resp, err := httpClient.DoJSON(ctx, model)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := ExpectStatus(resp, http.StatusCreated, http.StatusOK); err != nil {
		return nil, err
	}
	var out TranslationResponse
	if err := DecodeJSON(resp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *TranslationsService) Get(ctx context.Context, localeID string) (*TranslationResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/translations", localeID)
	if err != nil {
		return nil, err
	}
	httpClient := NewHTTPClient(endpoint, http.MethodGet, s.cfg.AccessToken)
	resp, err := httpClient.DoJSON(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := ExpectStatus(resp, http.StatusOK); err != nil {
		return nil, err
	}
	var out TranslationResponse
	if err := DecodeJSON(resp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *TranslationsService) Update(ctx context.Context, localeID string, model TranslationModel) (*TranslationResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/translations", localeID)
	if err != nil {
		return nil, err
	}
	if model.Locale == "" {
		model.Locale = localeID
	}
	httpClient := NewHTTPClient(endpoint, http.MethodPut, s.cfg.AccessToken)
	resp, err := httpClient.DoJSON(ctx, model)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := ExpectStatus(resp, http.StatusOK); err != nil {
		return nil, err
	}
	var out TranslationResponse
	if err := DecodeJSON(resp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *TranslationsService) Delete(ctx context.Context, localeID string) error {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/translations", localeID)
	if err != nil {
		return err
	}
	httpClient := NewHTTPClient(endpoint, http.MethodDelete, s.cfg.AccessToken)
	resp, err := httpClient.DoJSON(ctx, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := ExpectStatus(resp, http.StatusOK, http.StatusNoContent); err != nil {
		return fmt.Errorf("delete translation: %w", err)
	}
	return nil
}
