package group

import (
	"context"
	"fmt"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/base"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//nolint:revive
type GroupVerificationFilterResource struct {
	base.BaseResource
}

func NewGroupVerificationFilterResource() resource.Resource {
	return &GroupVerificationFilterResource{
		BaseResource: base.NewBaseResource(
			base.BaseResourceConfig{
				Name:   base.RESOURCE_GROUP_VERIFICATION_FILTER,
				Schema: &groupVerificationFilterSchema,
			},
		),
	}
}

type roleFilterModel struct {
	Roles          types.List   `tfsdk:"roles"`
	MatchCondition types.String `tfsdk:"match_condition"`
}

type groupVerificationFilterItemModel struct {
	GroupID    types.String     `tfsdk:"group_id"`
	GroupType  types.String     `tfsdk:"group_type"`
	RoleFilter *roleFilterModel `tfsdk:"role_filter"`
}

type groupVerificationFilterModel struct {
	ID             types.String                       `tfsdk:"id"`
	Description    types.String                       `tfsdk:"description"`
	MatchCondition types.String                       `tfsdk:"match_condition"`
	Filters        []groupVerificationFilterItemModel `tfsdk:"filters"`
	CreatedAt      types.String                       `tfsdk:"created_at"`
	UpdatedAt      types.String                       `tfsdk:"updated_at"`
}

var groupVerificationFilterSchema = schema.Schema{
	MarkdownDescription: "Resource for managing group verification request filters in cidaas v4.x (Trustdesk).",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The unique identifier of the group verification filter. If omitted, a unique ID will be auto-generated.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"description": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Description of the functionality this group verification filter restricts (max 600 characters).",
			Validators: []validator.String{
				stringvalidator.LengthAtMost(600),
			},
		},
		"match_condition": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Match condition for top-level filters. Must be either `AND` or `OR`.",
			Validators: []validator.String{
				stringvalidator.OneOf("AND", "OR", "and", "or"),
			},
		},
		"created_at": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Timestamp when the verification filter was created.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Timestamp when the verification filter was last updated.",
		},
	},
	Blocks: map[string]schema.Block{
		"filters": schema.ListNestedBlock{
			MarkdownDescription: "List of group verification filters.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"group_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Group identifier filter.",
					},
					"group_type": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Group type identifier filter.",
					},
				},
				Blocks: map[string]schema.Block{
					"role_filter": schema.SingleNestedBlock{
						MarkdownDescription: "Role filter criteria for the group.",
						Attributes: map[string]schema.Attribute{
							"roles": schema.ListAttribute{
								ElementType:         types.StringType,
								Optional:            true,
								MarkdownDescription: "List of role names to verify.",
							},
							"match_condition": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Match condition for roles (`AND` or `OR`).",
								Validators: []validator.String{
									stringvalidator.OneOf("AND", "OR", "and", "or"),
								},
							},
						},
					},
				},
			},
		},
	},
}

func (r *GroupVerificationFilterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
			"The resource `cidaas_group_verification_filter` is only supported on cidaas v4.x (Trustdesk), but the provider is configured for v3.x. Please set cidaas_version = \"4.x\" in your provider block.",
		)
		return
	}
	r.BaseResource.Configure(ctx, req, resp)
}

//nolint:dupl
func (r *GroupVerificationFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupVerificationFilterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := groupVerificationFilterModelToAPI(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.CidaasClient.GroupVerificationFilter.Create(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create group verification filter", util.FormatErrorMessage(err))
		return
	}

	state := groupVerificationFilterModelFromAPI(ctx, res.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//nolint:dupl
func (r *GroupVerificationFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupVerificationFilterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.CidaasClient.GroupVerificationFilter.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read group verification filter", util.FormatErrorMessage(err))
		return
	}

	newState := groupVerificationFilterModelFromAPI(ctx, res.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

//nolint:dupl
func (r *GroupVerificationFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan groupVerificationFilterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := groupVerificationFilterModelToAPI(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	apiReq.ID = plan.ID.ValueString()

	res, err := r.CidaasClient.GroupVerificationFilter.Update(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("failed to update group verification filter", util.FormatErrorMessage(err))
		return
	}

	state := groupVerificationFilterModelFromAPI(ctx, res.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//nolint:dupl
func (r *GroupVerificationFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupVerificationFilterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.CidaasClient.GroupVerificationFilter.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete group verification filter", util.FormatErrorMessage(err))
	}
}

func (r *GroupVerificationFilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func groupVerificationFilterModelToAPI(ctx context.Context, m groupVerificationFilterModel, diags *diag.Diagnostics) cidaas.GroupVerificationRequestModel {
	var filtersWire []cidaas.GroupVerificationFilterItemWire
	for _, f := range m.Filters {
		var roleFilterWire *cidaas.RoleVerificationFilterWire
		if f.RoleFilter != nil {
			var roles []string
			if !f.RoleFilter.Roles.IsNull() && !f.RoleFilter.Roles.IsUnknown() {
				diags.Append(f.RoleFilter.Roles.ElementsAs(ctx, &roles, false)...)
			}
			roleFilterWire = &cidaas.RoleVerificationFilterWire{
				Roles:          roles,
				MatchCondition: strings.ToLower(f.RoleFilter.MatchCondition.ValueString()),
			}
		}

		filtersWire = append(filtersWire, cidaas.GroupVerificationFilterItemWire{
			GroupID:    f.GroupID.ValueString(),
			GroupType:  f.GroupType.ValueString(),
			RoleFilter: roleFilterWire,
		})
	}

	id := strings.ToLower(m.ID.ValueString())
	if id == "" {
		id = strings.ToLower(util.GenerateUUID())
	}

	return cidaas.GroupVerificationRequestModel{
		ID:             id,
		Description:    m.Description.ValueString(),
		MatchCondition: strings.ToLower(m.MatchCondition.ValueString()),
		Filters:        filtersWire,
	}
}

func groupVerificationFilterModelFromAPI(ctx context.Context, data cidaas.GroupVerificationRequestModel, diags *diag.Diagnostics) groupVerificationFilterModel {
	var m groupVerificationFilterModel
	m.ID = util.StringValueOrNull(&data.ID)
	m.Description = util.StringValueOrNull(&data.Description)
	m.MatchCondition = util.StringValueOrNull(&data.MatchCondition)

	var filtersList []groupVerificationFilterItemModel
	for _, fWire := range data.Filters {
		var rf *roleFilterModel
		if fWire.RoleFilter != nil && (len(fWire.RoleFilter.Roles) > 0 || fWire.RoleFilter.MatchCondition != "") {
			rolesList, setDiags := types.ListValueFrom(ctx, types.StringType, fWire.RoleFilter.Roles)
			diags.Append(setDiags...)
			rf = &roleFilterModel{
				Roles:          rolesList,
				MatchCondition: util.StringValueOrNull(&fWire.RoleFilter.MatchCondition),
			}
		}

		filtersList = append(filtersList, groupVerificationFilterItemModel{
			GroupID:    util.StringValueOrNull(&fWire.GroupID),
			GroupType:  util.StringValueOrNull(&fWire.GroupType),
			RoleFilter: rf,
		})
	}

	m.Filters = filtersList
	m.CreatedAt = util.StringValueOrNull(&data.CreatedAt)
	m.UpdatedAt = util.StringValueOrNull(&data.UpdatedAt)
	return m
}
