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

const WEBHOOK_EVENTS_DATASOURCE = "cidaas_webhook_events" // nolint:stylecheck

type webhookEventsDataSource struct {
	BaseDataSource
}

func NewWebhookEvents() datasource.DataSource {
	return &webhookEventsDataSource{
		BaseDataSource: NewBaseDataSource(BaseDataSourceConfig{
			Name:   WEBHOOK_EVENTS_DATASOURCE,
			Schema: &webhookEventsSchema,
		}),
	}
}

var webhookEventsSchema = schema.Schema{
	MarkdownDescription: "The data source `cidaas_webhook_events` returns webhook-capable event IDs from your Cidaas instance " +
		"(`GET /webhook-srv/eventdescriptions?category=webhook`). Use these values in `cidaas_webhook.events`." +
		"\n\n Ensure that the below scope is assigned to the client:" +
		"\n- cidaas:webhook_read",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Stable datasource instance id (random UUID).",
		},
		"events": schema.ListNestedAttribute{
			Computed:            true,
			MarkdownDescription: "Webhook-capable event descriptions.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Event identifier (`_id`), e.g. `ACCOUNT_MODIFIED`.",
					},
					"object_type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Object type the event applies to, e.g. `users`.",
					},
				},
			},
		},
	},
}

type webhookEventsModel struct {
	ID     types.String `tfsdk:"id"`
	Events types.List   `tfsdk:"events"`
}

func (d *webhookEventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config webhookEventsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := d.Client.Webhook.ListEventDescriptions(ctx, "webhook")
	if err != nil {
		resp.Diagnostics.AddError("failed to list webhook events", util.FormatErrorMessage(err))
		return
	}

	attrTypes := map[string]attr.Type{
		"id":          types.StringType,
		"object_type": types.StringType,
	}
	elemType := types.ObjectType{AttrTypes: attrTypes}
	elems := make([]attr.Value, 0, len(list))
	for _, ed := range list {
		o, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"id":          types.StringValue(ed.ID),
			"object_type": types.StringValue(ed.ObjectType),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, o)
	}
	events, diags := types.ListValue(elemType, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out := webhookEventsModel{
		ID:     types.StringValue(uuid.New().String()),
		Events: events,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
