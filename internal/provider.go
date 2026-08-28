package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/Cidaas/terraform-provider-cidaas/helpers/util"
	cidaasDataSources "github.com/Cidaas/terraform-provider-cidaas/internal/datasources"
	cidaasResource "github.com/Cidaas/terraform-provider-cidaas/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type cidaasProvider struct {
	version string
}

type Model struct {
	BaseURL                  types.String `tfsdk:"base_url"`
	NotificationsContextPath types.String `tfsdk:"notifications_context_path"`
	ClientID                 types.String `tfsdk:"client_id"`
	ClientSecret             types.String `tfsdk:"client_secret"`
}

func Cidaas(version string) func() provider.Provider {
	return func() provider.Provider {
		return &cidaasProvider{
			version: version,
		}
	}
}

func (p *cidaasProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cidaas"
	resp.Version = "dev"
}

func (p *cidaasProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Required:    true,
				Description: "The base url of the Terraform client",
			},
			"notifications_context_path": schema.StringAttribute{
				Optional: true,
				Description: "URL path segment for notification-srv APIs (default: `notifications-srv`). " +
					"Used by notification-srv resources and datasources (`cidaas_notifications_template_group`, `cidaas_notification_template`, `cidaas_notification_service_setup`, `cidaas_notification_provider_config`, service setups, graph datasources). " +
					"Legacy `cidaas_template` / `cidaas_template_group` use `templates-srv` and ignore this setting.",
			},
			"client_id": schema.StringAttribute{
				Optional: true,
				Description: "The client ID of a non-interactive cidaas client used by Terraform to authenticate with cidaas. " +
					"Can also be set via the `TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID` environment variable. The provider configuration value takes precedence.",
			},
			"client_secret": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "The client secret of a non-interactive cidaas client used by Terraform to authenticate with cidaas. " +
					"Can also be set via the `TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET` environment variable. The provider configuration value takes precedence.",
			},
		},
	}
}

func (p *cidaasProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		cidaasDataSources.NewRole,
		cidaasDataSources.NewGroupType,
		cidaasDataSources.NewScope,
		cidaasDataSources.NewScopeGroup,
		cidaasDataSources.NewSystemTemplateOption,
		cidaasDataSources.NewConsent,
		cidaasDataSources.NewSocialProvider,
		cidaasDataSources.NewCustomProvider,
		cidaasDataSources.NewRegistrationField,
		cidaasDataSources.NewNotificationServiceSetups,
		cidaasDataSources.NewNotificationServiceSetup,
		cidaasDataSources.NewNotificationTemplates,
		cidaasDataSources.NewNotificationTemplateGroupsGraph,
		cidaasDataSources.NewWebhookEvents,
	}
}

func (p *cidaasProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		cidaasResource.NewRoleResource,
		cidaasResource.NewCustomProvider,
		cidaasResource.NewSocialProvider,
		cidaasResource.NewScopeResource,
		cidaasResource.NewScopeGroupResource,
		cidaasResource.NewConsentGroupResource,
		cidaasResource.NewGroupTypeResource,
		cidaasResource.NewUserGroupResource,
		cidaasResource.NewHostedPageResource,
		cidaasResource.NewWebhookResource,
		cidaasResource.NewAppResource,
		cidaasResource.NewRegFieldResource,
		cidaasResource.NewTemplateGroupResource,
		cidaasResource.NewNotificationsTemplateGroupResource,
		cidaasResource.NewNotificationsTemplateGroupLocaleResource,
		cidaasResource.NewTemplateResource,
		cidaasResource.NewNotificationTemplateTypeResource,
		cidaasResource.NewNotificationTemplateResource,
		cidaasResource.NewNotificationServiceSetupResource,
		cidaasResource.NewNotificationProviderConfigResource,
		cidaasResource.NewPasswordPolicy,
		cidaasResource.NewSecuritySettings,
		cidaasResource.NewConsentResource,
		cidaasResource.NewConsentVersionResource,
	}
}

func (p *cidaasProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Debug(ctx, "Starting provider configuration")

	var data Model
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to get provider config data", util.H{
			"errors": resp.Diagnostics.Errors(),
		})
		return
	}
	tflog.Debug(ctx, "Successfully retrieved provider configuration", util.H{
		"base_url": data.BaseURL.ValueString(),
	})

	clientID := data.ClientID.ValueString()
	if clientID == "" {
		clientID = os.Getenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID")
	}
	clientSecret := data.ClientSecret.ValueString()
	if clientSecret == "" {
		clientSecret = os.Getenv("TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET")
	}

	if clientID == "" || clientSecret == "" {
		tflog.Error(ctx, "Missing required client credentials", util.H{
			"client_id_set":     clientID != "",
			"client_secret_set": clientSecret != "",
		})
		resp.Diagnostics.AddError(
			"missing client credentials",
			"client_id or client_secret is missing. Set them in the provider configuration "+
				"(client_id / client_secret) or via the TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID / "+
				"TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET environment variables (deprecated). "+
				"Please check the document https://registry.terraform.io/providers/Cidaas/cidaas/latest/docs")
		return
	}
	tflog.Debug(ctx, "Successfully retrieved client credentials", util.H{
		"client_id_length": len(clientID),
		"base_url":         data.BaseURL.ValueString(),
	})

	clientConfig := cidaas.ClientConfig{
		ClientID:                 clientID,
		ClientSecret:             clientSecret,
		BaseURL:                  data.BaseURL.ValueString(),
		NotificationsContextPath: data.NotificationsContextPath.ValueString(),
	}

	tflog.Info(ctx, "Creating cidaas client", util.H{
		"base_url": data.BaseURL.ValueString(),
	})
	client, err := cidaas.NewClient(ctx, clientConfig)
	if err != nil {
		tflog.Error(ctx, "Failed to create cidaas client", util.H{
			"base_url": data.BaseURL.ValueString(),
			"error":    err.Error(),
		})
		resp.Diagnostics.AddError("provide configuration failed", fmt.Sprintf("failed to create cidaas client %s", err.Error()))
		return
	}

	resp.ResourceData = client
	resp.DataSourceData = client
	tflog.Info(ctx, "Provider configured successfully", util.H{
		"base_url": data.BaseURL.ValueString(),
	})
}
