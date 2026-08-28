package cidaas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type PasswordPolicyModel struct {
	ID             string  `json:"_id"`
	PolicyName     string  `json:"policy_name"`
	PasswordPolicy *Policy `json:"passwordPolicy"`
}

type Policy struct {
	BlockCompromised  bool              `json:"blockCompromised"`
	DenyUsageCount    int64             `json:"denyUsageCount"`
	StrengthRegexes   []string          `json:"strengthRegexes"`
	ChangeEnforcement ChangeEnforcement `json:"changeEnforcement"`
}

type ChangeEnforcement struct {
	ExpirationInDays       int64 `json:"expirationInDays"`
	NotifyUserBeforeInDays int64 `json:"notifyUserBeforeInDays"`
}

type PasswordPolicyResponse struct {
	Success bool                `json:"success"`
	Status  int                 `json:"status"`
	Data    PasswordPolicyModel `json:"data,omitempty"`
}

type PasswordPolicyUpdateResponse struct {
	Success bool `json:"success"`
	Status  int  `json:"status"`
	Data    bool `json:"data"`
}

type PasswordPolicy struct {
	ClientConfig
}

func NewPasswordPolicy(clientConfig ClientConfig) *PasswordPolicy {
	return &PasswordPolicy{clientConfig}
}

func (p *PasswordPolicy) Get(ctx context.Context, id string) (*PasswordPolicyResponse, error) { //nolint:dupl
	var response PasswordPolicyResponse
	url := fmt.Sprintf("%s/%s/%s", p.BaseURL, "verification-actions-srv/policies", id)
	client, err := util.NewHTTPClient(url, http.MethodGet, p.AccessToken)
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

func (p *PasswordPolicy) Create(ctx context.Context, payload PasswordPolicyModel) (*PasswordPolicyResponse, error) { //nolint:dupl
	var response PasswordPolicyResponse
	url := fmt.Sprintf("%s/%s", p.BaseURL, "verification-actions-srv/policies")
	client, err := util.NewHTTPClient(url, http.MethodPost, p.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, payload)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (p *PasswordPolicy) Update(ctx context.Context, payload PasswordPolicyModel) (*PasswordPolicyUpdateResponse, error) { //nolint:dupl
	var response PasswordPolicyUpdateResponse
	url := fmt.Sprintf("%s/%s", p.BaseURL, "verification-actions-srv/policies")
	client, err := util.NewHTTPClient(url, http.MethodPut, p.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, payload)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (p *PasswordPolicy) Delete(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s/%s", p.BaseURL, "verification-actions-srv/policies", id)
	client, err := util.NewHTTPClient(url, http.MethodDelete, p.AccessToken)
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
