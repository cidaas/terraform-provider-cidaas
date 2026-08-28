package cidaas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

// DefaultNotificationsContextPath is the default first path segment for notification-srv (see notification-srv CONTEXT_PATH).
const DefaultNotificationsContextPath = "notifications-srv"

// NotificationsSrvTemplateGroup calls notification-srv templategroups REST API (not legacy templates-srv/groups).
type NotificationsSrvTemplateGroup struct {
	ClientConfig
	ContextPath string
}

// NewNotificationsSrvTemplateGroup builds a client for notification-srv template group endpoints.
func NewNotificationsSrvTemplateGroup(cfg ClientConfig) *NotificationsSrvTemplateGroup {
	return &NotificationsSrvTemplateGroup{
		ClientConfig: cfg,
		ContextPath:  NormalizeNotificationsContextPath(cfg),
	}
}

func (t *NotificationsSrvTemplateGroup) segmentURL(parts ...string) string {
	return SegmentNotificationsURL(t.ClientConfig, parts...)
}

// --- Request / response DTOs (aligned with notification-srv template.Group + dto.TemplateGroupRequest)

// NotificationsSrvCommSetting maps template.CommSetting JSON.
type NotificationsSrvCommSetting struct {
	CommunicationMethod string `json:"communicationMethod"`
	ServiceSetupID      string `json:"serviceSetupId"`
	SenderName          string `json:"senderName,omitempty"`
	SenderAddress       string `json:"senderAddress,omitempty"`
	ReplyTo             string `json:"replyTo,omitempty"`
	HasRemoteTemplates  *bool  `json:"hasRemoteTemplates,omitempty"`
}

// NotificationsSrvLocaleMapping maps dto.LocaleMapping.
type NotificationsSrvLocaleMapping struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

// NotificationsSrvCopy maps dto.Copy.
type NotificationsSrvCopy struct {
	FromGroupID string                          `json:"fromGroupID,omitempty"`
	Locale      []NotificationsSrvLocaleMapping `json:"locale,omitempty"`
}

// NotificationsSrvTemplateGroupRequest is the JSON body for POST/PUT templategroups.
type NotificationsSrvTemplateGroupRequest struct {
	ID            string                                 `json:"_id,omitempty"`
	TGType        string                                 `json:"tgType,omitempty"`
	Description   string                                 `json:"description,omitempty"`
	Owner         string                                 `json:"owner,omitempty"`
	DefaultLocale string                                 `json:"defaultLocale,omitempty"`
	CommSettings  map[string]NotificationsSrvCommSetting `json:"commSettings,omitempty"`
	Copy          *NotificationsSrvCopy                  `json:"copy,omitempty"`
}

// NotificationsSrvCopyStats maps dto.CopyStats (response).
type NotificationsSrvCopyStats struct {
	FromLocale      string `json:"fromLocale"`
	FromLocaleCount int64  `json:"fromLocaleCount"`
	ToLocale        string `json:"toLocale"`
	ToLocaleCount   int64  `json:"toLocaleCount"`
}

// NotificationsSrvTemplateGroupData is the `data` object returned by the API (Group + optional CopyStats).
type NotificationsSrvTemplateGroupData struct {
	ID            string                                 `json:"_id"`
	TGType        string                                 `json:"tgType,omitempty"`
	Description   string                                 `json:"description,omitempty"`
	Owner         string                                 `json:"owner,omitempty"`
	DefaultLocale string                                 `json:"defaultLocale,omitempty"`
	CommSettings  map[string]NotificationsSrvCommSetting `json:"commSettings,omitempty"`
	CopyStats     []NotificationsSrvCopyStats            `json:"CopyStats,omitempty"`
}

func parseNotificationSrvResponse(body []byte, statusCode int) (*NotificationsSrvTemplateGroupData, error) {
	return ParseNotificationSrvData[NotificationsSrvTemplateGroupData](body, statusCode)
}

func truncateBody(b []byte) string {
	if len(b) > 200 {
		return string(b[:200]) + "..."
	}
	return string(b)
}

// Create posts a new template group (POST). Triggers copy-from-source when the API rules apply.
func (t *NotificationsSrvTemplateGroup) Create(ctx context.Context, req NotificationsSrvTemplateGroupRequest) (*NotificationsSrvTemplateGroupData, error) {
	return t.post(ctx, t.segmentURL("templategroups"), req)
}

// Get loads a template group by id (GET).
func (t *NotificationsSrvTemplateGroup) Get(ctx context.Context, groupID string) (*NotificationsSrvTemplateGroupData, error) {
	escaped := url.PathEscape(groupID)
	urlStr := t.segmentURL("templategroups", escaped)
	client, err := util.NewHTTPClient(urlStr, http.MethodGet, t.AccessToken)
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
		return nil, fmt.Errorf("failed to read templategroup response body: %w", err)
	}
	return parseNotificationSrvResponse(bodyBytes, res.StatusCode)
}

