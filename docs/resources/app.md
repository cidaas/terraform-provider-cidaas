---
page_title: "cidaas_app Resource - cidaas"
subcategory: ""
description: |-
  The App resource allows creation and management of clients in Cidaas system. When creating a client with a custom client_id and client_secret you can include the configuration in the resource. If not provided, Cidaas will generate a set for you. client_secret is sensitive data. Refer to the article Terraform Sensitive Variables https://developer.hashicorp.com/terraform/tutorials/configuration-language/sensitive-variables to properly handle sensitive information.
  Ensure that the below scopes are assigned to the client with the specified client_id:
  cidaas:apps_readcidaas:apps_writecidaas:apps_delete
  -> Note: Write-Only argument client_secret_wo is available to use in place of client_secret. Write-only arguments are supported in HashiCorp Terraform 1.11.0 and later. Learn more https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments.
---

# cidaas_app (Resource)

The App resource allows creation and management of clients in Cidaas system. When creating a client with a custom `client_id` and `client_secret` you can include the configuration in the resource. If not provided, Cidaas will generate a set for you. `client_secret` is sensitive data. Refer to the article [Terraform Sensitive Variables](https://developer.hashicorp.com/terraform/tutorials/configuration-language/sensitive-variables) to properly handle sensitive information.

 Ensure that the below scopes are assigned to the client with the specified `client_id`:
- cidaas:apps_read
- cidaas:apps_write
- cidaas:apps_delete

-> **Note:** Write-Only argument `client_secret_wo` is available to use in place of `client_secret`. Write-only arguments are supported in HashiCorp Terraform 1.11.0 and later. [Learn more](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments).

From version 3.3.0, the attribute `common_configs` is not supported anymore. Instead, we encourage you to use the custom module **terraform-cidaas-app**.
The module provides a variable with the same name `common_configs` which
supports all the attributes in the resource app except `client_name`. With this module you can avoid the repeated configuration and assign the common properties
of multiple apps to a common variable and inherit the properties.

Link to the custom module https://github.com/cidaas/terraform-cidaas-app

##### Module usage:

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
  accent_color                    = "#ef4923"                        // Default: #ef4923
  primary_color                   = "#ef4923"                        // Default: #f7941d
  media_type                      = "IMAGE"                          // Default: IMAGE
  allow_login_with                = ["EMAIL", "MOBILE", "USER_NAME"] // Default: ["EMAIL", "MOBILE", "USER_NAME"]
  redirect_uris                   = ["https://cidaas.com"]
  allowed_logout_urls             = ["https://cidaas.com"]
  enable_deduplication            = true      // Default: false
  auto_login_after_register       = true      // Default: false
  enable_passwordless_auth        = false     // Default: true
  register_with_login_information = false     // Default: false
  hosted_page_group               = "default" // Default: default
  company_name                    = "Widas ID GmbH"
  company_address                 = "01"
  company_website                 = "https://cidaas.com"
  allowed_scopes                  = ["openid", "cidaas:register", "profile"]
  client_display_name             = "Display Name of the app" // unique
  content_align                   = "CENTER"                  // Default: CENTER
  post_logout_redirect_uris       = ["https://cidaas.com"]
  logo_align                      = "CENTER" // Default: CENTER
  allow_disposable_email          = false    // Default: false
  validate_phone_number           = false    // Default: false
  additional_access_token_payload = ["sample_payload"]
  required_fields                 = ["email"]
  mobile_settings = {
    team_id      = "sample-team-id"
    bundle_id    = "sample-bundle-id"
    package_name = "sample-package-name"
    key_hash     = "sample-key-hash"
  }
  // For custom client credentials set client_id and either client_secret or client_secret_wo.
  // Omit both to have cidaas auto-generate client_secret (only with the non-Write-Only attribute).
  //
  // variable "app_client_secret" {
  //   type      = string
  //   sensitive = true
  //   ephemeral = true
  // }
  # client_id                       = ""
  # client_secret                   = "" # stored in the state file
  # client_secret_wo                = var.app_client_secret # Write-Only (Terraform 1.11+)
  # client_secret_wo_version        = "1"
  policy_uri                        = "https://cidaas.com"
  tos_uri                           = "https://cidaas.com"
  imprint_uri                       = "https://cidaas.com"
  contacts                          = ["support@cidas.de"]
  token_endpoint_auth_method        = "client_secret_post" // Default: client_secret_post
  token_endpoint_auth_signing_alg   = "RS256"              // Default: RS256
  default_acr_values                = ["default"]
  web_message_uris                  = ["https://cidaas.com"]
  allowed_fields                    = ["email"]
  smart_mfa                         = false // Default: false
  captcha_ref                       = "sample-captcha-ref"
  captcha_refs                      = ["sample"]
  consent_refs                      = ["sample"]
  communication_medium_verification = "email_verification_required_on_usage"
  enable_bot_detection              = false // Default: false
  allow_guest_login_groups = [{
    group_id      = "developer101"
    roles         = ["developer", "qa", "admin"]
    default_roles = ["developer"]
  }]
  is_login_success_page_enabled    = false // Default: false
  is_register_success_page_enabled = false // Default: false
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

> **NOTE**: [Write-only arguments](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments) are supported in Terraform 1.11 and later.

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
- `client_secret` (String, Sensitive) The client secret of the app. If not provided and `client_secret_wo` is not set, cidaas will generate one for you and the state will be updated with the same. Note that this will be stored in the state file.
- `client_secret_wo` (String, Sensitive, [Write-only](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments)) Write-Only equivalent of `client_secret`. The value is sent to cidaas on create and update but is not stored in plan or state. Cannot be set together with `client_secret`. Must be set together with `client_secret_wo_version`. When set, the auto-generation of `client_secret` is disabled and the value must be supplied by the user. Write-only arguments are supported in HashiCorp Terraform 1.11.0 and later.
- `client_secret_wo_version` (String) Used together with `client_secret_wo` to trigger an update. Increment this value when an update to `client_secret_wo` is required.
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
- `disable_insecure_pkce_method` (Boolean) When `true`, rejects PKCE authorization requests that use the insecure `plain` code challenge method (RFC 7636), including requests that send `code_challenge` without `code_challenge_method` (which defaults to `plain`). Only `S256` is accepted (AUTH10048). Enforced during `/authz-srv/authz` and PAR. When omitted, `plain` remains allowed for backward compatibility.
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
- `is_group_login_selection_enabled` (Boolean)
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
- `oauth_standard` (String) Specifies the OAuth standard version to use. Allowed values: 'OAuth2.1', 'OAuth2.0'
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
- `registration_client_uri` (String)
- `request_object_encryption_alg` (String)
- `request_object_encryption_enc` (String)
- `request_object_signing_alg` (String)
- `request_uris` (Set of String)
- `require_auth_time` (Boolean) Boolean flag to specify whether the auth_time claim is REQUIRED in a id token.
- `require_pkce` (Boolean) When `true`, the application requires PKCE for authorization requests: clients must send `code_challenge` (and typically `code_challenge_method`) on `/authz-srv/authz` and PAR. Requests without `code_challenge` are rejected (AUTH10063). Recommended for public clients (SPA, mobile apps). When omitted, PKCE is optional.
- `required_fields` (Set of String) The required fields while registering to the client.
- `response_types` (Set of String) The response types of the client. The default value is set to `['code','token', 'id_token']`
- `role` (String)
- `saml_providers` (Attributes List) A list of SAML identity providers that users can authenticate with. (see [below for nested schema](#nestedatt--saml_providers))
- `sector_identifier_uri` (String)
- `smart_mfa` (Boolean)
- `social_providers` (Attributes List) A list of social identity providers that users can authenticate with. Examples: Google, Facebook etc... (see [below for nested schema](#nestedatt--social_providers))
- `sub` (String)
- `subject_type` (String)
- `suggest_mfa` (Set of String)
- `suggest_verification_methods` (Attributes) Configuration for verification methods. (see [below for nested schema](#nestedatt--suggest_verification_methods))
- `template_group_id` (String) The id of the template group to be configured for commenication. Default is set to the system default group.
- `token_endpoint_auth_method` (String)
- `token_endpoint_auth_signing_alg` (String)
- `token_lifetime_in_seconds` (Number) The lifetime of the token in seconds. Default is 86400 seconds (24 hours).
- `tos_uri` (String) The URL to the TOS of a client.
- `user_consent` (Boolean) Specifies whether user consent is required or not. Default is `false`
- `userinfo_encrypted_response_alg` (String)
- `userinfo_encrypted_response_enc` (String)
- `userinfo_signed_response_alg` (String)
- `validate_phone_number` (Boolean) if enabled, phone number is validaed. Default is set to `false` while creating an app.
- `video_url` (String) The URL to the video of the client.
- `web_message_uris` (Set of String) A list of URLs for web messages used.
- `webfinger` (String)

### Read-Only

- `id` (String) The ID of the resource.

<a id="nestedatt--ad_providers"></a>
### Nested Schema for `ad_providers`

Optional:

- `display_name` (String)
- `domains` (Set of String)
- `is_provider_visible` (Boolean)
- `logo_url` (String)
- `provider_name` (String)
- `type` (String)


<a id="nestedatt--allow_guest_login_groups"></a>
### Nested Schema for `allow_guest_login_groups`

Optional:

- `default_roles` (Set of String)
- `group_id` (String)
- `group_type` (String)
- `roles` (Set of String)


<a id="nestedatt--allowed_groups"></a>
### Nested Schema for `allowed_groups`

Optional:

- `default_roles` (Set of String)
- `group_id` (String)
- `group_type` (String)
- `roles` (Set of String)


<a id="nestedatt--custom_providers"></a>
### Nested Schema for `custom_providers`

Optional:

- `display_name` (String)
- `domains` (Set of String)
- `is_provider_visible` (Boolean)
- `logo_url` (String)
- `provider_name` (String)
- `type` (String)


<a id="nestedatt--group_role_restriction"></a>
### Nested Schema for `group_role_restriction`

Required:

- `filters` (Attributes List) An array of group role filters. (see [below for nested schema](#nestedatt--group_role_restriction--filters))
- `match_condition` (String) The match condition for the role restriction

Optional:

- `hints` (Set of String) Optional response hints for group verification: groupIds, rolesOfGroup, or allowedGroups.

<a id="nestedatt--group_role_restriction--filters"></a>
### Nested Schema for `group_role_restriction.filters`

Optional:

- `group_id` (String) The unique ID of the user group.
- `group_type` (String) The unique ID of the user group.
- `role_filter` (Attributes) A filter for roles within the group. (see [below for nested schema](#nestedatt--group_role_restriction--filters--role_filter))

<a id="nestedatt--group_role_restriction--filters--role_filter"></a>
### Nested Schema for `group_role_restriction.filters.role_filter`

Optional:

- `match_condition` (String) The match condition for the roles (AND or OR).
- `roles` (Set of String) An array of role names.




<a id="nestedatt--group_selection"></a>
### Nested Schema for `group_selection`

Optional:

- `always_show_group_selection` (Boolean)
- `selectable_group_types` (Set of String)
- `selectable_groups` (Set of String)


<a id="nestedatt--login_spi"></a>
### Nested Schema for `login_spi`

Optional:

- `oauth_client_id` (String)
- `spi_url` (String)


<a id="nestedatt--mfa"></a>
### Nested Schema for `mfa`

Optional:

- `allowed_methods` (Set of String) Optional set of allowed MFA methods.
- `setting` (String) Specifies the Multi-Factor Authentication (MFA) setting. Allowed values are 'OFF', 'ALWAYS', 'SMART', 'TIME_BASED' and 'SMART_PLUS_TIME_BASED'.
- `time_interval_in_seconds` (Number) Optional time interval in seconds for time-based Multi-Factor Authentication.


<a id="nestedatt--mobile_settings"></a>
### Nested Schema for `mobile_settings`

Optional:

- `bundle_id` (String)
- `key_hash` (String)
- `package_name` (String)
- `team_id` (String)


<a id="nestedatt--operations_allowed_groups"></a>
### Nested Schema for `operations_allowed_groups`

Optional:

- `default_roles` (Set of String)
- `group_id` (String)
- `group_type` (String)
- `roles` (Set of String)


<a id="nestedatt--saml_providers"></a>
### Nested Schema for `saml_providers`

Optional:

- `display_name` (String)
- `domains` (Set of String)
- `is_provider_visible` (Boolean)
- `logo_url` (String)
- `provider_name` (String)
- `type` (String)


<a id="nestedatt--social_providers"></a>
### Nested Schema for `social_providers`

Optional:

- `provider_name` (String)
- `social_id` (String)


<a id="nestedatt--suggest_verification_methods"></a>
### Nested Schema for `suggest_verification_methods`

Optional:

- `mandatory_config` (Attributes) Configuration for mandatory verification methods. (see [below for nested schema](#nestedatt--suggest_verification_methods--mandatory_config))
- `optional_config` (Attributes) Configuration for optional verification methods (see [below for nested schema](#nestedatt--suggest_verification_methods--optional_config))
- `skip_duration_in_days` (Number) The number of days for which the verification methods can be skipped (default is 7 days).

<a id="nestedatt--suggest_verification_methods--mandatory_config"></a>
### Nested Schema for `suggest_verification_methods.mandatory_config`

Optional:

- `methods` (Set of String) List of mandatory verification methods.
- `range` (String) The range type for mandatory methods. Allowed value is one of ALLOF or ONEOF.
- `skip_until` (String) The date and time until which the mandatory methods can be skipped.


<a id="nestedatt--suggest_verification_methods--optional_config"></a>
### Nested Schema for `suggest_verification_methods.optional_config`

Optional:

- `methods` (Set of String) List of optional verification methods.

## Import

Import is supported using the following syntax:

```shell
# The import identifier in this command is the client_id of the app to be imported.

terraform import cidaas_app.sample client_id
```