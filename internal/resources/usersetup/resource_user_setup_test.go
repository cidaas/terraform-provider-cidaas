package usersetup

import (
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
)

func TestMissingRequiredFields(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		allowed  []string
		required []string
		want     []string
	}{
		{
			name:     "subset ok",
			allowed:  []string{"email", "given_name"},
			required: []string{"email"},
			want:     nil,
		},
		{
			name:     "missing field",
			allowed:  []string{"email"},
			required: []string{"email", "mobile_number"},
			want:     []string{"mobile_number"},
		},
		{
			name:     "allowed empty required not",
			allowed:  nil,
			required: []string{"email"},
			want:     []string{"email"},
		},
		{
			name:     "both empty",
			allowed:  nil,
			required: nil,
			want:     nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := missingRequiredFields(tc.allowed, tc.required)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("got[%d]=%s want %s", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestIsBuiltinAllowLogin(t *testing.T) {
	t.Parallel()

	for _, key := range []string{"EMAIL", "email", "MOBILE", "USER_NAME", "user-name"} {
		if !isBuiltinAllowLogin(key) {
			t.Fatalf("%q should be builtin", key)
		}
	}
	if isBuiltinAllowLogin("custom_field") {
		t.Fatal("custom_field should not be builtin")
	}
}

func TestFieldKeysMissingFromSetup(t *testing.T) {
	t.Parallel()

	fields := []client.FieldSetupEntry{
		{FieldKey: "email", Enabled: true},
		{FieldKey: "given_name", Enabled: true},
		{FieldKey: "disabled", Enabled: false},
	}

	attr, msg := fieldKeysMissingFromSetup(
		[]string{"email", "given_name"},
		[]string{"email"},
		[]string{"EMAIL", "custom_login"},
		fields,
	)
	if attr != "allow_login_with" || msg == "" {
		t.Fatalf("attr=%q msg=%q", attr, msg)
	}

	attr, msg = fieldKeysMissingFromSetup(
		[]string{"email", "missing"},
		[]string{"email"},
		[]string{"EMAIL"},
		fields,
	)
	if attr != "allowed_fields" {
		t.Fatalf("attr=%q msg=%q", attr, msg)
	}

	attr, msg = fieldKeysMissingFromSetup(
		[]string{"email", "given_name"},
		[]string{"email"},
		[]string{"EMAIL"},
		fields,
	)
	if attr != "" || msg != "" {
		t.Fatalf("attr=%q msg=%q", attr, msg)
	}
}
