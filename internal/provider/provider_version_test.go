package provider

import (
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
)

// TestNormalizeVersion verifies that version string inputs are properly mapped to canonical 3.x and 4.x formats.
func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"v3", "3.x"},
		{"3", "3.x"},
		{"3.x", "3.x"},
		{"v3.x", "3.x"},
		{"v3.102.8", "3.x"},
		{"v4", "4.x"},
		{"4", "4.x"},
		{"4.x", "4.x"},
		{"v4.x", "4.x"},
		{"v4.0.2", "4.x"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual := client.NormalizeVersion(tt.input)
			if actual != tt.expected {
				t.Errorf("NormalizeVersion(%q) = %q, expected %q", tt.input, actual, tt.expected)
			}
		})
	}
}
