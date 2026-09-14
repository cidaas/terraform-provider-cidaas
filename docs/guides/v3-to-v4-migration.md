---
page_title: "Cidaas v3 to v4 Migration Guide"
---

# Cidaas v3 to v4 Migration Guide

This guide details how to migrate your Terraform configurations from **Cidaas v3** to **Cidaas v4 (Trustdesk)**.

## Overview

The `cidaas` provider supports both **v3** and **v4** environments in a single provider binary. You select the platform version in your provider configuration:

```hcl
provider "cidaas" {
  base_url       = "https://your-tenant.dev.cidaas.eu"
  cidaas_version = "4.x" # or "3.x"
}
```

## Strategy: Standardized v4 Resources vs. Legacy v3 Compatibility

The provider provides two operational pathways when managing resources in a v4 environment:

1. **Legacy & Deprecated Resources**: `cidaas_app` is deprecated in v4 in favor of `cidaas_app_configuration`. `cidaas_social_provider` and `cidaas_custom_provider` are removed in favor of `cidaas_federation_provider`.
2. **Native v4 Standardized Resources**: Native v4 resources (`cidaas_app_configuration`, `cidaas_hosted_page`, `cidaas_federation_provider`, `cidaas_user_setup`, `cidaas_verification_options`) offer enhanced fine-grained microservice capabilities on `/apps-srv`, `/hostedpages-srv`, `/federation/providers`, `/usersetup-srv`, and `/verification-srv`.

## Resource Mapping Matrix

| Category                  | Old Resource (v3 / Legacy)                              | New Resource (v4 / Standardized)                                                             | Migration Notes                                                                                                                                                                                    |
| :------------------------ | :------------------------------------------------------ | :------------------------------------------------------------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Applications**          | `cidaas_app`                                            | `cidaas_app_configuration`                                                                   | Deprecated in v4. Migrate to `cidaas_app_configuration` for native `/app-srv/apps` app management.                                                                                                 |
| **Hosted Pages**          | `cidaas_hosted_page`                                    | `cidaas_hosted_page`                                                                         | Still exists on v3 and v4. Reserved system groups (`default`, `admin`) are protected against API deletion.                                                                                         |
| **Hosted page layout, theme & translations** | N/A                                                     | `cidaas_hosted_page_layout`<br>`cidaas_theme`<br>`cidaas_translations`                       | Additional v4-only resources. They do not replace `cidaas_hosted_page`. Apps reference a layout via `hosted_pages_layout_id`.                                                                      |
| **Identity Providers**    | `cidaas_social_provider`<br>`cidaas_custom_provider`    | `cidaas_federation_provider`                                                                 | Removed in v4. Standardize exclusively on `cidaas_federation_provider` (`OAUTH2`, `OPENID_CONNECT`, `SAML`, `LDAP`).                                                                               |
| **Scopes & Roles**        | `cidaas_scope`<br>`cidaas_scope_group`<br>`cidaas_role` | `cidaas_scope`<br>`cidaas_scope_group`<br>`cidaas_role`                                      | Standardized across v3 and v4.                                                                                                                                                                     |
| **Groups**                | `cidaas_group_type`<br>`cidaas_user_groups`             | `cidaas_group_type`<br>`cidaas_user_groups`                                                  | Unchanged. Both resources still exist and are shared across v3 and v4.                                                                                                                             |
| **Group selection & filters** | Nested on `cidaas_app` (`group_selection`, `group_role_restriction`) | `cidaas_group_selection`<br>`cidaas_group_verification_filter`                               | Additional v4-only resources. They do not replace `cidaas_group_type` / `cidaas_user_groups`.                                                                                                       |
| **User Setup & MFA**      | N/A                                                     | `cidaas_user_setup`<br>`cidaas_verification_options`<br>`cidaas_suggest_verification_method` | Native v4 user setup & MFA option features.                                                                                                                                                        |

---

## Application Migration (`cidaas_app` ➔ `cidaas_app_configuration`)

Below is the complete, zero-omissions mapping table showing how all fields from legacy `cidaas_app` map to `cidaas_app_configuration` in v4.

