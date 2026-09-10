package group

import (
	"context"
	"fmt"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//nolint:revive
type GroupSelectionResource struct {
	base.BaseResource
}

func NewGroupSelectionResource() resource.Resource {
	return &GroupSelectionResource{
		BaseResource: base.NewBaseResource(
			base.BaseResourceConfig{
				Name:   base.RESOURCE_GROUP_SELECTION,
				Schema: &groupSelectionSchema,
			},
		),
	}
}

type groupSelectionModel struct {
	ID                           types.String `tfsdk:"id"`
	Name                         types.String `tfsdk:"name"`
	Description                  types.String `tfsdk:"description"`
	IsGroupLoginSelectionEnabled types.Bool   `tfsdk:"is_group_login_selection_enabled"`
	AlwaysShowGroupSelection     types.Bool   `tfsdk:"always_show_group_selection"`
	SelectableGroups             types.Set    `tfsdk:"selectable_groups"`
	SelectableGroupTypes         types.Set    `tfsdk:"selectable_group_types"`
	CreatedAt                    types.String `tfsdk:"created_at"`
	UpdatedAt                    types.String `tfsdk:"updated_at"`
}

var groupSelectionSchema = schema.Schema{
	MarkdownDescription: "Resource for managing group selection configurations in cidaas v4.x (Trustdesk).",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The unique identifier of the group selection configuration.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The name of the group selection configuration.",
		},
		"description": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The description of the group selection configuration.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"is_group_login_selection_enabled": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(true),
			MarkdownDescription: "Whether group login selection is enabled.",
		},
		"always_show_group_selection": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Whether to always display group selection UI during login.",
		},
		"selectable_groups": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "List of allowed user group `groupId` values. Groups must exist on the tenant.",
		},
		"selectable_group_types": schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "List of allowed group type identifiers. Group types must exist on the tenant.",
		},
		"created_at": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Timestamp when the group selection was created.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Timestamp when the group selection was last updated.",
		},
	},
}

func (r *GroupSelectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}
	if c.Capabilities.TargetVersion == "3.x" {
		resp.Diagnostics.AddError(
			"Resource Not Supported on cidaas v3.x",
			"The resource `cidaas_group_selection` is only supported on cidaas v4.x (Trustdesk), but the provider is configured for v3.x. Please set cidaas_version = \"4.x\" in your provider block.",
		)
		return
	}
	r.BaseResource.Configure(ctx, req, resp)
}

//nolint:dupl
func (r *GroupSelectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupSelectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := groupSelectionModelToAPI(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.CidaasClient.GroupSelection.Upsert(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create group selection", util.FormatErrorMessage(err))
		return
	}

	state := groupSelectionModelFromAPI(ctx, res.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//nolint:dupl
func (r *GroupSelectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupSelectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.CidaasClient.GroupSelection.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read group selection", util.FormatErrorMessage(err))
		return
	}

	newState := groupSelectionModelFromAPI(ctx, res.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

//nolint:dupl
func (r *GroupSelectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan groupSelectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := groupSelectionModelToAPI(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	apiReq.ID = plan.ID.ValueString()

	res, err := r.CidaasClient.GroupSelection.Upsert(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("failed to update group selection", util.FormatErrorMessage(err))
		return
	}

	state := groupSelectionModelFromAPI(ctx, res.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//nolint:dupl
func (r *GroupSelectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupSelectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.CidaasClient.GroupSelection.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete group selection", util.FormatErrorMessage(err))
	}
}

func (r *GroupSelectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func groupSelectionModelToAPI(ctx context.Context, m groupSelectionModel, diags *diag.Diagnostics) cidaas.GroupSelectionModel {
	var selectableGroups, selectableGroupTypes []string
	if !m.SelectableGroups.IsNull() && !m.SelectableGroups.IsUnknown() {
		diags.Append(m.SelectableGroups.ElementsAs(ctx, &selectableGroups, false)...)
	}
	if !m.SelectableGroupTypes.IsNull() && !m.SelectableGroupTypes.IsUnknown() {
		diags.Append(m.SelectableGroupTypes.ElementsAs(ctx, &selectableGroupTypes, false)...)
	}

	isEnabled := true
	if !m.IsGroupLoginSelectionEnabled.IsNull() && !m.IsGroupLoginSelectionEnabled.IsUnknown() {
		isEnabled = m.IsGroupLoginSelectionEnabled.ValueBool()
	}

	alwaysShow := false
	if !m.AlwaysShowGroupSelection.IsNull() && !m.AlwaysShowGroupSelection.IsUnknown() {
		alwaysShow = m.AlwaysShowGroupSelection.ValueBool()
	}

	return cidaas.GroupSelectionModel{
		ID:          m.ID.ValueString(),
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		GroupSelection: cidaas.GroupSelectionDetails{
			IsGroupLoginSelectionEnabled: isEnabled,
			AlwaysShowGroupSelection:     alwaysShow,
			SelectableGroups:             selectableGroups,
			SelectableGroupTypes:         selectableGroupTypes,
		},
	}
}

func groupSelectionModelFromAPI(ctx context.Context, data cidaas.GroupSelectionModel, diags *diag.Diagnostics) groupSelectionModel {
	var m groupSelectionModel
	m.ID = util.StringValueOrNull(&data.ID)
	m.Name = util.StringValueOrNull(&data.Name)
	m.Description = util.StringValueOrNull(&data.Description)
	m.IsGroupLoginSelectionEnabled = types.BoolValue(data.GroupSelection.IsGroupLoginSelectionEnabled)
	m.AlwaysShowGroupSelection = types.BoolValue(data.GroupSelection.AlwaysShowGroupSelection)

	selectableGroupsSet, setDiags := types.SetValueFrom(ctx, types.StringType, data.GroupSelection.SelectableGroups)
	diags.Append(setDiags...)
	m.SelectableGroups = selectableGroupsSet

	selectableGroupTypesSet, setDiags := types.SetValueFrom(ctx, types.StringType, data.GroupSelection.SelectableGroupTypes)
	diags.Append(setDiags...)
	m.SelectableGroupTypes = selectableGroupTypesSet

	m.CreatedAt = util.StringValueOrNull(&data.CreatedAt)
	m.UpdatedAt = util.StringValueOrNull(&data.UpdatedAt)
	return m
}
