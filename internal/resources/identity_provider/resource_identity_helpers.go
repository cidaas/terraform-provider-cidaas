//nolint:revive
package identity_provider

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// readHandleNotFound drops the instance from state when the API reports the object is gone.
func readHandleNotFound(ctx context.Context, resp *resource.ReadResponse, err error) bool {
	if err == nil || !util.IsResourceNotFound(err) {
		return false
	}
	tflog.Info(ctx, "resource not found in remote; removing from Terraform state")
	resp.State.RemoveResource(ctx)
	return true
}

func listStringOrNull(values []string) types.List {
	if len(values) == 0 {
		return types.ListNull(types.StringType)
	}
	elems := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elems = append(elems, types.StringValue(v))
	}
	return types.ListValueMust(types.StringType, elems)
}
