package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type HostedPageGroupService struct {
	cfg Config
}

func NewHostedPageGroupService(cfg Config) *HostedPageGroupService {
	return &HostedPageGroupService{cfg: cfg}
}

type HostedPageData struct {
	HostedPageID string `json:"hosted_page_id"`
	Locale       string `json:"locale"`
	URL          string `json:"url"`
	Content      string `json:"content,omitempty"`
}

type HostedPageGroupModel struct {
	ID            string           `json:"_id,omitempty"`
	GroupOwner    string           `json:"groupOwner,omitempty"`
	DefaultLocale string           `json:"default_locale,omitempty"`
	HostedPages   []HostedPageData `json:"hosted_pages,omitempty"`
	CreatedTime   string           `json:"createdTime,omitempty"`
	UpdatedTime   string           `json:"updatedTime,omitempty"`
}

type HostedPageGroupResponse struct {
	Success bool                 `json:"success"`
	Status  int                  `json:"status"`
	Data    HostedPageGroupModel `json:"data"`
}

func (s *HostedPageGroupService) Upsert(ctx context.Context, model HostedPageGroupModel) (*HostedPageGroupResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/hpgroup")
	if err != nil {
		return nil, err
	}
	var out HostedPageGroupResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPost, endpoint, model, &out, http.StatusOK, http.StatusCreated); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *HostedPageGroupService) Get(ctx context.Context, id string) (*HostedPageGroupResponse, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/hpgroup", id)
	if err != nil {
		return nil, err
	}
	var out HostedPageGroupResponse
	if err := requestJSON(ctx, s.cfg, http.MethodGet, endpoint, nil, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *HostedPageGroupService) Delete(ctx context.Context, id string) error {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/hpgroup", id)
	if err != nil {
		return err
	}
	if err := requestJSON(ctx, s.cfg, http.MethodDelete, endpoint, nil, nil, http.StatusOK, http.StatusNoContent); err != nil {
		return fmt.Errorf("delete hosted page group: %w", err)
	}
	return nil
}
