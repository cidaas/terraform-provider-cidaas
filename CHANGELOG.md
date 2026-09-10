# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [4.0.0]

### Added

- **Dual Platform Support (v3 & v4)**: Single provider binary supporting both cidaas v3 (Legacy) and v4 (Trustdesk) platforms.
- **Provider Version Resolution**: Target platform selection precedence: HCL `cidaas_version` → `TERRAFORM_PROVIDER_CIDAAS_VERSION` → `CIDAAS_VERSION`.
- **v4 Trustdesk App Configuration**: `cidaas_app_configuration` resource for appv3 client configuration (`/app-srv/apps`).
- **v4 Hosted Pages & Branding Suite**:
  - `cidaas_hosted_page_group`: Hosted page group management (`/hostedpages-srv/hpgroup`).
  - `cidaas_hosted_page_layout`: Layout configurations and theme/translation resource mappings (`/hostedpages-srv/hosted-page-layouts`).
  - `cidaas_theme`: Custom CSS theme uploads (`/hostedpages-srv/themes`).
  - `cidaas_translations`: Multilingual translation set management (`/hostedpages-srv/translations`).
- **v4 User Setup & Profile Management**: `cidaas_user_setup` managing registration flows, deduplication, allowed fields, and group role assignments (`/user-srv/usersetup`).
- **v4 Verification & MFA Workflow Engine**:
  - `cidaas_suggest_verification_method`: Verification method suggestions (`/verification-actions-srv/suggest-verification-configs`).
  - `cidaas_verification_options`: Verification options and password policy linking (`/verification-actions-srv/verification-options`).
- **v4 Group Access & Filters**:
  - `cidaas_group_selection`: Group selection configurations.
  - `cidaas_group_verification_filter`: Group verification request filters.
- **v4 Identity Federation**: `cidaas_federation_provider` managing enterprise OIDC/SAML federation setups.

### Changed

- **Framework Migration**: Fully migrated from Terraform Plugin SDKv2 to modern **Terraform Plugin Framework** (`github.com/hashicorp/terraform-plugin-framework`).
- **Domain Package Restructuring**: Reorganized internal package layout from flat files into domain-driven subpackages under `internal/resources/` (`app`, `consent`, `group`, `hostedpages`, `identity_provider`, `notification`, `registration_field`, `role`, `scope`, `security`, `user_group`, `usersetup`, `verification`, `webhook`).
- **Shared Resources & Dual-Version Compatibility**: Restored and validated shared resources (`cidaas_scope`, `cidaas_scope_group`, `cidaas_role`, `cidaas_user_groups`, `cidaas_group_type`, `cidaas_registration_field`, `cidaas_password_policy`, `cidaas_security_settings`, `cidaas_webhook`, `cidaas_notifications_template_group`) as well as legacy resources (`cidaas_hosted_page`, `cidaas_social_provider`, `cidaas_custom_provider`) to function on both `3.x` and `4.x` target versions without requiring breaking changes in customer HCL configurations.

### Deprecated

- `cidaas_app`: Legacy v3 application resource (backed by appv1). Shows a deprecation message directing v4 users to `cidaas_app_configuration` (backed by appv3 `/app-srv/apps`).
- `cidaas_template` / `cidaas_template_group`: Legacy v3 templates-srv resources (use `cidaas_notifications_template_group` for v4).

### Fixed

- **Resource Leak Fixes**: Resolved `deferInLoop` memory leaks in API retry handlers (`security_settings.go`).
- **Lint & Code Formatting**: Resolved all `gofmt` formatting errors and `revive` unused parameter warnings (**0 lint warnings**).
- **Documentation Alignment**: Aligned HCL usage examples across guides and resources (`cidaas_scope`, `cidaas_hosted_page`, `cidaas_notification_template_type`).
