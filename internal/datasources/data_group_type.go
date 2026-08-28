package datasources

import (
	"context"
	"fmt"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type GroupTypeDataSource struct {
	BaseDataSource
}

type GroupTypeFilterModel struct {
	BaseModel
	GroupType []GroupType `tfsdk:"group_type"`
}

type GroupType struct {
	ID           types.String `tfsdk:"id"`
	RoleMode     types.String `tfsdk:"role_mode"`
	Description  types.String `tfsdk:"description"`
	GroupType    types.String `tfsdk:"group_type"`
	AllowedRoles types.Set    `tfsdk:"allowed_roles"`
}

var groupTypeFilter = FilterConfig{
	"group_type":    {TypeFunc: FilterTypeString},
	"role_mode":     {TypeFunc: FilterTypeString},
	"allowed_roles": {TypeFunc: FilterTypeString},
}

var groupTypeSchema = schema.Schema{
	MarkdownDescription: fmt.Sprintf("The data source `%s` returns a list of group types available in your Cidaas instance."+
		"\nYou can apply filters using the `filter` block in your Terraform configuration.", GROUP_TYPE_DATASOURCE),
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The data source's unique ID.",
			Computed:    true,
		},
	},
	Blocks: map[string]schema.Block{
		"filter": groupTypeFilter.Schema(),
		"group_type": schema.ListNestedBlock{
			Description: "The returned list of group types.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The identifier of the group type.",
					},
					"group_type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The unique identifier of the group type.",
					},
					"role_mode": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Determines the role mode for the user group type.",
					},
					"description": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The description of the group type",
					},
					"allowed_roles": schema.SetAttribute{
						ElementType:         types.StringType,
						Computed:            true,
						MarkdownDescription: "List of allowed roles in a group type.",
					},
				},
			},
		},
	},
}

func NewGroupType() datasource.DataSource {
	return &GroupTypeDataSource{
		BaseDataSource: NewBaseDataSource(
			BaseDataSourceConfig{
				Name:   GROUP_TYPE_DATASOURCE,
				Schema: &groupTypeSchema,
			},
		),
	}
}

func (d *GroupTypeDataSource) Read( //nolint:dupl
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data GroupTypeFilterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get config data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	data.ID = types.StringValue(uuid.New().String())
	result, diag := groupTypeFilter.GetAndFilter(ctx, d.Client, data.Filters, listGroupTypes)
	if diag != nil {
		tflog.Error(ctx, "failed to filter group_type data", util.H{
			"error":  diag.Summary(),
			"detail": diag.Detail(),
		})
		resp.Diagnostics.Append(diag)
		return
	}
	tflog.Debug(ctx, "successfully filtered group_type data")

	data.GroupType = parseModel(
		AnySliceToTyped[cidaas.GroupTypeData](result),
		parseGroupTypes,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Info(ctx, "successfully read group_type data source")
}

func listGroupTypes(ctx context.Context, client *cidaas.Client) ([]any, error) {
	groupTypes, err := client.GroupType.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return TypedSliceToAny(groupTypes), nil
}

func parseGroupTypes(groupType cidaas.GroupTypeData) GroupType {
	return GroupType{
		ID:           types.StringValue(groupType.ID),
		RoleMode:     types.StringValue(groupType.RoleMode),
		GroupType:    types.StringValue(groupType.GroupType),
		Description:  types.StringValue(groupType.Description),
		AllowedRoles: util.SetValueOrNull(groupType.AllowedRoles),
	}
}
