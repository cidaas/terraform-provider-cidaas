---
page_title: "cidaas_auth_setup_defaults Resource - cidaas"
subcategory: "Apps"
description: |-
  Manages the tenant Default Authentication Setup (Trustdesk / v4) via app-srv/apps/auth-setup-defaults.
  Requires admin roles and OAuth scopes cidaas:apps_read and cidaas:apps_write. Provider cidaas_version must be 4.x.
---

# cidaas_auth_setup_defaults (Resource)

Manages the tenant **Default Authentication Setup** (Trustdesk) via `GET/PUT app-srv/apps/auth-setup-defaults`.

This is a **v4 / Trustdesk-only** singleton (`id` is always `default`). Create updates the seeded record; Destroy only removes the resource from Terraform state (tenant defaults are not deleted in cidaas).

Apps may override the same bool flags under [`cidaas_app_configuration.authentication_setup`](app_configuration.md#nestedatt--authentication_setup) (omit a bool to inherit these defaults). `login_spi` is **app-only**; `net_id` is **tenant-only**.

## Requirements

| Item | Value |
|------|--------|
| Provider | `cidaas_version = "4.x"` (or `v4.x`) |
| OAuth scopes | `cidaas:apps_read`, `cidaas:apps_write` |
| Roles | Admin / app-manager roles that can manage Trustdesk Default Authentication Setup |
| API | `GET` / `PUT` `{base_url}/app-srv/apps/auth-setup-defaults` |

## Example Usage

```terraform
# Tenant Default Authentication Setup (Trustdesk / v4)
# Requires: cidaas_version = "4.x", scopes cidaas:apps_read + cidaas:apps_write

resource "cidaas_auth_setup_defaults" "tenant" {
  name                            = "default"
  description                     = "Trustdesk default authentication setup"
  auto_login_after_register       = false
  register_with_login_information = false
  enable_password_less_auth       = true
  allow_user_level_multi_provider = true
  social_business_ids             = false
  net_id                          = false
}
```

## Argument Reference

- `name` (Optional, String) Display name of the defaults record (usually `default`).
- `description` (Optional, String) Optional description.
- `auto_login_after_register` (Optional, Boolean) Automatically log the user in after registration.
- `register_with_login_information` (Optional, Boolean) Allow registration using login information.
- `enable_password_less_auth` (Optional, Boolean) Enable passwordless authentication methods (magic link / OTP).
- `allow_user_level_multi_provider` (Optional, Boolean) Allow users to link multiple identity providers.
- `social_business_ids` (Optional, Boolean) Enable social business IDs.
- `net_id` (Optional, Boolean) Enable the NetID special social provider (Active flag). Tenant-level only; not available on per-app `authentication_setup`.

## Attribute Reference

- `id` (String) Always `default`.
- `created_time` (String)
- `updated_time` (String)

## Import

```shell
terraform import cidaas_auth_setup_defaults.tenant default
```

Import id must be exactly `default`.

## See also

- Per-app overrides: [`cidaas_app_configuration.authentication_setup`](app_configuration.md#nestedatt--authentication_setup)
- Migration: [v3 → v4 application migration](../guides/v3-to-v4-migration.md#exhaustive-application-attribute-mapping)
