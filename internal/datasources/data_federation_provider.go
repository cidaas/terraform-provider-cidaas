package datasources

import (
	"context"
	"fmt"
	"strings"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	identityprovider "github.com/Cidaas/terraform-provider-cidaas/internal/resources/identity_provider"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure        = (*FederationProviderDataSource)(nil)
	_ datasource.DataSourceWithConfigValidators = (*FederationProviderDataSource)(nil)
)

// federationProviderListPageSize is the maximum page size accepted by GET /federation/providers.
const federationProviderListPageSize = 100

// FederationProviderDataSource is the cidaas_federation_provider data source.
type FederationProviderDataSource struct {
	BaseDataSource
}

// federationProviderLookup holds the lookup keys (exactly one is set).
type federationProviderLookup struct {
	ID           types.String `tfsdk:"id"`
	ProviderName types.String `tfsdk:"provider_name"`
}

var federationProviderDataSourceSchema = buildFederationProviderDataSourceSchema()

// buildFederationProviderDataSourceSchema derives the data source schema from the resource schema; id and provider_name become lookup keys.
func buildFederationProviderDataSourceSchema() schema.Schema {
	resourceSchema := identityprovider.FederationProviderResourceSchema()
	attrs := make(map[string]schema.Attribute, len(resourceSchema.Attributes))
	for name, a := range resourceSchema.Attributes {
		if ds, ok := toDataSourceAttribute(name, a); ok {
			attrs[name] = ds
		}
	}
	attrs["id"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "ID of the federation provider. Exactly one of `id` or `provider_name` must be set.",
	}
	attrs["provider_name"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Unique provider name. Exactly one of `id` or `provider_name` must be set; the name must match exactly one provider.",
	}
	return schema.Schema{
		MarkdownDescription: fmt.Sprintf("The data source `%s` looks up a single federated identity provider on cidaas v4 (Trustdesk) "+
			"by `id` or `provider_name` and exposes the same attributes as the `cidaas_federation_provider` resource. "+
			"Secrets are marked sensitive and are null when the API does not return them.", FEDERATION_PROVIDER_DATASOURCE),
		Attributes: attrs,
	}
}

// toDataSourceAttribute converts a resource attribute into a computed data source attribute.
// Write-only attributes (and their version markers) have no meaning for a data source and are dropped.
func toDataSourceAttribute(name string, a rschema.Attribute) (schema.Attribute, bool) {
	if a.IsWriteOnly() || strings.HasSuffix(name, "_wo_version") {
		return nil, false
	}
	desc := a.GetMarkdownDescription()
	sensitive := a.IsSensitive()
	switch t := a.(type) {
	case rschema.StringAttribute:
		return schema.StringAttribute{Computed: true, Sensitive: sensitive, MarkdownDescription: desc}, true
	case rschema.BoolAttribute:
		return schema.BoolAttribute{Computed: true, Sensitive: sensitive, MarkdownDescription: desc}, true
	case rschema.Int64Attribute:
		return schema.Int64Attribute{Computed: true, Sensitive: sensitive, MarkdownDescription: desc}, true
	case rschema.ListAttribute:
		return schema.ListAttribute{Computed: true, Sensitive: sensitive, ElementType: t.ElementType, MarkdownDescription: desc}, true
	case rschema.MapAttribute:
		return schema.MapAttribute{Computed: true, Sensitive: sensitive, ElementType: t.ElementType, MarkdownDescription: desc}, true
	case rschema.SingleNestedAttribute:
		return schema.SingleNestedAttribute{Computed: true, Sensitive: sensitive, Attributes: toDataSourceAttributes(t.Attributes), MarkdownDescription: desc}, true
	case rschema.ListNestedAttribute:
		return schema.ListNestedAttribute{
			Computed: true, Sensitive: sensitive, MarkdownDescription: desc,
			NestedObject: schema.NestedAttributeObject{Attributes: toDataSourceAttributes(t.NestedObject.Attributes)},
		}, true
	case rschema.MapNestedAttribute:
		return schema.MapNestedAttribute{
			Computed: true, Sensitive: sensitive, MarkdownDescription: desc,
			NestedObject: schema.NestedAttributeObject{Attributes: toDataSourceAttributes(t.NestedObject.Attributes)},
		}, true
	}
	panic(fmt.Sprintf("federation provider data source: unsupported attribute type %T for %q", a, name))
}

// toDataSourceAttributes converts nested resource attributes with toDataSourceAttribute.
func toDataSourceAttributes(in map[string]rschema.Attribute) map[string]schema.Attribute {
	out := make(map[string]schema.Attribute, len(in))
	for name, a := range in {
		if ds, ok := toDataSourceAttribute(name, a); ok {
			out[name] = ds
		}
	}
	return out
}

// NewFederationProvider returns the cidaas_federation_provider data source.
func NewFederationProvider() datasource.DataSource {
	return &FederationProviderDataSource{
		BaseDataSource: NewBaseDataSource(
			BaseDataSourceConfig{
				Name:   FEDERATION_PROVIDER_DATASOURCE,
				Schema: &federationProviderDataSourceSchema,
			},
		),
	}
}

// Configure blocks the data source unless the provider targets cidaas 4.x.
func (d *FederationProviderDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	if c, ok := req.ProviderData.(*client.Client); ok && !c.RequireV4(FEDERATION_PROVIDER_DATASOURCE, &resp.Diagnostics) {
		return
	}
	d.BaseDataSource.Configure(ctx, req, resp)
}

// ConfigValidators requires exactly one of id or provider_name.
func (d *FederationProviderDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("provider_name"),
		),
	}
}

// Read looks up the provider and stores it in state.
func (d *FederationProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var lookup federationProviderLookup
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &lookup.ID)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("provider_name"), &lookup.ProviderName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider, err := d.lookup(ctx, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read federation provider", err.Error())
		return
	}

	value, diags := identityprovider.FederationProviderDataSourceValue(ctx, provider, federationProviderDataSourceSchema.Type())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, value)...)
}

// lookup fetches the provider by id, or by provider_name requiring exactly one match.
func (d *FederationProviderDataSource) lookup(ctx context.Context, lookup federationProviderLookup) (*cidaas.FederationProviderModel, error) {
	if !lookup.ID.IsNull() {
		res, err := d.Client.FederationProvider.Get(ctx, lookup.ID.ValueString())
		if err != nil {
			return nil, fmt.Errorf("federation provider with id %q: %s", lookup.ID.ValueString(), util.FormatErrorMessage(err))
		}
		return &res.Data, nil
	}

	name := lookup.ProviderName.ValueString()
	providers, err := d.Client.FederationProvider.List(ctx, cidaas.FederationProviderListParams{
		ProviderName: name,
		Size:         federationProviderListPageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("listing federation providers with provider_name %q: %s", name, util.FormatErrorMessage(err))
	}
	matches := make([]cidaas.FederationProviderModel, 0, 1)
	for _, p := range providers {
		if p.ProviderName == name {
			matches = append(matches, p)
		}
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no federation provider found with provider_name %q", name)
	case 1:
		return &matches[0], nil
	default:
		ids := make([]string, 0, len(matches))
		for _, m := range matches {
			ids = append(ids, m.ID)
		}
		return nil, fmt.Errorf("%d federation providers found with provider_name %q (ids: %s); use id instead", len(matches), name, strings.Join(ids, ", "))
	}
}
