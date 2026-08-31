package util

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{"found", []string{"a", "b", "c"}, "b", true},
		{"not found", []string{"a", "b", "c"}, "d", false},
		{"empty slice", []string{}, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("Contains() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStringValueOrNull(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected bool // true if should be null
	}{
		{"nil pointer", nil, true},
		{"empty string", stringPtr(""), true},
		{"valid string", stringPtr("test"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringValueOrNull(tt.input)
			if tt.expected && !result.IsNull() {
				t.Errorf("Expected null, got %v", result)
			}
			if !tt.expected && result.IsNull() {
				t.Errorf("Expected value, got null")
			}
		})
	}
}

func TestResponseToString(t *testing.T) {
	tests := []struct {
		name        string
		response    *http.Response
		expectError bool
	}{
		{"nil response", nil, true},
		{"valid response", &http.Response{
			Body: io.NopCloser(strings.NewReader("test")),
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResponseToString(tt.response)
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

// Helper function
func stringPtr(s string) *string { return &s }

func TestInt64ValueOrNull(t *testing.T) {
	tests := []struct {
		name     string
		input    *int64
		expected bool // true if should be null
	}{
		{"nil pointer", nil, true},
		{"zero value", int64Ptr(0), false},
		{"positive value", int64Ptr(123), false},
		{"negative value", int64Ptr(-456), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Int64ValueOrNull(tt.input)
			if tt.expected && !result.IsNull() {
				t.Errorf("Expected null, got %v", result)
			}
			if !tt.expected && result.IsNull() {
				t.Errorf("Expected value, got null")
			}
		})
	}
}

func TestSetValueOrNull(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected bool // true if should be null
	}{
		{"nil slice", nil, true},
		{"empty slice", []string{}, true},
		{"single item", []string{"test"}, false},
		{"multiple items", []string{"a", "b", "c"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SetValueOrNull(tt.input)
			if tt.expected && !result.IsNull() {
				t.Errorf("Expected null, got %v", result)
			}
			if !tt.expected && result.IsNull() {
				t.Errorf("Expected value, got null")
			}
		})
	}
}

func TestBoolValueOrNull(t *testing.T) {
	tests := []struct {
		name     string
		input    *bool
		expected bool // true if should be null
	}{
		{"nil pointer", nil, true},
		{"true value", boolPtr(true), false},
		{"false value", boolPtr(false), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BoolValueOrNull(tt.input)
			if tt.expected && !result.IsNull() {
				t.Errorf("Expected null, got %v", result)
			}
			if !tt.expected && result.IsNull() {
				t.Errorf("Expected value, got null")
			}
		})
	}
}

func TestTimeValueOrNull(t *testing.T) {
	now := time.Now()
	zeroTime := time.Time{}

	tests := []struct {
		name     string
		input    *time.Time
		expected bool // true if should be null
	}{
		{"nil pointer", nil, true},
		{"zero time", &zeroTime, true},
		{"valid time", &now, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimeValueOrNull(tt.input)
			if tt.expected && !result.IsNull() {
				t.Errorf("Expected null, got %v", result)
			}
			if !tt.expected && result.IsNull() {
				t.Errorf("Expected value, got null")
			}
		})
	}
}

func TestMapValueOrNull(t *testing.T) {
	tests := []struct {
		name        string
		input       *map[string]string
		expectNull  bool
		expectError bool
	}{
		{"nil pointer", nil, true, false},
		{"empty map", &map[string]string{}, true, false},
		{"valid map", &map[string]string{"key": "value"}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diag := MapValueOrNull(tt.input)

			if tt.expectError && !diag.HasError() {
				t.Error("Expected error, got none")
			}
			if !tt.expectError && diag.HasError() {
				t.Errorf("Expected no error, got %v", diag)
			}

			if tt.expectNull && !result.IsNull() {
				t.Errorf("Expected null, got %v", result)
			}
			if !tt.expectNull && result.IsNull() {
				t.Errorf("Expected value, got null")
			}
		})
	}
}

func TestProcessResponse(t *testing.T) {
	type testStruct struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}

	tests := []struct {
		name        string
		response    *http.Response
		target      interface{}
		expectError bool
	}{
		{
			"nil response",
			nil,
			&testStruct{},
			true,
		},
		{
			"nil target",
			&http.Response{Body: io.NopCloser(strings.NewReader(`{"message":"test"}`))},
			nil,
			false,
		},
		{
			"valid json",
			&http.Response{Body: io.NopCloser(strings.NewReader(`{"message":"test","code":200}`))},
			&testStruct{},
			false,
		},
		{
			"invalid json",
			&http.Response{Body: io.NopCloser(strings.NewReader(`{invalid json}`))},
			&testStruct{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ProcessResponse(tt.response, tt.target)
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestHandleResponseError(t *testing.T) {
	tests := []struct {
		name        string
		response    *http.Response
		inputError  error
		expectError bool
	}{
		{"no error", &http.Response{}, nil, false},
		{"with error and response", &http.Response{Body: io.NopCloser(strings.NewReader("test"))}, fmt.Errorf("test error"), true},
		{"with error no response", nil, fmt.Errorf("test error"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := HandleResponseError(tt.response, tt.inputError)
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestIsResourceNotFound(t *testing.T) {
	t.Parallel()
	if IsResourceNotFound(nil) {
		t.Error("nil should not be not-found")
	}
	if !IsResourceNotFound(fmt.Errorf("%w: body", ErrResourceNotFound)) {
		t.Error("wrapped ErrResourceNotFound should match")
	}
	if !IsResourceNotFound(fmt.Errorf("unexpected status code 404, response body: {}")) {
		t.Error("legacy 404 string should match")
	}
	if IsResourceNotFound(fmt.Errorf("unexpected status code 500")) {
		t.Error("500 should not match")
	}
}

func TestHTTPClientMakeRequest_GET404IsResourceNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"ok":false}`))
	}))
	t.Cleanup(srv.Close)
	client, err := NewHTTPClient(srv.URL, http.MethodGet)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.MakeRequest(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsResourceNotFound(err) {
		t.Fatalf("expected not-found, got: %v", err)
	}
	if !errors.Is(err, ErrResourceNotFound) {
		t.Fatalf("expected errors.Is ErrResourceNotFound, got: %v", err)
	}
}

func TestFormatErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		expected string
	}{
		{"nil error", nil, ""},
		{"valid error", fmt.Errorf("test error"), "Error: test error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatErrorMessage(tt.input)
			if result != tt.expected {
				t.Errorf("FormatErrorMessage() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestSetValueOrEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected bool // true if should be empty set (not null)
	}{
		{"nil slice", nil, true},
		{"empty slice", []string{}, true},
		{"single item", []string{"test"}, false},
		{"multiple items", []string{"a", "b", "c"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SetValueOrEmpty(tt.input)
			if result.IsNull() {
				t.Error("SetValueOrEmpty should never return null")
			}

			// Check if it's empty
			elements := result.Elements()
			isEmpty := len(elements) == 0

			if tt.expected && !isEmpty {
				t.Errorf("Expected empty set, got %d elements", len(elements))
			}
			if !tt.expected && isEmpty {
				t.Error("Expected non-empty set, got empty")
			}
		})
	}
}

// helper functions
func int64Ptr(i int64) *int64 { return &i }
func boolPtr(b bool) *bool    { return &b }