// Update updates an existing template group (PUT .../templategroups/:id).
func (t *NotificationsSrvTemplateGroup) Update(ctx context.Context, groupID string, req NotificationsSrvTemplateGroupRequest) (*NotificationsSrvTemplateGroupData, error) {
	req.ID = groupID
	escaped := url.PathEscape(groupID)
	urlStr := t.segmentURL("templategroups", escaped)
	client, err := util.NewHTTPClient(urlStr, http.MethodPut, t.AccessToken)
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
		return nil, fmt.Errorf("failed to read templategroup response body: %w", err)
	}
	return parseNotificationSrvResponse(bodyBytes, res.StatusCode)
}

func (t *NotificationsSrvTemplateGroup) post(ctx context.Context, urlStr string, body interface{}) (*NotificationsSrvTemplateGroupData, error) {
	client, err := util.NewHTTPClient(urlStr, http.MethodPost, t.AccessToken)
	if err != nil {
		return nil, err
	}
	res, err := client.MakeRequest(ctx, body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read templategroup response body: %w", err)
	}
	return parseNotificationSrvResponse(bodyBytes, res.StatusCode)
}

// Delete removes a template group (DELETE .../templategroups/:id).
func (t *NotificationsSrvTemplateGroup) Delete(ctx context.Context, groupID string) error {
	escaped := url.PathEscape(groupID)
	urlStr := t.segmentURL("templategroups", escaped)
	client, err := util.NewHTTPClient(urlStr, http.MethodDelete, t.AccessToken)
	if err != nil {
		return err
	}
	res, err := client.MakeRequest(ctx, nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read templategroup delete body: %w", err)
	}
	var env notificationSrvEnvelope
	if err := json.Unmarshal(bodyBytes, &env); err == nil && !env.Success && env.ErrorMsg != "" {
		return fmt.Errorf("notification-srv delete error: %s", env.ErrorMsg)
	}
	return nil
}

// FindGraphGroups POST /graph/templategroups/ with graph filter body.
func (t *NotificationsSrvTemplateGroup) FindGraphGroups(ctx context.Context, filter json.RawMessage) ([]NotificationsSrvTemplateGroupData, error) { //nolint:dupl
	urlStr := t.segmentURL("graph", "templategroups")
	client, err := util.NewHTTPClient(urlStr, http.MethodPost, t.AccessToken)
	if err != nil {
		return nil, err
	}
	var body interface{}
	if len(filter) > 0 {
		body = filter
	}
	res, err := client.MakeRequest(ctx, body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read graph/templategroups body: %w", err)
	}
	out, err := ParseNotificationSrvData[[]NotificationsSrvTemplateGroupData](bodyBytes, res.StatusCode)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, nil
	}
	return *out, nil
}

// GetTemplateFilters GET /templategroups/:id/templatefilters — returns raw JSON `data` payload.
func (t *NotificationsSrvTemplateGroup) GetTemplateFilters(ctx context.Context, groupID string) (json.RawMessage, error) {
	escaped := url.PathEscape(groupID)
	urlStr := t.segmentURL("templategroups", escaped, "templatefilters")
	client, err := util.NewHTTPClient(urlStr, http.MethodGet, t.AccessToken)
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
		return nil, fmt.Errorf("failed to read templatefilters body: %w", err)
	}
	var env notificationSrvEnvelope
	if err := json.Unmarshal(bodyBytes, &env); err != nil {
		return nil, fmt.Errorf("failed to parse templatefilters response: %w", err)
	}
	if !env.Success {
		errMsg := env.ErrorMsg
		if errMsg == "" {
			errMsg = env.ErrorAlt
		}
		if errMsg != "" {
			return nil, fmt.Errorf("notification-srv templatefilters error: %s", errMsg)
		}
		return nil, fmt.Errorf("notification-srv templatefilters unsuccessful: %s", string(bodyBytes))
	}
	return env.Data, nil
}

// NotificationsSrvTemplateFiltersData is the `data` object from GET templategroups/:id/templatefilters.
type NotificationsSrvTemplateFiltersData struct {
	Locales []string `json:"locales"`
}

// ParseTemplateFiltersLocales unmarshals templatefilters `data` and returns locale codes.
func ParseTemplateFiltersLocales(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var data NotificationsSrvTemplateFiltersData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("failed to parse templatefilters data: %w", err)
	}
	return data.Locales, nil
}

// ListTemplateFiltersLocales GET …/templategroups/:id/templatefilters and returns locale codes.
func (t *NotificationsSrvTemplateGroup) ListTemplateFiltersLocales(ctx context.Context, groupID string) ([]string, error) {
	raw, err := t.GetTemplateFilters(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return ParseTemplateFiltersLocales(raw)
}

// CopyLocales PUT …/templategroups/:id with copy.locale[] to seed templates for target locales.
func (t *NotificationsSrvTemplateGroup) CopyLocales(ctx context.Context, groupID string, localeCopy NotificationsSrvCopy) error {
	_, err := t.Update(ctx, groupID, NotificationsSrvTemplateGroupRequest{
		ID:   groupID,
		Copy: &localeCopy,
	})
	return err
}

// IsNotificationSrvTemplatesAlreadyExistError reports API errors when templates already exist for target locales.
func IsNotificationSrvTemplatesAlreadyExistError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already templates found") || strings.Contains(msg, "already template")
}
