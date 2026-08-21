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

type CustomProviderDataSource struct {
	BaseDataSource
}

type CustomProviderFilterModel struct {
	BaseModel
	CustomProvider []CustomProvider `tfsdk:"custom_provider"`
}

type CustomProvider struct {
	ID           types.String `tfsdk:"id"`
	ProviderName types.String `tfsdk:"provider_name"`
	StandardType types.String `tfsdk:"standard_type"`
	Domains      types.Set    `tfsdk:"domains"`
}

var customProviderFilter = FilterConfig{
	"provider_name": {TypeFunc: FilterTypeString},
	"standard_type": {TypeFunc: FilterTypeString},
}

func NewCustomProvider() datasource.DataSource {
	return &CustomProviderDataSource{
		BaseDataSource: NewBaseDataSource(
			BaseDataSourceConfig{
				Name:   CUSTOM_PROVIDER_DATASOURCE,
				Schema: &customProviderSchema,
			},
		),
	}
}

var customProviderSchema = schema.Schema{
	MarkdownDescription: fmt.Sprintf("The data source `%s` returns a list of custom providers available in your Cidaas instance."+
		"\nYou can apply filters using the `filter` block in your Terraform configuration.", CUSTOM_PROVIDER_DATASOURCE),
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The data source's unique ID.",
			Computed:    true,
		},
	},
	Blocks: map[string]schema.Block{
		"filter": customProviderFilter.Schema(),
		"custom_provider": schema.ListNestedBlock{
			Description: "The returned list of custom providers.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The unique identifier of the custom provider.",
					},
					"provider_name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The name of the custom provider.",
					},
					"standard_type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Type of standard. `OAUTH2` or `OPENID_CONNECT`.",
					},
					"domains": schema.SetAttribute{
						ElementType:         types.StringType,
						Computed:            true,
						MarkdownDescription: "The domains of the provider.",
					},
				},
			},
		},
	},
}

func (d *CustomProviderDataSource) Read( //nolint:dupl
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data CustomProviderFilterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get config data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	data.ID = types.StringValue(uuid.New().String())
	result, diag := customProviderFilter.GetAndFilter(ctx, d.Client, data.Filters, listCustomProviders)
	if diag != nil {
		tflog.Error(ctx, "failed to filter custom_provider data", util.H{
			"error":  diag.Summary(),
			"detail": diag.Detail(),
		})
		resp.Diagnostics.Append(diag)
		return
	}
	tflog.Debug(ctx, "successfully filtered custom_provider data")

	data.CustomProvider = parseModel(
		AnySliceToTyped[cidaas.CustomProviderModel](result),
		parseCustomProvider,
	)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Info(ctx, "successfully read custom_provider data source")
}

func listCustomProviders(ctx context.Context, client *cidaas.Client) ([]any, error) {
	cps, err := client.CustomProvider.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return TypedSliceToAny(cps), nil
}

func parseCustomProvider(cp cidaas.CustomProviderModel) CustomProvider {
	return CustomProvider{
		ID:           types.StringValue(cp.ID),
		ProviderName: types.StringValue(cp.ProviderName),
		Domains:      util.SetValueOrNull(cp.Domains),
		StandardType: types.StringValue(cp.StandardType),
	}
}
