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

1. **Dual-Version Compatibility**: `cidaas_app` and `cidaas_hosted_page` operate on both `3.x` and `4.x` target versions.
2. **Native v4 Trustdesk Features**: Native v4 resources (`cidaas_app_configuration`, `cidaas_hosted_page_layout`, `cidaas_theme`, `cidaas_translations`, `cidaas_federation_provider`, `cidaas_user_setup`, `cidaas_verification_options`) offer enhanced fine-grained microservice capabilities on `/apps-srv`, `/hostedpages-srv`, `/federation/providers`, `/usersetup-srv`, and `/verification-srv`.

## Resource Mapping Matrix

| Category                  | Shared / Legacy Resource                                               | v4 Native Resource                                                                                   | Supported Target Versions | Guidance                                                                                                                     |
| :------------------------ | :--------------------------------------------------------------------- | :--------------------------------------------------------------------------------------------------- | :------------------------ | :--------------------------------------------------------------------------------------------------------------------------- |
| **Applications**          | `cidaas_app`                                                           | `cidaas_app_configuration`                                                                           | Both (3.x & 4.x)          | `cidaas_app` works on both v3 and v4. For v4 native features, migrate to `cidaas_app_configuration`.                         |
| **Hosted Pages**          | `cidaas_hosted_page`                                                   | `cidaas_hosted_page_layout`<br>`cidaas_theme`<br>`cidaas_translations`                               | Both (3.x & 4.x)          | `cidaas_hosted_page` is the single unified resource supporting inline theme, translations, and layout on both v3 and v4. System reserved groups (`default`, `admin`) are protected against accidental API deletion. |
| **Identity Providers**    | Legacy: `cidaas_social_provider`<br>`cidaas_custom_provider` (Removed) | `cidaas_federation_provider`                                                                         | 4.x only                  | `cidaas_social_provider` and `cidaas_custom_provider` are removed in v4. Standardize exclusively on `cidaas_federation_provider`. |
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

### Step 2: Test Existing Infrastructure

Run `terraform plan` to verify that existing dual-version resources (`cidaas_app`, `cidaas_hosted_page`) plan cleanly without errors.


### Step 3: Gradually Adopt Native v4 Resources

When you are ready to utilize Trustdesk features:

- Adopt `cidaas_app_configuration` for app management via `apps-srv`.
- Adopt `cidaas_hosted_page_layout`, `cidaas_theme`, and `cidaas_translations` alongside `cidaas_hosted_page` for modular UI styling.
- Adopt `cidaas_federation_provider` for native identity provider configuration.
