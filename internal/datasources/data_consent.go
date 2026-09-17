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

type ConsentDataSource struct {
	BaseDataSource
}

type ConsentFilterModel struct {
	BaseModel
	Consent []Consent `tfsdk:"consent"`
}

type Consent struct {
	ID          types.String `tfsdk:"id"`
	ConsentName types.String `tfsdk:"consent_name"`
}

var consentFilter = FilterConfig{
	"consent_name": {TypeFunc: FilterTypeString},
}

func NewConsent() datasource.DataSource {
	return &ConsentDataSource{
		BaseDataSource: NewBaseDataSource(
			BaseDataSourceConfig{
				Name:   CONSENT_DATASOURCE,
				Schema: &consentSchema,
			},
		),
	}
}

var consentSchema = schema.Schema{
	MarkdownDescription: fmt.Sprintf("The data source `%s` returns a list of consents available in your Cidaas instance."+
		"\nYou can apply filters using the `filter` block in your Terraform configuration.", CONSENT_DATASOURCE),
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The data source's unique ID.",
			Computed:    true,
		},
	},
	Blocks: map[string]schema.Block{
		"filter": consentFilter.Schema(),
		"consent": schema.ListNestedBlock{
			Description: "The returned list of consents.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The unique identifier of the consent.",
					},
					"consent_name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The name of the consent.",
					},
				},
			},
		},
	},
}

func (d *ConsentDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data ConsentFilterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to get config data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}

	data.ID = types.StringValue(uuid.New().String())
	result, diag := consentFilter.GetAndFilter(ctx, d.Client, data.Filters, listConsents)
	if diag != nil {
		tflog.Error(ctx, "failed to filter consent data", util.H{
			"error":  diag.Summary(),
			"detail": diag.Detail(),
		})
		resp.Diagnostics.Append(diag)
		return
	}
	tflog.Debug(ctx, "successfully filtered consent data")

	data.Consent = parseModel[cidaas.ConsentModel, Consent](AnySliceToTyped[cidaas.ConsentModel](result), parseConsent)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "failed to set state", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Info(ctx, "successfully read consent data source")
}

func listConsents(ctx context.Context, client *cidaas.Client) ([]any, error) {
	consents, err := client.Consent.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return TypedSliceToAny(consents), nil
}

func parseConsent(c cidaas.ConsentModel) Consent {
	return Consent{
		ID:          types.StringValue(c.ID),
		ConsentName: types.StringValue(c.ConsentName),
	}
}
