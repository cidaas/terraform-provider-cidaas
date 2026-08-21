package resources

import (
	"context"
	"fmt"

	"github.com/Cidaas/terraform-provider-cidaas/helpers/cidaas"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// nolint:revive
const (
	RESOURCE_APP                                 = "cidaas_app"                                 //nolint:revive
	RESOURCE_CONSENT_GROUP                       = "cidaas_consent_group"                       //nolint:revive
	RESOURCE_CONSENT_VERSION                     = "cidaas_consent_version"                     //nolint:revive
	RESOURCE_CONSENT                             = "cidaas_consent"                             //nolint:revive
	RESOURCE_CUSTOM_PROVIDER                     = "cidaas_custom_provider"                     //nolint:revive
	RESOURCE_GROUP_TYPE                          = "cidaas_group_type"                          //nolint:revive
	RESOURCE_HOSTED_PAGE                         = "cidaas_hosted_page"                         //nolint:revive
	RESOURCE_PASSWORD_POLICY                     = "cidaas_password_policy"                     //nolint:revive
	RESOURCE_SECURITY_SETTINGS                   = "cidaas_security_settings"                   //nolint:revive
	RESOURCE_REGISTRATION_FIELD                  = "cidaas_registration_field"                  //nolint:revive
	RESOURCE_ROLE                                = "cidaas_role"                                //nolint:revive
	RESOURCE_SCOPE_GROUP                         = "cidaas_scope_group"                         //nolint:revive
	RESOURCE_SCOPE                               = "cidaas_scope"                               //nolint:revive
	RESOURCE_SOCIAL_PROVIDER                     = "cidaas_social_provider"                     //nolint:revive
	RESOURCE_TEMPLATE_GROUP                      = "cidaas_template_group"                      //nolint:revive
	RESOURCE_NOTIFICATIONS_TEMPLATE_GROUP        = "cidaas_notifications_template_group"        //nolint:revive
	RESOURCE_NOTIFICATIONS_TEMPLATE_GROUP_LOCALE = "cidaas_notifications_template_group_locale" //nolint:revive
	RESOURCE_TEMPLATE                            = "cidaas_template"                            //nolint:revive
	RESOURCE_NOTIFICATION_TEMPLATE_TYPE          = "cidaas_notification_template_type"          //nolint:revive
	RESOURCE_NOTIFICATION_TEMPLATE               = "cidaas_notification_template"               //nolint:revive
	RESOURCE_NOTIFICATION_SERVICE_SETUP          = "cidaas_notification_service_setup"          //nolint:revive
	RESOURCE_NOTIFICATION_PROVIDER_CONFIG        = "cidaas_notification_provider_config"        //nolint:revive
	RESOURCE_USER_GROUP                          = "cidaas_user_groups"                         //nolint:revive
	RESOURCE_WEBHOOK                             = "cidaas_webhook"                             //nolint:revive
)

type BaseResourceConfig struct {
	Name   string
	Schema *schema.Schema
}

type BaseResource struct {
	Config       BaseResourceConfig
	cidaasClient *cidaas.Client
}

func NewBaseResource(cfg BaseResourceConfig) BaseResource {
	return BaseResource{
		Config: cfg,
	}
}

func (r *BaseResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured
	if req.ProviderData == nil {
		return
	}

	r.cidaasClient = GetResourceMeta(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
}

func GetResourceMeta(
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) *cidaas.Client {
	client, ok := req.ProviderData.(*cidaas.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected cidaas.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return nil
	}
	return client
}

func (r *BaseResource) Metadata(
	_ context.Context,
	_ resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = r.Config.Name
}

func (r *BaseResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	if r.Config.Schema == nil {
		resp.Diagnostics.AddError(
			"Missing Schema",
			"Base resource was not provided a schema. "+
				"Please provide a Schema config attribute or implement, the Schema(...) function.",
		)
		return
	}
	resp.Schema = *r.Config.Schema
}
