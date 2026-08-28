package cidaas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type GroupTypeData struct {
	ID           string   `json:"_id,omitempty"`
	RoleMode     string   `json:"roleMode,omitempty"`
	GroupType    string   `json:"groupType,omitempty"`
	Description  string   `json:"description,omitempty"`
	AllowedRoles []string `json:"allowedRoles,omitempty"`
	CreatedTime  string   `json:"createdTime,omitempty"`
	UpdatedTime  string   `json:"updatedTime,omitempty"`
}

type GroupTypeResponse struct {
	Success bool          `json:"success,omitempty"`
	Status  int           `json:"status,omitempty"`
	Data    GroupTypeData `json:"data,omitempty"`
}

type AllGroupTypeResponse struct {
	Success bool            `json:"success,omitempty"`
	Status  int             `json:"status,omitempty"`
	Data    []GroupTypeData `json:"data,omitempty"`
}

type GroupType struct {
	ClientConfig
}

func NewGroupType(clientConfig ClientConfig) *GroupType {
	return &GroupType{clientConfig}
}

func (c *GroupType) Create(ctx context.Context, gt GroupTypeData) (*GroupTypeResponse, error) { //nolint:dupl
	var response GroupTypeResponse
	url := fmt.Sprintf("%s/%s", c.BaseURL, "groups-srv/grouptypes")
	client, err := util.NewHTTPClient(url, http.MethodPost, c.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, gt)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *GroupType) Get(ctx context.Context, groupType string) (*GroupTypeResponse, error) { //nolint:dupl
	var response GroupTypeResponse
	url := fmt.Sprintf("%s/%s?groupType=%s", c.BaseURL, "groups-srv/grouptypes", groupType)
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

func (c *GroupType) Update(ctx context.Context, gt GroupTypeData) error {
	url := fmt.Sprintf("%s/%s", c.BaseURL, "groups-srv/grouptypes")
	client, err := util.NewHTTPClient(url, http.MethodPut, c.AccessToken)
	if err != nil {
		return err
	}
	res, err := client.MakeRequest(ctx, gt)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (c *GroupType) Delete(ctx context.Context, groupType string) error {
	url := fmt.Sprintf("%s/%s/%s", c.BaseURL, "groups-srv/grouptypes", groupType)
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

func (c *GroupType) GetAll(ctx context.Context) ([]GroupTypeData, error) {
	var response AllGroupTypeResponse
	url := fmt.Sprintf("%s/%s", c.BaseURL, "groups-srv/graph/grouptypes")
	client, err := util.NewHTTPClient(url, http.MethodPost, c.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, struct{}{})
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}
