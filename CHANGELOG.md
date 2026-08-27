# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Greenfield cidaas **v4** Terraform provider (`Cidaas/cidaas` 4.x line).
- `cidaas_theme` — CSS theme upload via multipart `POST /hostedpages-srv/themes`.
- `cidaas_translations` — locale translation CRUD via `/hostedpages-srv/translations` (API field `translation`).
- `cidaas_hosted_page_group` — hosted page group upsert/get/delete via `/hostedpages-srv/hpgroup`.
- `cidaas_hosted_page_layout` — hosted page layout CRUD via `/hostedpages-srv/hosted-page-layouts` (optional `resources` map).
- `cidaas_user_setup` — user setup profile CRUD via `/user-srv/usersetup` (PATCH update; top-level `consent_refs` maps to `user_setup.consent_refs`).
- `cidaas_suggest_verification_method` — Suggest Verification method via `/verification-actions-srv/suggest-verification-configs` (PUT update).
- `cidaas_verification_options` — verification options via `/verification-actions-srv/verification-options` (PUT update; optional `suggest_verification_method_id`).
- Provider-level v4 capability gate (`GET /public-srv/version`; 404 treated as v4-capable for this provider).
- GitLab `acceptance_test` job (`make test-ci`), `.releaserc`, `.goreleaser.yml`.

## [4.0.0-alpha.1] - Unreleased

Initial alpha scaffold for hosted pages Phase 1 (#2415).
