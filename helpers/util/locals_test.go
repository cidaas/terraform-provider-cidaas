package util

import "testing"

func TestGetLanguageForLocale(t *testing.T) {
	tests := []struct {
		locale   string
		expected string
	}{
		{"en-US", "en"},
		{"de-DE", "de"},
		{"invalid", "en"}, // default
		{"", "en"},        // default
	}

	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			result := GetLanguageForLocale(tt.locale)
			if result != tt.expected {
				t.Errorf("GetLanguageForLocale(%s) = %s, want %s", tt.locale, result, tt.expected)
			}
		})
	}
}
