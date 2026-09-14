package client

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ValidateResourceVersion gates Trustdesk (v4-only) resources when TargetVersion is 3.x.
//
// Call this from Configure on v4-only resources (app_configuration, user_setup,
// hostedpages Trustdesk resources, verification). Shared Phase 1 resources
// (role, scope, registration_field, etc.) must not call this — they work on
// both 3.x and 4.x.
//
// Default minimum is 4.x. The deprecated stub cidaas_app is treated as 3.x-compatible
// so Configure does not fail solely on version when TargetVersion is 3.x.
func (c *Client) ValidateResourceVersion(resourceName string, diags *diag.Diagnostics) bool {
	minVersion := "4.x"
	if resourceName == "cidaas_app" {
		minVersion = "3.x"
	}

	if c.Capabilities.TargetVersion == "3.x" && minVersion == "4.x" {
		diags.AddError(
			"Incompatible Resource",
			fmt.Sprintf("The resource `%s` is only supported on cidaas v4.x (Trustdesk), but the provider is configured for v3.x.", resourceName),
		)
		return false
	}

	return true
}
