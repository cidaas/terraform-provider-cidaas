package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// HostedPageLayoutService calls hostedpages-srv hosted-page-layouts APIs.
type HostedPageLayoutService struct {
	cfg Config
}

// NewHostedPageLayoutService returns a HostedPageLayoutService bound to cfg.
func NewHostedPageLayoutService(cfg Config) *HostedPageLayoutService {
	return &HostedPageLayoutService{cfg: cfg}
}

// LayoutDetail matches hostedpages-srv HostedPageLayout JSON.
type LayoutDetail struct {
	Theme           string `json:"theme,omitempty"`
	LogoURI         string `json:"logo_uri,omitempty"`
	PolicyURI       string `json:"policy_uri,omitempty"`
	TosURI          string `json:"tos_uri,omitempty"`
	ImprintURI      string `json:"imprint_uri,omitempty"`
	HostedPageGroup string `json:"hosted_page_group,omitempty"`
	PrimaryColor    string `json:"primary_color,omitempty"`
	AccentColor     string `json:"accent_color,omitempty"`
	BackgroundURI   string `json:"background_uri,omitempty"`
	ContentAlign    string `json:"content_align,omitempty"`
	LogoAlign       string `json:"logo_align,omitempty"`
	MediaType       string `json:"media_type,omitempty"`
	VideoURL        string `json:"video_url,omitempty"`
	FavIcon         string `json:"fav_icon,omitempty"`
}

// LayoutResourceConfig matches hostedpages-srv per-webapp overrides.
type LayoutResourceConfig struct {
	TranslationSet string `json:"translationSet"`
	Theme          string `json:"theme,omitempty"`
	Layout         string `json:"layout,omitempty"`
}

// HostedPageLayoutModel is the layout entity returned by hostedpages-srv.
type HostedPageLayoutModel struct {
	ID          string                          `json:"id,omitempty"`
	Fingerprint string                          `json:"fingerprint,omitempty"`
	Description string                          `json:"description"`
	Layout      LayoutDetail                    `json:"layout"`
	GroupID     string                          `json:"groupId,omitempty"`
	Owner       string                          `json:"owner,omitempty"`
	Resources   map[string]LayoutResourceConfig `json:"resources,omitempty"`
	CreatedTime string                          `json:"createdTime,omitempty"`
	UpdatedTime string                          `json:"updatedTime,omitempty"`
}

// HostedPageLayoutWrite is the POST/PUT body (groupId comes from token, not HCL).
type HostedPageLayoutWrite struct {
	Description string                          `json:"description"`
	Layout      LayoutDetail                    `json:"layout"`
	Resources   map[string]LayoutResourceConfig `json:"resources,omitempty"`
}

// HostedPageLayoutResponse is the standard {success,status,data} envelope.
type HostedPageLayoutResponse struct {
	Success bool                  `json:"success"`
	Status  int                   `json:"status"`
	Data    HostedPageLayoutModel `json:"data"`
}

// Create POSTs /hostedpages-srv/hosted-page-layouts.
func (s *HostedPageLayoutService) Create(ctx context.Context, model HostedPageLayoutWrite) (*HostedPageLayoutResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/hosted-page-layouts")
	if err != nil {
		return nil, err
	}
	var out HostedPageLayoutResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPost, endpoint, model, &out, http.StatusCreated); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get GETs /hostedpages-srv/hosted-page-layouts/{id}.
func (s *HostedPageLayoutService) Get(ctx context.Context, id string) (*HostedPageLayoutResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/hosted-page-layouts", id)
	if err != nil {
		return nil, err
	}
	var out HostedPageLayoutResponse
	if err := requestJSON(ctx, s.cfg, http.MethodGet, endpoint, nil, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update PUTs /hostedpages-srv/hosted-page-layouts/{id}.
func (s *HostedPageLayoutService) Update(ctx context.Context, id string, model HostedPageLayoutWrite) (*HostedPageLayoutResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/hosted-page-layouts", id)
	if err != nil {
		return nil, err
	}
	var out HostedPageLayoutResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPut, endpoint, model, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete DELETEs /hostedpages-srv/hosted-page-layouts/{id}.
func (s *HostedPageLayoutService) Delete(ctx context.Context, id string) error {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/hosted-page-layouts", id)
	if err != nil {
		return err
	}
	if err := requestJSON(ctx, s.cfg, http.MethodDelete, endpoint, nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("delete hosted page layout: %w", err)
	}
	return nil
}
