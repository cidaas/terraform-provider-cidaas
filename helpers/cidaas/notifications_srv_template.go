package cidaas

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

// ErrNotificationTemplateAlreadyExists is returned by Create when POST /templates/ responds with HTTP 409
// (template document id already present). Callers may retry with PUT /templates/:id using SyntheticTemplateDocumentID.
var ErrNotificationTemplateAlreadyExists = errors.New("notification-srv: template already exists (HTTP 409)")

// NotificationsSrvTemplate calls notification-srv /templates and POST graph/templates/ (not legacy templates-srv).
type NotificationsSrvTemplate struct {
	ClientConfig
}

// NewNotificationsSrvTemplate builds a client for notification-srv template endpoints.
func NewNotificationsSrvTemplate(cfg ClientConfig) *NotificationsSrvTemplate {
	return &NotificationsSrvTemplate{ClientConfig: cfg}
}

func (t *NotificationsSrvTemplate) segmentURL(parts ...string) string {
	return SegmentNotificationsURL(t.ClientConfig, parts...)
}

// NotificationsSrvTemplateModel maps notification-srv template.Template JSON for REST.
type NotificationsSrvTemplateModel struct {
	ID                  string   `json:"_id,omitempty"`
	GroupID             string   `json:"groupId,omitempty"`
	TemplateKey         string   `json:"templateKey,omitempty"`
	CommunicationMethod string   `json:"communicationMethod,omitempty"`
	ProcessingType      string   `json:"processingType,omitempty"`
	UsageType           string   `json:"usageType,omitempty"`
	VerificationType    string   `json:"verificationType,omitempty"`
	Locale              string   `json:"locale,omitempty"`
	Owner               string   `json:"owner,omitempty"`
	Number              *int     `json:"number,omitempty"`
	MessageFormat       string   `json:"messageFormat,omitempty"`
	Subject             string   `json:"subject,omitempty"`
	Content             string   `json:"content,omitempty"`
	ContentData         string   `json:"contentData,omitempty"`
	LastSeededBy        string   `json:"lastSeededBy,omitempty"`
	UserGroupIDs        []string `json:"userGroupIds,omitempty"`
	Enabled             bool     `json:"enabled"`
	Description         string   `json:"description,omitempty"`
	IsDraft             bool     `json:"isDraft,omitempty"`
}

// Get returns a template by document id (GET /templates/:id).
func (t *NotificationsSrvTemplate) Get(ctx context.Context, id string) (*NotificationsSrvTemplateModel, error) { //nolint:dupl
	escaped := url.PathEscape(id)
	urlStr := t.segmentURL("templates", escaped)
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
		return nil, fmt.Errorf("failed to read template response body: %w", err)
	}
	return ParseNotificationSrvDataOrNil[NotificationsSrvTemplateModel](bodyBytes, res.StatusCode)
}

// GetAllowNotFound returns the template for GET /templates/:id, or (nil, nil) when the server responds with HTTP 404.
func (t *NotificationsSrvTemplate) GetAllowNotFound(ctx context.Context, id string) (*NotificationsSrvTemplateModel, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("template id is empty")
	}
	escaped := url.PathEscape(id)
	urlStr := t.segmentURL("templates", escaped)
	client, err := util.NewHTTPClient(urlStr, http.MethodGet, t.AccessToken)
	if err != nil {
		return nil, err
	}
	status, bodyBytes, hdr, err := client.MakeRequestReadBody(ctx, nil)
	if err != nil {
		return nil, err
	}
	switch status {
	case http.StatusOK:
		return ParseNotificationSrvData[NotificationsSrvTemplateModel](bodyBytes, status)
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, fmt.Errorf("unexpected status code %d, response body: %s%s", status, truncateBody(bodyBytes), util.XRefNumberSuffixFromHeader(hdr))
	}
}

// Create POST /templates/
func (t *NotificationsSrvTemplate) Create(ctx context.Context, req NotificationsSrvTemplateModel) (*NotificationsSrvTemplateModel, error) {
	urlStr := t.segmentURL("templates")
	client, err := util.NewHTTPClient(urlStr, http.MethodPost, t.AccessToken)
	if err != nil {
		return nil, err
	}
	status, bodyBytes, hdr, err := client.MakeRequestReadBody(ctx, req)
	if err != nil {
		return nil, err
	}
	switch status {
	case http.StatusOK, http.StatusCreated:
		return ParseNotificationSrvData[NotificationsSrvTemplateModel](bodyBytes, status)
	case http.StatusConflict:
		return nil, fmt.Errorf("%w: %s%s", ErrNotificationTemplateAlreadyExists, truncateBody(bodyBytes), util.XRefNumberSuffixFromHeader(hdr))
	default:
		return nil, fmt.Errorf("unexpected status code %d, response body: %s%s", status, truncateBody(bodyBytes), util.XRefNumberSuffixFromHeader(hdr))
	}
}

