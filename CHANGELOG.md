## Changelog

### 3.5.19

#### Enhancements

- **`cidaas_registration_field`:** Added optional **`field_definition.regexes`** — list of Go `regexp` (RE2) patterns merged with AND into one API **`fieldDefinition.regex`** via supported validation shapes (length, charset, contains, no_leading), not concatenation or lookaheads. Unknown/unmergable shapes fail closed. Mutually exclusive with **`regex`**. Per-rule ErrorKeys are not supported. Same **`min_length_msg`** / **`max_length_msg`** requirements as **`regex`** (TEXT/URL only). Plan computes the composed **`regex`** so updates do not hit “inconsistent result after apply”.

### 3.5.18

#### Enhancements

- **`cidaas_notification_service_setup`:** New resource to create/update/delete communication provider service setups via notification-srv. **`status`** is computed (manual service-desk verify; Terraform does not call verify). **`parent_service_setup_id`** is optional+computed (may be auto-filled when omitted). Changing **`service_id`** or **`communication_methods`** forces replacement. Update sends **`name`** / **`description`** only. Destroy treats remote **404** as success. Active setups cannot be deleted until deactivated outside Terraform. Scopes: `cidaas:service_setups_read`, `cidaas:service_setups_write`, `cidaas:service_setups_delete`.
- **`cidaas_notification_provider_config`:** New resource for provider credentials (`config_data` or write-only **`config_data_wo`** + **`config_data_wo_version`**; Terraform **≥ 1.11** for write-only). Wizard-shaped JSON (`commProvider`, `commMethod`, `schemaData`) is supported; the provider injects **`configData.id`** from **`service_setup_id`** when omitted. Create/update upsert via `POST /providerconfigs`. Destroy removes Terraform state only (no remote DELETE). Scopes: `cidaas:service_setups_read`, `cidaas:provider_config_write`.
- **`data.cidaas_notification_service_setup`:** New data source for a single setup by id (any status, including `in-progress`).
- **`data.cidaas_notification_service_setups`:** List returns **active** setups only (suitable for wiring `comm_setting_*` after verify).

### 3.5.17

#### Enhancements

- **`cidaas_app`:** Added optional **`require_pkce`** and **`disable_insecure_pkce_method`**. `require_pkce` enforces `code_challenge` on authz/PAR (AUTH10063 when missing). `disable_insecure_pkce_method` rejects insecure PKCE `plain` method and implicit plain default; only `S256` allowed (AUTH10048).
- **`cidaas_registration_field`:** Added **`field_definition.match_with`** and **`local_texts.match_with_msg`** for the system **`password_echo`** field (password confirmation). Maps to API `fieldDefinition.matchWith` and `localeTexts.matchWith`. Only allowed when **`field_key`** is **`password_echo`**; when **`match_with`** is set, every locale must include **`match_with_msg`**.

### 3.5.16

#### Enhancements

- **`cidaas_registration_field`:** **`order`** is now a writable optional attribute (previously read-only/computed only). When set, the provider positions the field via fieldsetup-srv: on **update**, **`PATCH /fieldsetup-srv/fields/order`** runs before upsert when **`order`** changes (upsert alone does not change order); on **create**, reorder runs after upsert when an explicit **`order`** differs from the API-assigned value. Omit **`order`** to leave position unchanged or accept the next available slot on create.
- **`cidaas_registration_field`:** **`local_texts.required_msg`** may be set when **`required`** is **`false`**, so translations can be pre-defined for fields marked required at the application level. Validation still requires **`required_msg`** in every **`local_texts`** block when **`required`** is **`true`**.

### 3.5.15

#### Enhancements

- **cidaas_social_provider, cidaas_custom_provider, cidaas_app:** Added Write-Only argument `client_secret_wo` (with companion `client_secret_wo_version`) as a secure alternative to `client_secret`. Values supplied via `client_secret_wo` are sent to cidaas on create and update but are not stored in plan or state; increment `client_secret_wo_version` to trigger an update. `client_secret` and `client_secret_wo` cannot be set together. On `cidaas_social_provider` and `cidaas_custom_provider`, `client_secret` is now Optional (one of `client_secret` or `client_secret_wo` is required). On `cidaas_app`, using `client_secret_wo` also disables cidaas server-side secret auto-generation — the value must be supplied by the user. Write-only arguments are supported in HashiCorp Terraform 1.11.0 and later.

