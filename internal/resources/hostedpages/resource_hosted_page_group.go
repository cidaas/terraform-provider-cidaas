package hostedpages

import (
	"context"
	"fmt"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &hostedPageGroupResource{}
	_ resource.ResourceWithConfigure   = &hostedPageGroupResource{}
	_ resource.ResourceWithImportState = &hostedPageGroupResource{}
)

// hostedPageGroupResource implements cidaas_hosted_page_group.
type hostedPageGroupResource struct {
	client *client.Client
}

// hostedPageGroupModel is the Terraform state/plan model for cidaas_hosted_page_group.
type hostedPageGroupModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	GroupOwner    types.String `tfsdk:"group_owner"`
	DefaultLocale types.String `tfsdk:"default_locale"`
	HostedPages   types.Set    `tfsdk:"hosted_pages"`
	CreatedTime   types.String `tfsdk:"created_time"`
	UpdatedTime   types.String `tfsdk:"updated_time"`
}

// hostedPageNested is one hosted_pages set element.
type hostedPageNested struct {
	HostedPageID types.String `tfsdk:"hosted_page_id"`
	Locale       types.String `tfsdk:"locale"`
	URL          types.String `tfsdk:"url"`
	Content      types.String `tfsdk:"content"`
}

// NewHostedPageGroupResource returns the cidaas_hosted_page_group resource.
func NewHostedPageGroupResource() resource.Resource {
	return &hostedPageGroupResource{}
}

// Metadata sets the resource type name to cidaas_hosted_page_group.
func (r *hostedPageGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosted_page_group"
}

func hostedPageAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"hosted_page_id": types.StringType,
		"locale":         types.StringType,
		"url":            types.StringType,
		"content":        types.StringType,
	}
}

// Schema defines the Terraform schema for cidaas_hosted_page_group.
func (r *hostedPageGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a hosted page group via `POST/GET/DELETE /hostedpages-srv/hpgroup`. " +
			"Requires `cidaas:hosted_pages_write`, `cidaas:hosted_pages_read`, `cidaas:hosted_pages_delete`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Group id / name (`_id` in the API). Must be unique.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_owner": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("client"),
			},
			"default_locale": schema.StringAttribute{
				Required: true,
			},
			"hosted_pages": schema.SetNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"hosted_page_id": schema.StringAttribute{Required: true},
						"locale":         schema.StringAttribute{Required: true},
						"url":            schema.StringAttribute{Required: true},
						"content":        schema.StringAttribute{Optional: true},
					},
				},
			},
			"created_time": schema.StringAttribute{Computed: true},
			"updated_time": schema.StringAttribute{Computed: true},
		},
	}
}

// Configure injects the shared cidaas API client from the provider.
func (r *hostedPageGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *hostedPageGroupResource) toAPI(ctx context.Context, m hostedPageGroupModel) (client.HostedPageGroupModel, error) {
	out := client.HostedPageGroupModel{
		ID:            m.Name.ValueString(),
		GroupOwner:    m.GroupOwner.ValueString(),
		DefaultLocale: m.DefaultLocale.ValueString(),
	}
	var pages []hostedPageNested
	diags := m.HostedPages.ElementsAs(ctx, &pages, false)
	if diags.HasError() {
		return out, fmt.Errorf("%s", diags.Errors())
	}
	for _, p := range pages {
		out.HostedPages = append(out.HostedPages, client.HostedPageData{
			HostedPageID: p.HostedPageID.ValueString(),
			Locale:       p.Locale.ValueString(),
			URL:          p.URL.ValueString(),
			Content:      p.Content.ValueString(),
		})
	}
	return out, nil
}

func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func (r *hostedPageGroupResource) fromAPI(_ context.Context, data client.HostedPageGroupModel, state *hostedPageGroupModel) error {
	state.ID = types.StringValue(data.ID)
	state.Name = types.StringValue(data.ID)
	state.GroupOwner = types.StringValue(data.GroupOwner)
	state.DefaultLocale = types.StringValue(data.DefaultLocale)
	state.CreatedTime = stringOrNull(data.CreatedTime)
	state.UpdatedTime = stringOrNull(data.UpdatedTime)

	elems := make([]attr.Value, 0, len(data.HostedPages))
	for _, p := range data.HostedPages {
		// API returns content:"" when omitted; keep Null so set elements match plan.
		obj, diags := types.ObjectValue(hostedPageAttrTypes(), map[string]attr.Value{
			"hosted_page_id": types.StringValue(p.HostedPageID),
			"locale":         types.StringValue(p.Locale),
			"url":            types.StringValue(p.URL),
			"content":        stringOrNull(p.Content),
		})
		if diags.HasError() {
			return fmt.Errorf("%s", diags.Errors())
		}
		elems = append(elems, obj)
	}
	set, diags := types.SetValue(types.ObjectType{AttrTypes: hostedPageAttrTypes()}, elems)
	if diags.HasError() {
		return fmt.Errorf("%s", diags.Errors())
	}
	state.HostedPages = set
	return nil
}

// upsert creates or updates via POST /hostedpages-srv/hpgroup (srv upsert).
func (r *hostedPageGroupResource) upsert(ctx context.Context, plan *hostedPageGroupModel) error {
	apiModel, err := r.toAPI(ctx, *plan)
	if err != nil {
		return fmt.Errorf("invalid hosted pages: %w", err)
	}
	out, err := r.client.HostedPages.Upsert(ctx, apiModel)
	if err != nil {
		return err
	}
	return r.fromAPI(ctx, out.Data, plan)
}

// Create upserts the hosted page group via POST /hostedpages-srv/hpgroup.
func (r *hostedPageGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hostedPageGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.upsert(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Create hosted page group failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes state via GET /hostedpages-srv/hpgroup/{name}.
func (r *hostedPageGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) { //nolint:dupl // matches sibling hostedpages Read
	var state hostedPageGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.HostedPages.Get(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read hosted page group failed", err.Error())
		return
	}
	if err := r.fromAPI(ctx, out.Data, &state); err != nil {
		resp.Diagnostics.AddError("Map response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update upserts changes via POST /hostedpages-srv/hpgroup.
func (r *hostedPageGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan hostedPageGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.upsert(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Update hosted page group failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes via DELETE /hostedpages-srv/hpgroup/{name}.
func (r *hostedPageGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hostedPageGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.HostedPages.Delete(ctx, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete hosted page group failed", err.Error())
		return
	}
}

// ImportState imports by group name (_id).
func (r *hostedPageGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
