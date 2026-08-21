package cidaas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
)

const fraudDetectionSettingsPath = "fraud-detection-srv/settings"

// BlockingSetting is the fraud-detection blocking list / allowlist configuration (JSON: blockingSetting).
// Slice fields use pointers so PATCH JSON can represent “clear list” as [] while omitting unset fields (nil pointer).
type BlockingSetting struct {
	Enabled                     *bool     `json:"enabled,omitempty"`
	BlackListedEmailDomains     *[]string `json:"blackListedEmailDomains,omitempty"`
	BlackListedIPs              *[]string `json:"blackListedIps,omitempty"`
	ExcludedEmailsFromBlackList *[]string `json:"excludedEmailsFromBlackList,omitempty"`
	ExcludedIPsFromBlackList    *[]string `json:"excludedIpsFromBlackList,omitempty"`
	Subs                        *[]string `json:"subs,omitempty"`
	WhiteListedEmailDomains     *[]string `json:"whiteListedEmailDomains,omitempty"`
	WhiteListedIPs              *[]string `json:"whiteListedIps,omitempty"`
	BlackListedIdentifiers      *[]string `json:"blackListedIdentifiers,omitempty"`
}

// RepeatedLoginBlockingMechanism configures brute-force style lockouts (JSON: repeatedLoginBlockingMechanism).
type RepeatedLoginBlockingMechanism struct {
	BlockingDurationInMin     *int64 `json:"blockingDurationInMin,omitempty"`
	BlockedCount              *int64 `json:"blockedCount,omitempty"`
	BlockedCountUnknownDevice *int64 `json:"blockedCountUnknownDevice,omitempty"`
}

// RuleConfiguration is the subset of ruleConfiguration exposed by the Terraform resource (GET/PATCH).
// Other API fields are ignored on decode and are not sent from the provider.
type RuleConfiguration struct {
	RepeatedLoginBlockingMechanismEnabled *bool `json:"repeatedLoginBlockingMechanismEnabled,omitempty"`
}

// SecuritySettingsPatch is the PATCH body for fraud-detection-srv settings (partial update).
type SecuritySettingsPatch struct {
	BlockingSetting                *BlockingSetting                `json:"blockingSetting,omitempty"`
	RepeatedLoginBlockingMechanism *RepeatedLoginBlockingMechanism `json:"repeatedLoginBlockingMechanism,omitempty"`
	RuleConfiguration              *RuleConfiguration              `json:"ruleConfiguration,omitempty"`
}

// SecuritySettingsGetResponse wraps GET fraud-detection-srv/settings.
type SecuritySettingsGetResponse struct {
	Success bool                  `json:"success"`
	Status  int                   `json:"status"`
	Data    *SecuritySettingsData `json:"data,omitempty"`
}

// SecuritySettingsData is the payload returned for tenant security / fraud settings.
// Extra JSON keys from the API (e.g. cspBotDetection) are ignored by encoding/json.
type SecuritySettingsData struct {
	BlockingSetting                *BlockingSetting                `json:"blockingSetting"`
	RepeatedLoginBlockingMechanism *RepeatedLoginBlockingMechanism `json:"repeatedLoginBlockingMechanism"`
	RuleConfiguration              *RuleConfiguration              `json:"ruleConfiguration"`
}

// SecuritySettingsPatchResponse is a typical success wrapper for PATCH.
type SecuritySettingsPatchResponse struct {
	Success *bool           `json:"success"`
	Status  int             `json:"status"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// SecuritySettings calls fraud-detection-srv settings APIs.
type SecuritySettings struct {
	ClientConfig
}

// NewSecuritySettings constructs the API helper.
func NewSecuritySettings(clientConfig ClientConfig) *SecuritySettings {
	return &SecuritySettings{clientConfig}
}

// Get loads current fraud-detection settings.
func (s *SecuritySettings) Get(ctx context.Context) (*SecuritySettingsGetResponse, error) {
	var response SecuritySettingsGetResponse
	url := fmt.Sprintf("%s/%s", s.BaseURL, fraudDetectionSettingsPath)
	client, err := util.NewHTTPClient(url, http.MethodGet, s.AccessToken)
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
	if !response.Success {
		return nil, fmt.Errorf("fraud-detection settings GET was not successful (success=false, status=%d)", response.Status)
	}
	return &response, nil
}

// Patch applies a partial update (PATCH) to fraud-detection settings.
func (s *SecuritySettings) Patch(ctx context.Context, patch SecuritySettingsPatch) error {
	url := fmt.Sprintf("%s/%s", s.BaseURL, fraudDetectionSettingsPath)
	client, err := util.NewHTTPClient(url, http.MethodPatch, s.AccessToken)
	if err != nil {
		return err
	}
	res, err := client.MakeRequest(ctx, patch)
	if err := util.HandleResponseError(res, err); err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read fraud-detection settings PATCH response: %w", err)
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}

	var response SecuritySettingsPatchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to decode fraud-detection settings PATCH response: %w", err)
	}
	if response.Success != nil && !*response.Success {
		return fmt.Errorf("fraud-detection settings PATCH was not successful (success=false, status=%d)", response.Status)
	}
	return nil
}
