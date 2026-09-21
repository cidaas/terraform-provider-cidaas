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

func buildFederationProviderPayload(pc *ProviderConfigModel, id string) map[string]interface{} {
	if pc.Owner == "" {
		pc.Owner = "client"
	}
	if pc.Provider == "" {
		pc.Provider = "custom"
	}
	if pc.StandardType == "" {
		pc.StandardType = "OAUTH2"
	}

	clientID := pc.ClientID
	clientSecret := pc.ClientSecret
	authEndpoint := pc.AuthorizationEndpoint
	tokenEndpoint := pc.TokenEndpoint
	userinfoEndpoint := pc.UserinfoEndpoint
	if pc.OAuth2 != nil {
		if clientID == "" {
			clientID = pc.OAuth2.ClientID
		}
		if clientSecret == "" {
			clientSecret = pc.OAuth2.ClientSecret
		}
		if authEndpoint == "" {
			authEndpoint = pc.OAuth2.AuthorizationEndpoint
		}
		if tokenEndpoint == "" {
			tokenEndpoint = pc.OAuth2.TokenEndpoint
		}
		if userinfoEndpoint == "" {
			userinfoEndpoint = pc.OAuth2.UserinfoEndpoint
		}
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
			"client_id":              clientID,
			"client_secret":          clientSecret,
			"authorization_endpoint": authEndpoint,
			"token_endpoint":         tokenEndpoint,
			"userinfo_endpoint":      userinfoEndpoint,
			"userinfoSource":         "USERINFOENDPOINT",
		},
	}
	if id != "" {
		payload["id"] = id
	}
	return payload
}

func unwrapOAuth2Config(data *ProviderConfigModel) {
	if data.OAuth2 != nil {
		data.ClientID = data.OAuth2.ClientID
		data.ClientSecret = data.OAuth2.ClientSecret
		data.AuthorizationEndpoint = data.OAuth2.AuthorizationEndpoint
		data.TokenEndpoint = data.OAuth2.TokenEndpoint
		data.UserinfoEndpoint = data.OAuth2.UserinfoEndpoint
	}
}

func (f *FederationProvider) Create(ctx context.Context, pc *ProviderConfigModel) (*ProviderConfigResponse, error) {
	payload := buildFederationProviderPayload(pc, "")

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
	unwrapOAuth2Config(&response.Data)
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
	unwrapOAuth2Config(&response.Data)
	return &response, nil
}

//nolint:dupl
func (f *FederationProvider) Update(ctx context.Context, id string, pc *ProviderConfigModel) (*ProviderConfigResponse, error) {
	payload := buildFederationProviderPayload(pc, id)

	var response ProviderConfigResponse
	url := fmt.Sprintf("%s/%s/%s", f.BaseURL, "federation/providers", id)
	client, err := util.NewHTTPClient(url, http.MethodPut, f.AccessToken)
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
	unwrapOAuth2Config(&response.Data)
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
