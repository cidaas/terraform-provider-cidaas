package util //nolint:revive

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	ISO8601TimeFormat = "2006-01-02T15:04:05Z"
)

// ResponseToString reads and converts HTTP response body to string.
// Returns error if response or body is nil, or if reading fails.
func ResponseToString(resp *http.Response) (string, error) {
	if resp == nil || resp.Body == nil {
		return "", fmt.Errorf("invalid response or body")
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

// Contains checks if an item exists in a slice of comparable elements.
// It performs a linear search through the slice and returns true if the item is found.
func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Int64ValueOrNull converts an int64 pointer to Terraform Int64 type.
// Returns Int64Null if the pointer is nil.
func Int64ValueOrNull(value *int64) types.Int64 {
	if value != nil {
		return types.Int64Value(*value)
	}
	return types.Int64Null()
}

// StringValueOrNull converts a string pointer to Terraform String type.
// Returns StringNull if the pointer is nil or points to an empty string.
func StringValueOrNull(value *string) types.String {
	if value != nil {
		if *value == "" {
			return types.StringNull()
		}
		return types.StringValue(*value)
	}
	return types.StringNull()
}

// SetValueOrNull converts a string slice to Terraform Set type.
// Returns SetNull if the slice is empty or nil.
func SetValueOrNull(values []string) types.Set {
	if len(values) == 0 {
		return types.SetNull(types.StringType)
	}
	elements := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	return types.SetValueMust(types.StringType, elements)
}

// BoolValueOrNull converts a boolean pointer to Terraform Bool type.
// Returns BoolNull if the pointer is nil.
func BoolValueOrNull(value *bool) types.Bool {
	if value != nil {
		return types.BoolValue(*value)
	}
	return types.BoolNull()
}

// TimeValueOrNull converts a time pointer to Terraform String type in ISO8601 format.
// Returns StringNull if the pointer is nil or points to zero time.
func TimeValueOrNull(value *time.Time) types.String {
	if value != nil && !value.IsZero() {
		return types.StringValue(value.Format(ISO8601TimeFormat))
	}
	return types.StringNull()
}

// MapValueOrNull converts a string map pointer to Terraform Map type.
// Returns MapNull if the pointer is nil or points to an empty map.
func MapValueOrNull(value *map[string]string) (types.Map, diag.Diagnostics) {
	if value == nil || len(*value) < 1 {
		return types.MapNull(types.StringType), nil
	}
	mapAttributes := map[string]attr.Value{}
	for key, val := range *value {
		mapAttributes[key] = types.StringValue(val)
	}
	cf, diag := types.MapValue(types.StringType, mapAttributes)
	return cf, diag
}

// ProcessResponse decodes HTTP response body into the target interface.
// Returns error if JSON decoding fails.
func ProcessResponse(res *http.Response, target interface{}) error {
	if res == nil || res.Body == nil {
		return fmt.Errorf("invalid response or body")
	}
	if target == nil {
		return nil
	}
	err := json.NewDecoder(res.Body).Decode(target)
	if err != nil {
		return fmt.Errorf("failed to decode response body, %w", err)
	}
	return nil
}

// HandleResponseError handles HTTP response errors and ensures proper resource cleanup.
// Closes response body on error and returns the original error.
func HandleResponseError(res *http.Response, err error) error {
	if err != nil {
		if res != nil && res.Body != nil {
			defer func() { _ = res.Body.Close() }()
		}
		return err
	}
	return nil
}

// FormatErrorMessage formats an error into a standardized error message string.
// Returns empty string if error is nil.
func FormatErrorMessage(err error) string {
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return ""
}

// SetValueOrEmpty converts a string slice to Terraform Set type, preserving empty slices.
// Unlike SetValueOrNull, this returns an empty Set instead of null for empty slices.
func SetValueOrEmpty(values []string) types.Set {
	if len(values) == 0 {
		return types.SetValueMust(types.StringType, []attr.Value{})
	}
	elements := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	return types.SetValueMust(types.StringType, elements)
}

// LogFields is a convenience type for structured logging, similar to gin.H
type LogFields map[string]interface{}

// H is a shorthand alias for LogFields (like gin.H)
type H map[string]interface{}
