package datasources

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const notificationServiceSetupsDataSourceName = "cidaas_notification_service_setups"

type notificationServiceSetupsDataSource struct {
	BaseDataSource
}

func NewNotificationServiceSetups() datasource.DataSource {
	return &notificationServiceSetupsDataSource{
		BaseDataSource: NewBaseDataSource(BaseDataSourceConfig{
			Name:   notificationServiceSetupsDataSourceName,
			Schema: &notificationServiceSetupsSchema,
		}),
	}
}

var notificationServiceSetupsSchema = schema.Schema{
	MarkdownDescription: "Lists **active** service setups from notification-srv `GET /{notifications_context_path}/servicesetups/` " +
		"so you can reference `service_setup_id` values in `cidaas_notifications_template_group` comm settings. " +
		"Setups that are still `in-progress` (not yet verified) do not appear; use `data.cidaas_notification_service_setup` or the managed resource to read by id.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Stable datasource instance id (random UUID).",
		},
		"setups": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Service setup `_id` (use as `service_setup_id` in comm settings).",
					},
					"name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Human-readable name.",
					},
					"status": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Setup status.",
					},
					"description": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Description.",
					},
					"parent_service_setup_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Parent setup id when applicable.",
					},
					"has_remote_templates": schema.BoolAttribute{
						Computed:            true,
						MarkdownDescription: "Whether templates are remote.",
					},
					"service_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Service id from `serviceDescInfo`.",
					},
					"service_category": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Service category from `serviceDescInfo`.",
					},
				},
			},
		},
	},
}

type notificationServiceSetupsModel struct {
	ID     types.String `tfsdk:"id"`
	Setups types.List   `tfsdk:"setups"`
}

func (d *notificationServiceSetupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config notificationServiceSetupsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	list, err := d.Client.NotificationsSrvServiceSetup.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to list service setups", util.FormatErrorMessage(err))
		return
	}
	attrTypes := map[string]attr.Type{
		"id":                      types.StringType,
		"name":                    types.StringType,
		"status":                  types.StringType,
		"description":             types.StringType,
		"parent_service_setup_id": types.StringType,
		"has_remote_templates":    types.BoolType,
		"service_id":              types.StringType,
		"service_category":        types.StringType,
	}
	elemType := types.ObjectType{AttrTypes: attrTypes}
	elems := make([]attr.Value, 0, len(list))
	for _, s := range list {
		o, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"id":                      types.StringValue(s.ID),
			"name":                    types.StringValue(s.Name),
			"status":                  types.StringValue(s.Status),
			"description":             types.StringValue(s.Description),
			"parent_service_setup_id": types.StringValue(s.ParentServiceSetupID),
			"has_remote_templates":    types.BoolValue(s.HasRemoteTemplates),
			"service_id":              types.StringValue(s.ServiceDescInfo.ServiceID),
			"service_category":        types.StringValue(s.ServiceDescInfo.ServiceCategory),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, o)
	}
	setups, diags := types.ListValueFrom(ctx, elemType, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	out := notificationServiceSetupsModel{
		ID:     types.StringValue(uuid.New().String()),
		Setups: setups,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
