package cidaas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type RoleModel struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Role        string `json:"role,omitempty"`
}

type RoleResponse struct {
	Success bool      `json:"success,omitempty"`
	Status  int       `json:"status,omitempty"`
	Data    RoleModel `json:"data,omitempty"`
}

type AllRoleResponse struct {
	Success bool        `json:"success,omitempty"`
	Status  int         `json:"status,omitempty"`
	Data    []RoleModel `json:"data,omitempty"`
}

type Role struct {
	ClientConfig
}

func NewRole(clientConfig ClientConfig) *Role {
	return &Role{clientConfig}
}

const rolesEndpoint = "roles-srv/role"

func (r *Role) UpsertRole(ctx context.Context, role RoleModel) (*RoleResponse, error) {
	res, err := r.makeRequest(ctx, http.MethodPost, rolesEndpoint, role)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert role: %w", err)
	}
	defer res.Body.Close()

	var response RoleResponse
	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (r *Role) GetRole(ctx context.Context, role string) (*RoleResponse, error) { //nolint:dupl
	if role == "" {
		return nil, fmt.Errorf("role cannot be empty")
	}
	endpoint := fmt.Sprintf("%s?role=%s", rolesEndpoint, role)
	res, err := r.makeRequest(ctx, http.MethodGet, endpoint, role)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	defer res.Body.Close()

	var response RoleResponse
	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (r *Role) DeleteRole(ctx context.Context, role string) error {
	if role == "" {
		return fmt.Errorf("role cannot be empty")
	}
	endpoint := fmt.Sprintf("%s?role=%s", rolesEndpoint, role)
	res, err := r.makeRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	defer res.Body.Close()
	return nil
}

func (r *Role) GetAll(ctx context.Context) ([]RoleModel, error) {
	endpoint := "groups-srv/graph/roles"
	res, err := r.makeRequest(ctx, http.MethodPost, endpoint, struct{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}
	defer res.Body.Close()

	var response AllRoleResponse
	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}