| Legacy Field (`cidaas_app`)                                                                            | Standardized v4 Field (`cidaas_app_configuration`)   | Attribute Type     | Description & Migration Notes                                                                                                            |
| :----------------------------------------------------------------------------------------------------- | :--------------------------------------------------- | :----------------- | :--------------------------------------------------------------------------------------------------------------------------------------- |
| `client_id`                                                                                            | `client_id`                                          | String             | Unique OAuth client identifier. Auto-generated on create if omitted.                                                                     |
| `client_name`                                                                                          | `client_name`                                        | String             | Name of the client application.                                                                                                          |
| `client_display_name`                                                                                  | `client_display_name`                                | String             | Human-readable display name.                                                                                                             |
| `client_type`                                                                                          | `client_type`                                        | String             | Application client type (`NON_INTERACTIVE`, `SINGLE_PAGE`, `REGULAR_WEB`, `NATIVE`, `MOBILE`, `DESKTOP`, `THIRD_PARTY`, `DEVICE`, etc.). |
| `description`                                                                                          | `description`                                        | String             | Description of the application.                                                                                                          |
| `enabled`                                                                                              | `enabled`                                            | Bool               | Enables or disables the client application (default: `true`).                                                                            |
| `company_name`                                                                                         | `ownership_details.company_name`                     | String             | Moved into `ownership_details` block.                                                                                                    |
| `company_address`                                                                                      | `ownership_details.company_address`                  | String             | Moved into `ownership_details` block.                                                                                                    |
| `company_website`                                                                                      | `ownership_details.company_website`                  | String             | Moved into `ownership_details` block.                                                                                                    |
| `allowed_scopes`                                                                                       | `scopes.allowed_scopes`                              | List(String)       | Moved into `scopes` block.                                                                                                               |
| `default_scopes`                                                                                       | `scopes.default_scopes`                              | List(String)       | Moved into `scopes` block.                                                                                                               |
| `redirect_uris`                                                                                        | `redirect_uris.redirect_uris`                        | List(String)       | Moved into `redirect_uris` block.                                                                                                        |
| `allowed_logout_urls`                                                                                  | `redirect_uris.allowed_logout_urls`                  | List(String)       | Moved into `redirect_uris` block.                                                                                                        |
| `post_logout_redirect_uris`                                                                            | `redirect_uris.post_logout_redirect_uris`            | List(String)       | Moved into `redirect_uris` block.                                                                                                        |
| `grant_types`                                                                                          | `grant_types`                                        | List(String)       | Allowed OAuth 2.0 grant types (`authorization_code`, `client_credentials`, `refresh_token`, etc.).                                       |
| `response_types`                                                                                       | `response_types`                                     | List(String)       | Allowed OAuth 2.0 response types (`code`, `token`, `id_token`).                                                                          |
| `token_lifetime_in_seconds`                                                                            | `token_lifetimes.access_token_lifetime_in_seconds`   | Int64              | Moved into `token_lifetimes` block. Access token validity in seconds.                                                                    |
| `id_token_lifetime_in_seconds`                                                                         | `token_lifetimes.id_token_lifetime_in_seconds`       | Int64              | Moved into `token_lifetimes` block. ID token validity in seconds.                                                                        |
| `refresh_token_lifetime_in_seconds`                                                                    | `token_lifetimes.refresh_token_lifetime_in_seconds`  | Int64              | Moved into `token_lifetimes` block. Refresh token validity in seconds.                                                                   |
| `require_pkce`                                                                                         | `require_pkce`                                       | Bool               | Requires PKCE for authorization code flow.                                                                                               |
| `disable_insecure_pkce_method`                                                                         | `disable_insecure_pkce_method`                       | Bool               | Disables insecure `plain` PKCE challenge method (requires `S256`).                                                                       |
| `token_endpoint_auth_method`                                                                           | `client_auth_config.token_endpoint_auth_method`      | String             | Moved into `client_auth_config` block (`none`, `client_secret_basic`, `private_key_jwt`, `tls_client_auth`).                             |
| `social_providers`, `custom_providers`, `saml_providers`, `ad_providers`                               | `identity_providers`                                 | List(Object)       | Standardized under `identity_providers` list in v4, linking `cidaas_federation_provider` instances.                                      |
| `hosted_page_group`                                                                                    | `hosted_pages_layout_id`                             | String (Extdep ID) | References the `cidaas_hosted_page_layout` that includes `cidaas_hosted_page` ID.                                                                    |
| `auto_login_after_register`, `enable_deduplication`, `allow_disposable_email`, `validate_phone_number` | `user_setup_id`                                      | String (Extdep ID) | In v4, user registration & deduplication policies are managed via `cidaas_user_setup` resource.                                          |
| `mfa`, `suggest_verification_methods`, `smart_mfa`, `allowed_mfa`                                      | `authentication_setup.verification_options_id`       | String (Extdep ID) | In v4, MFA settings and policies are managed via `cidaas_verification_options` resource.                                                 |
| `group_selection`                                                                                      | `authentication_setup.group_selection_id`            | String (Extdep ID) | In v4, group selection policies are managed via `cidaas_group_selection` resource.                                                       |
| `group_role_restriction`                                                                               | `authentication_setup.group_verification_request_id` | String (Extdep ID) | In v4, group verification filters are managed via `cidaas_group_verification_filter` resource.                                           |

