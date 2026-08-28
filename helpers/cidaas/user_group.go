package cidaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type UserGroupData struct {
	ID                          string            `json:"_id,omitempty"`
	GroupType                   string            `json:"groupType,omitempty"`
	GroupID                     string            `json:"groupId,omitempty"`
	GroupName                   string            `json:"groupName,omitempty"`
	ParentID                    string            `json:"parentId,omitempty"`
	LogoURL                     string            `json:"logoUrl,omitempty"`
	Description                 string            `json:"description,omitempty"`
	MakeFirstUserAdmin          bool              `json:"make_first_user_admin,omitempty"`
	MemberProfileVisibility     string            `json:"memberProfileVisibility,omitempty"`
	NoneMemberProfileVisibility string            `json:"noneMemberProfileVisibility,omitempty"`
	GroupOwner                  string            `json:"groupOwner,omitempty"`
	CustomFields                map[string]string `json:"customFields,omitempty"`
	CreatedTime                 string            `json:"createdTime,omitempty"`
	UpdatedTime                 string            `json:"updatedTime,omitempty"`
}

type UserGroupResponse struct {
	Success bool          `json:"success,omitempty"`
	Status  int           `json:"status,omitempty"`
	Data    UserGroupData `json:"data,omitempty"`
}

type SubGroupResponse struct {
	Success bool `json:"success,omitempty"`
	Status  int  `json:"status,omitempty"`
	Data    struct {
		Groups []UserGroupData `json:"groups"`
	} `json:"data,omitempty"`
}

type UserGroup struct {
	ClientConfig
}

func NewUserGroup(clientConfig ClientConfig) *UserGroup {
	return &UserGroup{clientConfig}
}

const userGroupsEndpoint = "groups-srv/usergroups"

func (c *UserGroup) Create(ctx context.Context, ug UserGroupData) (*UserGroupResponse, error) {
	res, err := c.makeRequest(ctx, http.MethodPost, userGroupsEndpoint, ug)
	if err != nil {
		return nil, fmt.Errorf("failed to create user group: %w", err)
	}
	defer res.Body.Close()

	var response UserGroupResponse
	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *UserGroup) Get(ctx context.Context, groupID string) (*UserGroupResponse, error) { //nolint:dupl
	if groupID == "" {
		return nil, fmt.Errorf("groupID cannot be empty")
	}
	endpoint := fmt.Sprintf("%s/%s", userGroupsEndpoint, groupID)
	res, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ger user group: %w", err)
	}
	defer res.Body.Close()

	var response UserGroupResponse
	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *UserGroup) Update(ctx context.Context, ug UserGroupData) (*UserGroupResponse, error) {
	res, err := c.makeRequest(ctx, http.MethodPut, userGroupsEndpoint, ug)
	if err != nil {
		return nil, fmt.Errorf("failed to update user group: %w", err)
	}
	defer res.Body.Close()

	var response UserGroupResponse
	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *UserGroup) Delete(ctx context.Context, groupID string) error {
	if groupID == "" {
		return fmt.Errorf("groupID cannot be empty")
	}
	endpoint := fmt.Sprintf("%s/%s", userGroupsEndpoint, groupID)
	res, err := c.makeRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to delete user group: %w", err)
	}
	defer res.Body.Close()

	if err := util.HandleResponseError(res, err); err != nil {
		return err
	}
	return nil
}

func (c *UserGroup) GetSubGroups(ctx context.Context, parentID string) ([]UserGroupData, error) {
	endpoint := "groups-srv/graph/usergroups"
	payload := map[string]string{"parentId": parentID}
	res, err := c.makeRequest(ctx, http.MethodPost, endpoint, payload)
	if res.StatusCode == http.StatusNoContent {
		return []UserGroupData{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sub groups: %w", err)
	}
	defer res.Body.Close()

	var response SubGroupResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode subgroup response: %w", err)
	}
	return response.Data.Groups, nil
}
