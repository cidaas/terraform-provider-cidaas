package cidaas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type ConsentModel struct {
	ID             string `json:"_id,omitempty"`
	ConsentGroupID string `json:"consent_group_id,omitempty"`
	ConsentName    string `json:"consent_name,omitempty"`
	Enabled        bool   `json:"enabled"`
	CreatedTime    string `json:"createdTime,omitempty"`
	UpdatedTime    string `json:"updatedTime,omitempty"`
}

type ConsentResponse struct {
	Success bool         `json:"success,omitempty"`
	Status  int          `json:"status,omitempty"`
	Data    ConsentModel `json:"data,omitempty"`
}

type ConsentInstanceResponse struct {
	Success bool           `json:"success,omitempty"`
	Status  int            `json:"status,omitempty"`
	Data    []ConsentModel `json:"data,omitempty"`
}

type Consent struct {
	ClientConfig
}

func NewConsent(clientConfig ClientConfig) *Consent {
	return &Consent{clientConfig}
}

func (c *Consent) Upsert(ctx context.Context, consentConfig ConsentModel) (*ConsentResponse, error) { //nolint:dupl
	var response ConsentResponse
	url := fmt.Sprintf("%s/%s", c.BaseURL, "consent-management-srv/v2/consent/instance")
	client, err := util.NewHTTPClient(url, http.MethodPost, c.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, consentConfig)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Consent) GetConsentInstances(ctx context.Context, consentGroupID string) (*ConsentInstanceResponse, error) {
	var response ConsentInstanceResponse
	url := fmt.Sprintf("%s/%s/%s", c.BaseURL, "consent-management-srv/v2/consent/instance", consentGroupID)
	client, err := util.NewHTTPClient(url, http.MethodGet, c.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, nil)
	if res.StatusCode == http.StatusNoContent {
		return &ConsentInstanceResponse{
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

func (c *Consent) Delete(ctx context.Context, consentID string) error {
	url := fmt.Sprintf("%s/%s/%s", c.BaseURL, "consent-management-srv/v2/consent/instance", consentID)
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

func (c *Consent) GetAll(ctx context.Context) ([]ConsentModel, error) { //nolint:dupl
	var response ConsentInstanceResponse
	url := fmt.Sprintf("%s/%s", c.BaseURL, "consent-management-srv/v2/consent/instance/all/list")
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
