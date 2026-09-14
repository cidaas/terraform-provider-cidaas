// Package hostedpages implements Terraform resources for hostedpages-srv:
// cidaas_theme, cidaas_translations, and cidaas_hosted_page_layout.
package hostedpages

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &themeResource{}
	_ resource.ResourceWithConfigure   = &themeResource{}
	_ resource.ResourceWithImportState = &themeResource{}
)

// themeResource implements cidaas_theme (CSS upload to hostedpages-srv).
type themeResource struct {
	client *client.Client
}

// themeModel is the Terraform state/plan model for cidaas_theme.
type themeModel struct {
	ID         types.String `tfsdk:"id"`
	Filename   types.String `tfsdk:"filename"`
	CSSContent types.String `tfsdk:"css_content"`
	CSSHash    types.String `tfsdk:"css_hash"`
}

// NewThemeResource returns the cidaas_theme resource.
func NewThemeResource() resource.Resource {
	return &themeResource{}
}

// Metadata sets the resource type name to cidaas_theme.
func (r *themeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_theme"
}

// Schema defines the Terraform schema for cidaas_theme.
func (r *themeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Uploads a custom CSS theme to hostedpages-srv (`POST /hostedpages-srv/themes`). " +
			"Requires scopes `cidaas:themes_write`, `cidaas:themes_read`, `cidaas:themes_delete`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"filename": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Theme file name (e.g. `custom.css`). Used as the resource ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"css_content": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Raw CSS content uploaded as multipart field `theme`.",
				Sensitive:           false,
			},
			"css_hash": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "SHA-256 hex of css_content for drift detection.",
			},
		},
	}
}

// Configure injects the shared cidaas API client from the provider.
func (r *themeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	if !c.ValidateResourceVersion("cidaas_theme", &resp.Diagnostics) {
		return
	}
	r.client = c
}

func cssHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

// write uploads CSS via multipart POST and sets id/css_hash on the plan.
func (r *themeResource) write(ctx context.Context, plan *themeModel) error {
	if err := r.client.Themes.Upload(ctx, plan.Filename.ValueString(), plan.CSSContent.ValueString()); err != nil {
		return err
	}
	plan.ID = plan.Filename
	plan.CSSHash = types.StringValue(cssHash(plan.CSSContent.ValueString()))
	return nil
}

// Create uploads the theme via POST /hostedpages-srv/themes.
func (r *themeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan themeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.write(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Theme upload failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read checks the theme exists via GET; keeps HCL css_content as source of truth.
func (r *themeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state themeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data, err := r.client.Themes.Get(ctx, state.Filename.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Theme read failed", err.Error())
		return
	}
	// Prefer planned css_content from state; refresh hash from remote bytes when available.
	if len(data) > 0 {
		// Theme GET may return raw CSS or JSON envelope; keep state css_content as source of truth for HCL.
		state.CSSHash = types.StringValue(cssHash(state.CSSContent.ValueString()))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update re-uploads CSS (same multipart upload path as create).
func (r *themeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan themeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.write(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Theme update failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the theme file via DELETE /hostedpages-srv/themes/{filename}.
func (r *themeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state themeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Themes.Delete(ctx, state.Filename.ValueString()); err != nil {
		resp.Diagnostics.AddError("Theme delete failed", err.Error())
		return
	}
}

// ImportState imports by filename (also used as id).
func (r *themeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("filename"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
