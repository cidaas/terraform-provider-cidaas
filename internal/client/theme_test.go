package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestThemeUploadGetListDelete(t *testing.T) {
	t.Parallel()
	const filename = "custom.css"
	const css = "body{color:red}"

	mux := http.NewServeMux()
	mux.HandleFunc("/hostedpages-srv/themes/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/hostedpages-srv/themes/":
			if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
				t.Fatalf("missing auth: %q", r.Header.Get("Authorization"))
			}
			mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
				t.Fatalf("content-type=%q err=%v", r.Header.Get("Content-Type"), err)
			}
			mr := multipart.NewReader(r.Body, params["boundary"])
			part, err := mr.NextPart()
			if err != nil {
				t.Fatal(err)
			}
			if part.FormName() != "theme" {
				t.Fatalf("form name=%q", part.FormName())
			}
			if part.FileName() != filename {
				t.Fatalf("filename=%q", part.FileName())
			}
			got, _ := io.ReadAll(part)
			if string(got) != css {
				t.Fatalf("css=%q", got)
			}
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodGet && r.URL.Path == "/hostedpages-srv/themes/":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ThemeListResponse{
				Success: true,
				Status:  200,
				Data:    []string{filename, "other.css"},
			})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})
	mux.HandleFunc("/hostedpages-srv/themes/"+filename, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(css))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method", http.StatusMethodNotAllowed)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	svc := NewThemeService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	if err := svc.Upload(context.Background(), filename, css); err != nil {
		t.Fatal(err)
	}
	got, err := svc.Get(context.Background(), filename)
	if err != nil || string(got) != css {
		t.Fatalf("get: err=%v data=%q", err, got)
	}
	list, err := svc.List(context.Background())
	if err != nil || len(list) != 2 || list[0] != filename {
		t.Fatalf("list: err=%v data=%v", err, list)
	}
	if err := svc.Delete(context.Background(), filename); err != nil {
		t.Fatal(err)
	}
}

func TestThemeUploadUnauthorizedHint(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	svc := NewThemeService(Config{BaseURL: srv.URL, AccessToken: "tok"})
	err := svc.Upload(context.Background(), "x.css", "a{}")
	if err == nil || !strings.Contains(err.Error(), "cidaas:themes_write") {
		t.Fatalf("err=%v", err)
	}
}

func TestThemeUpload_PreservesAuthOnRedirect(t *testing.T) {
	t.Parallel()
	var sawAuthOnFinal bool
	mux := http.NewServeMux()
	mux.HandleFunc("/hostedpages-srv/themes", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/hostedpages-srv/themes/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/hostedpages-srv/themes/", func(w http.ResponseWriter, r *http.Request) {
		sawAuthOnFinal = r.Header.Get("Authorization") == "Bearer tok"
		if !sawAuthOnFinal {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Force a no-trailing-slash URL so the server redirects (simulates gateway behavior).
	endpoint := srv.URL + "/hostedpages-srv/themes"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("theme", "x.css")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("body{}"))
	_ = writer.Close()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer tok")
	resp, err := authClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d authKept=%v", resp.StatusCode, sawAuthOnFinal)
	}
}
