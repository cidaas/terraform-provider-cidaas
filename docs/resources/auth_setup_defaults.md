# cidaas_auth_setup_defaults

Manages the tenant **Default Authentication Setup** (Trustdesk) via `GET/PUT app-srv/apps/auth-setup-defaults`.

This is a singleton (`id` is always `default`). Create updates the seeded record; Destroy only removes the resource from Terraform state.

```terraform
resource "cidaas_auth_setup_defaults" "tenant" {
  name                             = "default"
  description                      = "Trustdesk default authentication setup"
  auto_login_after_register        = false
  register_with_login_information  = false
  enable_password_less_auth        = true
  allow_user_level_multi_provider  = true
  social_business_ids              = false
  net_id                           = false
}
```

## Argument Reference

- `name` (Optional, String) Display name (usually `default`).
- `description` (Optional, String) Optional description.
- `auto_login_after_register` (Optional, Boolean)
- `register_with_login_information` (Optional, Boolean)
- `enable_password_less_auth` (Optional, Boolean)
- `allow_user_level_multi_provider` (Optional, Boolean)
- `social_business_ids` (Optional, Boolean)
- `net_id` (Optional, Boolean) NetID Active flag.

## Attribute Reference

- `id` (String) Always `default`.
- `created_time` (String)
- `updated_time` (String)

## Import

```shell
terraform import cidaas_auth_setup_defaults.tenant default
```

Requires admin roles and scopes `cidaas:apps_read` / `cidaas:apps_write`. Per-app overrides live under `cidaas_app_configuration.authentication_setup`.
