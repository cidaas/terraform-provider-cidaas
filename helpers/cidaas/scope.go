package cidaas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type ScopeModel struct {
	ID                    string                  `json:"_id,omitempty"`
	LocaleWiseDescription []ScopeLocalDescription `json:"localeWiseDescription,omitempty"`
	SecurityLevel         string                  `json:"securityLevel,omitempty"`
	ScopeKey              string                  `json:"scopeKey,omitempty"`
	RequiredUserConsent   bool                    `json:"requiredUserConsent"`
	GroupName             []string                `json:"group_name,omitempty"`
	ScopeOwner            string                  `json:"scopeOwner,omitempty"`
}

type ScopeLocalDescription struct {
	Locale      string `json:"locale,omitempty"`
	Language    string `json:"language,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

type AllScopeResp struct {
	Success bool         `json:"success,omitempty"`
	Status  int          `json:"status,omitempty"`
	Data    []ScopeModel `json:"data,omitempty"`
}

type ScopeResponse struct {
	Success bool       `json:"success,omitempty"`
	Status  int        `json:"status,omitempty"`
	Data    ScopeModel `json:"data,omitempty"`
	Error   string     `json:"error,omitempty"`
}

type Scope struct {
	ClientConfig
}

func NewScope(clientConfig ClientConfig) *Scope {
	return &Scope{clientConfig}
}

func (c *Scope) Upsert(ctx context.Context, sc ScopeModel) (*ScopeResponse, error) { //nolint:dupl
	var response ScopeResponse
	url := fmt.Sprintf("%s/%s", c.BaseURL, "scopes-srv/scope")
	client, err := util.NewHTTPClient(url, http.MethodPost, c.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, sc)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Scope) Get(ctx context.Context, scopeKey string) (*ScopeResponse, error) { //nolint:dupl
	var response ScopeResponse
	url := fmt.Sprintf("%s/%s?scopekey=%s", c.BaseURL, "scopes-srv/scope", scopeKey)
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

func (c *Scope) Delete(ctx context.Context, scopeKey string) error {
	url := fmt.Sprintf("%s/%s/%s", c.BaseURL, "scopes-srv/scope", scopeKey)
	client, err := util.NewHTTPClient(url, http.MethodDelete, c.AccessToken)
	if err != nil {
		return err
	}
	res, err := client.MakeRequest(ctx, nil)
	if err := util.HandleResponseError(res, err); err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (c *Scope) GetAll(ctx context.Context) ([]ScopeModel, error) { //nolint:dupl
	var response AllScopeResp
	url := fmt.Sprintf("%s/%s", c.BaseURL, "scopes-srv/scope/list")
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
	return response.Data, nil
}
