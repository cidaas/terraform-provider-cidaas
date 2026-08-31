package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/Cidaas/terraform-provider-cidaas/internal/client"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources/app"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources/hostedpages"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources/notification"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources/scope"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources/security"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources/registration_field"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources/role"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources/user_group"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources/usersetup"
	"github.com/Cidaas/terraform-provider-cidaas/internal/resources/verification"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &cidaasProvider{}

type cidaasProvider struct {
	version string
}

type providerModel struct {
	BaseURL       types.String `tfsdk:"base_url"`
	CidaasVersion types.String `tfsdk:"cidaas_version"`
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
		MarkdownDescription: "The cidaas provider manages cidaas **v3** and **v4 (Trustdesk)** resources. " +
			"Authenticate with `TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID` and `TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET`.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "cidaas instance base URL, e.g. `https://your-tenant.cidaas.eu`.",
			},
			"cidaas_version": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Target cidaas version (e.g. `3.x`, `v3.x`, `4.x`, `v4.x`). " +
					"If unset, uses `TERRAFORM_PROVIDER_CIDAAS_VERSION` or `CIDAAS_VERSION`, then live `/public-srv/version`, else `4.x`.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`(?i)^v?[34](\..*)?$`),
						"must specify major version 3 or 4 (e.g. '3.x', '4.x', 'v4.0.4-alpha.1')",
					),
				},
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

	// Resolve target cidaas version in order of precedence:
	// 1. Explicit HCL provider block attribute (cidaas_version)
	// 2. TERRAFORM_PROVIDER_CIDAAS_VERSION environment variable
	// 3. CIDAAS_VERSION environment variable
	targetVersion := ""
	if !cfg.CidaasVersion.IsNull() && !cfg.CidaasVersion.IsUnknown() && cfg.CidaasVersion.ValueString() != "" {
		targetVersion = cfg.CidaasVersion.ValueString()
	} else if envVer := os.Getenv("TERRAFORM_PROVIDER_CIDAAS_VERSION"); envVer != "" {
		targetVersion = envVer
	} else if envVer := os.Getenv("CIDAAS_VERSION"); envVer != "" {
		targetVersion = envVer
	}

	if targetVersion == "" {
		resp.Diagnostics.AddError(
			"Missing required cidaas_version",
			"cidaas_version is not configured. You must specify cidaas_version in the provider block (e.g. cidaas_version = \"3.x\" or \"4.x\") or set the TERRAFORM_PROVIDER_CIDAAS_VERSION or CIDAAS_VERSION environment variable.",
		)
		return
	}

	normVersion := client.NormalizeVersion(targetVersion)
	if normVersion != "3.x" && normVersion != "4.x" {
		resp.Diagnostics.AddError(
			"Invalid cidaas_version",
			fmt.Sprintf("Invalid cidaas_version %q. Supported versions are '3.x' (v3) or '4.x' (v4).", targetVersion),
		)
		return
	}

	c.Capabilities.TargetVersion = normVersion
	c.Capabilities.SupportsV4 = (normVersion == "4.x")

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *cidaasProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// v4 Trustdesk & Hosted Pages Resources
		hostedpages.NewThemeResource,
		hostedpages.NewTranslationsResource,
		hostedpages.NewHostedPageGroupResource,
		hostedpages.NewHostedPageLayoutResource,
		usersetup.NewUserSetupResource,
		app.NewAppConfigurationResource,
		app.NewLegacyAppResource,
		verification.NewSuggestVerificationMethodResource,
		verification.NewVerificationOptionsResource,

		// Shared v3 Priority Resources organized in domain subpackages
		registrationfield.NewRegFieldResource,
		role.NewRoleResource,
		usergroup.NewUserGroupResource,
		usergroup.NewGroupTypeResource,
		scope.NewScopeResource,
		scope.NewScopeGroupResource,
		security.NewPasswordPolicy,
		security.NewSecuritySettings,
		notification.NewTemplateResource,
		notification.NewNotificationServiceSetupResource,
	}
}

func (p *cidaasProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
