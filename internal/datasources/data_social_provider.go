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

type SocialProviderDataSource struct {
	BaseDataSource
}

type SocialProviderFilterModel struct {
	BaseModel
	SocialProvider []SocialProvider `tfsdk:"social_provider"`
}

type SocialProvider struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	ProviderName          types.String `tfsdk:"provider_name"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	ClientID              types.String `tfsdk:"client_id"`
	ClientSecret          types.String `tfsdk:"client_secret"`
	EnabledForAdminPortal types.Bool   `tfsdk:"enabled_for_admin_portal"`
	Scopes                types.Set    `tfsdk:"scopes"`
}

var socialProviderFilter = FilterConfig{
	"name":                     {TypeFunc: FilterTypeString},
	"provider_name":            {TypeFunc: FilterTypeString},
	"enabled":                  {TypeFunc: FilterTypeString},
	"enabled_for_admin_portal": {TypeFunc: FilterTypeString},
}

var socialProviderSchema = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "The unique identifier of the social provider",
	},
	"name": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "The name of the social provider configuration.",
	},
	"provider_name": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "The name of the social provider e.g; `google`, `facebook`, `linkedin` etc.",
	},
	"enabled": schema.BoolAttribute{
		Computed:            true,
		MarkdownDescription: "A flag to identify if a provider is enabled.",
	},
	"client_id": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "The client ID of the social provider.",
	},
	"client_secret": schema.StringAttribute{
		Computed:            true,
		Sensitive:           true,
		MarkdownDescription: "The client secret of the social provider.",
	},
	"scopes": schema.SetAttribute{
		ElementType:         types.StringType,
		Computed:            true,
		MarkdownDescription: "A list of scopes of the social provider.",
	},
	"enabled_for_admin_portal": schema.BoolAttribute{
		Computed:            true,
		MarkdownDescription: "A flag to identify if a social provider is enabled for the admin portal.",
	},
}

var socialProviderDataSourceSchema = schema.Schema{
	MarkdownDescription: fmt.Sprintf("The data source `%s` returns a list of social providers available in your Cidaas instance."+
		"\nYou can apply filters using the `filter` block in your Terraform configuration.", SOCIAL_PROVIDER_DATASOURCE),
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The data source's unique ID.",
			Computed:    true,
		},
	},
	Blocks: map[string]schema.Block{
		"filter": socialProviderFilter.Schema(),
		"social_provider": schema.ListNestedBlock{
			Description: "The returned list of social providers.",
			NestedObject: schema.NestedBlockObject{
				Attributes: socialProviderSchema,
			},
		},
	},
}

func NewSocialProvider() datasource.DataSource {
	return &SocialProviderDataSource{
		BaseDataSource: NewBaseDataSource(
			BaseDataSourceConfig{
				Name:   SOCIAL_PROVIDER_DATASOURCE,
				Schema: &socialProviderDataSourceSchema,
			},
		),
	}
}

func (d *SocialProviderDataSource) Read( //nolint:dupl
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data SocialProviderFilterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get config data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	data.ID = types.StringValue(uuid.New().String())
	result, diag := socialProviderFilter.GetAndFilter(ctx, d.Client, data.Filters, listSocialProviders)
	if diag != nil {
		tflog.Error(ctx, "failed to filter social_provider data", util.H{
			"error":  diag.Summary(),
			"detail": diag.Detail(),
		})
		resp.Diagnostics.Append(diag)
		return
	}
	tflog.Debug(ctx, "successfully filtered social_provider data")

	data.SocialProvider = parseModel(AnySliceToTyped[cidaas.SocialProviderModel](result), parseSocialProvider)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Info(ctx, "successfully read social_provider data source")
}

func listSocialProviders(ctx context.Context, client *cidaas.Client) ([]any, error) {
	sps, err := client.SocialProvider.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return TypedSliceToAny(sps), nil
}

func parseSocialProvider(sp cidaas.SocialProviderModel) SocialProvider {
	return SocialProvider{
		ID:                    types.StringValue(sp.ID),
		ClientID:              types.StringValue(sp.ClientID),
		ClientSecret:          types.StringValue(sp.ClientSecret),
		Name:                  types.StringValue(sp.Name),
		ProviderName:          types.StringValue(sp.ProviderName),
		Enabled:               types.BoolValue(sp.Enabled),
		EnabledForAdminPortal: types.BoolValue(sp.EnabledForAdminPortal),
		Scopes:                util.SetValueOrNull(sp.Scopes),
	}
}
