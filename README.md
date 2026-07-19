![Logo](https://raw.githubusercontent.com/cidaas/terraform-provider-cidaas/master/logo.jpg)

## About cidaas

[cidaas](https://www.cidaas.com)
 is a fast and secure Cloud Identity & Access Management solution that standardises what’s important and simplifies what’s complex.

## Feature set includes

* Single Sign On (SSO) based on OAuth 2.0, OpenID Connect, SAML 2.0
* Multi-Factor-Authentication with more than 14 authentication methods, including TOTP and FIDO2
* Passwordless Authentication
* Social Login (e.g. Facebook, Google, LinkedIn and more) as well as Enterprise Identity Provider (e.g. SAML or AD)
* Security in Machine-to-Machine (M2M) and IoT

<a href="https://terraform.io">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset=".github/terraform_logo_dark.svg">
    <source media="(prefers-color-scheme: light)" srcset=".github/terraform_logo_light.svg">
    <img src=".github/terraform_logo_light.svg" alt="Terraform logo" title="Terraform" align="right" height="50">
  </picture>
</a>

# Terraform Provider for cidaas

The Terraform provider for cidaas enables interaction with cidaas instances that allows to perform CRUD operations on applications, custom providers, registration fields and many other functionalities. From managing applications to configuring custom providers, the Terraform provider enhances the user's capacity to define, provision and manipulate their cidaas resources.

## Prerequisites

* Ensure Terraform is installed on your local machine. Find installation instructions for different operating systems [here](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli).
* [Go](https://go.dev/doc/install) (1.21)

## Documentation

Official documentation on how to use this provider can be found on the
[Terraform Registry](https://registry.terraform.io/providers/cidaas/cidaas/latest/docs). Detailed explanations of the resources can also be found in the [Supported Resources](#supported-resources) section.

## Example Usage

Below is a step-by-step guide to help you set up the provider, configure essential environment variables and integrate the provider into your configuration:

### 1. Terraform Provider Declaration

Begin by specifying the cidaas provider in your `terraform` block in your Terraform configuration file:

```hcl
terraform {
    required_providers {
      cidaas = {
        version = "3.5.7"
        source  = "Cidaas/cidaas"
      }
    }
}
```

Terraform pulls the version configured of the cidaas provider for your infrastructure.

### 2. Setup Environment Variables

To authenticate and authorize Terraform operations with cidaas, set the necessary environment variables. These variables include your cidaas client credentials, allowing the Terraform provider to complete the client credentials flow and generate an access_token. Execute the following commands in your terminal, replacing placeholders with your actual cidaas client ID and client secret.

### For Linux and MacOS

```bash
export TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID="ENTER CIDAAS CLIENT ID"
export TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET="ENTER CIDAAS CLIENT SECRET"
```

### For Windows

```bash
Set-Item -Path env:TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID -Value “ENTER CIDAAS CLIENT ID“
Set-Item -Path env:TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET -Value “ENTER CIDAAS CLIENT SECRET“
```

You can get a set of client credentials from the cidaas Admin UI by creating a new client. Simply go to the `Apps` > `App Settings` > `Create New App`. It's important to note that when creating the client, you must select the app type as **Non-Interactive**.

### 3. Add cidaas Provider Configuration

Next, add the cidaas provider configuration to your Terraform configuration file. Specify the `base_url` parameter to point to your cidaas instance. For reference, check the example folder.

```hcl
provider "cidaas" {
  base_url = "https://cidaas.de"
}
```

**Note:** Starting from version 2.5.1, the `redirect_url` is no longer supported in the provider configuration. Ensure that you adjust your configuration accordingly.

By following these steps, you integrate the cidaas Terraform provider enabling you to manage your cidaas resources with Terraform.

### Notification templates (notification-srv)

Configure the provider with `base_url` and optional `notifications_context_path` (default `notifications-srv`). **New** template workflows should use **notification-srv** REST under `/{ctx}/…`, not legacy **templates-srv**.

| Scenario | Terraform building blocks |
| --- | --- |
| Update templates in the default template group | `data.cidaas_notification_templates` (POST `graph/templates/` with a graph filter) to discover rows; `cidaas_notification_template` for each `PUT /templates/:id` |
| Create a new template group and configure templates | `cidaas_notifications_template_group` (`POST/PUT templategroups`); `data.cidaas_notification_service_setups` to resolve `serviceSetupId` per channel; `cidaas_notification_template` per row (`POST /templates/`) |
| Add another locale and translated content | Copy options on `cidaas_notifications_template_group` and/or additional `cidaas_notification_template` instances per locale |
| Point the group at another comm / service setup | `data.cidaas_notification_service_setups` + `cidaas_notifications_template_group` `comm_setting_*` |
| Provision a communication provider (service setup + credentials) | `cidaas_notification_service_setup` + `cidaas_notification_provider_config` (`config_data_wo` + `schemaData`); verify in service-desk, then refresh `status` |
| Developer template type and template body | `cidaas_notification_template_type` (`POST /templatetypes/`) + `cidaas_notification_template` with `group_id = "developer"` |

**Legacy:** `cidaas_template` and `cidaas_template_group` call **templates-srv**; they remain for existing configurations but are deprecated for greenfield IaC in favour of the resources above.

For limitations, authentication, and step-by-step workflows, see the [Notification service (notification-srv) guide](docs/guides/notification_srv.md).

## Supported Resources

The Terraform provider for cidaas supports a variety of resources that enables you to manage and configure different aspects of your cidaas environment. These resources are designed to integrate with Terraform workflows, allowing you to define, provision and manage your cidaas resources as code.

Explore the following resources to understand their attributes, functionalities and how to use them in your Terraform configurations:

* [cidaas_app](#cidaas_app-resource)
* [cidaas_consent](#cidaas_consent-resource)
* [cidaas_consent_group](#cidaas_consent_group-resource)
* [cidaas_consent_version](#cidaas_consent_version-resource)
* [cidaas_custom_provider](#cidaas_custom_provider-resource)
* [cidaas_group_type](#cidaas_group_type-resource-previously-cidaas_user_group_category)
* [cidaas_hosted_page](#cidaas_hosted_page-resource)
* [cidaas_password_policy](#cidaas_password_policy-resource)
* [cidaas_registration_field](#cidaas_registration_field-resource)
* [cidaas_role](#cidaas_role-resource)
* [cidaas_scope_group](#cidaas_scope_group-resource)
* [cidaas_scope](#cidaas_scope-resource)
* [cidaas_social_provider](#cidaas_social_provider-resource)
* [cidaas_template_group](#cidaas_template_group-resource) (legacy templates-srv)
* [cidaas_template](#cidaas_template-resource) (legacy templates-srv)
* [cidaas_notifications_template_group](#cidaas_notifications_template_group-resource) (notification-srv)
* [cidaas_notification_template](#cidaas_notification_template-resource) (notification-srv)
* [cidaas_notification_template_type](#cidaas_notification_template_type-resource) (notification-srv)
* [cidaas_notification_service_setup](#cidaas_notification_service_setup-resource) (notification-srv)
* [cidaas_notification_provider_config](#cidaas_notification_provider_config-resource) (notification-srv)
* [cidaas_user_groups](#cidaas_user_groups-resource)
* [cidaas_webhook](#cidaas_webhook-resource)

## Datasources

The provider also provides a list of datasources to fetch your required data that can be referenced in your terraform configuration.

Here is the list of the datasources the provider supports:

* [cidaas_consent](#cidaas_consent-data-source)
* [cidaas_custom_provider](#cidaas_custom_provider-data-source)
* [cidaas_group_type](#cidaas_group_type-data-source)
* [cidaas_registration_field](#cidaas_registration_field-data-source)
* [cidaas_role](#cidaas_role-data-source)
* [cidaas_scope_group](#cidaas_scope_group-data-source)
* [cidaas_scope](#cidaas_scope-data-source)
* [cidaas_social_provider](#cidaas_social_provider-data-source)
* [cidaas_system_template_option](#cidaas_system_template_option-data-source)
* [cidaas_notification_service_setup](#cidaas_notification_service_setup-data-source)
* [cidaas_notification_service_setups](#cidaas_notification_service_setups-data-source)
* [cidaas_notification_templates](#cidaas_notification_templates-data-source)
* [cidaas_notification_template_groups](#cidaas_notification_template_groups-data-source)
* [cidaas_webhook_events](docs/data-sources/webhook_events.md)

# cidaas_app (Resource)

The App resource allows creation and management of clients in cidaas system. When creating a client with a custom `client_id` and `client_secret` you can include the configuration in the resource. If not provided, cidaas will generate a set for you. `client_secret` is sensitive data. Refer to the article [Terraform Sensitive Variables](https://developer.hashicorp.com/terraform/tutorials/configuration-language/sensitive-variables) to properly handle sensitive information.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:apps_read
* cidaas:apps_write
* cidaas:apps_delete

From version 3.3.0, the attribute `common_configs` is not supported anymore. Instead, we encourage you to use the custom module **terraform-cidaas-app**.
The module provides a variable with the same name `common_configs` which
supports all the attributes in the resource app except `client_name`. With this module you can avoid the repeated configuration and assign the common properties
of multiple apps to a common variable and inherit the properties.

Link to the custom module <https://github.com/cidaas/terraform-cidaas-app>

##### Module usage

```hcl
// local.tfvars
common_configs = {
  client_type     = "SINGLE_PAGE"
  company_address = "Wimsheim"
  company_name    = "WidasConcepts GmbH"
  company_address = "Maybachstraße 2, 71299 Wimsheim, Germany"
  company_website = "https://widas.com"
  redirect_uris = [
    "https://cidaas.de/callback",
  ]
  allowed_logout_urls = [
    "https://cidaas.de/logout"
  ]
  allowed_scopes = [
    "openid",
  ]
}

// main.tf
provider "cidaas" {
  base_url = "https://cidaas.de"
}

module "app1" {
  source = "git@github.com:cidaas/terraform-cidaas-app.git"

  providers = {
    cidaas = cidaas
  }
  client_name    = "Demo App"
  common_configs = var.common_configs
}

module "app2" {
  source = "git@github.com:cidaas/terraform-cidaas-app.git"
  providers = {
    cidaas = cidaas
  }
  client_name    = "Demo IOS App"
  client_type    = "IOS"
  common_configs = var.common_configs
}
```

You can explore more on the module in the github repo.

## Example Usage

```terraform
resource "cidaas_app" "sample" {
  client_name                     = "Test Terraform Application" // unique
  client_type                     = "SINGLE_PAGE"
  accent_color                    = "#ef4923"
  primary_color                   = "#ef4923"
  media_type                      = "IMAGE"
  allow_login_with                = ["EMAIL", "MOBILE", "USER_NAME"]
  redirect_uris                   = ["https://cidaas.com"]
  allowed_logout_urls             = ["https://cidaas.com"]
  enable_deduplication            = true
  auto_login_after_register       = true
  enable_passwordless_auth        = false
  register_with_login_information = false
  hosted_page_group               = "default"
  company_name                    = "Widas ID GmbH"
  company_address                 = "01"
  company_website                 = "https://cidaas.com"
  allowed_scopes                  = ["openid", "cidaas:register", "profile"]
  client_display_name             = "Display Name of the app" // unique
  content_align                   = "CENTER"
  post_logout_redirect_uris       = ["https://cidaas.com"]
  logo_align                      = "CENTER"
  allow_disposable_email          = false
  validate_phone_number           = false
  additional_access_token_payload = ["sample_payload"]
  required_fields                 = ["email"]
  mobile_settings = {
    team_id      = "sample-team-id"
    bundle_id    = "sample-bundle-id"
    package_name = "sample-package-name"
    key_hash     = "sample-key-hash"
  }
  // for custom client credentials use client_id and client_secret, you can leave blank if you want cidaas to create a set for you
  # client_id                       = ""
  # client_secret                   = ""
  policy_uri                          = "https://cidaas.com"
  tos_uri                             = "https://cidaas.com"
  imprint_uri                         = "https://cidaas.com"
  contacts                            = ["support@cidas.de"]
  token_endpoint_auth_method          = "client_secret_post"
  token_endpoint_auth_signing_alg     = "RS256"
  default_acr_values                  = ["default"]
  web_message_uris                    = ["https://cidaas.com"]
  allowed_fields                      = ["email"]
  smart_mfa                           = false // Default: false
  captcha_ref                         = "sample-captcha-ref"
  captcha_refs                        = ["sample"]
  consent_refs                        = ["sample"]
  communication_medium_verification   = "email_verification_required_on_usage"
  enable_bot_detection                = false
  allow_guest_login_groups = [{
    group_id      = "developer101"
    roles         = ["developer", "qa", "admin"]
    default_roles = ["developer"]
  }]
  is_login_success_page_enabled    = false
  is_register_success_page_enabled = false
  group_ids                        = ["sample"]
  group_selection = {
    selectable_groups      = ["developer-users"]
    selectable_group_types = ["sample"]
  }
  group_types                     = ["sample"]
  logo_uri                        = "https://cidaas.com"
  initiate_login_uri              = "https://cidaas.com"
  registration_client_uri         = "https://cidaas.com"
  registration_access_token       = "registration access token"
  client_uri                      = "https://cidaas.com"
  jwks_uri                        = "https://cidaas.com"
  jwks                            = "https://cidaas.com/jwks"
  sector_identifier_uri           = "https://cidaas.com"
  subject_type                    = "sample subject type"
  id_token_signed_response_alg    = "RS256"
  id_token_encrypted_response_alg = "RS256"
  id_token_encrypted_response_enc = "example"
  userinfo_signed_response_alg    = "RS256"
  userinfo_encrypted_response_alg = "RS256"
  userinfo_encrypted_response_enc = "example"
  request_object_signing_alg      = "RS256"
  request_object_encryption_alg   = "RS256"
  request_object_encryption_enc   = "userinfo_encrypted_response_enc"
  request_uris                    = ["sample"]
  description                     = "app description"
  consent_page_group              = "sample-consent-page-group"
  password_policy_ref             = "password-policy-ref"
  blocking_mechanism_ref          = "blocking-mechanism-ref"
  sub                             = "sample-sub"
  role                            = "sample-role"
  mfa_configuration               = "sample-configuration"
  suggest_mfa                     = ["OFF"]
  login_spi = {
    oauth_client_id = "bcb-4a6b-9777-8a64abe6af"
    spi_url         = "https://cidaas.com/spi-url"
  }
  background_uri  = "https://cidaas.com"
  video_url       = "https://cidaas.com"
  bot_captcha_ref = "sample-bot-captcha-ref"
  application_meta_data = {
    status : "active"
    version : "1.0.0"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `allowed_scopes` (Set of String) The URL of the company website. allowed_scopes is a required attribute. It must be provided in the main config or common_config
- `client_name` (String) Name of the client.
- `client_type` (String) The type of the client. The allowed values are SINGLE_PAGE, REGULAR_WEB, NON_INTERACTIVEIOS, ANDROID, WINDOWS_MOBILE, DESKTOP, MOBILE, DEVICE and THIRD_PARTY
- `company_address` (String) The company address.
- `company_name` (String) The name of the company that the client belongs to.
- `company_website` (String) The URL of the company website.

### Optional

- `accent_color` (String) The accent color of the client. e.g., `#f7941d`. The value must be a valid hex colorThe default is set to `#ef4923`.
- `accept_roles_in_the_registration` (Boolean) A boolean flag that determines whether roles can be accepted during the registration process.
- `ad_providers` (Attributes List) A list of Active Directory identity providers that users can authenticate with. (see [below for nested schema](#nestedatt--ad_providers))
- `additional_access_token_payload` (Set of String) Access token payload definition.
- `allow_disposable_email` (Boolean) Allow disposable email addresses. Default is set to `false` while creating an app.
- `allow_guest_login` (Boolean) Flag to specify whether guest users are allowed to access functionalities of the client. Default is set to `false`
- `allow_guest_login_groups` (Attributes List) (see [below for nested schema](#nestedatt--allow_guest_login_groups))
- `allow_login_with` (Set of String) allow_login_with is used to specify the preferred methods of login allowed for a client.
- `allowed_fields` (Set of String)
- `allowed_groups` (Attributes List) (see [below for nested schema](#nestedatt--allowed_groups))
- `allowed_logout_urls` (Set of String) Allowed logout URLs for OAuth2 client.
- `allowed_mfa` (Set of String)
- `allowed_native_clients` (Set of String) Client IDs of native apps that may use session transfer (STT) with this client. Typically the browser or web client lists the mobile/native client's `client_id` so STTs issued for that native client are accepted.
- `allowed_origins` (Set of String) List of the origins allowed to access the client.
- `allowed_roles` (Set of String)
- `allowed_web_origins` (Set of String) List of the web origins allowed to access the client.
- `application_meta_data` (Map of String) A map to add metadata of a client.
- `auto_login_after_register` (Boolean) Automatically login after registration. Default is set to `false` while creating an app.
- `backchannel_logout_session_required` (Boolean) If enabled, client applications or RPs must support session management through backchannel logout.
- `backchannel_logout_uri` (String)
- `background_uri` (String) The URL to the background image of the client.
- `blocking_mechanism_ref` (String)
- `bot_captcha_ref` (String)
- `bot_provider` (String)
- `captcha_ref` (String)
- `captcha_refs` (Set of String)
- `client_display_name` (String) The display name of the client.
- `client_id` (String) The client_id is the unqique identifier of the app. It's an optional attribute. If not provided, cidaas will gererate one for you and the state will be updated with the same
- `client_secret` (String, Sensitive) The client_id is the unqique identifier of the app. It's an optional attribute. If not provided, cidaas will gererate one for you and the state will be updated with the same
- `client_uri` (String)
- `communication_medium_verification` (String)
- `consent_page_group` (String)
- `consent_refs` (Set of String)
- `contacts` (Set of String) The contacts of the client.
- `content_align` (String) The alignment of the content of the client. e.g., `CENTER`. Allowed values are CENTER, LEFT and RIGHTThe default is set to `CENTER`.
- `custom_providers` (Attributes List) A list of custom identity providers that users can authenticate with. A custom provider can be created with the help of the resource cidaas_custom_provider. (see [below for nested schema](#nestedatt--custom_providers))
- `default_acr_values` (Set of String)
- `default_max_age` (Number) The default maximum age for the token in seconds. Default is 86400 seconds (24 hours).
- `default_roles` (Set of String)
- `default_scopes` (Set of String)
- `description` (String)
- `enable_bot_detection` (Boolean)
- `enable_deduplication` (Boolean) Enable deduplication.
- `enable_login_spi` (Boolean) If enabled, the login service verifies whether login spi responsded with success only then it issues a token.
- `enable_passwordless_auth` (Boolean) Enable passwordless authentication. Default is set to `true` while creating an app.
- `enabled` (Boolean)
- `grant_types` (Set of String) The grant types of the client. The default value is set to `['implicit','authorization_code', 'password', 'refresh_token']`. Include `session_transfer` on the native client when using the session transfer token (STT) flow.
- `group_ids` (Set of String)
- `group_role_restriction` (Attributes) (see [below for nested schema](#nestedatt--group_role_restriction))
- `group_selection` (Attributes) (see [below for nested schema](#nestedatt--group_selection))
- `group_types` (Set of String)
- `hosted_page_group` (String) Hosted page group.
- `id_token_encrypted_response_alg` (String)
- `id_token_encrypted_response_enc` (String)
- `id_token_lifetime_in_seconds` (Number) The lifetime of the id_token in seconds. Default is 86400 seconds (24 hours).
- `id_token_signed_response_alg` (String)
- `imprint_uri` (String) The URL to the imprint page.
- `initiate_login_uri` (String)
- `is_group_login_selection_enabled` (Boolean) Flag to enable or disble group login selection
- `is_hybrid_app` (Boolean) Flag to set if your app is hybrid or not. Default is set to `false`. Set to `true` to make your app hybrid.
- `is_login_success_page_enabled` (Boolean)
- `is_register_success_page_enabled` (Boolean)
- `is_remember_me_selected` (Boolean)
- `jwe_enabled` (Boolean) Flag to specify whether JSON Web Encryption (JWE) should be enabled for encrypting data.
- `jwks` (String)
- `jwks_uri` (String)
- `login_providers` (Set of String) With this attribute one can setup login provider to the client.
- `login_spi` (Attributes) A map defining the Login SPI configuration. (see [below for nested schema](#nestedatt--login_spi))
- `logo_align` (String)
- `logo_uri` (String)
- `media_type` (String) The media type of the client. e.g., `IMAGE`. Allowed values are VIDEO and IMAGEThe default is set to `IMAGE`.
- `mfa` (Attributes) Configuration settings for Multi-Factor Authentication (MFA). (see [below for nested schema](#nestedatt--mfa))
- `mfa_configuration` (String)
- `mobile_settings` (Attributes) (see [below for nested schema](#nestedatt--mobile_settings))
- `oauth_standard` (String) Specifies the OAuth standard version to use. Allowed values: 'OAuth2.1', 'OAuth2.0'.
- `operations_allowed_groups` (Attributes List) (see [below for nested schema](#nestedatt--operations_allowed_groups))
- `password_policy_ref` (String)
- `pending_scopes` (Set of String)
- `policy_uri` (String) The URL to the policy of a client.
- `post_logout_redirect_uris` (Set of String)
- `primary_color` (String) The primary color of the client. e.g., `#ef4923`. The value must be a valid hex colorThe default is set to `#f7941d`.
- `redirect_uris` (Set of String) Redirect URIs for OAuth2 client.
- `refresh_token_lifetime_in_seconds` (Number) The lifetime of the refresh token in seconds. Default is 15780000 seconds.
- `register_with_login_information` (Boolean) Register with login information. Default is set to `false` while creating an app.
- `registration_access_token` (String)
* `registration_client_uri` (String)
* `request_object_encryption_alg` (String)
* `request_object_encryption_enc` (String)
* `request_object_signing_alg` (String)
* `request_uris` (Set of String)
* `require_auth_time` (Boolean) Boolean flag to specify whether the auth_time claim is REQUIRED in a id token.
* `required_fields` (Set of String) The required fields while registering to the client.
* `response_types` (Set of String) The response types of the client. The default value is set to `['code','token', 'id_token']`
* `role` (String)
* `saml_providers` (Attributes List) A list of SAML identity providers that users can authenticate with. (see [below for nested schema](#nestedatt--saml_providers))
* `sector_identifier_uri` (String)
* `smart_mfa` (Boolean)
* `social_providers` (Attributes List) A list of social identity providers that users can authenticate with. Examples: Google, Facebook etc... (see [below for nested schema](#nestedatt--social_providers))
* `sub` (String)
* `subject_type` (String)
* `suggest_mfa` (Set of String)
* `suggest_verification_methods` (Attributes) Configuration for verification methods. (see [below for nested schema](#nestedatt--suggest_verification_methods))
* `template_group_id` (String) The id of the template group to be configured for commenication. Default is set to the system default group.
* `token_endpoint_auth_method` (String)
* `token_endpoint_auth_signing_alg` (String)
* `token_lifetime_in_seconds` (Number) The lifetime of the token in seconds. Default is 86400 seconds (24 hours).
* `tos_uri` (String) The URL to the TOS of a client.
* `user_consent` (Boolean) Specifies whether user consent is required or not. Default is `false`
* `userinfo_encrypted_response_alg` (String)
* `userinfo_encrypted_response_enc` (String)
* `userinfo_signed_response_alg` (String)
* `validate_phone_number` (Boolean) if enabled, phone number is validaed. Default is set to `false` while creating an app.
* `video_url` (String) The URL to the video of the client.
* `web_message_uris` (Set of String) A list of URLs for web messages used.
* `webfinger` (String)

### Read-Only

* `id` (String) The ID of the resource.

<a id="nestedatt--ad_providers"></a>

### Nested Schema for `ad_providers`

Optional:

* `display_name` (String)
* `domains` (Set of String)
* `is_provider_visible` (Boolean)
* `logo_url` (String)
* `provider_name` (String)
* `type` (String)

<a id="nestedatt--allow_guest_login_groups"></a>

### Nested Schema for `allow_guest_login_groups`

Optional:

* `default_roles` (Set of String)
* `group_id` (String)
* `roles` (Set of String)

<a id="nestedatt--allowed_groups"></a>

### Nested Schema for `allowed_groups`

Optional:

* `default_roles` (Set of String)
* `group_id` (String)
* `roles` (Set of String)

<a id="nestedatt--custom_providers"></a>

### Nested Schema for `custom_providers`

Optional:

* `display_name` (String)
* `domains` (Set of String)
* `is_provider_visible` (Boolean)
* `logo_url` (String)
* `provider_name` (String)
* `type` (String)

<a id="nestedatt--group_role_restriction"></a>

### Nested Schema for `group_role_restriction`

Required:

* `filters` (Attributes List) An array of group role filters. (see [below for nested schema](#nestedatt--group_role_restriction--filters))
* `match_condition` (String) The match condition for the role restriction

<a id="nestedatt--group_role_restriction--filters"></a>

### Nested Schema for `group_role_restriction.filters`

Optional:

* `group_id` (String) The unique ID of the user group.
* `role_filter` (Attributes) A filter for roles within the group. (see [below for nested schema](#nestedatt--group_role_restriction--filters--role_filter))

<a id="nestedatt--group_role_restriction--filters--role_filter"></a>

### Nested Schema for `group_role_restriction.filters.role_filter`

Optional:

* `match_condition` (String) The match condition for the roles (AND or OR).
* `roles` (Set of String) An array of role names.

<a id="nestedatt--group_selection"></a>

### Nested Schema for `group_selection`

Optional:

* `always_show_group_selection` (Boolean)
* `selectable_group_types` (Set of String)
* `selectable_groups` (Set of String)

<a id="nestedatt--login_spi"></a>

### Nested Schema for `login_spi`

Optional:

* `oauth_client_id` (String)
* `spi_url` (String)

<a id="nestedatt--mfa"></a>

### Nested Schema for `mfa`

Optional:

* `allowed_methods` (Set of String) Optional set of allowed MFA methods.
* `setting` (String) Specifies the Multi-Factor Authentication (MFA) setting. Allowed values are 'OFF', 'ALWAYS', 'SMART', 'TIME_BASED' and 'SMART_PLUS_TIME_BASED'.
* `time_interval_in_seconds` (Number) Optional time interval in seconds for time-based Multi-Factor Authentication.

<a id="nestedatt--mobile_settings"></a>

### Nested Schema for `mobile_settings`

Optional:

* `bundle_id` (String)
* `key_hash` (String)
* `package_name` (String)
* `team_id` (String)

<a id="nestedatt--operations_allowed_groups"></a>

### Nested Schema for `operations_allowed_groups`

Optional:

* `default_roles` (Set of String)
* `group_id` (String)
* `roles` (Set of String)

<a id="nestedatt--saml_providers"></a>

### Nested Schema for `saml_providers`

Optional:

* `display_name` (String)
* `domains` (Set of String)
* `is_provider_visible` (Boolean)
* `logo_url` (String)
* `provider_name` (String)
* `type` (String)

<a id="nestedatt--social_providers"></a>

### Nested Schema for `social_providers`

Optional:

* `provider_name` (String)
* `social_id` (String)

<a id="nestedatt--suggest_verification_methods"></a>

### Nested Schema for `suggest_verification_methods`

Optional:

* `mandatory_config` (Attributes) Configuration for mandatory verification methods. (see [below for nested schema](#nestedatt--suggest_verification_methods--mandatory_config))
* `optional_config` (Attributes) Configuration for optional verification methods (see [below for nested schema](#nestedatt--suggest_verification_methods--optional_config))
* `skip_duration_in_days` (Number) The number of days for which the verification methods can be skipped (default is 7 days).

<a id="nestedatt--suggest_verification_methods--mandatory_config"></a>

### Nested Schema for `suggest_verification_methods.mandatory_config`

Optional:

* `methods` (Set of String) List of mandatory verification methods.
* `range` (String) The range type for mandatory methods. Allowed value is one of ALLOF or ONEOF.
* `skip_until` (String) The date and time until which the mandatory methods can be skipped.

<a id="nestedatt--suggest_verification_methods--optional_config"></a>

### Nested Schema for `suggest_verification_methods.optional_config`

Optional:

* `methods` (Set of String) List of optional verification methods.

## Import

Import is supported using the following syntax:

```shell
# The import identifier in this command is the client_id of the app to be imported.

terraform import cidaas_app.sample client_id
```

# cidaas_consent (Resource)

The Consent resource in the provider allows you to manage different consents within a specific consent group in cidaas.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:tenant_consent_read
* cidaas:tenant_consent_write
* cidaas:tenant_consent_delete

## Example Usage

```terraform
resource "cidaas_consent" "sample" {
  consent_group_id = cidaas_consent_group.sample.id
  name             = "sample_consent"
  enabled          = true # By default enabled is set to 'true'
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

* `consent_group_id` (String) The `consent_group_id` to which the consent belongs.
* `name` (String) The name of the consent.

### Optional

* `enabled` (Boolean) The flag to enable or disable a speicific consent. By default, the value is set to `true`

### Read-Only

* `created_at` (String) The timestamp when the consent version was created.
* `id` (String) The unique identifier of the consent resource.
* `updated_at` (String) The timestamp when the consent version was last updated.

## Import

In the import statement, the identifier is the combination of `consent_group_id` and `consent_name` joined by the special character ":".

Below is an exmaple of import command to import a consent:

```shell
terraform import cidaas_consent.sample a0508317-cec9-4f3e-afa4:sample_consent
```

# cidaas_consent_group (Resource)

The Consent Group resource in the provider allows you to define and manage consent groups in cidaas.
 Consent Groups are useful to organize and manage consents by grouping related consent items together.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:tenant_consent_read
* cidaas:tenant_consent_write
* cidaas:tenant_consent_delete

## Example Usage

```terraform
resource "cidaas_consent_group" "sample" {
  group_name  = "sample_consent_group"
  description = "sample description"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

* `group_name` (String) The name of the consent group.

### Optional

* `description` (String) Description of the consent group.

### Read-Only

* `created_at` (String) The timestamp when the consent group was created.
* `id` (String) The unique identifier of the consent group.
* `updated_at` (String) The timestamp when the consent group was last updated.

## Import

Import is supported using the following syntax:

```shell
terraform import cidaas_consent_group.sample id
```

# cidaas_consent_version (Resource)

The Consent Version resource in the provider allows you to manage different versions of a specific consent in cidaas.
 This resource also supports managing consent versions across multiple locales enabling different configurations such as URLs and content for each locale.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:tenant_consent_read
* cidaas:tenant_consent_write
* cidaas:tenant_consent_delete

## Example Usage

```terraform
# cidaas_consent_version sample for consent_type "SCOPES"
resource "cidaas_consent_version" "v1" {
  version         = 1
  consent_id      = cidaas_consent.sample.id
  consent_type    = "SCOPES"
  scopes          = ["developer"]
  required_fields = ["name"]
  consent_locales = [
    {
      content = "consent version in German"
      locale  = "de"
    },
    {
      content = "consent version in English"
      locale  = "en"
    }
  ]
}

# cidaas_consent_version sample for consent_type "URL"
resource "cidaas_consent_version" "v2" {
  version      = 2
  consent_id   = cidaas_consent.sample.id
  consent_type = "URL"
  consent_locales = [
    {
      content = "consent version in German"
      locale  = "de"
      url     = "https://cidaas.de/de"
    },
    {
      content = "consent version in English"
      locale  = "en"
      url     = "https://cidaas.de/en"
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

* `consent_id` (String) The `consent_id` for which the consent version is created. It can not be updated for a specific consent version.
* `consent_locales` (Attributes Set) (see [below for nested schema](#nestedatt--consent_locales))
* `version` (Number) The version number of the consent. It can not be updated for a specific consent version.

### Optional

* `consent_type` (String) Specifies the type of consent. The allowed values are `SCOPES` or `URL`. It can not be updated for a specific consent version.
* `required_fields` (Set of String) A set of fields that are required for the consent. It can not be updated for a specific consent version.
Note that the attribute `required_fields` is required only if the `consent_type` is set to **SCOPES**.
* `scopes` (Set of String) A set of scopes related to the consent. It can not be updated for a specific consent version.
Note that the attribute `scopes` is required only if the `consent_type` is set to **SCOPES**.

### Read-Only

* `id` (String) The unique identifier of the consent version.

<a id="nestedatt--consent_locales"></a>

### Nested Schema for `consent_locales`

Required:

* `locale` (String) The locale for which the consent version is created. e.g. `en-us`, `de`.

Optional:

* `content` (String) The content of the consent version associated with a specific locale.
* `url` (String) The url to the consent page of the created consent version.
Note that the attribute `url` is required only if the `consent_type` is set to **URL**.

## Import

In the import statement, the identifier is the combination of `consent_id`, `consent_version_id` and `locale` joined by the special character ":".
To import a consent version for multiple locales, you need to append the locales separated by ":".
For example, the identifier "3f453233-92d4-475b-b10e:813fbd47-6c50-4fc4-881a:en-us:de:en" imports the consent version for the locales `en-us`, `de` and `en`.

Below is an exmaple of import command to import a consent version:

```shell
terraform import cidaas_consent_version.v1 3f453233-92d4-475b-b10e:813fbd47-6c50-4fc4-881a:en-us
```

# cidaas_custom_provider (Resource)

This example demonstrates the configuration of a custom provider resource for interacting with cidaas.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:
- cidaas:providers_read
- cidaas:providers_write
- cidaas:providers_delete


## Example Usage

```terraform
resource "cidaas_custom_provider" "sample" {
  standard_type          = "OAUTH2"
  authorization_endpoint = "https://cidaas.de/authz-srv/authz"
  token_endpoint         = "https://cidaas.de/token-srv/token"
  provider_name          = "terraform-sample"
  display_name           = "Terraform"
  logo_url               = "https://cidaas.de/logo"
  userinfo_endpoint      = "https://cidaas.de/users-srv/userinfo"
  scope_display_label    = "terraform sample scope display name"
  client_id              = "acb-4a6b-9777-8a64abe6af"
  client_secret          = "zcb-4a6b-9777-8a64abe6ay"
  domains                = ["cidaas.de", "cidaas.org"]

  scopes = [
    {
      recommended = true
      required    = true
      scope_name  = "email"
    },
    {
      recommended = true
      required    = true
      scope_name  = "openid"
    },
  ]

  userinfo_fields = {
    family_name        = { "ext_field_key" = "cp_family_name" }
    address            = { "ext_field_key" = "cp_address" }
    birthdate          = { "ext_field_key" = "01-01-2000" }
    email              = { "ext_field_key" = "cp@cidaas.de" }
    email_verified     = { "ext_field_key" = "email_verified", "default" = false }
    gender             = { "ext_field_key" = "male" }
    nickname           = { "ext_field_key" = "nickname" }
    given_name         = { "ext_field_key" = "cp_given_name" }
    locale             = { "ext_field_key" = "cp_locale" }
    middle_name        = { "ext_field_key" = "cp_middle_name" }
    mobile_number      = { "ext_field_key" = "100000000" }
    phone_number       = { "ext_field_key" = "10000000" }
    picture            = { "ext_field_key" = "https://cidaas.de/image.jpg" }
    preferred_username = { "ext_field_key" = "cp_preferred_username" }
    profile            = { "ext_field_key" = "cp_profile" }
    updated_at         = { "ext_field_key" = "01-01-01" }
    website            = { "ext_field_key" = "https://cidaas.de" }
    zoneinfo           = { "ext_field_key" = "cp_zone_info" }
    custom_fields = {
      zipcode         = "123456"
      alternate_phone = "1234567890"
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `authorization_endpoint` (String) The URL for authorization of the provider.
- `client_id` (String) The client ID of the provider.
- `client_secret` (String, Sensitive) The client secret of the provider.
- `display_name` (String) The display name of the provider.
- `provider_name` (String) The unique identifier of the custom provider. This cannot be updated for an existing state.
- `token_endpoint` (String) The URL to generate token with this provider.
- `userinfo_endpoint` (String) The URL to fetch user details using this provider.

### Optional

- `amr_config` (Attributes List) AMR configuration mapping. (see [below for nested schema](#nestedatt--amr_config))
- `apikey_details` (Attributes) Configuration for API key-based authentication. It's a **required** parameter when the auth_type is APIKEY. (see [below for nested schema](#nestedatt--apikey_details))
- `auth_type` (String) Type of authentication. Allowed values `APIKEY`, `CIDAAS_OAUTH2` and `TOTP`.
- `cidaas_auth_details` (Attributes) Configuration for cidaas authentication. It's a **required** parameter when the auth_type is CIDAAS_OAUTH2. (see [below for nested schema](#nestedatt--cidaas_auth_details))
- `domains` (Set of String) The domains of the provider.
- `logo_url` (String) The URL for the provider's logo.
- `pkce` (Boolean) The flag to enable or disable pkce flow. By default, the value is set to `false`
- `scope_display_label` (String) Display label for the scope of the provider.
- `scopes` (Attributes List) List of scopes of the provider with details (see [below for nested schema](#nestedatt--scopes))
- `standard_type` (String) Type of standard. Allowed values `OAUTH2` and `OPENID_CONNECT`.
- `totp_details` (Attributes) Configuration for TOTP based authentication.  It's a **required** parameter when the auth_type is TOTP. (see [below for nested schema](#nestedatt--totp_details))
- `userinfo_fields` (Attributes) Object containing various user information fields with their values. The userinfo_fields section includes specific fields such as name, family_name, address, etc., along with custom_fields allowing additional user information customization (see [below for nested schema](#nestedatt--userinfo_fields))
- `userinfo_source` (String) Source of userinfo. Allowed values are `IDTOKEN` and `USERINFOENDPOINT`.

### Read-Only

- `id` (String) The ID of the resource.

<a id="nestedatt--amr_config"></a>
### Nested Schema for `amr_config`

Required:

- `amr_value` (String)
- `ext_amr_value` (String)


<a id="nestedatt--apikey_details"></a>
### Nested Schema for `apikey_details`

Required:

- `key` (String) The API key that will be used to authenticate the webhook request.The key that will be passed in the request header or in query param as configured in the attribute `placement`
- `placeholder` (String) The attribute is the placeholder for the key which need to be passed as a query parameter or in the request header.
- `placement` (String) The placement of the API key in the request (e.g., query).The allowed value are `header` and `query`.


<a id="nestedatt--cidaas_auth_details"></a>
### Nested Schema for `cidaas_auth_details`

Required:

- `client_id` (String) The client ID for cidaas authentication.


<a id="nestedatt--scopes"></a>
### Nested Schema for `scopes`

Optional:

- `recommended` (Boolean) Indicates if the scope is recommended.
- `required` (Boolean) Indicates if the scope is required.
- `scope_name` (String) The name of the scope, e.g., `openid`, `profile`.


<a id="nestedatt--totp_details"></a>
### Nested Schema for `totp_details`

Required:

- `key` (String) The key used for TOTP authentication.
- `placeholder` (String) A placeholder value for the TOTP.
- `placement` (String) The placement of the TOTP in the request.The allowed value are `header` and `query`.


<a id="nestedatt--userinfo_fields"></a>
### Nested Schema for `userinfo_fields`

Optional:

- `address` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--address))
- `birthdate` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--birthdate))
- `custom_fields` (Map of String)
- `email` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--email))
- `email_verified` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--email_verified))
- `family_name` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--family_name))
- `gender` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--gender))
- `given_name` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--given_name))
- `groups` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--groups))
- `locale` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--locale))
- `middle_name` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--middle_name))
- `mobile_number` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--mobile_number))
- `name` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--name))
- `nickname` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--nickname))
- `phone_number` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--phone_number))
- `picture` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--picture))
- `preferred_username` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--preferred_username))
- `profile` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--profile))
- `sub` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--sub))
- `updated_at` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--updated_at))
- `website` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--website))
- `zoneinfo` (Attributes) (see [below for nested schema](#nestedatt--userinfo_fields--zoneinfo))

<a id="nestedatt--userinfo_fields--address"></a>
### Nested Schema for `userinfo_fields.address`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--birthdate"></a>
### Nested Schema for `userinfo_fields.birthdate`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--email"></a>
### Nested Schema for `userinfo_fields.email`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--email_verified"></a>
### Nested Schema for `userinfo_fields.email_verified`

Optional:

- `default` (Boolean)
- `ext_field_key` (String)


<a id="nestedatt--userinfo_fields--family_name"></a>
### Nested Schema for `userinfo_fields.family_name`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--gender"></a>
### Nested Schema for `userinfo_fields.gender`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--given_name"></a>
### Nested Schema for `userinfo_fields.given_name`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--groups"></a>
### Nested Schema for `userinfo_fields.groups`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--locale"></a>
### Nested Schema for `userinfo_fields.locale`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--middle_name"></a>
### Nested Schema for `userinfo_fields.middle_name`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--mobile_number"></a>
### Nested Schema for `userinfo_fields.mobile_number`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--name"></a>
### Nested Schema for `userinfo_fields.name`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--nickname"></a>
### Nested Schema for `userinfo_fields.nickname`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--phone_number"></a>
### Nested Schema for `userinfo_fields.phone_number`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--picture"></a>
### Nested Schema for `userinfo_fields.picture`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--preferred_username"></a>
### Nested Schema for `userinfo_fields.preferred_username`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--profile"></a>
### Nested Schema for `userinfo_fields.profile`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--sub"></a>
### Nested Schema for `userinfo_fields.sub`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--updated_at"></a>
### Nested Schema for `userinfo_fields.updated_at`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--website"></a>
### Nested Schema for `userinfo_fields.website`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)


<a id="nestedatt--userinfo_fields--zoneinfo"></a>
### Nested Schema for `userinfo_fields.zoneinfo`

Required:

- `ext_field_key` (String)

Optional:

- `default` (String)

## Import

Import is supported using the following syntax:

```shell
terraform import cidaas_custom_provider.resource_name provider_name
```

# cidaas_group_type (Resource)

The Group Type, managed through the `cidaas_group_type` resource in the provider defines and configures categories for user groups within the cidaas system.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:group_type_read
* cidaas:group_type_write
* cidaas:group_type_delete

## Example Usage

```terraform
resource "cidaas_group_type" "sample" {
  role_mode     = "no_roles"
  group_type    = "TerraformSampleGroupType"
  description   = "terraform user group category description"
  allowed_roles = ["developer"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

* `group_type` (String) The unique identifier of the group type. This cannot be updated for an existing state.
* `role_mode` (String) Determines the role mode for the user group type. Allowed values are `any_roles`, `no_roles`, `roles_required` and `allowed_roles`

### Optional

* `allowed_roles` (Set of String) List of allowed roles in this group type.
* `description` (String) The `description` attribute provides details about the group type, explaining its purpose.

### Read-Only

* `created_at` (String) The timestamp when the resource was created.
* `id` (String) The ID of the resource.
* `updated_at` (String) The timestamp when the resource was last updated.

## Import

Import is supported using the following syntax:

```shell
terraform import cidaas_group_type.resource_name group_type
```

# cidaas_hosted_page (Resource)

The Hosted Page resource in the provider allows you to define and manage hosted pages within the cidaas system.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:hosted_pages_write
* cidaas:hosted_pages_read
* cidaas:hosted_pages_delete

## Example Usage

```terraform
resource "cidaas_hosted_page" "sample" {
  hosted_page_group_name = "terraform-sample-hosted-page"
  default_locale         = "en-IN"
  hosted_pages = [
    {
      hosted_page_id = "register_success"
      locale         = "en-US"
      url            = "https://cidaas.de/register_success_hosted_page"
      content        = "content"
    },
    {
      hosted_page_id = "register_success"
      locale         = "en-IN"
      url            = "https://cidaas.de/register_success_hosted_page"
      content        = "content"
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

* `hosted_page_group_name` (String) The name of the hosted page group. This must be unique across the cidaas system and cannot be updated for an existing state.
* `hosted_pages` (Attributes Set) List of hosted pages with their respective attributes (see [below for nested schema](#nestedatt--hosted_pages))

### Optional

* `default_locale` (String) The default locale for hosted pages e.g. `en-US`.

### Read-Only

* `created_at` (String) The timestamp when the resource was created.
* `id` (String) The ID of the resource.
* `updated_at` (String) The timestamp when the resource was last updated.

<a id="nestedatt--hosted_pages"></a>

### Nested Schema for `hosted_pages`

Required:

* `hosted_page_id` (String) The identifier for the hosted page, e.g., `register_success`.
* `url` (String) The URL for the hosted page.

Optional:

* `content` (String) The conent of the hosted page.
* `locale` (String) The locale for the hosted page, e.g., `en-US`.

## Import

Import is supported using the following syntax:

```shell
terraform import cidaas_hosted_page.resource_name hosted_page_id
```

# cidaas_password_policy (Resource)

The Password Policy resource in the provider allows you to manage the password policy within the cidaas.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:password_policy_read
* cidaas:password_policy_write
* cidaas:password_policy_delete

## Example Usage

```terraform
resource "cidaas_password_policy" "sample" {
  policy_name = "sample_terraform_policy"
  password_policy = {
    block_compromised = false,
    deny_usage_count  = 3,
    strength_regexes = [
      "^(?=.*[A-Za-z])(?!.*\\s).{6,15}$"
    ],
    change_enforcement = {
      expiration_in_days         = 90
      notify_user_before_in_days = 7
    }
  }
}
```

## Schema

### Required

- `password_policy` (Attributes) The password policy configuration. All attributes are optional except strength_regexes. If not provided, default values will be applied. (see [below for nested schema](#nestedatt--password_policy))
- `policy_name` (String) The name of the password policy.

### Read-Only

- `id` (String) Unique identifier of the password policy.

<a id="nestedatt--password_policy"></a>
### Nested Schema for `password_policy`

Required:

- `strength_regexes` (Set of String) The regular expression to enforce the minimum and maximum character count, minimum number of numeric and special characters and whether to include lowercase or uppercase letters in a password.

Optional:

- `block_compromised` (Boolean) Flag to block passwords that have been compromised.
- `change_enforcement` (Attributes) (see [below for nested schema](#nestedatt--password_policy--change_enforcement))
- `deny_usage_count` (Number) The reuse limit specifies the maximum number of times a user can reuse a previous password.

<a id="nestedatt--password_policy--change_enforcement"></a>
### Nested Schema for `password_policy.change_enforcement`

Optional:

- `expiration_in_days` (Number) The number of days allowed before a password must be changed.
- `notify_user_before_in_days` (Number) Number of days before password expiry to notify the user.

## Import

Import is supported using the following syntax:

```shell
terraform import cidaas_password_policy.resource_name id
```

# cidaas_registration_field (Resource)

The `cidaas_registration_page_field` in the provider allows management of registration fields in the cidaas system. This resource enables you to configure and customize the fields displayed during user registration.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:field_setup_read
* cidaas:field_setup_write
* cidaas:field_setup_delete

## Example Usage

```terraform
resource "cidaas_registration_field" "text" {
  data_type                                      = "TEXT"
  field_key                                      = "sample_text_field"
  field_type                                     = "CUSTOM"  // CUSTOM and SYSTEM, SYSTEM can not be created but modified
  internal                                       = true      // Default: false
  required                                       = true      // Default: false
  read_only                                      = true      // Default: false
  is_group                                       = false     // Default: false
  unique                                         = true      // Default: false
  overwrite_with_null_value_from_social_provider = false     // Default: true
  is_searchable                                  = true      // Default: true
  enabled                                        = true      // Default: true
  claimable                                      = true      // Default: true
  order                                          = 1         // Default: 1
  parent_group_id                                = "DEFAULT" // Default: DEFAULT
  scopes                                         = ["profile"]
  local_texts = [
    {
      locale         = "en-US"
      name           = "Sample Field"
      required_msg   = "The field is required"
      max_length_msg = "Maximum 99 chars allowed"
      min_length_msg = "Minimum 99 chars allowed"
    },
    {
      locale         = "de-DE"
      name           = "Beispielfeld"
      required_msg   = "Dieses Feld ist erforderlich"
      max_length_msg = "DE maximum 99 chars allowed"
      min_length_msg = "DE minimum 10 chars allowed"
    }
  ]
  field_definition = {
    regex = "^.{10,100}$"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

* `data_type` (String) The data type of the field. This cannot be modified for an existing resource. Allowed values are `TEXT`,`NUMBER`,`SELECT`,`MULTISELECT`,`RADIO`,`CHECKBOX`,`PASSWORD`,`DATE`,`URL`,`EMAIL`,`TEXTAREA`,`MOBILE`,`CONSENT`,`JSON_STRING`,`USERNAME`,`ARRAY`,`GROUPING`,`DAYDATE`,
* `field_key` (String) The unique identifier of the registration field. This cannot be modified for an existing resource.
* `local_texts` (Attributes List) The localized detail of the registration field. (see [below for nested schema](#nestedatt--local_texts))

### Optional

* `claimable` (Boolean) Flag to mark if a field is claimable. Defaults set to `true`
* `consent_refs` (Set of String) List of consents(the ids of the consent in cidaas must be passed) in registration. The data type must be `CONSENT` in this case
* `enabled` (Boolean) Flag to mark if a field is enabled. Defaults set to `true`
* `field_definition` (Attributes) (see [below for nested schema](#nestedatt--field_definition))
* `field_type` (String) Specifies whether the field type is `SYSTEM` or `CUSTOM`. Defaults to `CUSTOM`. This cannot be modified for an existing resource. `SYSTEM` fields cannot be created but can be modified. To modify an existing field import it first and then update.
* `internal` (Boolean) Flag to mark if a field is internal. Defaults set to `false`
* `is_group` (Boolean) Setting is_group to `true` creates a registration field group. Defaults set to `false` The data_type attribute must be set to TEXT when is_group is true.
* `is_list` (Boolean)
* `is_searchable` (Boolean) Flag to mark if a field is searchable. Defaults set to `true`
* `order` (Number) The attribute order is used to set the order of the Field in the UI. Defaults set to `1`
* `overwrite_with_null_value_from_social_provider` (Boolean) Set to true if you want the value should be reset by identity provider. Defaults set to `false`
* `parent_group_id` (String) The ID of the parent registration group. Defaults to `DEFAULT` if not provided.
* `read_only` (Boolean) Flag to mark if a field is read only. Defaults set to `false`
* `required` (Boolean) Flag to mark if a field is required in registration. Defaults set to `false`
* `scopes` (Set of String) The scopes of the registration field.
* `unique` (Boolean) Flag to mark if a field is unique. Defaults set to `false`

### Read-Only

* `base_data_type` (String) The base data type of the field. This is computed property.
* `id` (String) The ID of the resource

<a id="nestedatt--local_texts"></a>

### Nested Schema for `local_texts`

Required:

* `name` (String) The name of the field in the local configured. for example: in **en-US** the name is `Sample Field` in de-DE `Beispielfeld`.

Optional:

* `attributes` (Attributes List) The field attributes must be provided for the data_type SELECT, MULTISELECT and RADIO. it's an array of key value pairs. Example provided in the example section. (see [below for nested schema](#nestedatt--local_texts--attributes))
* `consent_label` (Attributes) required when data_type is CONSENT. Example provided in the example section. (see [below for nested schema](#nestedatt--local_texts--consent_label))
* `locale` (String) The locale of the field. example: de-DE.
* `max_length_msg` (String) warning/error msg to show to the user when user exceeds the maximum character configured. This is applicable only for the attributes of base_data_type string.
* `min_length_msg` (String) warning/error msg to show to the user when user don't provide the minimum character required. This is applicable only for the attributes of base_data_type string.
* `required_msg` (String) When the flag required is set to true the required_msg must be provided. required_msg is shown if user does not provide a required field.

<a id="nestedatt--local_texts--attributes"></a>

### Nested Schema for `local_texts.attributes`

Required:

* `key` (String)
* `value` (String)

<a id="nestedatt--local_texts--consent_label"></a>

### Nested Schema for `local_texts.consent_label`

Required:

* `label` (String)
* `label_text` (String)

<a id="nestedatt--field_definition"></a>

### Nested Schema for `field_definition`

Optional:

* `initial_date` (String) The initial date. Applicable only for DATE attributes. Example format: `2024-06-28T18:30:00Z`.
* `initial_date_view` (String) The view of the calender. Applicable only for DATE attributes. Allowed values: `month`, `year` and `multi-year`
* `max_date` (String) The maximum date a user can select. Applicable only for DATE attributes. Example format: `2024-06-28T18:30:00Z`.
* `max_length` (Number) The maximum length of a string type attribute.
* `min_date` (String) The earliest date a user can select. Applicable only for DATE attributes. Example format: `2024-06-28T18:30:00Z`.
* `min_length` (Number) The minimum length of a string type attribute
* `regex` (String) The regex for max_length and min_length for the data types TEXT and URL.

## Import

Import is supported using the following syntax:

```shell
terraform import cidaas_registration_page_field.resource_name field_key
```

# cidaas_role (Resource)

The cidaas_role resource in Terraform facilitates the management of roles in cidaas system. This resource allows you to configure and define custom roles to suit your application's specific access control requirements.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:roles_read
* cidaas:roles_write
* cidaas:roles_delete

## Example Usage

```terraform
resource "cidaas_role" "sample" {
  role        = "terraform_sample_role"
  name        = "Terraform Sample Role"
  description = "The sample is designed to demonstrate the configuration of the terraform cidaas_role resource."
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

* `role` (String) The unique identifier of the role. The role name must be unique across the cidaas system and cannot be updated for an existing state.

### Optional

* `description` (String) The `description` attribute provides details about the role, explaining its purpose.
* `name` (String) The name of the role.

### Read-Only

* `id` (String) The ID of the role resource.

## Import

Import is supported using the following syntax:

```shell
terraform import cidaas_role.resource_name role
```

# cidaas_scope_group (Resource)

The cidaas_scope_group resource in the provider allows to manage Scope Groups in cidaas system. Scope Groups help organize and group related scopes for better categorization and access control.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:scopes_read
* cidaas:scopes_write
* cidaas:scopes_delete

## Example Usage

```terraform
resource "cidaas_scope_group" "sample" {
  group_name  = "terraform-sample-scope"
  description = "The sample is designed to demonstrate the configuration of the terraform cidaas_scope_group resource."
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

* `group_name` (String) The name of the group. The group name must be unique across the cidaas system and cannot be updated for an existing state.

### Optional

* `description` (String) The `description` attribute provides details about the scope of the group, explaining its purpose.

### Read-Only

* `created_at` (String) The timestamp when the resource was created.
* `id` (String) The ID of th resource.
* `updated_at` (String) The timestamp when the resource was last updated.

## Import

Import is supported using the following syntax:

```shell
terraform import cidaas_scope_group.resource_name group_name
```

# cidaas_scope (Resource)

The Scope resource allows to manage scopes in cidaas system. Scopes define the level of access and permissions granted to an application (client).

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:scopes_read
* cidaas:scopes_write
* cidaas:scopes_delete

## Example Usage

```terraform
resource "cidaas_scope" "sample" {
  security_level        = "CONFIDENTIAL"
  scope_key             = "terraform-sample-scope"
  required_user_consent = false
  group_name            = []
  localized_descriptions = [
    {
      title       = "cidaas Scope Tunisia Title"
      locale      = "ar-TN"
      description = "This is scope in local ar-TN"
    },
    {
      title       = "cidaas Scope German Title"
      locale      = "de-DE"
      description = "This is scope in local de-DE"
    },
    {
      title       = "cidaas Scope India Title"
      locale      = "en-IN"
      description = "This is scope in local en-IN"
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

* `scope_key` (String) Unique identifier for the scope. This cannot be updated for an existing state.

### Optional

* `group_name` (Set of String) List of scope_groups to associate the scope with.
* `localized_descriptions` (Attributes List) (see [below for nested schema](#nestedatt--localized_descriptions))
* `required_user_consent` (Boolean) Indicates whether user consent is required for the scope.
* `scope_owner` (String) The owner of the scope. e.g. `ADMIN`
* `security_level` (String) The security level of the scope, e.g., `PUBLIC`. Allowed values are `PUBLIC` and `CONFIDENTIAL`

### Read-Only

* `id` (String) The ID of the resource.

<a id="nestedatt--localized_descriptions"></a>

### Nested Schema for `localized_descriptions`

Required:

* `title` (String) The title of the scope in the configured locale.

Optional:

* `description` (String) The description of the scope in the configured locale.
* `locale` (String) The locale for the scope, e.g., `en-US`.

## Import

Import is supported using the following syntax:

```shell
terraform import cidaas_scope.resource_name scope_key
```

# cidaas_social_provider (Resource)

The `cidaas_social_provider` resource allows you to configure and manage social login providers within cidaas.
 Social login providers enable users to authenticate using their existing accounts from popular social platforms such as Google, Facebook, LinkedIn and others.

 Ensure that the below scopes are assigned to the client:

* cidaas:providers_read
* cidaas:providers_write
* cidaas:providers_delete

## Example Usage

```terraform
resource "cidaas_social_provider" "sample" {
  name                     = "Sample Social Provider"
  provider_name            = "google"
  enabled                  = true
  client_id                = "8d789b3d-b312"
  client_secret            = "96ae-ea2e8d8e6708"
  scopes                   = ["profile", "email"]
  enabled_for_admin_portal = true
  claims = {
    required_claims = {
      user_info = ["name"]
      id_token  = ["phone_number"]
    }
    optional_claims = {
      user_info = ["website"]
      id_token  = ["street_address"]
    }
  }
  userinfo_fields = [
    {
      inner_key       = "sample_custom_field"
      external_key    = "external_sample_cf"
      is_custom_field = true
      is_system_field = false
    },
    {
      inner_key       = "sample_system_field"
      external_key    = "external_sample_sf"
      is_custom_field = false
      is_system_field = true
    }
  ]
}
```

### Configuring a Social Provider to a Client

To configure a social provider for a client in your Terraform configuration, you need to update the `cidaas_app` resources with the details from the `cidaas_social_provider` resource. Below is an example demonstrating how you can configure a social provider for a client:

```terraform
resource "cidaas_app" "app_sample" {
  ...
  social_providers = [
    {
      provider_name = cidaas_social_provider.sample.provider_name
      social_id     = cidaas_social_provider.sample.id
      display_name  = "google"
    }
  ]
...
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

* `client_id` (String) The client ID provided by the social provider. This is used to authenticate your application with the social provider.
* `client_secret` (String, Sensitive) The client secret provided by the social provider. This is used alongside the client ID to authenticate your application with the social provider.
* `name` (String) The name of the social provider configuration. This should be unique within your cidaas environment.
* `provider_name` (String) The name of the social provider. Supported values include `google`, `facebook`, `linkedin` etc.

### Optional

* `claims` (Attributes) A map defining required and optional claims to be requested from the social provider. (see [below for nested schema](#nestedatt--claims))
* `enabled` (Boolean) A flag to enable or disable the social provider configuration. Set to `true` to enable and `false` to disable.
* `enabled_for_admin_portal` (Boolean) A flag to enable or disable the social provider for the admin portal. Set to `true` to enable and `false` to disable.
* `scopes` (Set of String) A list of scopes of the social provider.
* `userinfo_fields` (Attributes List) A list of user info fields to be mapped between the social provider and cidaas. (see [below for nested schema](#nestedatt--userinfo_fields))

### Read-Only

* `id` (String) The unique identifier of the social provider

<a id="nestedatt--claims"></a>

### Nested Schema for `claims`

Optional:

* `optional_claims` (Attributes) Defines the claims that are optional from the social provider. (see [below for nested schema](#nestedatt--claims--optional_claims))
* `required_claims` (Attributes) Defines the claims that are required from the social provider. (see [below for nested schema](#nestedatt--claims--required_claims))

<a id="nestedatt--claims--optional_claims"></a>

### Nested Schema for `claims.optional_claims`

Optional:

* `id_token` (Set of String) A list of ID token claims that are optional.
* `user_info` (Set of String) A list of user information claims that are optional.

<a id="nestedatt--claims--required_claims"></a>

### Nested Schema for `claims.required_claims`

Optional:

* `id_token` (Set of String) A list of ID token claims that are required.
* `user_info` (Set of String) A list of user information claims that are required.

<a id="nestedatt--userinfo_fields"></a>

### Nested Schema for `userinfo_fields`

Required:

* `external_key` (String) The external key used by the social provider.
* `inner_key` (String) The internal key used by cidaas.
* `is_custom_field` (Boolean) A flag indicating whether the field is a custom field. Set to `true` if it is a custom field.
* `is_system_field` (Boolean) A flag indicating whether the field is a system field. Set to `true` if it is a system field.

## Import

The import identifier of resource social provider is a combination of **provider_name** and **provider_id** joined by the special character ":".
For example, if the resource name is `sample` with provider_name `google` and provider_id `8d789b3d-b312-4251`, the import statement would be:

```shell
terraform import cidaas_social_provider.sample google:8d789b3d-b312-4251
```

# cidaas_template_group (Resource)

The cidaas_template_group resource in the provider is used to define and manage templates groups within the cidaas system. Template Groups categorize your communication templates allowing you to map preferred templates to specific clients effectively.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:templates_read
* cidaas:templates_write
* cidaas:templates_delete

## Example Usage

```terraform
// To create a template group, only the attribute group_id is required in the configuration.
// The attributes shown in sample-tg-2 are optional and can be configured as needed.
// If these properties are not configured in the .tf file, the provider/cidaas will compute
// and assign values to them.

// sample1
resource "cidaas_template_group" "sample-tg-1" {
  group_id = "sample_group"
}

// sample2
resource "cidaas_template_group" "sample-tg-2" {
  group_id = "group_another"
  email_sender_config = {
    from_email = "noreply@cidaas.de"
    from_name  = "Kube-dev"
    reply_to   = "noreply@cidaas.de"
    sender_names = [
      "System",
    ]
  }
  ivr_sender_config = {
    sender_names = [
      "System",
    ]
  }
  push_sender_config = {
    sender_names = [
      "System",
    ]
  }
  sms_sender_config = {
    sender_names = [
      "System",
    ]
  }
}
```

Warning: Using "SYSTEM" instead of "System" as `sender_names` will cause Terraform to fail with 500 error (code: 35001).

<!-- schema generated by tfplugindocs -->
## Schema

### Required

* `group_id` (String) The group_id of the Template Group. The group_id is used to import an existing template group. The maximum allowed length of a group_id is **15** characters.

### Optional

* `email_sender_config` (Attributes) The `email_sender_config` is used to configure your email sender. (see [below for nested schema](#nestedatt--email_sender_config))
* `ivr_sender_config` (Attributes) The configuration of the IVR sender. (see [below for nested schema](#nestedatt--ivr_sender_config))
* `push_sender_config` (Attributes) The configuration of the PUSH notification sender. (see [below for nested schema](#nestedatt--push_sender_config))
* `sms_sender_config` (Attributes) The configuration of the SMS sender. (see [below for nested schema](#nestedatt--sms_sender_config))

### Read-Only

* `id` (String) The ID of the resource

<a id="nestedatt--email_sender_config"></a>

### Nested Schema for `email_sender_config`

Optional:

* `from_email` (String) The email from address from which the emails will be sent when the specific group is configured.
* `from_name` (String) The `from_name` attribute is the display name that appears in the 'From' field of the emails.
* `reply_to` (String) The `reply_to` attribute is the email address where replies should be directed.
* `sender_names` (Set of String) The `sender_names` attribute defines the names associated with email senders.

<a id="nestedatt--ivr_sender_config"></a>

### Nested Schema for `ivr_sender_config`

Optional:

* `sender_names` (Set of String)

<a id="nestedatt--push_sender_config"></a>

### Nested Schema for `push_sender_config`

Optional:

* `sender_names` (Set of String)

<a id="nestedatt--sms_sender_config"></a>

### Nested Schema for `sms_sender_config`

Optional:

* `from_name` (String)
* `sender_names` (Set of String)

## Import

Import is supported using the following syntax:

```shell
terraform import cidaas_template_group.resource_name group_id
```

# cidaas_template (Resource)

The Template resource in the provider is used to define and manage templates within the cidaas system. Templates are used for emails, SMS, IVR, and push notifications.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:templates_read
* cidaas:templates_write
* cidaas:templates_delete

### Managing System Templates

* To create system templates, set the **is_system_template** flag to `true`.
By default, this value is `false` and creates custom templates when applied.
* When creating system templates validation checks are applied and suggestions are
provided in error messages to assist users in creating system templates.
* System templates cannot be imported using the standard Terraform import command. Instead, users
must create a configuration that matches the existing system template and run terraform apply.

## Example Usage

```terraform
// custom template example
resource "cidaas_template" "custom-template-1" {
  locale        = "en-US"
  template_key  = "TERRAFORM_TEMPLATE"
  template_type = "EMAIL"
  content       = "Sample email body in US English"
  subject       = "Email custom template subject (en-US)"
}

// custom template example with same template_key as custom-template-1 but different template_type and locale
resource "cidaas_template" "custom-template-2" {
  locale        = "de-DE"
  template_key  = "TERRAFORM_TEMPLATE"
  template_type = "SMS"
  content       = "Sample SMS template content in German English"
}

// custom template example with same template_key and template_type as custom-template-2 but different locale
resource "cidaas_template" "custom-template-3" {
  locale        = "en-US"
  template_key  = "TERRAFORM_TEMPLATE"
  template_type = "SMS"
  content       = "Sample SMS template content in US English"
}


// System templates are created by setting the flag is_system_template to true.
// By default, this value is false and creates custom templates when applied.
// Validation checks are applied, and suggestions are provided in error messages to assist users in creating system templates.
// System templates cannot be imported using the standard Terraform import command.
// Instead, users must create a configuration that matches the existing system template and run terraform apply.

// Example of a system template for the template group "sample_group":
resource "cidaas_template" "system-template-1" {
  locale             = "en-US"
  template_key       = "VERIFY_USER"
  template_type      = "SMS"
  content            = "Hi {{name}}, here is the {{code}} to verify the user"
  is_system_template = true
  group_id           = "sample_group"
  processing_type    = "GENERAL"
  verification_type  = "SMS"
  usage_type         = "VERIFICATION_CONFIGURATION"
}

// Example of a  system template for the system default template_group "default"
resource "cidaas_template" "system-template-2" {
  locale             = "en-US"
  template_key       = "NOTIFY_COMMUNICATION_CHANGE"
  template_type      = "SMS"
  content            = "Your mobile number changed in {{account_name}}-account to {{communication_medium_value}}."
  is_system_template = true
  group_id           = "default"
  processing_type    = "GENERAL"
  usage_type         = "GENERAL"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

* `content` (String) The content of the template.
* `locale` (String) The BCP47 locale of the template (e.g. `en-US`, `en-GB`, `de-DE`, or a language-only tag like `en`). Use lowercase for the language subtag and uppercase for the region when present. Find the allowed locales in the Allowed Locales section below. It cannot be updated for an existing state.
* `template_key` (String) The unique name of the template. It cannot be updated for an existing state.
* `template_type` (String) The type of the template. Allowed template_types are EMAIL, SMS, IVR and PUSH. Template types are case sensitive. It cannot be updated for an existing state.

### Optional

* `group_id` (String) The `group_id` under which the configured template will be categorized. Only applicable for SYSTEM templates.
* `is_system_template` (Boolean) A boolean flag to decide between SYSTEM and CUSTOM template. When set to true the provider creates a SYSTEM template else CUSTOM
* `processing_type` (String) The processing_type attribute specifies the method by which the template information is processed and delivered. Only applicable for SYSTEM templates. It should be set to `GENERAL` when cidaas does not provide an allowed list of values.
* `subject` (String) Applicable only for template_type EMAIL. It represents the subject of an email.
* `usage_type` (String) The usage_type attribute specifies the specific use case or application for the template. Only applicable for SYSTEM templates. It should be set to `GENERAL` when cidaas does not provide an allowed list of values.
* `verification_type` (String) The verification_type attribute defines the method used for verification. Only applicable for SYSTEM templates.
* `enabled` (Boolean) A boolean flag enable or disable a template.

### Read-Only

* `id` (String) The unique identifier of the template resource.
* `language` (String) The language based on the local provided in the configuration.
* `template_owner` (String) The template owner of the template.

## Import

Import is supported using the following syntax:

```shell
# System templates cannot be imported using the standard Terraform import command.
# Instead, users must create a configuration that matches the existing system template and run terraform apply.

# V3 Change Note: The format of the import identifier is changed in V3. In V2, the import identifier was joined by the chracter "-"
# However in V3, it is replaced by the chracter ":". Example: TERRAFORM_TEMPLATE:SMS:en-US

# Below is the command to import a custom template
# Here, template_key:template_type:locale is a combination of template_key, template_type and locale, joined by the special character ":".
# For example, if the resource name is "sample" with template_key as "TERRAFORM_TEMPLATE", template_type as "SMS" and locale as "de-DE", the import statement would be:

terraform import cidaas_template.sample TERRAFORM_TEMPLATE:SMS:de-DE
```

# cidaas_notifications_template_group (Resource)

Manages a **template group** via notification-srv `POST/GET/PUT/DELETE …/templategroups/`. Requires provider `notifications_context_path` (default `notifications-srv`). Use `data.cidaas_notification_service_setups` to resolve `service_setup_id` values for `comm_setting_*` blocks.

See [Notification templates (notification-srv)](#notification-templates-notification-srv). Full schema is generated for the registry; run `make docs` if your fork uses `tfplugindocs`.

# cidaas_notification_template (Resource)

Manages a **single template document** via notification-srv `GET/POST/PUT/DELETE …/templates/…`. Import with the template document id (the `_id` value; the locale segment is often lowercased in that id, e.g. `en-us`, while `locale` in configuration should use canonical BCP47 such as `en-US`):

```shell
terraform import cidaas_notification_template.example "<template_document_id>"
```

# cidaas_notification_template_type (Resource)

Manages **template types** via notification-srv `…/templatetypes/…` (same `{ctx}` as other notification-srv resources). Scopes: `cidaas:templates_*` as documented on the resource.

# cidaas_notification_service_setup (Resource)

Manages a **communication provider service setup** via notification-srv `…/servicesetups/`. **`status`** is computed (manual service-desk verify). Pair with **`cidaas_notification_provider_config`**. Scopes: `cidaas:service_setups_read`, `cidaas:service_setups_write`, `cidaas:service_setups_delete`. See the [Notification service guide](docs/guides/notification_srv.md).

# cidaas_notification_provider_config (Resource)

Stores **provider credentials** for a service setup via `POST …/providerconfigs` (`config_data` or write-only `config_data_wo` + `config_data_wo_version`; Terraform **≥ 1.11**). Destroy removes state only. Scopes: `cidaas:service_setups_read`, `cidaas:provider_config_write`.

# cidaas_notification_service_setup (Data source)

`data.cidaas_notification_service_setup` reads one setup by id (`GET …/servicesetups/{id}`), including **`in-progress`** status for polling after create.

# cidaas_notification_service_setups (Data source)

`data.cidaas_notification_service_setups` calls `GET …/servicesetups/` and exposes **active** setups (`id`, `name`, `service_id`, etc.) for wiring `comm_setting_*` blocks. Newly created setups stay out of this list until verified/`active`.

# cidaas_notification_templates (Data source)

`data.cidaas_notification_templates` requires `graph_filter` (JSON string) and returns `templates` from `POST …/graph/templates/`.

# cidaas_notification_template_groups (Data source)

`data.cidaas_notification_template_groups` requires `graph_filter` (JSON string) and returns `groups` from `POST …/graph/templategroups/`.

# cidaas_user_groups (Resource)

The cidaas_user_groups resource enables the creation of user groups in the cidaas system. These groups allow users to be organized and assigned group-specific roles.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:groups_write
* cidaas:groups_read
* cidaas:groups_delete

## Example Usage

```terraform
# In the below examples, 'parent-user-group' is the top-level group, and its group_id is passed as parent_id in the 'child-user-group' resource.

resource "cidaas_user_groups" "parent-user-group" {
  group_type                     = "test_terraform"
  group_id                       = "sample-group-id"
  group_name                     = "sample-group-name"
  logo_url                       = "https://cidaas.de/logo"
  description                    = "sample parent user groups description"
  custom_fields                  = {}
  make_first_user_admin          = true
  member_profile_visibility      = "full"
  none_member_profile_visibility = "public"
}


resource "cidaas_user_groups" "child-user-group" {
  group_type  = "test_terraform"
  group_id    = "sample-child-group-id-sub"
  group_name  = "sample-child-group-name"
  logo_url    = "https://cidaas.de/logo"
  description = "sample child user groups description"
  custom_fields = {
    first_name  = "cidaas"
    family_name = "widaas"
  }
  parent_id = cidaas_user_groups.sample.group_id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

* `group_id` (String) Identifier for the user group.
* `group_name` (String) Name of the user group.

### Optional

* `custom_fields` (Map of String) Custom fields for the user group.
* `description` (String) Description of the user group.
* `group_type` (String) Type of the user group.
* `logo_url` (String) URL for the user group's logo
* `make_first_user_admin` (Boolean) Indicates whether the first user should be made an admin.
* `member_profile_visibility` (String) Visibility of member profiles. Allowed values `public` or `full`.
* `none_member_profile_visibility` (String) Visibility of non-member profiles. Allowed values `none` or `public`.
* `parent_id` (String) Identifier of the parent user group.

### Read-Only

* `created_at` (String) The timestamp when the resource was created.
* `id` (String) The unique identifier of the user group resource.
* `updated_at` (String) The timestamp when the resource was last updated.

## Import

Import is supported using the following syntax:

```shell
terraform import cidaas_user_groups.resource_name group_id
```

# cidaas_webhook (Resource)

The Webhook resource in the provider facilitates integration of webhooks in the cidaas system. This resource allows you to configure webhooks with different authentication options.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:

* cidaas:webhook_read
* cidaas:webhook_write
* cidaas:webhook_delete

## Example Usage

```terraform
# This is a sample configuration for setting up a webhook with multiple authentication options.
# The available authentication types include apikey_config, totp_config, and cidaas_auth_config.

# When the auth_type is set to "APIKEY", the apikey_config is required, while totp_config and cidaas_auth_config are optional.
# These optional configurations can be removed if not needed. However, by including them, you can easily switch the auth_type 
# to other options by simply updating the auth_type value without needing to modify other parts of the configuration.

resource "cidaas_webhook" "sample_webhook" {
  auth_type = "APIKEY"
  url       = "https://cidaas.de/webhook-srv/webhook"
  events = [
    "ACCOUNT_MODIFIED"
  ]
  apikey_config = {
    key         = "api-key"
    placeholder = "test-apikey-placeholder"
    placement   = "query"
  }
  totp_config = {
    key         = "totp-key"
    placeholder = "test-totp-placeholder"
    placement   = "query"
  }
  cidaas_auth_config = {
    client_id = "ce90d6ba-9a5a-49b6-9a50-b8db759e9b90"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

* `auth_type` (String) The attribute auth_type is to define how this url is secured from your end.The allowed values are `APIKEY`, `TOTP` and `CIDAAS_OAUTH2`
* `events` (Set of String) A set of webhook-capable event IDs that trigger the webhook. Use the `cidaas_webhook_events` data source (or GET /webhook-srv/eventdescriptions?category=webhook) to discover valid values for your tenant.
* `url` (String) The webhook url that needs to be called when an event occurs.

### Optional

* `apikey_config` (Attributes) Configuration for API key-based authentication. It's a **required** parameter when the auth_type is APIKEY. (see [below for nested schema](#nestedatt--apikey_config))
* `cidaas_auth_config` (Attributes) Configuration for cidaas authentication. It's a **required** parameter when the auth_type is CIDAAS_OAUTH2. (see [below for nested schema](#nestedatt--cidaas_auth_config))
* `disable` (Boolean) Flag to disable the webhook.
* `totp_config` (Attributes) Configuration for TOTP based authentication.  It's a **required** parameter when the auth_type is TOTP. (see [below for nested schema](#nestedatt--totp_config))

### Read-Only

* `created_at` (String) The timestamp when the webhook was created.
* `id` (String) The unique identifier of the webhook resource.
* `updated_at` (String) The timestamp when the webhook was last updated.

<a id="nestedatt--apikey_config"></a>

### Nested Schema for `apikey_config`

Required:

* `key` (String) The API key that will be used to authenticate the webhook request.The key that will be passed in the request header or in query param as configured in the attribute `placement`
* `placeholder` (String) The attribute is the placeholder for the key which need to be passed as a query parameter or in the request header.
* `placement` (String) The placement of the API key in the request (e.g., query).The allowed value are `header` and `query`.

<a id="nestedatt--cidaas_auth_config"></a>

### Nested Schema for `cidaas_auth_config`

Required:

* `client_id` (String) The client ID for cidaas authentication.

<a id="nestedatt--totp_config"></a>

### Nested Schema for `totp_config`

Required:

* `key` (String) The key used for TOTP authentication.
* `placeholder` (String) A placeholder value for the TOTP.
* `placement` (String) The placement of the TOTP in the request.The allowed value are `header` and `query`.

## Import

Import is supported using the following syntax:

```shell
# The import identifier in this command is the ID of the webhook to be imported.

terraform import cidaas_webhook.sample ae90d6ba-9a5a-49b6-9a50-b8db759e9b90
```

# cidaas_consent (Data Source)

The data source `cidaas_consent` returns a list of consents available in your cidaas instance.
You can apply filters using the `filter` block in your Terraform configuration.

## Example Usage

```terraform
data "cidaas_consent" "example" {
  filter {
    name     = "consent_name"
    values   = ["terraform"]
    match_by = "substring"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

* `filter` (Block Set) (see [below for nested schema](#nestedblock--filter))

### Read-Only

* `consent` (Block List) The returned list of consents. (see [below for nested schema](#nestedblock--consent))
* `id` (String) The data source's unique ID.

<a id="nestedblock--filter"></a>

### Nested Schema for `filter`

Required:

* `name` (String) The name of the attribute to filter on.
* `values` (Set of String) The value(s) to be used in the filter.

Optional:

* `match_by` (String) The type of comparison to use for this filter. Allowed values `exact`, `substring` and `regex`

<a id="nestedblock--consent"></a>

### Nested Schema for `consent`

Read-Only:

* `consent_name` (String) The name of the consent.
* `id` (String) The unique identifier of the consent.

## Filterable Fields

* `consent_name`

# cidaas_custom_provider (Data Source)

The data source `cidaas_custom_provider` returns a list of custom providers available in your cidaas instance.
You can apply filters using the `filter` block in your Terraform configuration.

## Example Usage

```terraform
data "cidaas_custom_provider" "example" {
  filter {
    name     = "provider_name"
    values   = ["dev"]
    match_by = "substring"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

* `filter` (Block Set) (see [below for nested schema](#nestedblock--filter))

### Read-Only

* `custom_provider` (Block List) The returned list of custom providers. (see [below for nested schema](#nestedblock--custom_provider))
* `id` (String) The data source's unique ID.

<a id="nestedblock--filter"></a>

### Nested Schema for `filter`

Required:

* `name` (String) The name of the attribute to filter on.
* `values` (Set of String) The value(s) to be used in the filter.

Optional:

* `match_by` (String) The type of comparison to use for this filter. Allowed values `exact`, `substring` and `regex`

<a id="nestedblock--custom_provider"></a>

### Nested Schema for `custom_provider`

Read-Only:

* `domains` (Set of String) The domains of the provider.
* `id` (String) The unique identifier of the custom provider.
* `provider_name` (String) The name of the custom provider.
* `standard_type` (String) Type of standard. `OAUTH2` or `OPENID_CONNECT`.

## Filterable Fields

* `provider_name`
* `standard_type`

# cidaas_group_type (Data Source)

The data source `cidaas_group_type` returns a list of group types available in your cidaas instance.
You can apply filters using the `filter` block in your Terraform configuration.

## Example Usage

```terraform
data "cidaas_group_type" "example" {
  filter {
    name   = "role_mode"
    values = ["roles_required"]
  }
  filter {
    name   = "allowed_roles"
    values = ["DEVELOPER"]
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

* `filter` (Block Set) (see [below for nested schema](#nestedblock--filter))

### Read-Only

* `group_type` (Block List) The returned list of group types. (see [below for nested schema](#nestedblock--group_type))
* `id` (String) The data source's unique ID.

<a id="nestedblock--filter"></a>

### Nested Schema for `filter`

Required:

* `name` (String) The name of the attribute to filter on.
* `values` (Set of String) The value(s) to be used in the filter.

Optional:

* `match_by` (String) The type of comparison to use for this filter. Allowed values `exact`, `substring` and `regex`

<a id="nestedblock--group_type"></a>

### Nested Schema for `group_type`

Read-Only:

* `allowed_roles` (Set of String) List of allowed roles in a group type.
* `description` (String) The description of the group type
* `group_type` (String) The unique identifier of the group type.
* `id` (String) The identifier of the group type.
* `role_mode` (String) Determines the role mode for the user group type.

## Filterable Fields

* `group_type`
* `role_mode`
* `allowed_roles`

# cidaas_registration_field (Data Source)

The data source `cidaas_registration_field` returns a list of registration fields available in your cidaas instance.
You can apply filters using the `filter` block in your Terraform configuration.

## Example Usage

```terraform
data "cidaas_registration_field" "example" {
  filter {
    name   = "field_type"
    values = ["CUSTOM"]
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

* `filter` (Block Set) (see [below for nested schema](#nestedblock--filter))
* `registration_field` (Block List) The returned list of registration fields. (see [below for nested schema](#nestedblock--registration_field))

### Read-Only

* `id` (String) The data source's unique ID.

<a id="nestedblock--filter"></a>

### Nested Schema for `filter`

Required:

* `name` (String) The name of the attribute to filter on.
* `values` (Set of String) The value(s) to be used in the filter.

Optional:

* `match_by` (String) The type of comparison to use for this filter. Allowed values `exact`, `substring` and `regex`

<a id="nestedblock--registration_field"></a>

### Nested Schema for `registration_field`

Required:

* `field_key` (String) The unique name of the registration field.

Read-Only:

* `data_type` (String) The data type of the field.
* `enabled` (Boolean) Flag to identify if a field is enabled.
* `field_type` (String) Specifies whether the field type is `SYSTEM` or `CUSTOM`.
* `id` (String) The unique identifier of the group type.
* `internal` (Boolean) Flag to identify if a field is internal.
* `is_group` (Boolean) Flag to identify if a field is group field.
* `order` (Number) The order of the Field in the UI.
* `parent_group_id` (String) The ID of the parent registration group.
* `read_only` (Boolean) Flag to identify if a field is read only.
* `required` (Boolean) Flag to identify if a field is required in registration.

## Filterable Fields

* `parent_group_id`
* `field_type`
* `data_type`
* `field_key`
* `required`
* `internal`
* `read_only`
* `is_group`
* `enabled`

# cidaas_role (Data Source)

The data source `cidaas_role` returns a list of roles available in your cidaas instance.
You can apply filters using the `filter` block in your Terraform configuration.

## Example Usage

```terraform
data "cidaas_role" "example" {
  filter {
    name   = "name"
    values = ["DEVELOPER"]
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

* `filter` (Block Set) (see [below for nested schema](#nestedblock--filter))

### Read-Only

* `id` (String) The data source's unique ID.
* `role` (Block List) The returned list of roles. (see [below for nested schema](#nestedblock--role))

<a id="nestedblock--filter"></a>

### Nested Schema for `filter`

Required:

* `name` (String) The name of the attribute to filter on.
* `values` (Set of String) The value(s) to be used in the filter.

Optional:

* `match_by` (String) The type of comparison to use for this filter. Allowed values `exact`, `substring` and `regex`

<a id="nestedblock--role"></a>

### Nested Schema for `role`

Read-Only:

* `description` (String) The `description` of the role
* `name` (String) The name of the role.
* `role` (String) The unique identifier of the role.

## Filterable Fields

* `role`
* `name`

# cidaas_scope_group (Data Source)

The data source `cidaas_scope_group` returns a list of scope groups available in your cidaas instance.
You can apply filters using the `filter` block in your Terraform configuration.

## Example Usage

```terraform
data "cidaas_scope_group" "example" {
  filter {
    name     = "group_name"
    values   = ["terraform"]
    match_by = "substring"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

* `filter` (Block Set) (see [below for nested schema](#nestedblock--filter))

### Read-Only

* `id` (String) The data source's unique ID.
* `scope_group` (Block List) The returned list of scope groups (see [below for nested schema](#nestedblock--scope_group))

<a id="nestedblock--filter"></a>

### Nested Schema for `filter`

Required:

* `name` (String) The name of the attribute to filter on.
* `values` (Set of String) The value(s) to be used in the filter.

Optional:

* `match_by` (String) The type of comparison to use for this filter. Allowed values `exact`, `substring` and `regex`

<a id="nestedblock--scope_group"></a>

### Nested Schema for `scope_group`

Read-Only:

* `description` (String) The `description` attribute provides details about the scope of the group, explaining its purpose.
* `group_name` (String) The name of the group.
* `id` (String) The ID of th resource.

## Filterable Fields

* `group_name`

# cidaas_scope (Data Source)

The data source `cidaas_scope` returns a list of scopes available in your cidaas instance.
You can apply filters using the `filter` block in your Terraform configuration.

## Example Usage

```terraform
data "cidaas_scope" "example" {
  filter {
    name   = "security_level"
    values = ["CONFIDENTIAL"]
  }
  filter {
    name     = "scope_key"
    values   = ["cidaas"]
    match_by = "substring"
  }
  filter {
    name   = "required_user_consent"
    values = [false]
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

* `filter` (Block Set) (see [below for nested schema](#nestedblock--filter))
* `scope` (Block List) The returned list of scopes. (see [below for nested schema](#nestedblock--scope))

### Read-Only

* `id` (String) The data source's unique ID.

<a id="nestedblock--filter"></a>

### Nested Schema for `filter`

Required:

* `name` (String) The name of the attribute to filter on.
* `values` (Set of String) The value(s) to be used in the filter.

Optional:

* `match_by` (String) The type of comparison to use for this filter. Allowed values `exact`, `substring` and `regex`

<a id="nestedblock--scope"></a>

### Nested Schema for `scope`

Optional:

* `localized_descriptions` (Attributes List) (see [below for nested schema](#nestedatt--scope--localized_descriptions))

Read-Only:

* `group_name` (Set of String) List of scope_groups associated with the scope.
* `id` (String) The ID of the scope.
* `required_user_consent` (Boolean) Indicates whether user consent is required for the scope.
* `scope_key` (String) Unique identifier(name) for the scope.
* `scope_owner` (String) The owner of the scope. e.g. `ADMIN`.
* `security_level` (String) The security level of the scope, `PUBLIC` or `CONFIDENTIAL`.

<a id="nestedatt--scope--localized_descriptions"></a>

### Nested Schema for `scope.localized_descriptions`

Read-Only:

* `description` (String) The description of the scope in the configured locale.
* `locale` (String) The locale for the scope, e.g., `en-US`.
* `title` (String) The title of the scope in the configured locale.

## Filterable Fields

* `scope_key`
* `security_level`
* `group_name`
* `required_user_consent`

# cidaas_social_provider (Data Source)

The data source `cidaas_social_provider` returns a list of social providers available in your cidaas instance.
You can apply filters using the `filter` block in your Terraform configuration.

## Example Usage

```terraform
data "cidaas_social_provider" "example" {
  filter {
    name   = "enabled_for_admin_portal"
    values = ["true"]
  }
  filter {
    name   = "enabled"
    values = ["true"]
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

* `filter` (Block Set) (see [below for nested schema](#nestedblock--filter))

### Read-Only

* `id` (String) The data source's unique ID.
* `social_provider` (Block List) The returned list of social providers. (see [below for nested schema](#nestedblock--social_provider))

<a id="nestedblock--filter"></a>

### Nested Schema for `filter`

Required:

* `name` (String) The name of the attribute to filter on.
* `values` (Set of String) The value(s) to be used in the filter.

Optional:

* `match_by` (String) The type of comparison to use for this filter. Allowed values `exact`, `substring` and `regex`

<a id="nestedblock--social_provider"></a>

### Nested Schema for `social_provider`

Read-Only:

* `client_id` (String) The client ID of the social provider.
* `client_secret` (String, Sensitive) The client secret of the social provider.
* `enabled` (Boolean) A flag to identify if a provider is enabled.
* `enabled_for_admin_portal` (Boolean) A flag to identify if a social provider is enabled for the admin portal.
* `id` (String) The unique identifier of the social provider
* `name` (String) The name of the social provider configuration.
* `provider_name` (String) The name of the social provider e.g; `google`, `facebook`, `linkedin` etc.
* `scopes` (Set of String) A list of scopes of the social provider.

## Filterable Fields

* `name`
* `provider_name`
* `enabled`
* `enabled_for_admin_portal`

# cidaas_system_template_option (Data Source)

The data source `cidaas_system_template_option` returns a list of system templates optionsa that can be
configured to create a system template in your cidaas instance.
You can apply filters using the `filter` block in your Terraform configuration.

## Example Usage

```terraform
data "cidaas_system_template_option" "example" {
  filter {
    name   = "template_key"
    values = ["UN_REGISTER_USER_ALERT"]
  }
  filter {
    name   = "role"
    values = ["USER_CREATE"]
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

* `filter` (Block Set) (see [below for nested schema](#nestedblock--filter))
* `system_template_option` (Block List) The returned list of system template options. (see [below for nested schema](#nestedblock--system_template_option))

### Read-Only

* `id` (String) The data source's unique ID.

<a id="nestedblock--filter"></a>

### Nested Schema for `filter`

Required:

* `name` (String) The name of the attribute to filter on.
* `values` (Set of String) The value(s) to be used in the filter.

Optional:

* `match_by` (String) The type of comparison to use for this filter. Allowed values `exact`, `substring` and `regex`

<a id="nestedblock--system_template_option"></a>

### Nested Schema for `system_template_option`

Required:

* `enabled` (Boolean) The flag to identify if a system template is enabled.
* `template_key` (String) The key of the template.

Optional:

* `template_types` (Attributes List) (see [below for nested schema](#nestedatt--system_template_option--template_types))

<a id="nestedatt--system_template_option--template_types"></a>

### Nested Schema for `system_template_option.template_types`

Optional:

* `processing_types` (Attributes List) (see [below for nested schema](#nestedatt--system_template_option--template_types--processing_types))

Read-Only:

* `template_type` (String) The type of the template. e.g. `EMAIL`

<a id="nestedatt--system_template_option--template_types--processing_types"></a>

### Nested Schema for `system_template_option.template_types.processing_types`

Optional:

* `verification_types` (Attributes List) (see [below for nested schema](#nestedatt--system_template_option--template_types--processing_types--verification_types))

Read-Only:

* `processing_type` (String) The processing type of the template. e.g. `LINK` or `CODE`
* `supported_tags` (Attributes) (see [below for nested schema](#nestedatt--system_template_option--template_types--processing_types--supported_tags))

<a id="nestedatt--system_template_option--template_types--processing_types--verification_types"></a>

### Nested Schema for `system_template_option.template_types.processing_types.verification_types`

Read-Only:

* `usage_types` (Set of String) The usage type of the template. e.g. `MULTIFACTOR_AUTHENTICATION`
* `verification_type` (String) The verification type of the template. e.g. `EMAIL`

<a id="nestedatt--system_template_option--template_types--processing_types--supported_tags"></a>

### Nested Schema for `system_template_option.template_types.processing_types.supported_tags`

Read-Only:

* `optional` (Set of String) This lists provides the optional tags supported in a template content.
* `required` (Set of String) The required tags in a template. While creating a templates the required tags must be part of the content.

## Filterable Fields

* `template_key`
* `enabled`
