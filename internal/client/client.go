package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Client is the root API client for the cidaas v4 Terraform provider.
type Client struct {
	Config                    Config
	Capabilities              Capabilities
	HostedPages               *HostedPageGroupService
	Themes                    *ThemeService
	Translations              *TranslationsService
	Layouts                   *HostedPageLayoutService
	UserSetup                 *UserSetupService
	FieldSetup                *FieldSetupService
	SuggestVerificationMethod *SuggestVerificationMethodService
	VerificationOptions       *VerificationOptionsService
}

type Config struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
	AccessToken  string
}

type Capabilities struct {
	SupportsV4 bool
	Version    string
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type versionAPIResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Version string `json:"version"`
	} `json:"data"`
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
	defer resp.Body.Close()
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

	caps, err := probeCapabilities(ctx, cfg)
	if err != nil {
		return nil, err
	}

	c := &Client{
		Config:       cfg,
		Capabilities: caps,
	}
	c.HostedPages = NewHostedPageGroupService(cfg)
	c.Themes = NewThemeService(cfg)
	c.Translations = NewTranslationsService(cfg)
	c.Layouts = NewHostedPageLayoutService(cfg)
	c.UserSetup = NewUserSetupService(cfg)
	c.FieldSetup = NewFieldSetupService(cfg)
	c.SuggestVerificationMethod = NewSuggestVerificationMethodService(cfg)
	c.VerificationOptions = NewVerificationOptionsService(cfg)
	return c, nil
}

func probeCapabilities(ctx context.Context, cfg Config) (Capabilities, error) {
	// Documented probe: GET /public-srv/version → data.version (e.g. "4.0.2").
	// When the endpoint is absent (404), this v4-only provider continues with SupportsV4=true.
	versionURL, err := url.JoinPath(cfg.BaseURL, "public-srv/version")
	if err != nil {
		return Capabilities{}, err
	}
	httpClient := NewHTTPClient(versionURL, http.MethodGet, cfg.AccessToken)
	resp, err := httpClient.DoJSON(ctx, nil)
	if err != nil {
		return Capabilities{SupportsV4: true}, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return Capabilities{SupportsV4: true, Version: "unknown"}, nil
	}
	if err := ExpectStatus(resp, http.StatusOK); err != nil {
		return Capabilities{}, fmt.Errorf("version probe: %w", err)
	}
	var out versionAPIResponse
	if err := DecodeJSON(resp, &out); err != nil {
		return Capabilities{}, fmt.Errorf("version decode: %w", err)
	}
	ver := strings.TrimSpace(out.Data.Version)
	major, err := majorVersion(ver)
	if err != nil {
		return Capabilities{}, fmt.Errorf("parse version %q: %w", ver, err)
	}
	return Capabilities{
		SupportsV4: major >= 4,
		Version:    ver,
	}, nil
}

var versionMajorRE = regexp.MustCompile(`^v?(\d+)`)

func majorVersion(v string) (int, error) {
	m := versionMajorRE.FindStringSubmatch(strings.TrimSpace(v))
	if len(m) < 2 {
		return 0, fmt.Errorf("invalid version")
	}
	return strconv.Atoi(m[1])
}
