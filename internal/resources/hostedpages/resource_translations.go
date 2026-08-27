package hostedpages

import (
	"context"
	"fmt"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &translationsResource{}
	_ resource.ResourceWithConfigure   = &translationsResource{}
	_ resource.ResourceWithImportState = &translationsResource{}
)

// translationsResource implements cidaas_translations.
type translationsResource struct {
	client *client.Client
}

// translationsModel is the Terraform state/plan model for cidaas_translations.
type translationsModel struct {
	ID           types.String `tfsdk:"id"`
	LocaleID     types.String `tfsdk:"locale_id"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Translations types.Map    `tfsdk:"translations"`
}

// NewTranslationsResource returns the cidaas_translations resource.
func NewTranslationsResource() resource.Resource {
	return &translationsResource{}
}

// Metadata sets the resource type name to cidaas_translations.
func (r *translationsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_translations"
}

// Schema defines the Terraform schema for cidaas_translations.
func (r *translationsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages custom locale translations via hostedpages-srv (`/hostedpages-srv/translations`). " +
			"API JSON field is `translation` (map). Requires `cidaas:hosted_pages_*` scopes.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"locale_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Locale id (e.g. `fr`, `de`). Maps to API `locale`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"translations": schema.MapAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Key/value translation strings. Sent as API field `translation`.",
			},
		},
	}
}

// Configure injects the shared cidaas API client from the provider.
func (r *translationsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *translationsResource) toAPI(ctx context.Context, m translationsModel) (client.TranslationModel, error) {
	out := client.TranslationModel{
		Locale:      m.LocaleID.ValueString(),
		Enabled:     m.Enabled.ValueBool(),
		Translation: map[string]any{},
	}
	var raw map[string]string
	diags := m.Translations.ElementsAs(ctx, &raw, false)
	if diags.HasError() {
		return out, fmt.Errorf("%s", diags.Errors())
	}
	for k, v := range raw {
		out.Translation[k] = v
	}
	return out, nil
}

func (r *translationsResource) fromAPI(_ context.Context, locale string, enabled bool, translation map[string]any, state *translationsModel) error {
	state.ID = types.StringValue(locale)
	state.LocaleID = types.StringValue(locale)
	state.Enabled = types.BoolValue(enabled)
	elems := map[string]attr.Value{}
	for k, v := range translation {
		elems[k] = types.StringValue(fmt.Sprint(v))
	}
	m, diags := types.MapValue(types.StringType, elems)
	if diags.HasError() {
		return fmt.Errorf("%s", diags.Errors())
	}
	state.Translations = m
	return nil
}

// Create creates via POST /hostedpages-srv/translations (falls back to PUT on 409).
func (r *translationsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan translationsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiModel, err := r.toAPI(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid translations", err.Error())
		return
	}
	out, err := r.client.Translations.Create(ctx, apiModel)
	if err != nil {
		// Locale may already exist (HTTP 409); Terraform create should converge via update.
		if !strings.Contains(err.Error(), "409") {
			resp.Diagnostics.AddError("Create translation failed", err.Error())
			return
		}
		out, err = r.client.Translations.Update(ctx, plan.LocaleID.ValueString(), apiModel)
		if err != nil {
			resp.Diagnostics.AddError("Create translation failed", err.Error())
			return
		}
	}
	if err := r.fromAPI(ctx, out.Data.Locale, out.Data.Enabled, out.Data.Translation, &plan); err != nil {
		resp.Diagnostics.AddError("Map response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes state via GET /hostedpages-srv/translations/{locale}.
func (r *translationsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state translationsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.Translations.Get(ctx, state.LocaleID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read translation failed", err.Error())
		return
	}
	if err := r.fromAPI(ctx, out.Data.Locale, out.Data.Enabled, out.Data.Translation, &state); err != nil {
		resp.Diagnostics.AddError("Map response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update applies changes via PUT /hostedpages-srv/translations/{locale}.
func (r *translationsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan translationsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiModel, err := r.toAPI(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid translations", err.Error())
		return
	}
	out, err := r.client.Translations.Update(ctx, plan.LocaleID.ValueString(), apiModel)
	if err != nil {
		resp.Diagnostics.AddError("Update translation failed", err.Error())
		return
	}
	if err := r.fromAPI(ctx, out.Data.Locale, out.Data.Enabled, out.Data.Translation, &plan); err != nil {
		resp.Diagnostics.AddError("Map response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes via DELETE /hostedpages-srv/translations/{locale}.
func (r *translationsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state translationsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Translations.Delete(ctx, state.LocaleID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete translation failed", err.Error())
		return
	}
}

// ImportState imports by locale_id.
func (r *translationsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("locale_id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
