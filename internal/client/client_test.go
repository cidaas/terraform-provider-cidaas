package client

import (
	"testing"
)

func TestNormalizeVersion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want string
	}{
		{"v3", "3.x"},
		{"3", "3.x"},
		{"3.x", "3.x"},
		{"v3.x", "3.x"},
		{"v4", "4.x"},
		{"4", "4.x"},
		{"4.x", "4.x"},
		{"v4.x", "4.x"},
		{"", ""},
	}
	for _, tc := range cases {
		got := NormalizeVersion(tc.in)
		if got != tc.want {
			t.Fatalf("NormalizeVersion(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}