### Application Before / After HCL Comparison

#### Legacy `cidaas_app` (v3 Monolithic)

```hcl
resource "cidaas_app" "legacy_full_app" {
  client_name                      = "sample_app_v3_full"
  client_display_name              = "Sample v3 Monolithic App"
  client_type                      = "SINGLE_PAGE"
  company_name                     = "Widas Concepts GmbH"
  company_address                  = "Maybachstraße 2, 71229 Leonberg"
  company_website                  = "https://widas.de"
  allowed_scopes                   = ["openid", "profile", "email", "cidaas:user_read"]
  default_scopes                   = ["openid", "profile"]
  grant_types                      = ["authorization_code", "refresh_token"]
  response_types                   = ["code"]
  redirect_uris                    = ["https://example.com/callback"]
  allowed_logout_urls              = ["https://example.com/logout"]
  post_logout_redirect_uris        = ["https://example.com/post-logout"]
  token_lifetime_in_seconds        = 86400
  id_token_lifetime_in_seconds     = 86400
  refresh_token_lifetime_in_seconds = 15780000
  require_pkce                     = true
  disable_insecure_pkce_method    = true
  token_endpoint_auth_method       = "client_secret_basic"
  hosted_page_group                = "default"
  auto_login_after_register        = false
  enable_deduplication             = true
}
```

#### Standardized `cidaas_app_configuration` (v4 Decoupled Microservice)

```hcl
resource "cidaas_app_configuration" "v4_full_app" {
  client_name         = "sample_app_v3_full"
  client_display_name = "Sample v4 Decoupled App"
  client_type         = "SINGLE_PAGE"
  enabled             = true

  grant_types    = ["authorization_code", "refresh_token"]
  response_types = ["code"]

  require_pkce                 = true
  disable_insecure_pkce_method = true

  ownership_details = {
    company_name    = "Widas Concepts GmbH"
    company_address = "Maybachstraße 2, 71229 Leonberg"
    company_website = "https://widas.de"
  }

  scopes = {
    allowed_scopes = ["openid", "profile", "email", "cidaas:user_read"]
    default_scopes = ["openid", "profile"]
  }

  redirect_uris = {
    redirect_uris             = ["https://example.com/callback"]
    allowed_logout_urls       = ["https://example.com/logout"]
    post_logout_redirect_uris = ["https://example.com/post-logout"]
  }

  token_lifetimes = {
    access_token_lifetime_in_seconds  = 86400
    id_token_lifetime_in_seconds      = 86400
    refresh_token_lifetime_in_seconds = 15780000
  }

  client_auth_config = {
    token_endpoint_auth_method = "client_secret_basic"
  }

  # Decoupled microservice resource ID references (Extdeps)
  hosted_pages_layout_id = cidaas_hosted_page_layout.example.id
  user_setup_id          = cidaas_user_setup.sample.id

  authentication_setup = {
    verification_options_id       = cidaas_verification_options.web.id
    group_selection_id            = cidaas_group_selection.sample.id
    group_verification_request_id = cidaas_group_verification_filter.sample.id
  }
}
```

---

## Identity Provider Migration (`cidaas_social_provider` / `cidaas_custom_provider` ➔ `cidaas_federation_provider`)

In Cidaas v4 (Trustdesk), legacy `cidaas_social_provider` and `cidaas_custom_provider` resources are **removed**. All identity providers (Social, Custom OAuth2, OpenID Connect, SAML, LDAP) are standardized under **`cidaas_federation_provider`** backed by `identityprovider-srv` (`/federation/providers`).

