package cidaas

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type Scopes struct {
	DisplayLabel string       `json:"display_label"`
	Scopes       []ScopeChild `json:"scopes,omitempty"`
}

type ScopeChild struct {
	ScopeName   string `json:"scope_name,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Recommended bool   `json:"recommended,omitempty"`
}

type CustomProviderModel struct {
	ID                    string                 `json:"_id,omitempty"`
	ClientID              string                 `json:"client_id,omitempty"`
	ClientSecret          string                 `json:"client_secret,omitempty"`
	DisplayName           string                 `json:"display_name,omitempty"`
	StandardType          string                 `json:"standard_type,omitempty"`
	AuthorizationEndpoint string                 `json:"authorization_endpoint,omitempty"`
	TokenEndpoint         string                 `json:"token_endpoint,omitempty"`
	ProviderName          string                 `json:"provider_name,omitempty"`
	LogoURL               string                 `json:"logo_url"`
	UserinfoEndpoint      string                 `json:"userinfo_endpoint,omitempty"`
	UserinfoFields        map[string]interface{} `json:"userInfoFields,omitempty"`
	Scopes                Scopes                 `json:"scopes,omitempty"`
	Domains               []string               `json:"domains,omitempty"`
	AmrConfig             []AmrConfig            `json:"amrConfig,omitempty"`
	UserInfoSource        string                 `json:"userInfoSource,omitempty"`
	Pkce                  bool                   `json:"pkce"`
	AuthType              string                 `json:"auth_type,omitempty"`
	APIKeyDetails         APIKeyDetails          `json:"apikeyDetails,omitempty"`
	TotpDetails           TotpDetails            `json:"totpDetails,omitempty"`
	CidaasAuthDetails     AuthDetails            `json:"cidaasAuthDetails,omitempty"`
}

type UserInfoField struct {
	ExtFieldKey string `json:"extFieldKey"`
	Default     string `json:"default,omitempty"`
}

type AmrConfig struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type CustomProviderResponse struct {
	Success bool                `json:"success,omitempty"`
	Status  int                 `json:"status,omitempty"`
	Data    CustomProviderModel `json:"data,omitempty"`
}

type AllCustomProviderResponse struct {
	Success bool                  `json:"success,omitempty"`
	Status  int                   `json:"status,omitempty"`
	Data    []CustomProviderModel `json:"data,omitempty"`
}

type CustomProvider struct {
	ClientConfig
}

func NewCustomProvider(clientConfig ClientConfig) *CustomProvider {
	return &CustomProvider{clientConfig}
}

func (c *CustomProvider) CreateCustomProvider(ctx context.Context, cp *CustomProviderModel) (*CustomProviderResponse, error) {
	if len(cp.UserinfoFields) == 0 {
		cp.UserinfoFields = map[string]interface{}{
			"sub": map[string]string{"extFieldKey": "sub"},
		}
	}
	var response CustomProviderResponse
	url := fmt.Sprintf("%s/%s", c.BaseURL, "providers-srv/custom")
	client, err := util.NewHTTPClient(url, http.MethodPost, c.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, cp)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *CustomProvider) UpdateCustomProvider(ctx context.Context, cp *CustomProviderModel) error {
	url := fmt.Sprintf("%s/%s", c.BaseURL, "providers-srv/custom")
	client, err := util.NewHTTPClient(url, http.MethodPut, c.AccessToken)
	if err != nil {
		return err
	}
	res, err := client.MakeRequest(ctx, cp)
	if err := util.HandleResponseError(res, err); err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	return nil
}

//nolint:dupl
func (c *CustomProvider) GetCustomProvider(ctx context.Context, providerName string) (*CustomProviderResponse, error) {
	var response CustomProviderResponse
	url := fmt.Sprintf("%s/%s/%s", c.BaseURL, "providers-srv/custom", providerName)
	client, err := util.NewHTTPClient(url, http.MethodGet, c.AccessToken)
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
	return &response, nil
}

func (c *CustomProvider) DeleteCustomProvider(ctx context.Context, providerName string) error {
	url := fmt.Sprintf("%s/%s/%s", c.BaseURL, "providers-srv/custom", strings.ToLower(providerName))
	client, err := util.NewHTTPClient(url, http.MethodDelete, c.AccessToken)
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