### 3.5.14

#### Breaking changes

- **`cidaas_notifications_template_group`:** Removed **`copy_from_group_id`** and **`copy_locale_mappings`**. Locale copy and per-locale template deletion are handled by **`cidaas_notifications_template_group_locale`**. Group create no longer sends API `copy` (notification-srv may still seed locales from `default` per server rules).

**Upgrade steps:**

1. Remove **`copy_from_group_id`** and **`copy_locale_mappings`** from every `cidaas_notifications_template_group` block.
2. Add one **`cidaas_notifications_template_group_locale`** per locale (see [Migration: Template group locale copy](docs/guides/migration-notifications-template-group-locales.md)).
3. Set **`tg_type`** to match usage: **`cidaas`** for platform groups; **`developer`** for groups used with **`cidaas_notification_template`** / custom template types; **`reminder`** for reminder groups.
4. If templates already exist, **import** locale resources: `terraform import cidaas_notifications_template_group_locale.<name> <group_id>/<locale>`.

#### Enhancements

- **`cidaas_notifications_template_group_locale`:** New resource. **Create** copies one locale (`PUT` with `copy.locale[]`); **Read** checks `GET …/templatefilters`; **Destroy** bulk-deletes templates for that locale (blocked when `locale` equals the group's **`default_locale`** until you change **`default_locale`** on the group). **Import** id: `{group_id}/{locale}`. Extra API locales without a locale resource are not auto-removed.
### 3.5.13

#### Enhancements

- **cidaas_app:** Optional `allowed_native_clients` (set of client IDs) for the session transfer flow: web clients can list native app `client_id` values whose STTs are accepted. Maps to API field `allowed_native_clients` on `apps-srv/clients`.
- **All managed resources:** On refresh, if the remote object was deleted outside Terraform (e.g. HTTP **404** or equivalent not-found from the API), **Read** now removes the instance from state instead of failing the plan. The next apply can recreate the resource. Data sources are unchanged (they still error when the lookup fails).

### 3.5.12

#### Enhancements

- **cidaas_app:** Optional `allowed_native_clients` (set of client IDs) for the session transfer flow: web clients can list native app `client_id` values whose STTs are accepted. Maps to API field `allowed_native_clients` on `apps-srv/clients`.
- **All managed resources:** On refresh, if the remote object was deleted outside Terraform (e.g. HTTP **404** or equivalent not-found from the API), **Read** now removes the instance from state instead of failing the plan. The next apply can recreate the resource. Data sources are unchanged (they still error when the lookup fails).


### 3.5.11

#### Breaking changes

- **cidaas_notifications_template_group:** Removed the `enabled` argument; notification-srv template groups do not expose this field. Remove `enabled` from configuration if present.

### 3.5.10

#### Bug Fixes

- **cidaas_template:** `locale` validation now allows canonical BCP47 tags (e.g. `en-US`, `de-DE`, or language-only `en`) instead of an all-lowercase allow list, so region casing matches BCP47 and values such as `de-DE` are accepted. Import uses the same allow list. Update configurations that used all-lowercase locales (e.g. `en-us` → `en-US`).

### 3.5.9

#### Enhancements

- **cidaas_security_settings:** New resource for tenant fraud-detection settings via `fraud-detection-srv/settings` (HTTP **PATCH** for create/update, **GET** for read). It supports **`blocking_setting`**, **`repeated_login_blocking_mechanism`**, and **`rule_configuration`** with **`repeated_login_blocking_mechanism_enabled`** only. **Destroy** removes the resource from Terraform state only and does not reset remote settings. Other API fields under `ruleConfiguration`, `cspBotDetection`, and `repeatedFailedLoginAttempts` are not exposed on this resource.
- **OAuth scopes:** Managing this resource requires **`cidaas:fds_settings_read`** and **`cidaas:fds_settings_write`** on the Terraform client.
- **State merge:** After read/apply, state is merged with configuration so partial `.tf` matches apply/read (avoids “inconsistent result after apply” when the API returns more fields than configured). **PATCH** response handling and list **null** vs empty-set behavior are aligned with the API.
- **HTTP client:** **PATCH** responses with status **200** or **204** are treated as success.

### 3.5.8

#### Enhancements

- **cidaas_app:** Optional `hints` set under `group_role_restriction` for group verification response shaping. Allowed values: `groupIds`, `rolesOfGroup`, `allowedGroups` (API field `hints`). The sample module variable type includes optional `hints`.

#### Bug Fixes

- **Scope client:** `Get` and `Delete` no longer normalize scope keys to lowercase; the query parameter and path segment match the scope key passed in, so mixed- or upper-case canonical keys from cidaas are addressed correctly (see framework issue #1856).

### 3.5.7

#### Enhancements

- **Documentation:** Added the [Notification service (notification-srv)](docs/guides/notification_srv.md) guide (source: `templates/guides/notification_srv.md.tmpl`) covering legacy templates-srv vs notification-srv, `notifications_context_path`, scopes, use cases, API casing, and known limitations. Linked from the README notification section.
- **Examples:** Added examples for `cidaas_notification_template`, `cidaas_notifications_template_group`, `data.cidaas_notification_templates`, and `data.cidaas_notification_template_groups`; refreshed `cidaas_notification_template_type` examples to use lowercase `communication_methods` / `msg_formats` aligned with notification-srv JSON.
- **Registry docs:** Regenerated provider documentation (`go generate ./...`) including notification-srv resources and datasources; provider index template example provider version set to **3.5.7**.

#### Bug Fixes

- **Registration field:** Computed `base_data_type` is no longer unknown after apply for GROUPING fields when the API returns no base type; the provider sets empty string with fallbacks so state is always known.
- **App – `allow_guest_login_groups`:** `group_type` is now populated in state from the API.
- **App – MFA:** `time_interval_in_seconds` is stored as null when the API returns 0 or omits it, reducing plan drift.
- **`cidaas_notification_template_type`:** Sends a default **`owner`** of `client` on create/update when unset; **`communication_methods`** accept case-insensitive values matching notification-srv (`email`, `sms`, `ivr`, `push`) with **`ModifyPlan`** and state normalization to lowercase to avoid inconsistent plan/apply and API validation errors.

#### CI

- Added `.ci/lint/configs/golang/.golangci-standard.yml` so the shared GitLab lint template resolves the config.
- Fixed the lint:diff job in the pipeline.

### 3.5.6

#### Enhancements

- **Registration field data source:** The `cidaas_registration_field` data source now exposes the `enabled` attribute from the fieldsetup API, so you can filter and read whether each field is enabled.
- **Registration field resource:** Added support for remote field settings (GROUPING type) with `RemoteFieldSettings`, `ApiClientSetup`, `APIAccessSetup`, and auth detail types (APIKEY, TOTP, BASIC_AUTH, OAuth2). `RemoteSettings` is now part of `RegistrationFieldConfig`.

#### Bug Fixes

- **Registration field:** Removed the `is_group` attribute from the registration field data source and resource; it is no longer part of the fieldsetup/API contract.
- **Registration field / LocaleText:** The `Language` field has been removed from `LocaleText` and is no longer sent in registration field API payloads or set in the provider, aligning with the current cidaas backend.
- **Tests:** Updated test cases for registration fields and hosted page to match current API and schema behavior.

### 3.5.5

#### Enhancements

- **Registration field data source:** The list endpoint for registration fields now uses `fieldsetup-srv/graph/fields` instead of `registration-setup-srv/fields/list`, aligning with the current cidaas API.
- **Notification template type:** The cidaas client now exposes the template type service, enabling the `cidaas_notification_template_type` resource to manage template types via the provider.
- **Documentation:** Migration guide for classic `cidaas_template` / `cidaas_template_group` vs `cidaas_notification_template_type`: `docs/guides/migration-template-to-notification-template-type.md`.

#### Bug Fixes

- **Template type helper:** Corrected handling of `NewHTTPClient` return values (client and error) and updated all `MakeRequest` calls to pass `context.Context` as the first argument.
- **Hosted page:** Removed the `content` attribute from the `cidaas_hosted_page` resource and from the hosted page API payload. This attribute is no longer supported by the cidaas backend; existing configurations that set `content` should remove it to avoid schema errors.

#### Upgrade notes (3.5.4 → 3.5.5)

- If you use `cidaas_hosted_page` with a `content` attribute, remove that attribute from your configuration before upgrading. The provider schema no longer accepts `content`; after removing it, run `terraform plan` to confirm no other changes.

### 3.5.4

### Enhancements

- Added `group_type` attribute to the `group_role_restriction` in cidaas app resource.

### 3.5.3

### Enhancements

- Enhanced logging and error handling across all resources and data sources

### 3.5.2

### Enhancements

- Added `oauth_standard` attribute to the `cidaas_app` resource, allowing selection of the OAuth standard version between `OAuth2.0` and `OAuth2.1`.

### 3.5.1

### Bug Fixes

- Fixed issue where `redirect_uris`, `allowed_logout_urls`, and `grant_types` fields were not being properly set during resource import based on client type requirements.
- Changed `order` attribute from optional to computed-only as it is now automatically managed by the backend service and cannot be set or updated from the client side.
- vulnerability fix

### Security
- Upgraded Go toolchain from 1.21.0 to 1.24.4 for provider build to fix known standard library vulnerabilities

### 3.5.0
- Added context support for proper HTTP request cancellation and timeout handling
- Enhanced resource `cidaas_app` import to include all schema fields
- Added support for new fields in `cidaas_custom_provider` resource: `groups`, `pkce`, `auth_type`, `apikey_details`, `totp_details`, `cidaas_auth_details`

### 3.4.9

### Bug Fixes

- Added `is_group_login_selection_enabled` flag to the `cidaas_app` resource that removed earlier in v3.4.7, allowing to enable or disable group login selection

### 3.4.8

### Enhancements & Bug Fixes

- Added `enabled` flag to the `cidaas_template` resource, allowing to activate or deactivate a template
- Added support for custom values in `allow_login_with` attribute in `cidaas_app` resource
- Fixed handling of null set/list attributes in `cidaas_app` resource by sending empty arrays in API requests

### 3.4.7

### Enhancements & Bug Fixes

- Unsupported attributes removed from the `cidaas_app` resource schema. The following attributes were removed:

  - always_ask_mfa
  - editable
  - email_verification_required
  - enable_classical_provider
  - fds_enabled
  - is_group_login_selection_enabled
  - mobile_number_verification_required

- Fix provided for the issue in the `cidaas_webhook` resource where `placeholder` attribute was incorrectly rejecting valid values containing dashes (e.g. `test-apikey-placeholder`). It now correctly accepts placeholders using lowercase alphabets and dashes as intended.

- The attribute `group_type` is optional now in the `cidaas_user_groups` resource.

### 3.4.6

#### Enhancements

- The `cidaas_app` resource has been enhanced to behave more accurately based on the `client_type` attribute. With this update, Terraform configurations must now explicitly define values for all relevant attributes, as they are no longer treated as computed or automatically assigned defaults by the provider during resource creation.
For example, the `enabled` attribute was previously defaulted to `true` by the provider when creating an application. With this change, if you do not specify `enabled` in your configuration, the provider will omit it from the API request allowing the server to apply its own default behavior instead.
This ensures a more predictable and transparent configuration experience, aligning the provider behavior more closely with user intent and server-side defaults.

### 3.4.5

#### Bug Fixes

- The attribute `hosted_pages` in the resource `cidaas_hosted_page` has been updated to use an unordered list. This change resolves the issue where Terraform would incorrectly detect changes in the `hosted_pages` attribute, even when there were no actual modifications to the list, apart from reordering.

### 3.4.4

#### Bug Fixes

- Reduced plan time validation from resource `cidaas_template`.

### 3.4.3

#### Enhancements
- The `regex` field has been introduced in `field_definition` for the `cidaas_registration_field` resource (starting from cidaas version 3.101.5).
- This change **replaces** the `max_length` and `min_length` attributes **for `TEXT` and `URL` data types**.
- Instead of relying on fixed length constraints, validation for these field types will now be handled using **regular expressions (`regex`)**, providing more flexibility.

#### **Example of new regex-based validation**
```python
field_definition = {
    regex = "^(https?:\/\/)?([\da-z.-]+)\.([a-z.]{2,6})([\/\w .-]*)*\/?$"
}
```
#### Bug Fixes

- Fixed state consistency issues in resource `cidaas_template` and `cidaas_template_group`.

### 3.4.2

#### Bugfix

- The`cidaas_password_policy` resource has been updated to support the enhanced password policy introduced in cidaas version 3.100.x

### 3.4.1

#### Bugfix

- Attribute `scope_display_label` in resource **cidaas_custom_provider** marked optional. This fixes the state inconsistency issue when `scope_display_label` set to empty string.

### 3.4.0

#### Bugfix

- Resource **cidaas_social_provide** bug fix where empty `required_claims` and `optional_claims` provider plan error.
- Fixed the issue in **cidaas_app** resource  where the custom provider was not updated to an empty state after being removed from the config.


### 3.3.9

#### Enhancement

- Schema of the attribute `userinfo_fields` in resource **cidaas_custom_provider** changed to match cidaas api to suppoort external field key and default value.
- Attributses `amr_config` and `userinfo_source` are supported now in resource **cidaas_custom_provider**.

### 3.3.8

#### Bugfix

- Fixed issue where empty `consent_refs` array was being omitted from API requests causing state inconsistency

### 3.3.7

#### Enhancements

- Attribute basic_settings no longer supported in resource cidaas_app.

### 3.3.6

#### Enhancements

- Extend custom provider resource to support custom provider new api contract.

### 3.3.5

#### Enhancements

- Extend custom provider resource to support custom provider new api contract.

### 3.3.4

#### Enhancements

- Attribute `accept_roles_in_the_registration` added to the resource cidaas_app.

### 3.3.3

#### Enhancements

- `match_condition` and `filters` attributes in `group_role_restriction`(cidaas_app) are now required if `group_role_restriction` is declared in the configuration. This helps prevent misconfiguration.
- cidaas_app import now ignore empty `group_role_restriction` objects in the api response fixing schema mismarch issue.

### 3.3.2

#### Enhancements

- Enhanced validation on attributes processing_type and usage_type in resource cidaas_template

### 3.3.1

#### Enhanced Locale Support

The provider now includes additional locales `de-BE`, `id`, `zh-Hans` and `zh-Hant`.

### 3.3.0

#### Removed common_configs from resource app

The attribute `common_configs` is removed from the resource cidaas_app as we introduce [terraform-cidaas-app](https://github.com/cidaas/terraform-cidaas-app) module.

### 3.2.0

#### Addition of datasources

This release includes the following datasources:

- cidaas_consent
- cidaas_custom_provider
- cidaas_group_type
- cidaas_registration_field
- cidaas_role
- cidaas_scope_group
- cidaas_scope
- cidaas_social_provider
- cidaas_system_template_option

#### Additional attribute support in resource cidaas_app

The following attributes are added to the resource `cidaas_app`:

- require_auth_time
- enable_login_spi
- backchannel_logout_session_required
- suggest_verification_methods
- group_role_restriction
- basic_settings

#### Bug Fix

- Fixed the issue **Consent Not Found** when the name of the consent resource is in uppercase during update & destroy

### 3.1.2

#### Enhancements

- **Multiple Password Policy Support:** Password Policy resource changed to support multiple policies

### 3.1.1

#### Enhancements

- **Locale Support for Template Resource:** Added support for the **rm** & **rm-CH** language code in the template resource.

### 3.1.0

This release includes the below new resources

- cidaas_social_provider
- cidaas_password_policy
- cidaas_consent_group
- cidaas_consent
- cidaas_consent_version

Please find the readme [here](https://github.com/Cidaas/terraform-provider-cidaas/blob/v3.1.0/README.md) to explore more on the new resources.

### 3.0.5

#### Bug Fix

- password_policy_ref empty string or null values can be passed as "" when configured.
- Addressed the issue where computed attributes group_selection, login_spi & mobile_settings are not known after terraform apply, a default value {} is assigned to them.

### 3.0.4

#### Bug Fix

- **custom provider schema fix:** The issue with the sub attribute not aligning with the schema of the custom provider has been resolved.
- **app schema fix**: The app resource's list nested attributes are now updated to align with the cidaas API response.

### 3.0.3

#### Enhancements

- **Enhanced State Management:** Fixed state inconsistencies for attributes computed by cidaas APIs due to dependencies or API support changes.

### 3.0.2

#### Enhancements

- **Validation Update:** `group_id` is now required when `is_system_template=true` in resource cidaas_template.

#### Fixes

- **Validation Removed:** Removed the validation that checked the availability of template group by `group_id` in cidaas before creating a template as the api sometimes may fail to fetch the template group immediately after its creation.

### 3.0.1

#### Removed

- **URL Validation**: Removed strict URL validation that enforced URLs to start with `https://`.

### 3.0.0

This new release is based on Terraform Plugin Framework and is designed to be mostly backwards compatible with existing implementations. It offers several benefits including enhanced performance, improved debugging capabilities and streamlined development processes. Specific advantages include:

- **Simplified Resource Management**: More straightforward management through enhanced schemas.
- **Improved Error Handling**: Error handling has been revamped. Errors now includes suggestions that should assist you to manage your resources.
- **Enhanced Performance**: Custom plan-time validations across all the resources that helps faster plugin operations and dynamic provider configurations in resource app.

Despite these improvements, some breaking changes are present. Users need to be aware of the following modifications:

- **Schema Changes**: The new release includes changes in some of the existing schemas. Review and update schema definitions to align with the new structure.
- **Change in Import Statement**: The import statement of resource cidaas_template has been changed.
- **Resource Name Update**: The resource name of cidaas_user_group_category and cidaas_registration_page_field changed.

#### Additions

- A new resource `cidaas_template_group` has been added to support template groups,which are required for creating system templates.
- **SYSTEM** templates can now be created using the provider. Refer to the template section in the documentation for more details.
- Added support for internationalization in `cidaas_registration_field` and `cidaas_scope` with multi-language capabilities.
- `cidaas_registration_field` now supports all the datatypes that cidaas supports.

### 2.5.8

#### Additions

- resource cidaas_registration_page_field schema update to toggle `overwrite_with_null_value_from_social_provider` in SYSTEM fields.

### 2.5.7

#### Additions

- app_resource schema update to support `is_provider_visible` in customProviders, socialProviders & adProviders.

### 2.5.6

#### Additions

- `application_meta_data` added to support custom fields in cidaas_app resource.

- Validations added to prevent users from updating the locale and template_type of an existing `cidaas_template` state. This ensures data integrity and consistency.

- Enhanced error messages when users provide an incorrect locale,

- Updated the format of the `cidaas_template` ID from `templateKey_templateType` to `templateKey_templateType_locale`. Added validation checks for incorrect format types.

#### Bug Fixes

- Resolved the issue causing the error message `failed to unmarshal JSON body, EOF` during template deletion.

### 2.5.5

#### Bug Fixes

- Fixed the issue **subject can not be empty for template_key EMAIL** even though subject is available in the terraform config file

- app_key marked sensitive

- README updated with the instructions to guide Windows users to set env variables and scopes required for templates are added

### 2.5.4

#### Bug Fixes

- Fix added to address the issue where updating an existing cidaas_app without the `client_id` attribute throws error **client id is missing**.

- Improved error handling in terraform cidaas_app destroy. This solves the issue **invalid memory address or nil pointer dereference** while deleting client in cidaas.
