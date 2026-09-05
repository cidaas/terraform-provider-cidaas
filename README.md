![Logo](https://raw.githubusercontent.com/cidaas/terraform-provider-cidaas/master/logo.jpg)

# Terraform Provider for cidaas (v3 & v4)

Manage cidaas **v3** and **v4 (Trustdesk)** tenants with Terraform: applications, scopes, roles, registration fields, hosted pages, notifications, identity providers, and more.

Registry address: `registry.terraform.io/Cidaas/cidaas`

---

## Documentation

| Section | Path |
|---------|------|
| Provider overview | [docs/index.md](docs/index.md) |
| Resource reference | [docs/resources/](docs/resources/) |
| Guides | [docs/guides/](docs/guides/) |
| Examples (HCL) | [examples/](examples/) |
| Changelog | [CHANGELOG.md](CHANGELOG.md) |
| Contributing (docs workflow) | [CONTRIBUTING.md](CONTRIBUTING.md) |

**Guides**

- [Getting started](docs/guides/getting-started.md)
- [Authentication](docs/guides/authentication.md)
- [Version targeting (v3 vs v4)](docs/guides/version-targeting.md)
- [Resource dependency order](docs/guides/resource-dependencies.md)

Attribute schemas and Example Usage for each resource are generated into `docs/resources/` — do not treat this README as a substitute for those pages.

---

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- Go >= 1.25 (building from source)
- A cidaas **v3** or **v4** tenant and a non-interactive OAuth client (`client_id` / `client_secret`) with scopes for the resources you manage

---

## Quick start

### Credentials (required)

```bash
export TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID="your-client-id"
export TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET="your-client-secret"
```

### Target version (required)

Set via HCL or environment. Precedence: `cidaas_version` in the provider block → `TERRAFORM_PROVIDER_CIDAAS_VERSION` → `CIDAAS_VERSION`.

```hcl
terraform {
  required_providers {
    cidaas = {
      source  = "Cidaas/cidaas"
      version = ">= 4.0.0-alpha.1"
    }
  }
}

provider "cidaas" {
  base_url       = "https://your-tenant.cidaas.eu"
  cidaas_version = "4.x" # or "3.x"
}
```

Accepted version strings normalize to `3.x` or `4.x` (for example `3`, `v3`, `4.x`, `v4.x`).

Details: [Authentication](docs/guides/authentication.md) · [Version targeting](docs/guides/version-targeting.md) · [Provider schema](docs/index.md)

---

## Dual versioning

| Kind | Resources |
|------|-----------|
| **v3 only** | `cidaas_app`, `cidaas_hosted_page` |
| **v4 only** | `cidaas_app_configuration`, hosted page group/layout/theme/translations, user setup, verification resources, federation provider, group selection / verification filter |
| **Shared** | scopes, roles, groups, consent, notifications, webhooks, social/custom providers, registration fields, security settings, and more |

Using a v4-only resource with `cidaas_version = "3.x"` (or the reverse) fails with an explicit configure/plan error. See [Version targeting](docs/guides/version-targeting.md).

---

## Resources

Every registered resource is listed below. **Docs** = Registry page; **Example** = sample HCL.

### Trustdesk (v4 only)

| Resource | API / service | Description | Docs | Example |
|----------|---------------|-------------|------|---------|
| [`cidaas_app_configuration`](docs/resources/app_configuration.md) | `/app-srv/apps` | Application configuration (Trustdesk app model) | [docs](docs/resources/app_configuration.md) | [example](examples/resources/cidaas_app_configuration/) |
| [`cidaas_hosted_page_group`](docs/resources/hosted_page_group.md) | `/hostedpages-srv/hpgroup` | Hosted page groups | [docs](docs/resources/hosted_page_group.md) | [example](examples/resources/cidaas_hosted_page_group/) |
| [`cidaas_hosted_page_layout`](docs/resources/hosted_page_layout.md) | `/hostedpages-srv/hosted-page-layouts` | Hosted page layouts | [docs](docs/resources/hosted_page_layout.md) | [example](examples/resources/cidaas_hosted_page_layout/) |
| [`cidaas_theme`](docs/resources/theme.md) | `/hostedpages-srv/themes` | Custom CSS themes | [docs](docs/resources/theme.md) | [example](examples/resources/cidaas_theme/) |
| [`cidaas_translations`](docs/resources/translations.md) | `/hostedpages-srv/translations` | Hosted page locale strings | [docs](docs/resources/translations.md) | [example](examples/resources/cidaas_translations/) |
| [`cidaas_user_setup`](docs/resources/user_setup.md) | `/user-srv/usersetup` | Tenant user registration setup | [docs](docs/resources/user_setup.md) | [example](examples/resources/cidaas_user_setup/) |
| [`cidaas_suggest_verification_method`](docs/resources/suggest_verification_method.md) | `/verification-actions-srv/suggest-verification-configs` | Suggested verification methods | [docs](docs/resources/suggest_verification_method.md) | [example](examples/resources/cidaas_suggest_verification_method/) |
| [`cidaas_verification_options`](docs/resources/verification_options.md) | `/verification-actions-srv/verification-options` | Verification options | [docs](docs/resources/verification_options.md) | [example](examples/resources/cidaas_verification_options/) |
| [`cidaas_federation_provider`](docs/resources/federation_provider.md) | `/federation/providers` | Enterprise federation providers | [docs](docs/resources/federation_provider.md) | [example](examples/resources/cidaas_federation_provider/) |
| [`cidaas_group_selection`](docs/resources/group_selection.md) | `/groups-srv/selection` | Group login selection | [docs](docs/resources/group_selection.md) | [example](examples/resources/cidaas_group_selection/) |
| [`cidaas_group_verification_filter`](docs/resources/group_verification_filter.md) | `/groups-srv/verification-filter` | Group verification filters | [docs](docs/resources/group_verification_filter.md) | [example](examples/resources/cidaas_group_verification_filter/) |

### v3 only

| Resource | API / service | Description | Docs | Example |
|----------|---------------|-------------|------|---------|
| [`cidaas_app`](docs/resources/app.md) | app-srv (legacy) | Legacy v3 application (`cidaas_app_configuration` on v4) | [docs](docs/resources/app.md) | [example](examples/resources/cidaas_app/) |
| [`cidaas_hosted_page`](docs/resources/hosted_page.md) | hostedpages (legacy) | Legacy v3 hosted page (`cidaas_hosted_page_group` on v4) | [docs](docs/resources/hosted_page.md) | [example](examples/resources/cidaas_hosted_page/) |

### Shared (v3 & v4)

| Resource | API / service | Description | Docs | Example |
|----------|---------------|-------------|------|---------|
| [`cidaas_custom_provider`](docs/resources/custom_provider.md) | providers-srv | Custom OIDC / OAuth2 identity providers | [docs](docs/resources/custom_provider.md) | [example](examples/resources/cidaas_custom_provider/) |
| [`cidaas_social_provider`](docs/resources/social_provider.md) | providers-srv | Social identity providers (Google, Apple, …) | [docs](docs/resources/social_provider.md) | [example](examples/resources/cidaas_social_provider/) |
| [`cidaas_consent`](docs/resources/consent.md) | consent-management-srv | Consent definitions | [docs](docs/resources/consent.md) | [example](examples/resources/cidaas_consent/) |
| [`cidaas_consent_group`](docs/resources/consent_group.md) | consent-management-srv | Consent groups | [docs](docs/resources/consent_group.md) | [example](examples/resources/cidaas_consent_group/) |
| [`cidaas_consent_version`](docs/resources/consent_version.md) | consent-management-srv | Consent versions | [docs](docs/resources/consent_version.md) | [example](examples/resources/cidaas_consent_version/) |
| [`cidaas_registration_field`](docs/resources/registration_field.md) | registration-setup-srv | Registration fields & validators | [docs](docs/resources/registration_field.md) | [example](examples/resources/cidaas_registration_field/) |
| [`cidaas_role`](docs/resources/role.md) | roles-srv | Tenant roles | [docs](docs/resources/role.md) | [example](examples/resources/cidaas_role/) |
| [`cidaas_user_groups`](docs/resources/user_groups.md) | groups-srv | User groups | [docs](docs/resources/user_groups.md) | [example](examples/resources/cidaas_user_groups/) |
| [`cidaas_group_type`](docs/resources/group_type.md) | groups-srv | Group types & role modes | [docs](docs/resources/group_type.md) | [example](examples/resources/cidaas_group_type/) |
| [`cidaas_scope`](docs/resources/scope.md) | scopes-srv | OAuth scopes | [docs](docs/resources/scope.md) | [example](examples/resources/cidaas_scope/) |
| [`cidaas_scope_group`](docs/resources/scope_group.md) | scopes-srv | Scope groups | [docs](docs/resources/scope_group.md) | [example](examples/resources/cidaas_scope_group/) |
| [`cidaas_password_policy`](docs/resources/password_policy.md) | password-policy-srv | Password policies | [docs](docs/resources/password_policy.md) | [example](examples/resources/cidaas_password_policy/) |
| [`cidaas_security_settings`](docs/resources/security_settings.md) | security-srv | Tenant security settings | [docs](docs/resources/security_settings.md) | [example](examples/resources/cidaas_security_settings/) |
| [`cidaas_template`](docs/resources/template.md) | templates-srv | Legacy communication templates | [docs](docs/resources/template.md) | [example](examples/resources/cidaas_template/) |
| [`cidaas_template_group`](docs/resources/template_group.md) | templates-srv | Legacy template groups (prefer notifications template group on v4+) | [docs](docs/resources/template_group.md) | [example](examples/resources/cidaas_template_group/) |
| [`cidaas_notification_template`](docs/resources/notification_template.md) | templates-srv / notification | Notification templates | [docs](docs/resources/notification_template.md) | [example](examples/resources/cidaas_notification_template/) |
| [`cidaas_notification_template_type`](docs/resources/notification_template_type.md) | templates-srv / notification | Notification template types | [docs](docs/resources/notification_template_type.md) | [example](examples/resources/cidaas_notification_template_type/) |
| [`cidaas_notifications_template_group`](docs/resources/notifications_template_group.md) | notification-srv | Notification template groups | [docs](docs/resources/notifications_template_group.md) | [example](examples/resources/cidaas_notifications_template_group/) |
| [`cidaas_notifications_template_group_locale`](docs/resources/notifications_template_group_locale.md) | notification-srv | Template group locales | [docs](docs/resources/notifications_template_group_locale.md) | [example](examples/resources/cidaas_notifications_template_group_locale/) |
| [`cidaas_notification_service_setup`](docs/resources/notification_service_setup.md) | notification-srv | Communication provider service setups | [docs](docs/resources/notification_service_setup.md) | [example](examples/resources/cidaas_notification_service_setup/) |
| [`cidaas_webhook`](docs/resources/webhook.md) | webhook-srv | Webhooks and event subscriptions | [docs](docs/resources/webhook.md) | [example](examples/resources/cidaas_webhook/) |

Identity provider examples: [custom](examples/resources/cidaas_custom_provider/) · [social](examples/resources/cidaas_social_provider/) · [federation](examples/resources/cidaas_federation_provider/). On Trustdesk, `owner` defaults to `client` for Admin UI visibility.

---

## Local development

```bash
make build      # build provider binary
make install    # install into ~/.terraform.d/plugins
make test       # unit tests
make generate   # fmt examples + generate docs/

# Acceptance tests (requires tenant credentials)
CIDAAS_VERSION=3.x TF_ACC=1 BASE_URL=https://... make testacc
CIDAAS_VERSION=4.x TF_ACC=1 BASE_URL=https://... make testacc
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the documentation workflow.

---

## CI / CD

GitLab CI (`.gitlab-ci.yml`):

- `acceptance_test` — `make test-ci` (v3/v4 via `CIDAAS_VERSION`)

---

## License

See [LICENSE](LICENSE).
