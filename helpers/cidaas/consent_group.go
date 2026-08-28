package cidaas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type ConsentGroupConfig struct {
	ID          string `json:"_id,omitempty"`
	GroupName   string `json:"group_name,omitempty"`
	Description string `json:"description"` // description needs to set to empty string, so omitempty is removed here
	CreatedTime string `json:"createdTime,omitempty"`
	UpdatedTime string `json:"updatedTime,omitempty"`
}

type ConsentGroupResponse struct {
	Success bool               `json:"success,omitempty"`
	Status  int                `json:"status,omitempty"`
	Data    ConsentGroupConfig `json:"data,omitempty"`
}

type ConsentGroup struct {
	ClientConfig
}

func NewConsentGroup(clientConfig ClientConfig) *ConsentGroup {
	return &ConsentGroup{clientConfig}
}

func (c *ConsentGroup) Upsert(ctx context.Context, cg ConsentGroupConfig) (*ConsentGroupResponse, error) { //nolint:dupl
	var response ConsentGroupResponse
	url := fmt.Sprintf("%s/%s", c.BaseURL, "consent-management-srv/v2/groups")
	client, err := util.NewHTTPClient(url, http.MethodPost, c.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, cg)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *ConsentGroup) Get(ctx context.Context, consentGroupID string) (*ConsentGroupResponse, error) { //nolint:dupl
	var response ConsentGroupResponse
	url := fmt.Sprintf("%s/%s/%s", c.BaseURL, "consent-management-srv/v2/groups", consentGroupID)
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

func (c *ConsentGroup) Delete(ctx context.Context, consentGroupID string) error {
	url := fmt.Sprintf("%s/%s/%s", c.BaseURL, "consent-management-srv/v2/groups", consentGroupID)
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
