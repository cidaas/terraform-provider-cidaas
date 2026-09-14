package cidaas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type RoleVerificationFilterWire struct {
	Roles          []string `json:"roles,omitempty"`
	MatchCondition string   `json:"matchCondition,omitempty"`
}

type GroupVerificationFilterItemWire struct {
	GroupID    string                      `json:"groupId,omitempty"`
	GroupType  string                      `json:"groupType,omitempty"`
	RoleFilter *RoleVerificationFilterWire `json:"roleFilter,omitempty"`
}

type GroupVerificationRequestModel struct {
	ID             string                            `json:"id,omitempty"`
	Description    string                            `json:"description,omitempty"`
	MatchCondition string                            `json:"matchCondition,omitempty"`
	Filters        []GroupVerificationFilterItemWire `json:"filters,omitempty"`
	CreatedAt      string                            `json:"createdTime,omitempty"`
	UpdatedAt      string                            `json:"updatedTime,omitempty"`
}

type GroupVerificationRequestResponse struct {
	Success bool                          `json:"success,omitempty"`
	Status  int                           `json:"status,omitempty"`
	Data    GroupVerificationRequestModel `json:"data,omitempty"`
}

type GroupVerificationFilter struct {
	ClientConfig
}

func NewGroupVerificationFilter(clientConfig ClientConfig) *GroupVerificationFilter {
	return &GroupVerificationFilter{clientConfig}
}

func (g *GroupVerificationFilter) Create(ctx context.Context, config GroupVerificationRequestModel) (*GroupVerificationRequestResponse, error) {
	url := "groups-srv/verifications/requests"
	res, err := g.makeRequest(ctx, http.MethodPost, url, config)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	var response GroupVerificationRequestResponse
	err = util.ProcessResponse(res, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (g *GroupVerificationFilter) Update(ctx context.Context, config GroupVerificationRequestModel) (*GroupVerificationRequestResponse, error) {
	url := fmt.Sprintf("groups-srv/verifications/requests/%s", config.ID)
	res, err := g.makeRequest(ctx, http.MethodPut, url, config)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	var response GroupVerificationRequestResponse
	err = util.ProcessResponse(res, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (g *GroupVerificationFilter) Get(ctx context.Context, id string) (*GroupVerificationRequestResponse, error) {
	url := fmt.Sprintf("groups-srv/verifications/requests/%s", id)
	res, err := g.makeRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	var response GroupVerificationRequestResponse
	err = util.ProcessResponse(res, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (g *GroupVerificationFilter) Delete(ctx context.Context, id string) (bool, error) {
	url := fmt.Sprintf("groups-srv/verifications/requests/%s", id)
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