// Update PUT /templates/:id
func (t *NotificationsSrvTemplate) Update(ctx context.Context, id string, req NotificationsSrvTemplateModel) (*NotificationsSrvTemplateModel, error) {
	req.ID = id
	escaped := url.PathEscape(id)
	urlStr := t.segmentURL("templates", escaped)
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
		return nil, fmt.Errorf("failed to read template update body: %w", err)
	}
	return ParseNotificationSrvData[NotificationsSrvTemplateModel](bodyBytes, res.StatusCode)
}

// notificationSrvCodeTemplateCannotDelete is returned when DELETE is not allowed (e.g. system templates).
const notificationSrvCodeTemplateCannotDelete = "35013"

// Delete DELETE /templates/:id.
// If the server refuses deletion (HTTP 400, code 35013), returns privilegedNoOp=true and err=nil so Terraform
// can still remove the resource from state (e.g. tainted replace); the remote template remains.
func (t *NotificationsSrvTemplate) Delete(ctx context.Context, id string) (bool, error) {
	escaped := url.PathEscape(id)
	urlStr := t.segmentURL("templates", escaped)
	client, err := util.NewHTTPClient(urlStr, http.MethodDelete, t.AccessToken)
	if err != nil {
		return false, err
	}
	status, bodyBytes, hdr, err := client.MakeRequestReadBody(ctx, nil)
	if err != nil {
		return false, err
	}
	switch status {
	case http.StatusOK, http.StatusNoContent, http.StatusAccepted, http.StatusCreated:
		return false, nil
	case http.StatusBadRequest:
		if notificationSrvDeleteRefused(bodyBytes) {
			return true, nil
		}
		return false, fmt.Errorf("notification-srv delete template error: %s%s", truncateBody(bodyBytes), util.XRefNumberSuffixFromHeader(hdr))
	default:
		return false, fmt.Errorf("unexpected status code %d, response body: %s%s", status, truncateBody(bodyBytes), util.XRefNumberSuffixFromHeader(hdr))
	}
}

func notificationSrvDeleteRefused(body []byte) bool {
	var env notificationSrvEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return false
	}
	if env.Code == notificationSrvCodeTemplateCannotDelete {
		return true
	}
	msg := strings.ToLower(env.ErrorMsg + " " + env.ErrorAlt)
	return strings.Contains(msg, "can not be deleted") || strings.Contains(msg, "cannot be deleted")
}

// DeleteByGroupAndLocales DELETE /templates?groupId=&locale=
func (t *NotificationsSrvTemplate) DeleteByGroupAndLocales(ctx context.Context, groupID string, locales []string) error {
	base := t.segmentURL("templates")
	u, err := url.Parse(base)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("groupId", groupID)
	for _, loc := range locales {
		if strings.TrimSpace(loc) != "" {
			q.Add("locale", loc)
		}
	}
	u.RawQuery = q.Encode()
	client, err := util.NewHTTPClient(u.String(), http.MethodDelete, t.AccessToken)
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
		return fmt.Errorf("failed to read bulk delete body: %w", err)
	}
	var env notificationSrvEnvelope
	if err := json.Unmarshal(bodyBytes, &env); err == nil && !env.Success && env.ErrorMsg != "" {
		return fmt.Errorf("notification-srv bulk delete templates error: %s", env.ErrorMsg)
	}
	return nil
}

// FindGraph POST /graph/templates/ with a graph filter JSON body (GenericGraphFilter).
func (t *NotificationsSrvTemplate) FindGraph(ctx context.Context, filter json.RawMessage) ([]NotificationsSrvTemplateModel, error) { //nolint:dupl
	urlStr := t.segmentURL("graph", "templates")
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
		return nil, fmt.Errorf("failed to read graph/templates body: %w", err)
	}
	out, err := ParseNotificationSrvData[[]NotificationsSrvTemplateModel](bodyBytes, res.StatusCode)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, nil
	}
	return *out, nil
}
