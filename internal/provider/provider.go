package provider

import (
	"context"
	"os"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources/hostedpages"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &cidaasProvider{}

type cidaasProvider struct {
	version string
}

type providerModel struct {
	BaseURL types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &cidaasProvider{version: version}
	}
}

func (p *cidaasProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cidaas"
	resp.Version = p.version
}

func (p *cidaasProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The cidaas provider manages cidaas **v4 (Trustdesk)** resources. " +
			"Authenticate with `TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID` and `TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET`.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "cidaas instance base URL, e.g. `https://your-tenant.cidaas.eu`.",
			},
		},
	}
}

func (p *cidaasProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientID := os.Getenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID")
	clientSecret := os.Getenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		resp.Diagnostics.AddError(
			"Missing credentials",
			"Set TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID and TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET.",
		)
		return
	}

	c, err := client.NewClient(ctx, client.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		BaseURL:      cfg.BaseURL.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cidaas client", err.Error())
		return
	}
	if !c.Capabilities.SupportsV4 {
		resp.Diagnostics.AddError(
			"Incompatible tenant version",
			"This provider requires cidaas v4 (Trustdesk). Detected version: "+c.Capabilities.Version,
		)
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *cidaasProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		hostedpages.NewThemeResource,
		hostedpages.NewTranslationsResource,
		hostedpages.NewHostedPageGroupResource,
	}
}

func (p *cidaasProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
