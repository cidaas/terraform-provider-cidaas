package cidaas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type OAuth2Config struct {
	ClientID              string   `json:"client_id,omitempty"`
	ClientSecret          string   `json:"client_secret,omitempty"`
	AuthorizationEndpoint string   `json:"authorization_endpoint,omitempty"`
	TokenEndpoint         string   `json:"token_endpoint,omitempty"`
	UserinfoEndpoint      string   `json:"userinfo_endpoint,omitempty"`
	UserinfoSource        string   `json:"userinfoSource,omitempty"`
	Scopes                []string `json:"scopes,omitempty"`
}

type ProviderConfigModel struct {
	ID                    string                 `json:"id,omitempty"`
	MongoID               string                 `json:"_id,omitempty"`
	ClientID              string                 `json:"client_id,omitempty"`
	ClientSecret          string                 `json:"client_secret,omitempty"`
	DisplayName           string                 `json:"display_name,omitempty"`
	StandardType          string                 `json:"standard_type,omitempty"`
	AuthorizationEndpoint string                 `json:"authorization_endpoint,omitempty"`
	TokenEndpoint         string                 `json:"token_endpoint,omitempty"`
	ProviderName          string                 `json:"provider_name,omitempty"`
	LogoURL               string                 `json:"logo_url,omitempty"`
	UserinfoEndpoint      string                 `json:"userinfo_endpoint,omitempty"`
	UserinfoFields        map[string]interface{} `json:"userInfoFields,omitempty"`
	Scopes                Scopes                 `json:"scopes,omitempty"`
	Domains               []string               `json:"domains,omitempty"`
	Pkce                  bool                   `json:"pkce,omitempty"`
	AuthType              string                 `json:"auth_type,omitempty"`
	Owner                 string                 `json:"owner,omitempty"`
	Provider              string                 `json:"provider,omitempty"`
	OAuth2                *OAuth2Config          `json:"oauth2,omitempty"`
}

type ProviderConfigResponse struct {
	Success bool                `json:"success,omitempty"`
	Status  int                 `json:"status,omitempty"`
	Data    ProviderConfigModel `json:"data,omitempty"`
}

type AllProviderConfigResponse struct {
	Success bool                  `json:"success,omitempty"`
	Status  int                   `json:"status,omitempty"`
	Data    []ProviderConfigModel `json:"data,omitempty"`
}

type FederationProvider struct {
	ClientConfig
}

func NewFederationProvider(clientConfig ClientConfig) *FederationProvider {
	return &FederationProvider{clientConfig}
}

func (f *FederationProvider) Create(ctx context.Context, pc *ProviderConfigModel) (*ProviderConfigResponse, error) {
	if pc.Owner == "" {
		pc.Owner = "client"
	}
	if pc.Provider == "" {
		pc.Provider = "custom"
	}
	if pc.StandardType == "" {
		pc.StandardType = "OAUTH2"
	}

	payload := map[string]interface{}{
		"provider":      pc.Provider,
		"provider_name": pc.ProviderName,
		"display_name":  pc.DisplayName,
		"standard_type": pc.StandardType,
		"logo_url":      pc.LogoURL,
		"owner":         pc.Owner,
		"domains":       pc.Domains,
		"oauth2": map[string]interface{}{
			"client_id":              pc.ClientID,
			"client_secret":          pc.ClientSecret,
			"authorization_endpoint": pc.AuthorizationEndpoint,
			"token_endpoint":         pc.TokenEndpoint,
			"userinfo_endpoint":      pc.UserinfoEndpoint,
			"userinfoSource":         "USERINFOENDPOINT",
		},
	}

	var response ProviderConfigResponse
	url := fmt.Sprintf("%s/%s", f.BaseURL, "federation/providers")
	client, err := util.NewHTTPClient(url, http.MethodPost, f.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, payload)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	if response.Data.OAuth2 != nil {
		response.Data.ClientID = response.Data.OAuth2.ClientID
		response.Data.ClientSecret = response.Data.OAuth2.ClientSecret
		response.Data.AuthorizationEndpoint = response.Data.OAuth2.AuthorizationEndpoint
		response.Data.TokenEndpoint = response.Data.OAuth2.TokenEndpoint
		response.Data.UserinfoEndpoint = response.Data.OAuth2.UserinfoEndpoint
	}
	return &response, nil
}

func (f *FederationProvider) Get(ctx context.Context, id string) (*ProviderConfigResponse, error) {
	var response ProviderConfigResponse
	url := fmt.Sprintf("%s/%s/%s", f.BaseURL, "federation/providers", id)
	client, err := util.NewHTTPClient(url, http.MethodGet, f.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, nil)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	if response.Data.OAuth2 != nil {
		response.Data.ClientID = response.Data.OAuth2.ClientID
		response.Data.ClientSecret = response.Data.OAuth2.ClientSecret
		response.Data.AuthorizationEndpoint = response.Data.OAuth2.AuthorizationEndpoint
		response.Data.TokenEndpoint = response.Data.OAuth2.TokenEndpoint
		response.Data.UserinfoEndpoint = response.Data.OAuth2.UserinfoEndpoint
	}
	return &response, nil
}

//nolint:dupl
func (f *FederationProvider) Update(ctx context.Context, id string, pc *ProviderConfigModel) (*ProviderConfigResponse, error) {
	var response ProviderConfigResponse
	url := fmt.Sprintf("%s/%s/%s", f.BaseURL, "federation/providers", id)
	client, err := util.NewHTTPClient(url, http.MethodPut, f.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, pc)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (f *FederationProvider) Delete(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s/%s", f.BaseURL, "federation/providers", id)
	client, err := util.NewHTTPClient(url, http.MethodDelete, f.AccessToken)
	if err != nil {
		return err
	}
	res, err := client.MakeRequest(ctx, nil)
	if err := util.HandleResponseError(res, err); err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	return nil
}
