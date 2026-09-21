---
page_title: "cidaas Provider Configuration (v4)"
---

# Provider configuration

Provider **4.x targets cidaas v4 (Trustdesk)**. It is not a dual-compatibility product for remaining on cidaas v3 forever — those customers should keep **provider 3.5.x**.

Set `cidaas_version = "4.x"` when using this major version against Trustdesk.

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

## Version checks

- **New Trustdesk-only resources** call a `cidaas_version` gate at Configure. With `cidaas_version = "3.x"` they fail with a clear error (so a mistaken v3 target cannot use v4-only APIs).
- **Shared / legacy resources** (scopes, roles, groups, registration fields, consent, notification templates, hosted page groups, deprecated social/custom providers, etc.) are **not** version-gated. Existing customer HCL for those types must keep working; additive optional fields for Trustdesk structure are allowed.

## Resource availability (Trustdesk)

**v4-only** (require `cidaas_version = "4.x"`):

- **Applications**: `cidaas_app_configuration`
- **Hosted page branding**: `cidaas_hosted_page_layout`, `cidaas_theme`, `cidaas_translations`
- **Identity**: `cidaas_federation_provider`
- **User / MFA / groups**: `cidaas_user_setup`, `cidaas_verification_options`, `cidaas_suggest_verification_method`, `cidaas_group_selection`, `cidaas_group_verification_filter`

**Shared / legacy** (existing configs stay valid; prefer modern replacements where noted):

- `cidaas_hosted_page`, `cidaas_scope`, `cidaas_scope_group`, `cidaas_role`, `cidaas_group_type`, `cidaas_user_groups`, `cidaas_registration_field`, `cidaas_password_policy`, `cidaas_security_settings`, `cidaas_webhook`, notification template stack, consent resources
- Deprecated: `cidaas_app`, `cidaas_social_provider`, `cidaas_custom_provider`, `cidaas_template`, `cidaas_template_group`

See the [provider overview](../index.md), [v3 to v4 migration guide](v3-to-v4-migration.md), and resource pages for attribute-level details.
