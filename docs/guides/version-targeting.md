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

| Kind | Examples |
|------|----------|
| **v3 only** | `cidaas_app`, `cidaas_hosted_page` |
| **v4 only** | `cidaas_app_configuration`, `cidaas_hosted_page_group`, `cidaas_hosted_page_layout`, `cidaas_theme`, `cidaas_translations`, `cidaas_user_setup`, `cidaas_verification_options`, `cidaas_suggest_verification_method`, `cidaas_federation_provider`, `cidaas_group_selection`, `cidaas_group_verification_filter` |
| **Shared** | scopes, roles, groups, consent, notifications, webhooks, social/custom providers, and more |

Using a v4-only resource with `cidaas_version = "3.x"` (or the reverse) fails during provider configure / plan with a clear error.

## Migration notes

- Prefer `cidaas_app_configuration` on Trustdesk instead of `cidaas_app`.
- Prefer `cidaas_hosted_page_group` (+ layout/theme) instead of `cidaas_hosted_page`.
- Prefer `cidaas_notifications_template_group` instead of deprecated `cidaas_template_group` for new notification designs.

See the [provider overview](../index.md) and resource pages for attribute-level details.
