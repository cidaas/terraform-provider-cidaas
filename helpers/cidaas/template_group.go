package cidaas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type TemplateGroupResponse struct {
	Success bool               `json:"success"`
	Status  int                `json:"status"`
	Data    TemplateGroupModel `json:"data"`
}

type TemplateGroupModel struct {
	ID                string             `json:"id,omitempty"`
	GroupID           string             `json:"group_id,omitempty"`
	SenderConfig      *SenderConfig      `json:"sender_config,omitempty"`
	EmailSenderConfig *EmailSenderConfig `json:"email_sender_config,omitempty"`
	SMSSenderConfig   *SMSSenderConfig   `json:"sms_sender_config,omitempty"`
	IVRSenderConfig   *IVRSenderConfig   `json:"ivr_sender_config,omitempty"`
	PushSenderConfig  *IVRSenderConfig   `json:"push_sender_config,omitempty"`
}

type SenderConfig struct {
	ID        string `json:"id,omitempty"`
	FromEmail string `json:"from_email,omitempty"`
	FromName  string `json:"from_name,omitempty"`
}

type EmailSenderConfig struct {
	ID          string   `json:"id,omitempty"`
	FromEmail   string   `json:"from_email,omitempty"`
	FromName    string   `json:"from_name,omitempty"`
	ReplyTo     string   `json:"reply_to,omitempty"`
	SenderNames []string `json:"sender_names,omitempty"`
}

type SMSSenderConfig struct {
	ID          string   `json:"id,omitempty"`
	FromName    string   `json:"from_name,omitempty"`
	SenderNames []string `json:"sender_names,omitempty"`
}

type IVRSenderConfig struct {
	ID          string   `json:"id,omitempty"`
	SenderNames []string `json:"sender_names,omitempty"`
}

type TemplateGroup struct {
	ClientConfig
}

func NewTemplateGroup(clientConfig ClientConfig) *TemplateGroup {
	return &TemplateGroup{clientConfig}
}

func (t *TemplateGroup) Create(ctx context.Context, tg TemplateGroupModel) (*TemplateGroupResponse, error) { //nolint:dupl
	var response TemplateGroupResponse
	url := fmt.Sprintf("%s/%s", t.BaseURL, "templates-srv/groups")
	client, err := util.NewHTTPClient(url, http.MethodPost, t.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, tg)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (t *TemplateGroup) Update(ctx context.Context, tg TemplateGroupModel) (*TemplateGroupResponse, error) {
	var response TemplateGroupResponse
	url := fmt.Sprintf("%s/%s/%s", t.BaseURL, "templates-srv/groups", tg.GroupID)
	client, err := util.NewHTTPClient(url, http.MethodPut, t.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, tg)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (t *TemplateGroup) Get(ctx context.Context, groupID string) (*TemplateGroupResponse, error) {
	var response TemplateGroupResponse
	url := fmt.Sprintf("%s/%s/%s", t.BaseURL, "templates-srv/groups", groupID)
	client, err := util.NewHTTPClient(url, http.MethodGet, t.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, nil)
	if res != nil && res.StatusCode == http.StatusNoContent {
		resp := &TemplateGroupResponse{
			Status: http.StatusNoContent,
		}
		return resp, fmt.Errorf("%w: template group not found by the provider group_id  %s", util.ErrResourceNotFound, groupID)
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

func (t *TemplateGroup) Delete(ctx context.Context, groupID string) error {
	url := fmt.Sprintf("%s/%s/%s", t.BaseURL, "templates-srv/groups", groupID)
	client, err := util.NewHTTPClient(url, http.MethodDelete, t.AccessToken)
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
