package app_test

import (
	"testing"

	"github.com/Cidaas/terraform-provider-cidaas/internal/resources/app"
)

func TestLegacyAppResourceType(t *testing.T) {
	t.Parallel()
	if app.NewLegacyAppResource() == nil {
		t.Fatal("expected legacy app resource")
	}
}

func TestAppConfigurationResourceType(t *testing.T) {
	t.Parallel()
	if app.NewAppConfigurationResource() == nil {
		t.Fatal("expected app configuration resource")
	}
}
