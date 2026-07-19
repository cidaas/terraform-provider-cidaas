package datasources

import (
	"context"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const NOTIFICATION_SERVICE_SETUP_DATASOURCE = "cidaas_notification_service_setup" // nolint:stylecheck

type notificationServiceSetupDataSource struct {
	BaseDataSource
}

func NewNotificationServiceSetup() datasource.DataSource {
	return &notificationServiceSetupDataSource{
		BaseDataSource: NewBaseDataSource(BaseDataSourceConfig{
			Name:   NOTIFICATION_SERVICE_SETUP_DATASOURCE,
			Schema: &notificationServiceSetupDataSchema,
		}),
	}
}

var notificationServiceSetupDataSchema = schema.Schema{
	MarkdownDescription: "Reads a single **service setup** from notification-srv `GET /{notifications_context_path}/servicesetups/{id}`. " +
		"Unlike the list datasource, this returns any status (including `in-progress`) for status polling after create.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Stable datasource instance id (random UUID).",
		},
		"service_setup_id": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Service setup `_id` to read.",
		},
		"name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Human-readable name.",
		},
		"status": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Setup status (`in-progress`, `active`, `inactive`).",
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
}

type notificationServiceSetupDataModel struct {
	ID                   types.String `tfsdk:"id"`
	ServiceSetupID       types.String `tfsdk:"service_setup_id"`
	Name                 types.String `tfsdk:"name"`
	Status               types.String `tfsdk:"status"`
	Description          types.String `tfsdk:"description"`
	ParentServiceSetupID types.String `tfsdk:"parent_service_setup_id"`
	HasRemoteTemplates   types.Bool   `tfsdk:"has_remote_templates"`
	ServiceID            types.String `tfsdk:"service_id"`
	ServiceCategory      types.String `tfsdk:"service_category"`
}

func (d *notificationServiceSetupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config notificationServiceSetupDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.Client.NotificationsSrvServiceSetup.Get(ctx, config.ServiceSetupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get service setup", util.FormatErrorMessage(err))
		return
	}
	if got == nil {
		resp.Diagnostics.AddError("service setup not found", "no service setup matches service_setup_id")
		return
	}
	out := notificationServiceSetupDataModel{
		ID:                   types.StringValue(uuid.New().String()),
		ServiceSetupID:       types.StringValue(got.ID),
		Name:                 types.StringValue(got.Name),
		Status:               types.StringValue(got.Status),
		Description:          types.StringValue(got.Description),
		ParentServiceSetupID: types.StringValue(got.ParentServiceSetupID),
		HasRemoteTemplates:   types.BoolValue(got.HasRemoteTemplates),
		ServiceID:            types.StringValue(got.ServiceDescInfo.ServiceID),
		ServiceCategory:      types.StringValue(got.ServiceDescInfo.ServiceCategory),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
