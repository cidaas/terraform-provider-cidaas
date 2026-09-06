# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Dual **v3 / v4 (Trustdesk)** support: resolve target version from HCL `cidaas_version`, then `TERRAFORM_PROVIDER_CIDAAS_VERSION`, then `CIDAAS_VERSION`, then live `GET /public-srv/version`, else default `4.x` (`NormalizeVersion`).
- GitLab acceptance matrix: `acceptance_test_v3` (`CIDAAS_VERSION=3.x`) and `acceptance_test_v4` (`CIDAAS_VERSION=4.x`).
- Trustdesk resources (v4-only; fail Configure when target is `3.x`):
  - `cidaas_theme` — CSS theme upload via multipart `POST /hostedpages-srv/themes`.
  - `cidaas_translations` — locale translation CRUD via `/hostedpages-srv/translations` (API field `translation`).
  - `cidaas_hosted_page_group` — hosted page group upsert/get/delete via `/hostedpages-srv/hpgroup`.
  - `cidaas_hosted_page_layout` — hosted page layout CRUD via `/hostedpages-srv/hosted-page-layouts` (optional `resources` map).
  - `cidaas_user_setup` — user setup profile CRUD via `/user-srv/usersetup` (PATCH update; top-level `consent_refs` maps to `user_setup.consent_refs`).
  - `cidaas_suggest_verification_method` — Suggest Verification method via `/verification-actions-srv/suggest-verification-configs` (PUT update).
  - `cidaas_verification_options` — verification options via `/verification-actions-srv/verification-options` (PUT update; optional `suggest_verification_method_id`).
  - `cidaas_app_configuration` — appv3 app document via `/app-srv/apps` (extdep IDs only).
- Shared resources restored into domain packages (work on **both** `3.x` and `4.x`):
  - `cidaas_registration_field`, `cidaas_role`, `cidaas_user_groups`
  - `cidaas_scope`, `cidaas_scope_group`
  - `cidaas_password_policy`, `cidaas_security_settings`
  - `cidaas_template`, `cidaas_notification_service_setup`
- `cidaas_app` — deprecated migration stub pointing to `cidaas_app_configuration`.

## [4.0.0-alpha.1] - Unreleased

Initial alpha scaffold for hosted pages Phase 1 (#2415).
