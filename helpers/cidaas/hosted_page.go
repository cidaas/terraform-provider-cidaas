package cidaas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

type HostedPageModel struct {
	ID            string           `json:"_id,omitempty"`
	GroupOwner    string           `json:"groupOwner,omitempty"`
	DefaultLocale string           `json:"default_locale,omitempty"`
	HostedPages   []HostedPageData `json:"hosted_pages,omitempty"`
	CreatedTime   string           `json:"createdTime,omitempty"`
	UpdatedTime   string           `json:"updatedTime,omitempty"`
}

type HostedPageData struct {
	HostedPageID string `json:"hosted_page_id,omitempty"`
	Content      string `json:"content,omitempty"`
	Locale       string `json:"locale,omitempty"`
	URL          string `json:"url,omitempty"`
}

type HostedPageResponse struct {
	Success bool            `json:"success,omitempty"`
	Status  int             `json:"status,omitempty"`
	Data    HostedPageModel `json:"data,omitempty"`
}

type HostedPage struct {
	ClientConfig
}

func NewHostedPage(clientConfig ClientConfig) *HostedPage {
	return &HostedPage{clientConfig}
}

func (hp *HostedPage) Upsert(ctx context.Context, hpm HostedPageModel) (*HostedPageResponse, error) { //nolint:dupl
	var response HostedPageResponse
	url := fmt.Sprintf("%s/%s", hp.BaseURL, "hostedpages-srv/hpgroup")
	client, err := util.NewHTTPClient(url, http.MethodPost, hp.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, hpm)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (hp *HostedPage) Get(ctx context.Context, hpGroupName string) (*HostedPageResponse, error) { //nolint:dupl
	var response HostedPageResponse
	url := fmt.Sprintf("%s/%s/%s", hp.BaseURL, "hostedpages-srv/hpgroup", hpGroupName)
	client, err := util.NewHTTPClient(url, http.MethodGet, hp.AccessToken)
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

func (hp *HostedPage) Delete(ctx context.Context, hpGroupName string) error {
	url := fmt.Sprintf("%s/%s/%s", hp.BaseURL, "hostedpages-srv/hpgroup", hpGroupName)
	client, err := util.NewHTTPClient(url, http.MethodDelete, hp.AccessToken)
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
