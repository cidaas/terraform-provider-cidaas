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

type TemplateTypeModel struct {
	ID                   string            `json:"_id,omitempty"`
	Category             string            `json:"category,omitempty"`
	Owner                string            `json:"owner,omitempty"`
	Description          string            `json:"description,omitempty"`
	Deactivatable        bool              `json:"deactivatable,omitempty"`
	SystemAttributes     map[string]string `json:"systemAttributes,omitempty"`
	CustomAttributes     map[string]string `json:"customAttributes,omitempty"`
	ContextAttributes    map[string]string `json:"contextAttributes,omitempty"`
	ProcessingTypes      []string          `json:"processingTypes,omitempty"`
	UsageTypes           []string          `json:"usageTypes,omitempty"`
	VerificationTypes    []string          `json:"verificationTypes,omitempty"`
	CommunicationMethods []string          `json:"communicationMethods,omitempty"`
	TemplateGroupIDs     []string          `json:"templateGroupIds,omitempty"`
	MsgFormats           []string          `json:"msgFormats,omitempty"`
	CreatedTime          string            `json:"createdTime,omitempty"`
	UpdatedTime          string            `json:"updatedTime,omitempty"`
}

type TemplateTypePatchModel struct {
	ID               string             `json:"_id,omitempty"`
	CustomAttributes *map[string]string `json:"customAttributes,omitempty"`
}

type TemplateTypeResponse struct {
	Success bool              `json:"success,omitempty"`
	Status  int               `json:"status,omitempty"`
	Data    TemplateTypeModel `json:"data,omitempty"`
}

var _ TemplateTypeService = &TemplateTypeServiceImpl{}

type TemplateTypeServiceImpl struct {
	ClientConfig
}

type TemplateTypeService interface {
	Upsert(templateType TemplateTypeModel) (*TemplateTypeResponse, error)
	Get(id string) (*TemplateTypeResponse, error)
	Patch(patch TemplateTypePatchModel) (*TemplateTypeResponse, error)
	Delete(id string) error
}

func NewTemplateType(clientConfig ClientConfig) *TemplateTypeServiceImpl {
	return &TemplateTypeServiceImpl{clientConfig}
}

func (t *TemplateTypeServiceImpl) Upsert(templateType TemplateTypeModel) (*TemplateTypeResponse, error) {
	var response TemplateTypeResponse
	u := SegmentNotificationsURL(t.ClientConfig, "templatetypes")
	httpClient, err := util.NewHTTPClient(u, http.MethodPost, t.AccessToken)
	if err != nil {
		return nil, err
	}

	res, err := httpClient.MakeRequest(context.Background(), templateType)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read template type response body: %w", err)
	}
	bodyString := string(bodyBytes)
	if bodyString == "" {
		return nil, fmt.Errorf("response code %d with empty response body", res.StatusCode)
	}

	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json body %s, status code %d, error %s", bodyString, res.StatusCode, err.Error())
	}
	return &response, nil
}

func (t *TemplateTypeServiceImpl) Get(id string) (*TemplateTypeResponse, error) {
	var response TemplateTypeResponse
	id = strings.ToUpper(id)
	u := SegmentNotificationsURL(t.ClientConfig, "templatetypes", url.PathEscape(id))
	httpClient, err := util.NewHTTPClient(u, http.MethodGet, t.AccessToken)
	if err != nil {
		return nil, err
	}

	res, err := httpClient.MakeRequest(context.Background(), nil)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("template type not found with id %s", id)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read template type response body: %w", err)
	}
	bodyString := string(bodyBytes)
	if bodyString == "" {
		return nil, fmt.Errorf("response code %d with empty response body", res.StatusCode)
	}

	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json body %s, status code %d, error %s", bodyString, res.StatusCode, err.Error())
	}
	return &response, nil
}

func (t *TemplateTypeServiceImpl) Patch(patch TemplateTypePatchModel) (*TemplateTypeResponse, error) {
	var response TemplateTypeResponse
	id := strings.ToUpper(patch.ID)
	u := SegmentNotificationsURL(t.ClientConfig, "templatetypes", url.PathEscape(id))
	httpClient, err := util.NewHTTPClient(u, http.MethodPatch, t.AccessToken)
	if err != nil {
		return nil, err
	}

	res, err := httpClient.MakeRequest(context.Background(), patch)
	if err := util.HandleResponseError(res, err); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read template type response body: %w", err)
	}
	bodyString := string(bodyBytes)
	if bodyString == "" {
		return nil, fmt.Errorf("response code %d with empty response body", res.StatusCode)
	}

	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json body %s, status code %d, error %s", bodyString, res.StatusCode, err.Error())
	}
	return &response, nil
}

func (t *TemplateTypeServiceImpl) Delete(id string) error {
	id = strings.ToUpper(id)
	u := SegmentNotificationsURL(t.ClientConfig, "templatetypes", url.PathEscape(id))
	httpClient, err := util.NewHTTPClient(u, http.MethodDelete, t.AccessToken)
	if err != nil {
		return err
	}

	res, err := httpClient.MakeRequest(context.Background(), nil)
	if err := util.HandleResponseError(res, err); err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

// FindGraphTemplateTypes POST /graph/templatetypes/ with graph filter body.
func (t *TemplateTypeServiceImpl) FindGraphTemplateTypes(ctx context.Context, filter json.RawMessage) ([]TemplateTypeModel, error) {
	u := SegmentNotificationsURL(t.ClientConfig, "graph", "templatetypes")
	httpClient, err := util.NewHTTPClient(u, http.MethodPost, t.AccessToken)
	if err != nil {
		return nil, err
	}
	var body interface{}
	if len(filter) > 0 {
		body = filter
	}
	res, err := httpClient.MakeRequest(ctx, body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read graph/templatetypes body: %w", err)
	}
	out, err := ParseNotificationSrvData[[]TemplateTypeModel](bodyBytes, res.StatusCode)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, nil
	}
	return *out, nil
}
