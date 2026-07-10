package cidaas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

// NotificationsSrvProviderConfig calls notification-srv /providerconfigs.
type NotificationsSrvProviderConfig struct {
	ClientConfig
}

// NewNotificationsSrvProviderConfig builds a client for notification-srv provider config APIs.
func NewNotificationsSrvProviderConfig(cfg ClientConfig) *NotificationsSrvProviderConfig {
	return &NotificationsSrvProviderConfig{ClientConfig: cfg}
}

func (p *NotificationsSrvProviderConfig) segmentURL(parts ...string) string {
	return SegmentNotificationsURL(p.ClientConfig, parts...)
}

// NotificationsSrvProviderConfigModel maps notification-srv provider config JSON.
type NotificationsSrvProviderConfigModel struct {
	ID         string          `json:"_id"`
	ConfigData json.RawMessage `json:"configData"`
}

// Get returns provider config for a service setup id (GET /providerconfigs/:id).
func (p *NotificationsSrvProviderConfig) Get(ctx context.Context, serviceSetupID string) (*NotificationsSrvProviderConfigModel, error) {
	escaped := url.PathEscape(serviceSetupID)
	urlStr := p.segmentURL("providerconfigs", escaped)
	client, err := util.NewHTTPClient(urlStr, http.MethodGet, p.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read providerconfig response body: %w", err)
	}
	return ParseNotificationSrvDataOrNil[NotificationsSrvProviderConfigModel](bodyBytes, res.StatusCode)
}

// Create POST /providerconfigs/ (upsert by _id).
func (p *NotificationsSrvProviderConfig) Create(ctx context.Context, serviceSetupID string, configData json.RawMessage) (*NotificationsSrvProviderConfigModel, error) {
	req := NotificationsSrvProviderConfigModel{
		ID:         serviceSetupID,
		ConfigData: configData,
	}
	urlStr := p.segmentURL("providerconfigs")
	client, err := util.NewHTTPClient(urlStr, http.MethodPost, p.AccessToken)
	if err != nil {
		return nil, err
	}
	status, bodyBytes, hdr, err := client.MakeRequestReadBody(ctx, req)
	if err != nil {
		return nil, err
	}
	switch status {
	case http.StatusOK, http.StatusCreated:
		return ParseNotificationSrvData[NotificationsSrvProviderConfigModel](bodyBytes, status)
	default:
		return nil, fmt.Errorf("unexpected status code %d, response body: %s%s", status, truncateBody(bodyBytes), util.XRefNumberSuffixFromHeader(hdr))
	}
}
