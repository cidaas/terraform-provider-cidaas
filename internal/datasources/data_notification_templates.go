package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"sort"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const notificationTemplatesDataSourceName = "cidaas_notification_templates"

type notificationTemplatesDataSource struct {
	BaseDataSource
}

func NewNotificationTemplates() datasource.DataSource {
	return &notificationTemplatesDataSource{
		BaseDataSource: NewBaseDataSource(BaseDataSourceConfig{
			Name:   notificationTemplatesDataSourceName,
			Schema: &notificationTemplatesGraphSchema,
		}),
	}
}

var notificationTemplatesGraphSchema = schema.Schema{
	MarkdownDescription: "Runs **POST** `/{notifications_context_path}/graph/templates/` with a graph filter JSON body " +
		"and returns matching templates (e.g. filter by `groupId`).",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Stable datasource id derived from `graph_filter` (same filter → same id across runs).",
		},
		"graph_filter": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "JSON body for the graph filter (e.g. `{\"filter\":{...}}` per notification-srv / graphfilters).",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(2),
			},
		},
		"templates": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{Computed: true},
					"group_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "`groupId`",
					},
					"template_key":         schema.StringAttribute{Computed: true},
					"communication_method": schema.StringAttribute{Computed: true},
					"locale":               schema.StringAttribute{Computed: true},
					"enabled":              schema.BoolAttribute{Computed: true},
				},
			},
		},
	},
}

type notificationTemplatesGraphModel struct {
	ID          types.String `tfsdk:"id"`
	GraphFilter types.String `tfsdk:"graph_filter"`
	Templates   types.List   `tfsdk:"templates"`
}

func (d *notificationTemplatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config notificationTemplatesGraphModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	raw := json.RawMessage(config.GraphFilter.ValueString())
	list, err := d.Client.NotificationsSrvTemplate.FindGraph(ctx, raw)
	if err != nil {
		resp.Diagnostics.AddError("failed to query graph/templates", util.FormatErrorMessage(err))
		return
	}
	// Graph APIs often return rows in non-deterministic order; Terraform compares lists by index, so sort by id
	// or plans show spurious +/- churn when only order changes.
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	attrTypes := map[string]attr.Type{
		"id":                   types.StringType,
		"group_id":             types.StringType,
		"template_key":         types.StringType,
		"communication_method": types.StringType,
		"locale":               types.StringType,
		"enabled":              types.BoolType,
	}
	elemType := types.ObjectType{AttrTypes: attrTypes}
	elems := make([]attr.Value, 0, len(list))
	for _, t := range list {
		o, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"id":                   types.StringValue(t.ID),
			"group_id":             types.StringValue(t.GroupID),
			"template_key":         types.StringValue(t.TemplateKey),
			"communication_method": types.StringValue(t.CommunicationMethod),
			"locale":               types.StringValue(t.Locale),
			"enabled":              types.BoolValue(t.Enabled),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, o)
	}
	templates, diags := types.ListValueFrom(ctx, elemType, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	out := notificationTemplatesGraphModel{
		ID:          types.StringValue(notificationTemplatesDatasourceID(config.GraphFilter.ValueString())),
		GraphFilter: config.GraphFilter,
		Templates:   templates,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func notificationTemplatesDatasourceID(graphFilter string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(graphFilter))
	return fmt.Sprintf("notification-templates-%08x", h.Sum32())
}
