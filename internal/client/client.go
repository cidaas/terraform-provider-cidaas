package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
)

// Client is the root API client for the cidaas v4 Terraform provider.
type Client struct {
	Config                    Config
	Capabilities              Capabilities
	CidaasClient              *cidaas.Client
	HostedPages               *HostedPageGroupService
	Themes                    *ThemeService
	Translations              *TranslationsService
	Layouts                   *HostedPageLayoutService
	UserSetup                 *UserSetupService
	FieldSetup                *FieldSetupService
	SuggestVerificationMethod *SuggestVerificationMethodService
	VerificationOptions       *VerificationOptionsService
	AppConfiguration          *AppConfigurationService
}

type Config struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
	AccessToken  string
}

type Capabilities struct {
	SupportsV4    bool
	Version       string
	TargetVersion string
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

// NewClient authenticates via client_credentials and probes tenant version.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base_url is required")
	}
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, fmt.Errorf("client_id and client_secret are required")
	}

	tokenURL, err := url.JoinPath(cfg.BaseURL, "token-srv/token")
	if err != nil {
		return nil, fmt.Errorf("token url: %w", err)
	}
	httpClient := NewHTTPClient(tokenURL, http.MethodPost, "")
	resp, err := httpClient.DoJSON(ctx, map[string]string{
		"client_id":     cfg.ClientID,
		"client_secret": cfg.ClientSecret,
		"grant_type":    "client_credentials",
	})
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err := ExpectStatus(resp, http.StatusOK); err != nil {
		return nil, fmt.Errorf("token: %w", err)
	}
	var token TokenResponse
	if err := DecodeJSON(resp, &token); err != nil {
		return nil, fmt.Errorf("token decode: %w", err)
	}
	if token.AccessToken == "" {
		return nil, fmt.Errorf("token response missing access_token")
	}
	cfg.AccessToken = token.AccessToken

	caps := Capabilities{SupportsV4: true}

	cidaasClient, err := cidaas.NewClient(ctx, cidaas.ClientConfig{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		BaseURL:      cfg.BaseURL,
		AccessToken:  cfg.AccessToken,
	})
	if err != nil {
		return nil, fmt.Errorf("v3 cidaas client init: %w", err)
	}

	c := &Client{
		Config:       cfg,
		Capabilities: caps,
		CidaasClient: cidaasClient,
	}
	c.HostedPages = NewHostedPageGroupService(cfg)
	c.Themes = NewThemeService(cfg)
	c.Translations = NewTranslationsService(cfg)
	c.Layouts = NewHostedPageLayoutService(cfg)
	c.UserSetup = NewUserSetupService(cfg)
	c.FieldSetup = NewFieldSetupService(cfg)
	c.SuggestVerificationMethod = NewSuggestVerificationMethodService(cfg)
	c.VerificationOptions = NewVerificationOptionsService(cfg)
	c.AppConfiguration = NewAppConfigurationService(cfg)
	return c, nil
}

// NormalizeVersion converts inputs like "v3", "3", "3.x" to "3.x" and "v4", "4", "4.x" to "4.x".
func NormalizeVersion(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if strings.HasPrefix(v, "v3") || strings.HasPrefix(v, "3") {
		return "3.x"
	}
	if strings.HasPrefix(v, "v4") || strings.HasPrefix(v, "4") {
		return "4.x"
	}
	return v
}
