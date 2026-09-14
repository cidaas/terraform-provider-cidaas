package base

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ReadHandleNotFound drops the instance from state when the API reports the object is gone.
// It returns true if [err] was a not-found and the caller should return from Read.
func ReadHandleNotFound(ctx context.Context, resp *resource.ReadResponse, err error) bool {
	if err == nil || !util.IsResourceNotFound(err) {
		return false
	}
	tflog.Info(ctx, "resource not found in remote; removing from Terraform state")
	resp.State.RemoveResource(ctx)
	return true
}
