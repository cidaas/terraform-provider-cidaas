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
)

type ThemeService struct {
	cfg Config
}

func NewThemeService(cfg Config) *ThemeService {
	return &ThemeService{cfg: cfg}
}

type ThemeListResponse struct {
	Success bool     `json:"success"`
	Status  int      `json:"status"`
	Data    []string `json:"data"`
}

// Upload posts multipart form field "theme" with the given filename and CSS content.
func (s *ThemeService) Upload(ctx context.Context, filename, cssContent string) error {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/themes/")
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
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload theme: %w", err)
	}
	defer resp.Body.Close()
	return ExpectStatus(resp, http.StatusOK, http.StatusCreated)
}

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
	defer resp.Body.Close()
	if err := ExpectStatus(resp, http.StatusOK); err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

func (s *ThemeService) List(ctx context.Context) ([]string, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "hostedpages-srv/themes/")
	if err != nil {
		return nil, err
	}
	httpClient := NewHTTPClient(endpoint, http.MethodGet, s.cfg.AccessToken)
	resp, err := httpClient.DoJSON(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := ExpectStatus(resp, http.StatusOK); err != nil {
		return nil, err
	}
	var out ThemeListResponse
	if err := DecodeJSON(resp, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

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
	defer resp.Body.Close()
	return ExpectStatus(resp, http.StatusOK, http.StatusNoContent)
}
