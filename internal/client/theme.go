package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// ThemeService calls hostedpages-srv theme APIs (multipart upload).
type ThemeService struct {
	cfg Config
}

// NewThemeService returns a ThemeService bound to cfg.
func NewThemeService(cfg Config) *ThemeService {
	return &ThemeService{cfg: cfg}
}

// ThemeListResponse is the list themes envelope.
type ThemeListResponse struct {
	Success bool     `json:"success"`
	Status  int      `json:"status"`
	Data    []string `json:"data"`
}

// authClient preserves Authorization across same-host redirects.
// Go's default client strips Authorization on redirect, which turns
// POST /themes → 301 /themes/ into a 401 Unauthorized.
func authClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			if len(via) > 0 {
				if auth := via[0].Header.Get("Authorization"); auth != "" && req.Header.Get("Authorization") == "" {
					req.Header.Set("Authorization", auth)
				}
			}
			return nil
		},
	}
}

func themesCollectionURL(base string) (string, error) {
	endpoint, err := url.JoinPath(base, "hostedpages-srv/themes")
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}
	return endpoint, nil
}

// Upload posts multipart form field "theme" with the given filename and CSS content.
func (s *ThemeService) Upload(ctx context.Context, filename, cssContent string) error {
	endpoint, err := themesCollectionURL(s.cfg.BaseURL)
	if err != nil {
		return err
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("theme", path.Base(filename))
	if err != nil {
		return fmt.Errorf("form file: %w", err)
	}
	if _, err := io.WriteString(part, cssContent); err != nil {
		return fmt.Errorf("write css: %w", err)
	}
	if err := writer.Close(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+s.cfg.AccessToken)
	resp, err := authClient().Do(req)
	if err != nil {
		return fmt.Errorf("upload theme: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err := ExpectStatus(resp, http.StatusOK, http.StatusCreated); err != nil {
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("%w (check cidaas:themes_write on the OAuth client)", err)
		}
		return err
	}
	return nil
}

// Get GETs /hostedpages-srv/themes/{filename} and returns raw body bytes.
func (s *ThemeService) Get(ctx context.Context, filename string) ([]byte, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/themes", filename)
	if err != nil {
		return nil, err
	}
	httpClient := NewHTTPClient(endpoint, http.MethodGet, s.cfg.AccessToken)
	resp, err := httpClient.DoJSON(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := ExpectStatus(resp, http.StatusOK); err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

// List GETs /hostedpages-srv/themes/ and returns theme filenames.
func (s *ThemeService) List(ctx context.Context) ([]string, error) {
	endpoint, err := themesCollectionURL(s.cfg.BaseURL)
	if err != nil {
		return nil, err
	}
	httpClient := NewHTTPClient(endpoint, http.MethodGet, s.cfg.AccessToken)
	resp, err := httpClient.DoJSON(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := ExpectStatus(resp, http.StatusOK); err != nil {
		return nil, err
	}
	var out ThemeListResponse
	if err := DecodeJSON(resp, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// Delete DELETEs /hostedpages-srv/themes/{filename}.
func (s *ThemeService) Delete(ctx context.Context, filename string) error {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/themes", filename)
	if err != nil {
		return err
	}
	httpClient := NewHTTPClient(endpoint, http.MethodDelete, s.cfg.AccessToken)
	resp, err := httpClient.DoJSON(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return ExpectStatus(resp, http.StatusOK, http.StatusNoContent)
}
