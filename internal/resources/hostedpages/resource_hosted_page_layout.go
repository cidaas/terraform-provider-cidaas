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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = &hostedPageLayoutResource{}
	_ resource.ResourceWithConfigure   = &hostedPageLayoutResource{}
	_ resource.ResourceWithImportState = &hostedPageLayoutResource{}
)

type hostedPageLayoutResource struct {
	client *client.Client
}

type hostedPageLayoutModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Layout      types.Object `tfsdk:"layout"`
	Resources   types.Map    `tfsdk:"resources"`
	GroupID     types.String `tfsdk:"group_id"`
	Owner       types.String `tfsdk:"owner"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	CreatedTime types.String `tfsdk:"created_time"`
	UpdatedTime types.String `tfsdk:"updated_time"`
}

type layoutNested struct {
	HostedPageGroup types.String `tfsdk:"hosted_page_group"`
	Theme           types.String `tfsdk:"theme"`
	LogoURI         types.String `tfsdk:"logo_uri"`
	PolicyURI       types.String `tfsdk:"policy_uri"`
	TosURI          types.String `tfsdk:"tos_uri"`
	ImprintURI      types.String `tfsdk:"imprint_uri"`
	PrimaryColor    types.String `tfsdk:"primary_color"`
	AccentColor     types.String `tfsdk:"accent_color"`
	BackgroundURI   types.String `tfsdk:"background_uri"`
	ContentAlign    types.String `tfsdk:"content_align"`
	LogoAlign       types.String `tfsdk:"logo_align"`
	MediaType       types.String `tfsdk:"media_type"`
	VideoURL        types.String `tfsdk:"video_url"`
	FavIcon         types.String `tfsdk:"fav_icon"`
}

type layoutResourceNested struct {
	TranslationSet types.String `tfsdk:"translation_set"`
	Theme          types.String `tfsdk:"theme"`
	Layout         types.String `tfsdk:"layout"`
}

func NewHostedPageLayoutResource() resource.Resource {
	return &hostedPageLayoutResource{}
}

func (r *hostedPageLayoutResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosted_page_layout"
}

func layoutAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"hosted_page_group": types.StringType,
		"theme":             types.StringType,
		"logo_uri":          types.StringType,
		"policy_uri":        types.StringType,
		"tos_uri":           types.StringType,
		"imprint_uri":       types.StringType,
		"primary_color":     types.StringType,
		"accent_color":      types.StringType,
		"background_uri":    types.StringType,
		"content_align":     types.StringType,
		"logo_align":        types.StringType,
		"media_type":        types.StringType,
		"video_url":         types.StringType,
		"fav_icon":          types.StringType,
	}
}

func layoutResourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"translation_set": types.StringType,
		"theme":           types.StringType,
		"layout":          types.StringType,
	}
}

func (r *hostedPageLayoutResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a hosted page layout via `/hostedpages-srv/hosted-page-layouts`. " +
			"Links branding (`layout`) to a hosted page group and theme. " +
			"`group_id` is computed from the access token (not configurable). " +
			"Requires `cidaas:hosted_pages_*` and `cidaas:themes_*` scopes plus admin group claims for writes.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable layout description (max 500 characters).",
			},
			"layout": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"hosted_page_group": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Hosted page group id (`_id` from `cidaas_hosted_page_group`).",
					},
					"theme":           schema.StringAttribute{Optional: true},
					"logo_uri":        schema.StringAttribute{Optional: true},
					"policy_uri":      schema.StringAttribute{Optional: true},
					"tos_uri":         schema.StringAttribute{Optional: true},
					"imprint_uri":     schema.StringAttribute{Optional: true},
					"primary_color":   schema.StringAttribute{Optional: true},
					"accent_color":    schema.StringAttribute{Optional: true},
					"background_uri":  schema.StringAttribute{Optional: true},
					"content_align":   schema.StringAttribute{Optional: true},
					"logo_align":      schema.StringAttribute{Optional: true},
					"media_type":      schema.StringAttribute{Optional: true},
					"video_url":       schema.StringAttribute{Optional: true},
					"fav_icon":        schema.StringAttribute{Optional: true},
				},
			},
			"resources": schema.MapNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"translation_set": schema.StringAttribute{Required: true},
						"theme":           schema.StringAttribute{Optional: true},
						"layout":          schema.StringAttribute{Optional: true},
					},
				},
			},
			"group_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Developer/admin group id from the access token.",
			},
			"owner": schema.StringAttribute{
				Computed: true,
			},
			"fingerprint": schema.StringAttribute{
				Computed: true,
			},
			"created_time": schema.StringAttribute{Computed: true},
			"updated_time": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *hostedPageLayoutResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func layoutToAPI(n layoutNested) client.LayoutDetail {
	return client.LayoutDetail{
		HostedPageGroup: n.HostedPageGroup.ValueString(),
		Theme:           n.Theme.ValueString(),
		LogoURI:         n.LogoURI.ValueString(),
		PolicyURI:       n.PolicyURI.ValueString(),
		TosURI:          n.TosURI.ValueString(),
		ImprintURI:      n.ImprintURI.ValueString(),
		PrimaryColor:    n.PrimaryColor.ValueString(),
		AccentColor:     n.AccentColor.ValueString(),
		BackgroundURI:   n.BackgroundURI.ValueString(),
		ContentAlign:    n.ContentAlign.ValueString(),
		LogoAlign:       n.LogoAlign.ValueString(),
		MediaType:       n.MediaType.ValueString(),
		VideoURL:        n.VideoURL.ValueString(),
		FavIcon:         n.FavIcon.ValueString(),
	}
}

func (r *hostedPageLayoutResource) toAPI(ctx context.Context, m hostedPageLayoutModel) (client.HostedPageLayoutWrite, error) {
	out := client.HostedPageLayoutWrite{
		Description: m.Description.ValueString(),
	}
	var layout layoutNested
	diags := m.Layout.As(ctx, &layout, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return out, fmt.Errorf("%s", diags.Errors())
	}
	out.Layout = layoutToAPI(layout)

	if m.Resources.IsNull() || m.Resources.IsUnknown() {
		return out, nil
	}
	var raw map[string]layoutResourceNested
	if diags := m.Resources.ElementsAs(ctx, &raw, false); diags.HasError() {
		return out, fmt.Errorf("%s", diags.Errors())
	}
	out.Resources = make(map[string]client.LayoutResourceConfig, len(raw))
	for k, v := range raw {
		out.Resources[k] = client.LayoutResourceConfig{
			TranslationSet: v.TranslationSet.ValueString(),
			Theme:          v.Theme.ValueString(),
			Layout:         v.Layout.ValueString(),
		}
	}
	return out, nil
}

func layoutFromAPI(l client.LayoutDetail) (types.Object, error) {
	obj, diags := types.ObjectValue(layoutAttrTypes(), map[string]attr.Value{
		"hosted_page_group": stringOrNull(l.HostedPageGroup),
		"theme":             stringOrNull(l.Theme),
		"logo_uri":          stringOrNull(l.LogoURI),
		"policy_uri":        stringOrNull(l.PolicyURI),
		"tos_uri":           stringOrNull(l.TosURI),
		"imprint_uri":       stringOrNull(l.ImprintURI),
		"primary_color":     stringOrNull(l.PrimaryColor),
		"accent_color":      stringOrNull(l.AccentColor),
		"background_uri":    stringOrNull(l.BackgroundURI),
		"content_align":     stringOrNull(l.ContentAlign),
		"logo_align":        stringOrNull(l.LogoAlign),
		"media_type":        stringOrNull(l.MediaType),
		"video_url":         stringOrNull(l.VideoURL),
		"fav_icon":          stringOrNull(l.FavIcon),
	})
	if diags.HasError() {
		return types.ObjectNull(layoutAttrTypes()), fmt.Errorf("%s", diags.Errors())
	}
	return obj, nil
}

func resourcesFromAPI(resources map[string]client.LayoutResourceConfig) (types.Map, error) {
	if len(resources) == 0 {
		return types.MapNull(types.ObjectType{AttrTypes: layoutResourceAttrTypes()}), nil
	}
	elems := make(map[string]attr.Value, len(resources))
	for k, v := range resources {
		obj, diags := types.ObjectValue(layoutResourceAttrTypes(), map[string]attr.Value{
			"translation_set": types.StringValue(v.TranslationSet),
			"theme":             stringOrNull(v.Theme),
			"layout":            stringOrNull(v.Layout),
		})
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: layoutResourceAttrTypes()}), fmt.Errorf("%s", diags.Errors())
		}
		elems[k] = obj
	}
	m, diags := types.MapValue(types.ObjectType{AttrTypes: layoutResourceAttrTypes()}, elems)
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: layoutResourceAttrTypes()}), fmt.Errorf("%s", diags.Errors())
	}
	return m, nil
}

func (r *hostedPageLayoutResource) fromAPI(_ context.Context, data client.HostedPageLayoutModel, state *hostedPageLayoutModel) error {
	state.ID = types.StringValue(data.ID)
	state.Description = types.StringValue(data.Description)
	state.GroupID = stringOrNull(data.GroupID)
	state.Owner = stringOrNull(data.Owner)
	state.Fingerprint = stringOrNull(data.Fingerprint)
	state.CreatedTime = stringOrNull(data.CreatedTime)
	state.UpdatedTime = stringOrNull(data.UpdatedTime)
	layoutObj, err := layoutFromAPI(data.Layout)
	if err != nil {
		return err
	}
	state.Layout = layoutObj
	resources, err := resourcesFromAPI(data.Resources)
	if err != nil {
		return err
	}
	state.Resources = resources
	return nil
}

func (r *hostedPageLayoutResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hostedPageLayoutModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiModel, err := r.toAPI(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid hosted page layout", err.Error())
		return
	}
	out, err := r.client.Layouts.Create(ctx, apiModel)
	if err != nil {
		resp.Diagnostics.AddError("Create hosted page layout failed", err.Error())
		return
	}
	if err := r.fromAPI(ctx, out.Data, &plan); err != nil {
		resp.Diagnostics.AddError("Map response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *hostedPageLayoutResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hostedPageLayoutModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.Layouts.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read hosted page layout failed", err.Error())
		return
	}
	if err := r.fromAPI(ctx, out.Data, &state); err != nil {
		resp.Diagnostics.AddError("Map response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *hostedPageLayoutResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan hostedPageLayoutModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiModel, err := r.toAPI(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid hosted page layout", err.Error())
		return
	}
	out, err := r.client.Layouts.Update(ctx, plan.ID.ValueString(), apiModel)
	if err != nil {
		resp.Diagnostics.AddError("Update hosted page layout failed", err.Error())
		return
	}
	if err := r.fromAPI(ctx, out.Data, &plan); err != nil {
		resp.Diagnostics.AddError("Map response", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *hostedPageLayoutResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hostedPageLayoutModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Layouts.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete hosted page layout failed", err.Error())
		return
	}
}

func (r *hostedPageLayoutResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
