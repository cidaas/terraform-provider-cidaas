package acctest

import (
	"math/rand"
	"os"
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
}

func BaseURL() string {
	return os.Getenv("BASE_URL")
}

func ProviderConfig() string {
	return `
provider "cidaas" {
  base_url = "` + BaseURL() + `"
}
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