### Exhaustive Identity Provider Attribute Mapping

| Legacy Field (`social_provider` / `custom_provider`) | Standardized v4 Field (`cidaas_federation_provider`) | Attribute Type     | Description & Migration Guidance                                                                                                           |
| :--------------------------------------------------- | :--------------------------------------------------- | :----------------- | :----------------------------------------------------------------------------------------------------------------------------------------- |
| `provider_name`                                      | `provider_name`                                      | String             | Unique provider identifier (e.g. `google`, `facebook`, `custom_oidc`). Requires replace.                                                   |
| `display_name`                                       | `display_name`                                       | String             | Human-readable display name of the provider.                                                                                               |
| N/A                                                  | `standard_type`                                      | String             | Standard protocol type (`"OAUTH2"`, `"OPENID_CONNECT"`, `"SAML"`, `"LDAP"`). Set to `"OPENID_CONNECT"` or `"OAUTH2"` for social providers. |
| `client_id`                                          | `client_id`                                          | String             | OAuth/OIDC client ID issued by external IdP.                                                                                               |
| `client_secret`                                      | `client_secret`                                      | String (Sensitive) | Client secret issued by external IdP.                                                                                                      |
| `authorization_endpoint`                             | `authorization_endpoint`                             | String             | Authorization endpoint URL of the external IdP.                                                                                            |
| `token_endpoint`                                     | `token_endpoint`                                     | String             | Token endpoint URL of the external IdP.                                                                                                    |
| `userinfo_endpoint`                                  | `userinfo_endpoint`                                  | String             | Userinfo endpoint URL of the external IdP.                                                                                                 |
| `logo_url`                                           | `logo_url`                                           | String             | Icon or logo URL of the provider.                                                                                                          |
| `domains`                                            | `domains`                                            | List(String)       | Allowed email domains for domain-based routing.                                                                                            |
| N/A                                                  | `owner`                                              | String             | Owner identifier (defaults to `"client"` for Admin UI compatibility).                                                                      |

### Identity Provider Before / After HCL Comparison

#### Legacy v3 (`cidaas_social_provider` & `cidaas_custom_provider`)

```hcl
# Legacy Social Provider (v3)
resource "cidaas_social_provider" "google" {
  provider_name = "google"
  client_id     = "google-client-id-123"
  client_secret = "google-client-secret-456"
}

# Legacy Custom Provider (v3)
resource "cidaas_custom_provider" "custom_oidc" {
  provider_name          = "custom_idp"
  display_name           = "Corporate OIDC IdP"
  client_id              = "custom-client-id"
  client_secret          = "custom-client-secret"
  authorization_endpoint = "https://idp.example.com/oauth/authorize"
  token_endpoint         = "https://idp.example.com/oauth/token"
  userinfo_endpoint      = "https://idp.example.com/oauth/userinfo"
}
```

#### Standardized v4 (`cidaas_federation_provider`)

```hcl
# Standardized Social Provider (v4)
resource "cidaas_federation_provider" "google" {
  provider_name = "google"
  display_name  = "Google Login"
  standard_type = "OPENID_CONNECT"

  client_id     = "google-client-id-123"
  client_secret = "google-client-secret-456"

  authorization_endpoint = "https://accounts.google.com/o/oauth2/v2/auth"
  token_endpoint         = "https://oauth2.googleapis.com/token"
  userinfo_endpoint      = "https://openidconnect.googleapis.com/v1/userinfo"
  logo_url               = "https://cdn.example.com/google-logo.svg"
  owner                  = "client"
}

# Standardized Custom Provider (v4)
resource "cidaas_federation_provider" "custom_oidc" {
  provider_name = "custom_idp"
  display_name  = "Corporate OIDC IdP"
  standard_type = "OPENID_CONNECT"

  client_id     = "custom-client-id"
  client_secret = "custom-client-secret"

  authorization_endpoint = "https://idp.example.com/oauth/authorize"
  token_endpoint         = "https://idp.example.com/oauth/token"
  userinfo_endpoint      = "https://idp.example.com/oauth/userinfo"
  domains                = ["example.com"]
  owner                  = "client"
}
```

---

