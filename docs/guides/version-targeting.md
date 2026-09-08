---
page_title: "cidaas Version Targeting (v3 vs v4)"
---

# Version targeting

The provider supports **cidaas v3** and **cidaas v4 (Trustdesk)** in one binary. The target version is selected per Terraform configuration (not per resource).

## How to set the version

Precedence (highest first):

1. `cidaas_version` in the `provider "cidaas"` block
2. `TERRAFORM_PROVIDER_CIDAAS_VERSION`
3. `CIDAAS_VERSION`

```hcl
provider "cidaas" {
  base_url       = "https://your-tenant.cidaas.eu"
  cidaas_version = "4.x" # or "3.x"
}
```

## Resource availability

| Kind | Description & Examples |
|------|------------------------|
| **Shared (Dual-Version)** | Operates on both `3.x` and `4.x` target versions: `cidaas_app`, `cidaas_hosted_page`, `cidaas_social_provider`, `cidaas_custom_provider`, `cidaas_scope`, `cidaas_role`, `cidaas_group_type`, `cidaas_user_groups`, `cidaas_notification_service_setup`, `cidaas_webhook`, etc. |
| **v4 Native (Trustdesk)** | Requires `cidaas_version = "4.x"`: `cidaas_app_configuration`, `cidaas_federation_provider`, `cidaas_hosted_page_group`, `cidaas_hosted_page_layout`, `cidaas_theme`, `cidaas_translations`, `cidaas_user_setup`, `cidaas_verification_options`, `cidaas_suggest_verification_method`, `cidaas_group_selection`, `cidaas_group_verification_filter` |
| **v3 Only** | Requires `cidaas_version = "3.x"`: `cidaas_consent`, `cidaas_consent_group`, `cidaas_consent_version` |

Attempting to use a v4-only resource with `cidaas_version = "3.x"` (or a v3-only resource with `"4.x"`) returns a validation error.

## Migration notes

- **Applications**: `cidaas_app` supports both `3.x` and `4.x`. For v4 (Trustdesk) environments, `cidaas_app_configuration` is recommended as it natively exposes Trustdesk app capabilities (`app-srv/apps`).
- **Identity Providers**: `cidaas_social_provider` and `cidaas_custom_provider` work on both `3.x` and `4.x`. Use `cidaas_federation_provider` for v4-native federated provider integration.
- **Hosted Pages**: `cidaas_hosted_page` works on both `3.x` and `4.x`. Use `cidaas_hosted_page_group` (+ layout/theme/translations) for advanced v4 layout and styling capabilities.
- **Notifications**: Prefer `cidaas_notifications_template_group` instead of deprecated `cidaas_template_group` for new notification designs.

See the [Cidaas v3 to v4 migration guide](v3-to-v4-migration.md), [provider overview](../index.md), and resource pages for attribute-level details.
