package datasources

import (
	"context"
	"encoding/json"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const notificationTemplateGroupsGraphDataSourceName = "cidaas_notification_template_groups"

type notificationTemplateGroupsGraphDataSource struct {
	BaseDataSource
}

func NewNotificationTemplateGroupsGraph() datasource.DataSource {
	return &notificationTemplateGroupsGraphDataSource{
		BaseDataSource: NewBaseDataSource(BaseDataSourceConfig{
			Name:   notificationTemplateGroupsGraphDataSourceName,
			Schema: &notificationTemplateGroupsGraphSchema,
		}),
	}
}

var notificationTemplateGroupsGraphSchema = schema.Schema{
	MarkdownDescription: "Runs **POST** `/{notifications_context_path}/graph/templategroups/` with a graph filter JSON body " +
		"and returns matching template groups.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Stable datasource id (random UUID).",
		},
		"graph_filter": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "JSON body for the graph filter.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(2),
			},
		},
		"groups": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Template group `_id`.",
					},
					"tg_type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Template group type.",
					},
					"description": schema.StringAttribute{Computed: true},
					"default_locale": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Default locale.",
					},
				},
			},
		},
	},
}

type notificationTemplateGroupsGraphModel struct {
	ID          types.String `tfsdk:"id"`
	GraphFilter types.String `tfsdk:"graph_filter"`
	Groups      types.List   `tfsdk:"groups"`
}

func (d *notificationTemplateGroupsGraphDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config notificationTemplateGroupsGraphModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	raw := json.RawMessage(config.GraphFilter.ValueString())
	list, err := d.Client.NotificationsSrvTemplateGroup.FindGraphGroups(ctx, raw)
	if err != nil {
		resp.Diagnostics.AddError("failed to query graph/templategroups", util.FormatErrorMessage(err))
		return
	}
	attrTypes := map[string]attr.Type{
		"id":             types.StringType,
		"tg_type":        types.StringType,
		"description":    types.StringType,
		"default_locale": types.StringType,
	}
	elemType := types.ObjectType{AttrTypes: attrTypes}
	elems := make([]attr.Value, 0, len(list))
	for _, g := range list {
		o, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"id":             types.StringValue(g.ID),
			"tg_type":        types.StringValue(g.TGType),
			"description":    types.StringValue(g.Description),
			"default_locale": types.StringValue(g.DefaultLocale),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, o)
	}
	groups, diags := types.ListValueFrom(ctx, elemType, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	out := notificationTemplateGroupsGraphModel{
		ID:          types.StringValue(uuid.New().String()),
		GraphFilter: config.GraphFilter,
		Groups:      groups,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
