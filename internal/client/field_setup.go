package client

import (
	"context"
	"net/http"
	"net/url"
)

type FieldSetupService struct {
	cfg Config
}

func NewFieldSetupService(cfg Config) *FieldSetupService {
	return &FieldSetupService{cfg: cfg}
}

// FieldSetupEntry is the minimal fieldsetup-srv field row used for user setup validation.
type FieldSetupEntry struct {
	FieldKey string `json:"fieldKey,omitempty"`
	Enabled  bool   `json:"enabled"`
}

type fieldSetupListResponse struct {
	Success bool              `json:"success"`
	Status  int               `json:"status"`
	Data    []FieldSetupEntry `json:"data"`
}

// List returns all field setups via POST /fieldsetup-srv/graph/fields.
func (s *FieldSetupService) List(ctx context.Context) ([]FieldSetupEntry, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "fieldsetup-srv/graph/fields")
	if err != nil {
		return nil, err
	}
	var out fieldSetupListResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPost, endpoint, map[string]any{}, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return out.Data, nil
}
