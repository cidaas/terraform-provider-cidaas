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

type RoleDataSource struct {
	BaseDataSource
}

type RoleFilterModel struct {
	BaseModel
	Role []Role `tfsdk:"role"`
}

type Role struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Role        types.String `tfsdk:"role"`
}

var roleFilter = FilterConfig{
	"role": {TypeFunc: FilterTypeString},
	"name": {TypeFunc: FilterTypeString},
}

func NewRole() datasource.DataSource {
	return &RoleDataSource{
		BaseDataSource: NewBaseDataSource(
			BaseDataSourceConfig{
				Name:   ROLE_DATASOURCE,
				Schema: &roleDataSourceSchema,
			},
		),
	}
}

var roleDataSourceSchema = schema.Schema{
	MarkdownDescription: fmt.Sprintf("The data source `%s` returns a list of roles available in your Cidaas instance."+
		"\nYou can apply filters using the `filter` block in your Terraform configuration.", ROLE_DATASOURCE),
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The data source's unique ID.",
			Computed:    true,
		},
	},
	Blocks: map[string]schema.Block{
		"filter": roleFilter.Schema(),
		"role": schema.ListNestedBlock{
			Description: "The returned list of roles.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"role": schema.StringAttribute{
						Computed:    true,
						Description: "The unique identifier of the role.",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The name of the role.",
					},
					"description": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The `description` of the role",
					},
				},
			},
		},
	},
}

func (d *RoleDataSource) Read( //nolint:dupl
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data RoleFilterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get config data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	data.ID = types.StringValue(uuid.New().String())
	result, diag := roleFilter.GetAndFilter(ctx, d.Client, data.Filters, listRoles)
	if diag != nil {
		tflog.Error(ctx, "failed to filter role data", util.H{
			"error":  diag.Summary(),
			"detail": diag.Detail(),
		})
		resp.Diagnostics.Append(diag)
		return
	}
	tflog.Debug(ctx, "successfully filtered role data")

	data.Role = parseModel(AnySliceToTyped[cidaas.RoleModel](result), parseRole)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Info(ctx, "successfully read role data source")
}

func listRoles(ctx context.Context, client *cidaas.Client) ([]any, error) {
	roles, err := client.Roles.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return TypedSliceToAny(roles), nil
}

func parseRole(role cidaas.RoleModel) Role {
	return Role{
		Name:        types.StringValue(role.Name),
		Role:        types.StringValue(role.Role),
		Description: types.StringValue(role.Description),
	}
}
