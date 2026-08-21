package cidaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

var (
	AllowedAuthType          = []string{"APIKEY", "TOTP", "CIDAAS_OAUTH2"}
	AllowedKeyPlacementValue = []string{"query", "header"}
)

type WebhookModel struct {
	ID                string        `json:"_id,omitempty"`
	AuthType          string        `json:"auth_type,omitempty"`
	URL               string        `json:"url,omitempty"`
	Events            []string      `json:"events,omitempty"`
	APIKeyDetails     APIKeyDetails `json:"apikeyDetails,omitempty"`
	TotpDetails       TotpDetails   `json:"totpDetails,omitempty"`
	CidaasAuthDetails AuthDetails   `json:"cidaasAuthDetails,omitempty"`
	Disable           bool          `json:"disable"`
	CreatedTime       string        `json:"createdTime,omitempty"`
	UpdatedTime       string        `json:"updatedTime,omitempty"`
}

type APIKeyDetails struct {
	ApikeyPlaceholder string `json:"apikey_placeholder,omitempty"`
	ApikeyPlacement   string `json:"apikey_placement,omitempty"`
	Apikey            string `json:"apikey,omitempty"`
}

type TotpDetails struct {
	TotpPlaceholder string `json:"totp_placeholder,omitempty"`
	TotpPlacement   string `json:"totp_placement,omitempty"`
	TotpKey         string `json:"totpkey,omitempty"`
}
type AuthDetails struct {
	ClientID string `json:"client_id,omitempty"`
}

type WebhookResponse struct {
	Success bool         `json:"success,omitempty"`
	Status  int          `json:"status,omitempty"`
	Data    WebhookModel `json:"data,omitempty"`
}

// EventDescriptionModel is a subset of webhook-srv event description fields used by Terraform.
type EventDescriptionModel struct {
	ID             string `json:"_id"`
	ObjectType     string `json:"objectType"`
	GoodForWebhook bool   `json:"goodForWebhook"`
}

type EventDescriptionsResponse struct {
	Success bool                    `json:"success,omitempty"`
	Status  int                     `json:"status,omitempty"`
	Data    []EventDescriptionModel `json:"data,omitempty"`
}

type Webhook struct {
	ClientConfig
}

func NewWebhook(clientConfig ClientConfig) *Webhook {
	return &Webhook{clientConfig}
}

const (
	webhookEndpoint           = "webhook-srv/webhook"
	eventDescriptionsEndpoint = "webhook-srv/eventdescriptions"
	webhookEventCategory      = "webhook"
)

func (w *Webhook) Upsert(ctx context.Context, wb WebhookModel) (*WebhookResponse, error) {
	res, err := w.makeRequest(ctx, http.MethodPost, webhookEndpoint, wb)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert webhook: %w", err)
	}
	defer res.Body.Close()

	var response WebhookResponse
	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (w *Webhook) Get(ctx context.Context, id string) (*WebhookResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("webhook ID cannot be empty")
	}
	var response WebhookResponse
	endpoint := fmt.Sprintf("%s?id=%s", webhookEndpoint, id)
	res, err := w.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}
	defer res.Body.Close()

	if err := util.ProcessResponse(res, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (w *Webhook) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("webhook ID cannot be empty")
	}
	endpoint := fmt.Sprintf("%s/%s", webhookEndpoint, id)
	res, err := w.makeRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	defer res.Body.Close()
	return nil
}

// ListEventDescriptions returns event descriptions for the given category
// (e.g. "webhook"). A 204 No Content response is treated as an empty list.
func (w *Webhook) ListEventDescriptions(ctx context.Context, category string) ([]EventDescriptionModel, error) {
	q := url.Values{}
	if category != "" {
		q.Set("category", category)
	}
	endpoint := eventDescriptionsEndpoint
	if encoded := q.Encode(); encoded != "" {
		endpoint = fmt.Sprintf("%s?%s", eventDescriptionsEndpoint, encoded)
	}

	reqURL := fmt.Sprintf("%s/%s", w.BaseURL, endpoint)
	client, err := util.NewHTTPClient(reqURL, http.MethodGet, w.AccessToken)
	if err != nil {
		return nil, err
	}
	status, body, hdr, err := client.MakeRequestReadBody(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list event descriptions: %w", err)
	}
	switch status {
	case http.StatusNoContent:
		return nil, nil
	case http.StatusOK:
		var response EventDescriptionsResponse
		if len(body) == 0 {
			return nil, nil
		}
		if err = json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to decode event descriptions response: %w", err)
		}
		if len(response.Data) == 0 {
			return nil, nil
		}
		return response.Data, nil
	default:
		return nil, fmt.Errorf("unexpected status code %d, response body: %s%s", status, string(body), util.XRefNumberSuffixFromHeader(hdr))
	}
}

// ListWebhookEventIDs returns the _id values of webhook-capable event descriptions.
func (w *Webhook) ListWebhookEventIDs(ctx context.Context) ([]string, error) {
	eds, err := w.ListEventDescriptions(ctx, webhookEventCategory)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(eds))
	for _, ed := range eds {
		if ed.ID != "" {
			ids = append(ids, ed.ID)
		}
	}
	return ids, nil
}
