package cidaas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type GroupSelectionDetails struct {
	IsGroupLoginSelectionEnabled bool     `json:"isGroupLoginSelectionEnabled"`
	AlwaysShowGroupSelection     bool     `json:"alwaysShowGroupSelection"`
	SelectableGroups             []string `json:"selectableGroups,omitempty"`
	SelectableGroupTypes         []string `json:"selectableGroupTypes,omitempty"`
}

type GroupSelectionModel struct {
	ID             string                `json:"id,omitempty"`
	Name           string                `json:"name,omitempty"`
	Description    string                `json:"description,omitempty"`
	GroupSelection GroupSelectionDetails `json:"group_selection"`
	CreatedAt      string                `json:"createdTime,omitempty"`
	UpdatedAt      string                `json:"updatedTime,omitempty"`
}

type GroupSelectionResponse struct {
	Success bool                `json:"success,omitempty"`
	Status  int                 `json:"status,omitempty"`
	Data    GroupSelectionModel `json:"data,omitempty"`
}

type GroupSelection struct {
	ClientConfig
}

func NewGroupSelection(clientConfig ClientConfig) *GroupSelection {
	return &GroupSelection{clientConfig}
}

func (g *GroupSelection) Upsert(ctx context.Context, config GroupSelectionModel) (*GroupSelectionResponse, error) {
	var method string
	var url string
	if config.ID != "" {
		method = http.MethodPatch
		url = fmt.Sprintf("groups-srv/group-selection/%s", config.ID)
	} else {
		method = http.MethodPost
		url = "groups-srv/group-selection"
	}

	res, err := g.makeRequest(ctx, method, url, config)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	var response GroupSelectionResponse
	err = util.ProcessResponse(res, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (g *GroupSelection) Get(ctx context.Context, id string) (*GroupSelectionResponse, error) {
	url := fmt.Sprintf("groups-srv/group-selection/%s", id)
	res, err := g.makeRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	var response GroupSelectionResponse
	err = util.ProcessResponse(res, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (g *GroupSelection) Delete(ctx context.Context, id string) (bool, error) {
	url := fmt.Sprintf("groups-srv/group-selection/%s", id)
	res, err := g.makeRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = res.Body.Close() }()

	var response struct {
		Success bool `json:"success"`
	}
	err = util.ProcessResponse(res, &response)
	if err != nil {
		return false, err
	}
	return response.Success, nil
}
