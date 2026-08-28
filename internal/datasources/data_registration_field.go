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

type RegistrationFieldsDataSource struct {
	BaseDataSource
}

type RegistrationFieldsFilterModel struct {
	BaseModel
	RegistrationField []RegistrationField `tfsdk:"registration_field"`
}

type RegistrationField struct {
	ID            types.String `tfsdk:"id"`
	FieldKey      types.String `tfsdk:"field_key"`
	DataType      types.String `tfsdk:"data_type"`
	Internal      types.Bool   `tfsdk:"internal"`
	ReadOnly      types.Bool   `tfsdk:"read_only"`
	Required      types.Bool   `tfsdk:"required"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	ParentGroupID types.String `tfsdk:"parent_group_id"`
	FieldType     types.String `tfsdk:"field_type"`
	Order         types.Int64  `tfsdk:"order"`
}

var registrationFieldsFilter = FilterConfig{
	"parent_group_id": {TypeFunc: FilterTypeString},
	"field_type":      {TypeFunc: FilterTypeString},
	"data_type":       {TypeFunc: FilterTypeString},
	"field_key":       {TypeFunc: FilterTypeString},
	"required":        {TypeFunc: FilterTypeBool},
	"internal":        {TypeFunc: FilterTypeBool},
	"read_only":       {TypeFunc: FilterTypeBool},
	"enabled":         {TypeFunc: FilterTypeBool},
}

var fieldSchema = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "The unique identifier of the group type.",
	},
	"parent_group_id": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "The ID of the parent registration group.",
	},
	"field_type": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Specifies whether the field type is `SYSTEM` or `CUSTOM`.",
	},
	"data_type": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "The data type of the field.",
	},
	"field_key": schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "The unique name of the registration field.",
	},
	"required": schema.BoolAttribute{
		Computed:            true,
		MarkdownDescription: "Flag to identify if a field is required in registration.",
	},
	"internal": schema.BoolAttribute{
		Computed:            true,
		MarkdownDescription: "Flag to identify if a field is internal.",
	},
	"enabled": schema.BoolAttribute{
		Computed:            true,
		MarkdownDescription: "Flag to identify if a field is enabled.",
	},

	"read_only": schema.BoolAttribute{
		Computed:            true,
		MarkdownDescription: "Flag to identify if a field is read only.",
	},
	"order": schema.Int64Attribute{
		Computed:            true,
		MarkdownDescription: "The order of the Field in the UI.",
	},
}

var fieldDataSourceSchema = schema.Schema{
	MarkdownDescription: fmt.Sprintf("The data source `%s` returns a list of registration fields available in your Cidaas instance."+
		"\nYou can apply filters using the `filter` block in your Terraform configuration.", REG_FIELD_DATASOURCE),
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The data source's unique ID.",
			Computed:    true,
		},
	},
	Blocks: map[string]schema.Block{
		"filter": registrationFieldsFilter.Schema(),
		"registration_field": schema.ListNestedBlock{
			Description: "The returned list of registration fields.",
			NestedObject: schema.NestedBlockObject{
				Attributes: fieldSchema,
			},
		},
	},
}

func NewRegistrationField() datasource.DataSource {
	return &RegistrationFieldsDataSource{
		BaseDataSource: NewBaseDataSource(
			BaseDataSourceConfig{
				Name:   REG_FIELD_DATASOURCE,
				Schema: &fieldDataSourceSchema,
			},
		),
	}
}

func (d *RegistrationFieldsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) { //nolint:dupl
	var data RegistrationFieldsFilterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get config data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	data.ID = types.StringValue(uuid.New().String())
	result, diag := registrationFieldsFilter.GetAndFilter(ctx, d.Client, data.Filters, listRegistrationFieldss)
	if diag != nil {
		tflog.Error(ctx, "failed to filter registration_field data", util.H{
			"error":  diag.Summary(),
			"detail": diag.Detail(),
		})
		resp.Diagnostics.Append(diag)
		return
	}
	tflog.Debug(ctx, "successfully filtered registration_field data")

	data.RegistrationField = parseModel(
		AnySliceToTyped[cidaas.RegistrationFieldConfig](result),
		parseRegistrationField,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Info(ctx, "successfully read registration_field data source")
}

func listRegistrationFieldss(ctx context.Context, client *cidaas.Client) ([]any, error) {
	rfs, err := client.RegFields.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return TypedSliceToAny(rfs), nil
}

func parseRegistrationField(rf cidaas.RegistrationFieldConfig) RegistrationField {
	return RegistrationField{
		ID:            types.StringValue(rf.ID),
		FieldKey:      types.StringValue(rf.FieldKey),
		DataType:      types.StringValue(rf.DataType),
		FieldType:     types.StringValue(rf.FieldType),
		Required:      types.BoolValue(rf.Required),
		ReadOnly:      types.BoolValue(rf.ReadOnly),
		Internal:      types.BoolValue(rf.Internal),
		ParentGroupID: types.StringValue(rf.ParentGroupID),
		Order:         types.Int64Value(rf.Order),
	}
}
