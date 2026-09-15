package client

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// RequireV4 errors unless the provider TargetVersion is 4.x (Trustdesk).
// Call only from Configure on v4-only resources. Shared/legacy resources must not call this.
func (c *Client) RequireV4(resourceName string, diags *diag.Diagnostics) bool {
	if c.Capabilities.TargetVersion == "4.x" {
		return true
	}
	diags.AddError(
		"Incompatible Resource",
		fmt.Sprintf(
			"The resource `%s` is only supported on cidaas v4.x (Trustdesk), but the provider is configured for %s. Set cidaas_version = \"4.x\" in the provider block.",
			resourceName,
			c.Capabilities.TargetVersion,
		),
	)
	return false
}

// ValidateResourceVersion gates Trustdesk (v4-only) resources when TargetVersion is 3.x.
//
// Call this from Configure on v4-only resources (app_configuration, federation_provider,
// hostedpages Trustdesk resources, verification, group_selection/filter, user_setup).
// Shared/legacy resources (role, scope, registration_field, consent, social/custom, etc.)
// must not call this.
//
// The deprecated stub cidaas_app is treated as version-agnostic so Configure does not fail
// solely on version when TargetVersion is 3.x.
func (c *Client) ValidateResourceVersion(resourceName string, diags *diag.Diagnostics) bool {
	if resourceName == "cidaas_app" {
		return true
	}
	return c.RequireV4(resourceName, diags)
}