## Hosted Pages Migration

`cidaas_hosted_page` still manages the hosted page **group** (URLs and page content), on both v3 and v4. It was not split or replaced.

v4 adds separate resources around that group:

1. `cidaas_hosted_page` — group + `hosted_pages` entries (same role as in v3)
2. `cidaas_theme` — CSS upload
3. `cidaas_translations` — locale dictionaries
4. `cidaas_hosted_page_layout` — branding; **references** the group via `layout.hosted_page_group`

`cidaas_app_configuration` does not take a group name. It references the layout: `hosted_pages_layout_id`.

> [!IMPORTANT]
> **Reserved System Group Protection**: System reserved hosted page groups (`default` and `admin`) are hard-protected against accidental API deletion during `terraform destroy` or state teardowns.

### Hosted Pages Attribute Mapping

| v3 / Legacy                                              | v4                                                                 | Notes                                                                                          |
| :------------------------------------------------------- | :----------------------------------------------------------------- | :--------------------------------------------------------------------------------------------- |
| `cidaas_hosted_page.hosted_page_group_name`              | `cidaas_hosted_page.hosted_page_group_name`                        | Same resource. Referenced from the layout as `layout.hosted_page_group`.                       |
| `cidaas_hosted_page.default_locale`                      | `cidaas_hosted_page.default_locale`                                | Unchanged.                                                                                     |
| `cidaas_hosted_page.hosted_pages`                        | `cidaas_hosted_page.hosted_pages`                                  | Unchanged (`hosted_page_id`, `url`, `locale`, `content`).                                      |
| `cidaas_app.hosted_page_group`                           | `cidaas_app_configuration.hosted_pages_layout_id`                  | App now points at a `cidaas_hosted_page_layout` ID, not the group name.                        |
| App-level branding (`accent_color`, `logo_uri`, …)       | `cidaas_hosted_page_layout.layout`                                 | Additional v4 resource.                                                                        |
| N/A                                                      | `cidaas_theme`                                                     | Additional v4 resource. Filename is referenced from `layout.theme` / `resources.*.theme`.      |
| N/A                                                      | `cidaas_translations`                                              | Additional v4 resource. Locale set is referenced from `resources.*.translation_set`.           |

### Hosted Pages Before / After HCL Comparison

#### Legacy v3 (`cidaas_hosted_page` + group name on the app)

```hcl
resource "cidaas_hosted_page" "custom_group" {
  hosted_page_group_name = "custom_group"
  default_locale         = "en-US"

  hosted_pages = [
    {
      hosted_page_id = "login"
      locale         = "en-US"
      url            = "https://example.com/login"
    }
  ]
}

resource "cidaas_app" "legacy_app" {
  # ...
  hosted_page_group = cidaas_hosted_page.custom_group.hosted_page_group_name
}
```

#### v4 (`cidaas_hosted_page` plus layout, theme, translations)

```hcl
resource "cidaas_hosted_page" "custom_group" {
  hosted_page_group_name = "custom_group"
  default_locale         = "en-US"

  hosted_pages = [
    {
      hosted_page_id = "login"
      locale         = "en-US"
      url            = "https://example.com/login"
    }
  ]
}

resource "cidaas_theme" "custom_theme" {
  filename    = "custom.css"
  css_content = ".login-card { background-color: #ffffff; }"
}

resource "cidaas_translations" "en" {
  locale_id = "en-US"
  enabled   = true
  translations = {
    "login.title" = "Welcome to Example Portal"
  }
}

resource "cidaas_hosted_page_layout" "custom_layout" {
  description = "Example branding layout"

  layout = {
    hosted_page_group = cidaas_hosted_page.custom_group.hosted_page_group_name
    theme             = cidaas_theme.custom_theme.filename
    primary_color     = "#ef4923"
    accent_color      = "#f7941d"
    content_align     = "CENTER"
    media_type        = "IMAGE"
    logo_uri          = "https://example.com/logo.png"
  }

  resources = {
    "default-hosted-pages-webapp" = {
      translation_set = cidaas_translations.en.locale_id
      theme           = cidaas_theme.custom_theme.filename
    }
  }
}

resource "cidaas_app_configuration" "v4_app" {
  # ...
  hosted_pages_layout_id = cidaas_hosted_page_layout.custom_layout.id
}
```
