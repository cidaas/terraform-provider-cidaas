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

type NotificationsSrvTemplateGroup struct {
	ClientConfig
	ContextPath string
}

func NewNotificationsSrvTemplateGroup(cfg ClientConfig) *NotificationsSrvTemplateGroup {
	return &NotificationsSrvTemplateGroup{
		ClientConfig: cfg,
		ContextPath:  NormalizeNotificationsContextPath(cfg),
	}
}

func (t *NotificationsSrvTemplateGroup) segmentURL(parts ...string) string {
	return SegmentNotificationsURL(t.ClientConfig, parts...)
}

type NotificationsSrvCommSetting struct {
	CommunicationMethod string `json:"communicationMethod"`
	ServiceSetupID      string `json:"serviceSetupId"`
	SenderName          string `json:"senderName,omitempty"`
	SenderAddress       string `json:"senderAddress,omitempty"`
	ReplyTo             string `json:"replyTo,omitempty"`
	HasRemoteTemplates  *bool  `json:"hasRemoteTemplates,omitempty"`
}

type NotificationsSrvLocaleMapping struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

type NotificationsSrvCopy struct {
	FromGroupID string                          `json:"fromGroupID,omitempty"`
	Locale      []NotificationsSrvLocaleMapping `json:"locale,omitempty"`
}

type NotificationsSrvTemplateGroupRequest struct {
	ID            string                                 `json:"_id,omitempty"`
	TGType        string                                 `json:"tgType,omitempty"`
	Description   string                                 `json:"description,omitempty"`
	Owner         string                                 `json:"owner,omitempty"`
	DefaultLocale string                                 `json:"defaultLocale,omitempty"`
	CommSettings  map[string]NotificationsSrvCommSetting `json:"commSettings,omitempty"`
	Copy          *NotificationsSrvCopy                  `json:"copy,omitempty"`
}

type NotificationsSrvCopyStats struct {
	FromLocale      string `json:"fromLocale"`
	FromLocaleCount int64  `json:"fromLocaleCount"`
	ToLocale        string `json:"toLocale"`
	ToLocaleCount   int64  `json:"toLocaleCount"`
}

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

func (t *NotificationsSrvTemplateGroup) Create(ctx context.Context, req NotificationsSrvTemplateGroupRequest) (*NotificationsSrvTemplateGroupData, error) {
	return t.post(ctx, t.segmentURL("templategroups"), req)
}

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
	defer func() { _ = res.Body.Close() }()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read templategroup response body: %w", err)
	}
	return parseNotificationSrvResponse(bodyBytes, res.StatusCode)
}

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
	defer func() { _ = res.Body.Close() }()
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
	defer func() { _ = res.Body.Close() }()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read templategroup response body: %w", err)
	}
	return parseNotificationSrvResponse(bodyBytes, res.StatusCode)
}

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
	defer func() { _ = res.Body.Close() }()
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
	defer func() { _ = res.Body.Close() }()
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
	defer func() { _ = res.Body.Close() }()
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

type NotificationsSrvTemplateFiltersData struct {
	Locales []string `json:"locales"`
}

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

func (t *NotificationsSrvTemplateGroup) ListTemplateFiltersLocales(ctx context.Context, groupID string) ([]string, error) {
	raw, err := t.GetTemplateFilters(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return ParseTemplateFiltersLocales(raw)
}

func (t *NotificationsSrvTemplateGroup) CopyLocales(ctx context.Context, groupID string, localeCopy NotificationsSrvCopy) error {
	_, err := t.Update(ctx, groupID, NotificationsSrvTemplateGroupRequest{
		ID:   groupID,
		Copy: &localeCopy,
	})
	return err
}

func IsNotificationSrvTemplatesAlreadyExistError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already templates found") || strings.Contains(msg, "already template")
}
