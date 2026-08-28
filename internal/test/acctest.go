package acctest

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	provider "github.com/Cidaas/terraform-provider-cidaas/internal"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// TestAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cidaas": providerserver.NewProtocol6WithError(provider.Cidaas("test")()),
}

var (
	TestToken string
	BaseURL   string
)

func TestAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
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

	tokenURL := fmt.Sprintf("%s/%s", os.Getenv("BASE_URL"), "token-srv/token")
	client, err := util.NewHTTPClient(tokenURL, http.MethodPost)
	if err != nil {
		t.Error(err)
	}
	payload := map[string]string{
		"client_id":     os.Getenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID"),
		"client_secret": os.Getenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET"),
		"grant_type":    "client_credentials",
	}
	res, err := client.MakeRequest(context.Background(), payload)
	if err := util.HandleResponseError(res, err); err != nil {
		t.Fatalf("failed to generate access token %s", err.Error())
	}
	defer res.Body.Close()
	var response cidaas.TokenResponse
	if err := util.ProcessResponse(res, &response); err != nil {
		t.Fatalf("failed to generate access token %s", err.Error())
	}
	TestToken = response.AccessToken
	BaseURL = os.Getenv("BASE_URL")
}

// RandString generates a random string with the given length.
func RandString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	charset := "abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

func GetBaseURL() string {
	if BaseURL != "" {
		return BaseURL
	}
	return os.Getenv("BASE_URL")
}
