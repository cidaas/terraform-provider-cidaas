package cidaas

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

const serviceSetupStatusInProgress = "in-progress"

// NotificationsSrvServiceSetup calls notification-srv /servicesetups.
type NotificationsSrvServiceSetup struct {
	ClientConfig
}

// NewNotificationsSrvServiceSetup builds a client for notification-srv service setup APIs.
func NewNotificationsSrvServiceSetup(cfg ClientConfig) *NotificationsSrvServiceSetup {
	return &NotificationsSrvServiceSetup{ClientConfig: cfg}
}

func (s *NotificationsSrvServiceSetup) segmentURL(parts ...string) string {
	return SegmentNotificationsURL(s.ClientConfig, parts...)
}

// NotificationsSrvCommProvider is the commProv fragment on serviceDescInfo.
type NotificationsSrvCommProvider struct {
	CommMethods []string `json:"commMethods,omitempty"`
}

// NotificationsSrvServiceDescInfo is serviceDescInfo on a service setup.
type NotificationsSrvServiceDescInfo struct {
	ServiceID       string                       `json:"serviceId"`
	ServiceCategory string                       `json:"serviceCategory,omitempty"`
	Name            string                       `json:"name,omitempty"`
	CommProv        NotificationsSrvCommProvider `json:"commProv,omitempty"`
}

// NotificationsSrvServiceSetupModel maps notification-srv service setup JSON.
type NotificationsSrvServiceSetupModel struct {
	ID                   string                          `json:"_id"`
	Name                 string                          `json:"name,omitempty"`
	Status               string                          `json:"status,omitempty"`
	Description          string                          `json:"description,omitempty"`
	ParentServiceSetupID string                          `json:"parentServiceSetupId,omitempty"`
	HasRemoteTemplates   bool                            `json:"hasRemoteTemplates,omitempty"`
	ServiceDescInfo      NotificationsSrvServiceDescInfo `json:"serviceDescInfo"`
}

// NotificationsSrvCreateServiceSetupRequest is POST /servicesetups body (no saasEntity — server injects tenant).
type NotificationsSrvCreateServiceSetupRequest struct {
	ServiceSetup NotificationsSrvServiceSetupWrite `json:"serviceSetup"`
}

// NotificationsSrvServiceSetupWrite is the writable subset for create.
type NotificationsSrvServiceSetupWrite struct {
	Name                 string                          `json:"name"`
	Description          string                          `json:"description,omitempty"`
	Status               string                          `json:"status,omitempty"`
	ParentServiceSetupID string                          `json:"parentServiceSetupId,omitempty"`
	HasRemoteTemplates   bool                            `json:"hasRemoteTemplates,omitempty"`
	ServiceDescInfo      NotificationsSrvServiceDescInfo `json:"serviceDescInfo"`
}

// NotificationsSrvUpdateServiceSetupRequest is PATCH /servicesetups body.
type NotificationsSrvUpdateServiceSetupRequest struct {
	ServiceSetup NotificationsSrvServiceSetupUpdate `json:"serviceSetup"`
}

// NotificationsSrvServiceSetupUpdate is the mutable subset for update (no status — verify is manual).
type NotificationsSrvServiceSetupUpdate struct {
	ID          string `json:"_id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// NotificationsSrvDeleteServiceSetupRequest is DELETE /servicesetups body.
type NotificationsSrvDeleteServiceSetupRequest struct {
	ID string `json:"id"`
}

// Get returns a service setup by id (GET /servicesetups/:id).
func (s *NotificationsSrvServiceSetup) Get(ctx context.Context, id string) (*NotificationsSrvServiceSetupModel, error) { //nolint:dupl
	escaped := url.PathEscape(id)
	urlStr := s.segmentURL("servicesetups", escaped)
	client, err := util.NewHTTPClient(urlStr, http.MethodGet, s.AccessToken)
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
		return nil, fmt.Errorf("failed to read servicesetup response body: %w", err)
	}
	return ParseNotificationSrvDataOrNil[NotificationsSrvServiceSetupModel](bodyBytes, res.StatusCode)
}

// List GET /servicesetups/
func (s *NotificationsSrvServiceSetup) List(ctx context.Context) ([]NotificationsSrvServiceSetupModel, error) {
	urlStr := s.segmentURL("servicesetups")
	client, err := util.NewHTTPClient(urlStr, http.MethodGet, s.AccessToken)
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
		return nil, fmt.Errorf("failed to read servicesetups body: %w", err)
	}
	out, err := ParseNotificationSrvData[[]NotificationsSrvServiceSetupModel](bodyBytes, res.StatusCode)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, nil
	}
	return *out, nil
}

// Create POST /servicesetups/ (status defaults to in-progress; tenant is injected by notification-srv).
func (s *NotificationsSrvServiceSetup) Create(ctx context.Context, write NotificationsSrvServiceSetupWrite) (*NotificationsSrvServiceSetupModel, error) {
	if strings.TrimSpace(write.Status) == "" {
		write.Status = serviceSetupStatusInProgress
	}
	req := NotificationsSrvCreateServiceSetupRequest{ServiceSetup: write}
	urlStr := s.segmentURL("servicesetups")
	client, err := util.NewHTTPClient(urlStr, http.MethodPost, s.AccessToken)
	if err != nil {
		return nil, err
	}
	status, bodyBytes, hdr, err := client.MakeRequestReadBody(ctx, req)
	if err != nil {
		return nil, err
	}
	switch status {
	case http.StatusOK, http.StatusCreated:
		return ParseNotificationSrvData[NotificationsSrvServiceSetupModel](bodyBytes, status)
	default:
		return nil, fmt.Errorf("unexpected status code %d, response body: %s%s", status, truncateBody(bodyBytes), util.XRefNumberSuffixFromHeader(hdr))
	}
}

// Update PATCH /servicesetups/ (name and description only).
func (s *NotificationsSrvServiceSetup) Update(ctx context.Context, update NotificationsSrvServiceSetupUpdate) (*NotificationsSrvServiceSetupModel, error) {
	req := NotificationsSrvUpdateServiceSetupRequest{ServiceSetup: update}
	urlStr := s.segmentURL("servicesetups")
	client, err := util.NewHTTPClient(urlStr, http.MethodPatch, s.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read servicesetup update body: %w", err)
	}
	return ParseNotificationSrvData[NotificationsSrvServiceSetupModel](bodyBytes, res.StatusCode)
}

// Delete DELETE /servicesetups/ with JSON body { "id": "..." }.
func (s *NotificationsSrvServiceSetup) Delete(ctx context.Context, id string) error {
	req := NotificationsSrvDeleteServiceSetupRequest{ID: id}
	urlStr := s.segmentURL("servicesetups")
	client, err := util.NewHTTPClient(urlStr, http.MethodDelete, s.AccessToken)
	if err != nil {
		return err
	}
	status, bodyBytes, hdr, err := client.MakeRequestReadBody(ctx, req)
	if err != nil {
		return err
	}
	switch status {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("%w: notification-srv: %s", util.ErrResourceNotFound, truncateBody(bodyBytes))
	default:
		return fmt.Errorf("unexpected status code %d, response body: %s%s", status, truncateBody(bodyBytes), util.XRefNumberSuffixFromHeader(hdr))
	}
}
