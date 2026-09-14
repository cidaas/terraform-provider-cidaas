---
page_title: "cidaas Provider Configuration (v4)"
---

# Provider configuration

The provider targets **cidaas v4 (Trustdesk)**.

## How to set the version

Precedence (highest first):

1. `cidaas_version` in the `provider "cidaas"` block
2. `TERRAFORM_PROVIDER_CIDAAS_VERSION`
3. `CIDAAS_VERSION`

```hcl
provider "cidaas" {
  base_url       = "https://your-tenant.cidaas.eu"
  cidaas_version = "4.x"
}
```

## Resource availability

All resources natively support **cidaas v4 (Trustdesk)**:

- **Applications**: `cidaas_app_configuration` (`/app-srv/apps`).
- **Hosted Pages & Branding**: `cidaas_hosted_page`, `cidaas_hosted_page_layout`, `cidaas_theme`, `cidaas_translations`.
- **Identity Providers**: `cidaas_federation_provider`.
- **Security & User Setup**: `cidaas_user_setup`, `cidaas_verification_options`, `cidaas_suggest_verification_method`, `cidaas_group_selection`, `cidaas_group_verification_filter`.
- **Core Baseline**: `cidaas_scope`, `cidaas_role`, `cidaas_group_type`, `cidaas_user_groups`, `cidaas_registration_field`, `cidaas_password_policy`, `cidaas_security_settings`, `cidaas_notifications_template_group`, `cidaas_webhook`.

See the [provider overview](../index.md) and resource pages for attribute-level details.
