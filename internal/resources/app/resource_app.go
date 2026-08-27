package app

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

const legacyAppMigrationMsg = "Deprecated: use cidaas_app_configuration for cidaas v4 (Trustdesk). " +
	"cidaas_app targets the legacy appv1 model and is not supported on this provider version."

var (
	_ resource.Resource              = &legacyAppResource{}
	_ resource.ResourceWithConfigure = &legacyAppResource{}
)

type legacyAppResource struct{}

// NewLegacyAppResource returns deprecated cidaas_app (migration stub — ticket #2413 task 5).
func NewLegacyAppResource() resource.Resource {
	return &legacyAppResource{}
}

// Metadata sets the resource type name to cidaas_app.
func (r *legacyAppResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

// Schema defines a minimal deprecated schema so existing state can be detected; all CRUD returns migration errors.
func (r *legacyAppResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		DeprecationMessage: legacyAppMigrationMsg,
		MarkdownDescription: legacyAppMigrationMsg + "\n\n" +
			"Migrate HCL to [`cidaas_app_configuration`](app_configuration.md).",
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Legacy attribute — use `cidaas_app_configuration.client_id`.",
			},
			"client_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Legacy attribute — use `cidaas_app_configuration.client_name`.",
			},
		},
	}
}

// Configure is a no-op; this stub never calls the API.
func (r *legacyAppResource) Configure(_ context.Context, _ resource.ConfigureRequest, _ *resource.ConfigureResponse) {
}

// Create rejects with a migration message (ponytail: no legacy appv1 API on v4 provider).
func (r *legacyAppResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Deprecated resource",
		fmt.Sprintf("%s Create a `cidaas_app_configuration` resource instead.", legacyAppMigrationMsg),
	)
}

// Read rejects with a migration message.
func (r *legacyAppResource) Read(_ context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.AddError(
		"Deprecated resource",
		fmt.Sprintf("%s Use `terraform state rm` and migrate to `cidaas_app_configuration`.", legacyAppMigrationMsg),
	)
}

// Update rejects with a migration message.
func (r *legacyAppResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Deprecated resource",
		fmt.Sprintf("%s Update a `cidaas_app_configuration` resource instead.", legacyAppMigrationMsg),
	)
}

// Delete rejects with a migration message.
func (r *legacyAppResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Deprecated resource",
		fmt.Sprintf("%s Remove via `cidaas_app_configuration` or `terraform state rm`.", legacyAppMigrationMsg),
	)
}
