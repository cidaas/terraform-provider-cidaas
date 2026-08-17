package cidaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type ConsentVersionResponse struct {
	Success bool                `json:"success,omitempty"`
	Status  int                 `json:"status,omitempty"`
	Data    ConsentVersionModel `json:"data,omitempty"`
}

type ConsentVersionReadResponse struct {
	Success bool                  `json:"success,omitempty"`
	Status  int                   `json:"status,omitempty"`
	Data    []ConsentVersionModel `json:"data,omitempty"`
}

// consentVersionScopeWire matches consent-management-srv v2 responses where scopes are
// []DetailedFieldsByScope (scope key only is needed for Terraform state).
type consentVersionScopeWire struct {
	Scope string `json:"scope"`
}

type ConsentVersionModel struct {
	ID             string        `json:"_id,omitempty"`
	Version        float64       `json:"version,omitempty"`
	ConsentID      string        `json:"consent_id,omitempty"`
	ConsentType    string        `json:"consentType,omitempty"`
	Scopes         []string      `json:"scopes,omitempty"`
	RequiredFields []string      `json:"required_fields,omitempty"`
	ConsentLocale  ConsentLocale `json:"consent_locale,omitempty"`
	CreatedAt      string        `json:"created_at,omitempty"`
	UpdatedAt      string        `json:"updated_at,omitempty"`
}

// UnmarshalJSON accepts scopes as either []string (request/legacy) or []object with a
// "scope" field (consent-management-srv POST .../v2/consent/versions response).
func (cv *ConsentVersionModel) UnmarshalJSON(data []byte) error {
	type consentVersionModelAlias ConsentVersionModel
	aux := struct {
		Scopes json.RawMessage `json:"scopes,omitempty"`
		consentVersionModelAlias
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*cv = ConsentVersionModel(aux.consentVersionModelAlias)
	if len(aux.Scopes) == 0 || string(aux.Scopes) == "null" {
		return nil
	}
	var scopes []string
	if err := json.Unmarshal(aux.Scopes, &scopes); err == nil {
		if len(scopes) == 0 {
			cv.Scopes = nil
		} else {
			cv.Scopes = scopes
		}
		return nil
	}
	var scopeObjects []consentVersionScopeWire
	if err := json.Unmarshal(aux.Scopes, &scopeObjects); err != nil {
		return fmt.Errorf("decode consent version scopes: %w", err)
	}
	cv.Scopes = make([]string, 0, len(scopeObjects))
	for _, s := range scopeObjects {
		if s.Scope != "" {
			cv.Scopes = append(cv.Scopes, s.Scope)
		}
	}
	if len(cv.Scopes) == 0 {
		cv.Scopes = nil
	}
	return nil
}

type ConsentLocalResponse struct {
	Success bool              `json:"success,omitempty"`
	Status  int               `json:"status,omitempty"`
	Data    ConsentLocalModel `json:"data,omitempty"`
}

type ConsentLocalModel struct {
	ConsentVersionID string   `json:"consent_version_id,omitempty"`
	ConsentID        string   `json:"consent_id,omitempty"`
	Content          string   `json:"content,omitempty"`
	Locale           string   `json:"locale,omitempty"`
	URL              string   `json:"url,omitempty"`
	Scopes           []string `json:"scopes,omitempty"`
	RequiredFields   []string `json:"required_fields,omitempty"`
}
type ConsentLocale struct {
	Locale  string `json:"locale,omitempty"`
	Content string `json:"content,omitempty"`
	URL     string `json:"url,omitempty"`
}

type ConsentVersion struct {
	ClientConfig
}

func NewConsentVersion(clientConfig ClientConfig) *ConsentVersion {
	return &ConsentVersion{clientConfig}
}

// ponytail: 5 attempts, 1+2+4+8s sleeps (~15s). Consent-management returns 400/30001 until the parent consent is indexed. Raise maxAttempts if that lag grows.
const consentVersionUpsertMaxAttempts = 5

var consentVersionRetryDelay = func(attempt int) time.Duration {
	return time.Duration(1<<attempt) * time.Second
}

func isConsentVersionNotIndexed(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "status code 400") {
		return false
	}
	return strings.Contains(msg, "30001") || strings.Contains(msg, "consent version not found")
}

func (c *ConsentVersion) Upsert(ctx context.Context, consentVersionConfig ConsentVersionModel) (*ConsentVersionResponse, error) {
	var lastErr error
	for attempt := 0; attempt < consentVersionUpsertMaxAttempts; attempt++ {
		res, err := c.upsertOnce(ctx, consentVersionConfig)
		if err == nil {
			return res, nil
		}
		lastErr = err
		if !isConsentVersionNotIndexed(err) || attempt == consentVersionUpsertMaxAttempts-1 {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(consentVersionRetryDelay(attempt)):
		}
	}
	return nil, lastErr
}

func (c *ConsentVersion) upsertOnce(ctx context.Context, consentVersionConfig ConsentVersionModel) (*ConsentVersionResponse, error) { //nolint:dupl
	var response ConsentVersionResponse
	url := fmt.Sprintf("%s/%s", c.BaseURL, "consent-management-srv/v2/consent/versions")
	client, err := util.NewHTTPClient(url, http.MethodPost, c.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, consentVersionConfig)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *ConsentVersion) Get(ctx context.Context, consentID string) (*ConsentVersionReadResponse, error) { //nolint:dupl
	var response ConsentVersionReadResponse
	url := fmt.Sprintf("%s/%s/%s", c.BaseURL, "consent-management-srv/v2/consent/versions/list", consentID)
	client, err := util.NewHTTPClient(url, http.MethodGet, c.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, nil)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *ConsentVersion) UpsertLocal(ctx context.Context, consentLocal ConsentLocalModel) (*ConsentLocalResponse, error) { //nolint:dupl
	var response ConsentLocalResponse
	url := fmt.Sprintf("%s/%s", c.BaseURL, "consent-management-srv/v2/consent/locale")
	client, err := util.NewHTTPClient(url, http.MethodPost, c.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, consentLocal)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *ConsentVersion) GetLocal(ctx context.Context, consentVersionID string, locale string) (*ConsentLocalResponse, error) {
	var response ConsentLocalResponse
	url := fmt.Sprintf("%s/%s/%s?locale=%s", c.BaseURL, "consent-management-srv/v2/consent/locale", consentVersionID, locale)
	client, err := util.NewHTTPClient(url, http.MethodGet, c.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, nil)
	if res.StatusCode == http.StatusNoContent {
		return &ConsentLocalResponse{
			Success: false,
			Status:  http.StatusNoContent,
		}, nil
	}
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}
