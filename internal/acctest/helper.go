package acctest

import (
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Cidaas/terraform-provider-cidaas/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ProtoV6ProviderFactories is used by acceptance tests.
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cidaas": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func PreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID") == "" {
		t.Fatal("TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID must be set for acceptance tests")
	}
	if os.Getenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET") == "" {
		t.Fatal("TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET must be set for acceptance tests")
	}
	if os.Getenv("BASE_URL") == "" {
		t.Fatal("BASE_URL must be set for acceptance tests")
	}
	if os.Getenv("TERRAFORM_PROVIDER_CIDAAS_VERSION") == "" && os.Getenv("CIDAAS_VERSION") == "" {
		t.Fatal("TERRAFORM_PROVIDER_CIDAAS_VERSION or CIDAAS_VERSION must be set for acceptance tests")
	}
}

func BaseURL() string {
	return os.Getenv("BASE_URL")
}

func ProviderConfig() string {
	verAttr := ""
	if v := os.Getenv("TERRAFORM_PROVIDER_CIDAAS_VERSION"); v != "" {
		verAttr = `  cidaas_version = "` + v + `"` + "\n"
	} else if v := os.Getenv("CIDAAS_VERSION"); v != "" {
		verAttr = `  cidaas_version = "` + v + `"` + "\n"
	}
	return `
provider "cidaas" {
  base_url = "` + BaseURL() + `"
` + verAttr + `}
`
}

// RandString returns a lowercase random string of the given length.
func RandString(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func targetVersionEnv() string {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("TERRAFORM_PROVIDER_CIDAAS_VERSION")))
	if v == "" {
		v = strings.ToLower(strings.TrimSpace(os.Getenv("CIDAAS_VERSION")))
	}
	return v
}

// SkipIfV3 skips when the acceptance run targets cidaas v3.x (v4-only resources).
func SkipIfV3(t *testing.T) {
	t.Helper()
	v := targetVersionEnv()
	if strings.HasPrefix(v, "3") || v == "3.x" {
		t.Skip("Resource is supported on cidaas v4.x only; skipping on v3 environment")
	}
}

// SkipIfV4 skips when the acceptance run targets cidaas v4.x (v3-only / happy-path v3 CRUD).
func SkipIfV4(t *testing.T) {
	t.Helper()
	v := targetVersionEnv()
	if strings.HasPrefix(v, "4") || strings.HasPrefix(v, "v4") || v == "4.x" {
		t.Skip("Resource is supported on cidaas v3.x only; skipping on v4 environment")
	}
}
