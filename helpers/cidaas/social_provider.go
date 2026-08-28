package cidaas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type SocialProviderModel struct {
	ID                    string                `json:"id,omitempty"`
	ClientID              string                `json:"client_id,omitempty"`
	ClientSecret          string                `json:"client_secret,omitempty"`
	Name                  string                `json:"name,omitempty"`
	ProviderName          string                `json:"provider_name,omitempty"`
	Claims                *ClaimsModel          `json:"claims"`
	EnabledForAdminPortal bool                  `json:"enabled_for_admin_portal"`
	Enabled               bool                  `json:"enabled"`
	Scopes                []string              `json:"scopes"`
	UserInfoFields        []UserInfoFieldsModel `json:"userinfo_fields"`
}

type ClaimsModel struct {
	RequiredClaims RequiredClaimsModel `json:"required_claims,omitempty"`
	OptionalClaims OptionalClaimsModel `json:"optional_claims,omitempty"`
}

type RequiredClaimsModel struct {
	UserInfo []string `json:"user_info,omitempty"`
	IDToken  []string `json:"id_token,omitempty"`
}

type OptionalClaimsModel struct {
	UserInfo []string `json:"user_info,omitempty"`
	IDToken  []string `json:"id_token,omitempty"`
}

type UserInfoFieldsModel struct {
	InnerKey      string `json:"inner_key,omitempty" tfsdk:"inner_key"`
	ExternalKey   string `json:"external_key,omitempty" tfsdk:"external_key"`
	IsCustomField bool   `json:"is_custom_field,omitempty" tfsdk:"is_custom_field"`
	IsSystemField bool   `json:"is_system_field,omitempty" tfsdk:"is_system_field"`
}

type SocialProviderResponse struct {
	Success bool `json:"success,omitempty"`
	Status  int  `json:"status,omitempty"`
	Data    SocialProviderModel
}

type AllSocialProviderResponse struct {
	Success bool `json:"success,omitempty"`
	Status  int  `json:"status,omitempty"`
	Data    []SocialProviderModel
}

type SocialProvider struct {
	ClientConfig
}

func NewSocialProvider(clientConfig ClientConfig) *SocialProvider {
	return &SocialProvider{clientConfig}
}

func (s *SocialProvider) Upsert(ctx context.Context, sp *SocialProviderModel) (*SocialProviderResponse, error) { //nolint:dupl
	var response SocialProviderResponse
	url := fmt.Sprintf("%s/%s", s.BaseURL, "providers-srv/multi/providers")
	client, err := util.NewHTTPClient(url, http.MethodPost, s.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, sp)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *SocialProvider) Get(ctx context.Context, providerName, providerID string) (*SocialProviderResponse, error) {
	var response SocialProviderResponse
	url := fmt.Sprintf("%s/%s?provider_name=%s&provider_id=%s", s.BaseURL, "providers-srv/multi/providers", providerName, providerID)
	client, err := util.NewHTTPClient(url, http.MethodGet, s.AccessToken)
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

func (s *SocialProvider) Delete(ctx context.Context, providerName, providerID string) error {
	url := fmt.Sprintf("%s/%s/%s/%s", s.BaseURL, "providers-srv/multi/providers", providerName, providerID)
	client, err := util.NewHTTPClient(url, http.MethodDelete, s.AccessToken)
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

func (s *SocialProvider) GetAll(ctx context.Context) ([]SocialProviderModel, error) { //nolint:dupl
	var response AllSocialProviderResponse
	url := fmt.Sprintf("%s/%s", s.BaseURL, "providers-srv/providers/enabled/list")
	client, err := util.NewHTTPClient(url, http.MethodGet, s.AccessToken)
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
