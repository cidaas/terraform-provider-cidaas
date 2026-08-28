// nolint:dupl
package cidaas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type ScopeGroupConfig struct {
	ID          string `json:"_id,omitempty"`
	GroupName   string `json:"group_name,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedTime string `json:"createdTime,omitempty"`
	UpdatedTime string `json:"updatedTime,omitempty"`
}

type ScopeGroupResponse struct {
	Success bool             `json:"success,omitempty"`
	Status  int              `json:"status,omitempty"`
	Data    ScopeGroupConfig `json:"data,omitempty"`
}

type AllScopeGroupResp struct {
	Success bool               `json:"success,omitempty"`
	Status  int                `json:"status,omitempty"`
	Data    []ScopeGroupConfig `json:"data,omitempty"`
}

type ScopeGroup struct {
	ClientConfig
}

func NewScopeGroup(clientConfig ClientConfig) *ScopeGroup {
	return &ScopeGroup{clientConfig}
}

const scopeGroupEndpoint = "scopes-srv/group"

func (c *ScopeGroup) Upsert(ctx context.Context, sg ScopeGroupConfig) (*ScopeGroupResponse, error) {
	res, err := c.makeRequest(ctx, http.MethodPost, scopeGroupEndpoint, sg)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert scope group: %w", err)
	}
	defer res.Body.Close()

	var response ScopeGroupResponse
	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *ScopeGroup) Get(ctx context.Context, scopeGroupName string) (*ScopeGroupResponse, error) {
	if scopeGroupName == "" {
		return nil, fmt.Errorf("scopeGroupName cannot be empty")
	}
	endpoint := fmt.Sprintf("%s?group_name=%s", scopeGroupEndpoint, scopeGroupName)
	res, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get scope group: %w", err)
	}
	defer res.Body.Close()

	var response ScopeGroupResponse
	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *ScopeGroup) Delete(ctx context.Context, scopeGroupName string) error {
	if scopeGroupName == "" {
		return fmt.Errorf("scopeGroupName cannot be empty")
	}
	endpoint := fmt.Sprintf("%s/%s", scopeGroupEndpoint, scopeGroupName)
	res, err := c.makeRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to delete scope group: %w", err)
	}
	defer res.Body.Close()
	return nil
}

func (c *ScopeGroup) GetAll(ctx context.Context) ([]ScopeGroupConfig, error) {
	endpoint := "scopes-srv/group/list"
	res, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get scope group list: %w", err)
	}
	defer res.Body.Close()

	var response AllScopeGroupResp
	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}
