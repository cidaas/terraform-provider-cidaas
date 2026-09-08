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

## Strategy: Dual-Version Support vs. Native v4 Resources

The provider provides two operational pathways when managing resources in a v4 environment:

1. **Dual-Version Compatibility**: `cidaas_app`, `cidaas_hosted_page`, `cidaas_social_provider`, and `cidaas_custom_provider` operate on both `3.x` and `4.x` target versions without requiring breaking changes to existing HCL files.
2. **Native v4 Trustdesk Features**: Native v4 resources (`cidaas_app_configuration`, `cidaas_hosted_page_group`, `cidaas_hosted_page_layout`, `cidaas_theme`, `cidaas_translations`, `cidaas_federation_provider`, `cidaas_user_setup`, `cidaas_verification_options`) offer enhanced fine-grained microservice capabilities on `/apps-srv`, `/hostedpages-srv`, `/federation/providers`, `/usersetup-srv`, and `/verification-srv`.

## Resource Mapping Matrix

| Category                  | Legacy / Shared Resource                                               | v4 Native Resource                                                                                   | Supported Target Versions | Guidance                                                                                                                     |
| :------------------------ | :--------------------------------------------------------------------- | :--------------------------------------------------------------------------------------------------- | :------------------------ | :--------------------------------------------------------------------------------------------------------------------------- |
| **Applications**          | `cidaas_app`                                                           | `cidaas_app_configuration`                                                                           | Both (3.x & 4.x)          | `cidaas_app` works on both v3 and v4. For v4 native features, migrate to `cidaas_app_configuration`.                         |
| **Hosted Pages**          | `cidaas_hosted_page`                                                   | `cidaas_hosted_page_group`<br>`cidaas_hosted_page_layout`<br>`cidaas_theme`<br>`cidaas_translations` | Both (3.x & 4.x)          | `cidaas_hosted_page` works on both v3 and v4. Split into container, layouts, theme, and translations for v4 native features. |
| **Identity Providers**    | `cidaas_social_provider`<br>`cidaas_custom_provider`                   | `cidaas_federation_provider`                                                                         | Both (3.x & 4.x)          | `social_provider` and `custom_provider` work on both. Use `federation_provider` for native v4 IDP management.                |
| **Consents**              | `cidaas_consent`<br>`cidaas_consent_group`<br>`cidaas_consent_version` | N/A                                                                                                  | 3.x only                  | Requires `cidaas_version = "3.x"`.                                                                                           |
| **Scopes & Roles**        | `cidaas_scope`<br>`cidaas_scope_group`<br>`cidaas_role`                | Shared                                                                                               | Both (3.x & 4.x)          | Fully compatible across v3 and v4.                                                                                           |
| **Groups & Verification** | `cidaas_group_type`<br>`cidaas_user_groups`                            | `cidaas_group_selection`<br>`cidaas_group_verification_filter`                                       | Both (3.x & 4.x)          | Core groups shared; selection and verification filters are v4 native.                                                        |
| **User Setup & MFA**      | N/A                                                                    | `cidaas_user_setup`<br>`cidaas_verification_options`<br>`cidaas_suggest_verification_method`         | 4.x only                  | Native v4 features.                                                                                                          |

## Recommended Step-by-Step Migration

### Step 1: Update Target Version

In your provider configuration, set `cidaas_version` to `"4.x"`:

```hcl
provider "cidaas" {
  base_url       = var.cidaas_base_url
  cidaas_version = "4.x"
}
```

### Step 2: Test Existing Infrastructure (Zero Breaking Changes)

Run `terraform plan` to verify that existing dual-version resources (`cidaas_app`, `cidaas_hosted_page`, `cidaas_social_provider`, `cidaas_custom_provider`) plan cleanly without errors.

### Step 3: Gradually Adopt Native v4 Resources

When you are ready to utilize Trustdesk features:

- Adopt `cidaas_app_configuration` for app management via `apps-srv`.
- Adopt `cidaas_hosted_page_group`, `cidaas_hosted_page_layout`, `cidaas_theme`, and `cidaas_translations` for modular UI styling.
- Adopt `cidaas_federation_provider` for native identity provider configuration.
