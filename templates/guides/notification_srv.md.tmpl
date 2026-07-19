---
page_title: "Notification service (notification-srv)"
description: |-
  How to use the cidaas Terraform provider with notification-srv: template groups, template types, templates, service setups, graph datasources, limitations, and legacy templates-srv.
---

# Notification service (notification-srv)

This guide describes **notification-srv** resources and datasources in the cidaas Terraform provider, how they differ from **legacy templates-srv**, and how to complete common workflows.

## Two template stacks

| Stack | Provider resources | Backend |
| --- | --- | --- |
| **Notification service** | `cidaas_notifications_template_group`, `cidaas_notifications_template_group_locale`, `cidaas_notification_template_type`, `cidaas_notification_template`, service setup / provider config, graph datasources | **notification-srv** under `/{notifications_context_path}/…` |
| **Legacy** | `cidaas_template_group`, `cidaas_template` | **templates-srv** (separate API paths) |

- Set optional provider argument **`notifications_context_path`** (default: `notifications-srv`) so all notification-srv clients use the same URL prefix. Legacy `cidaas_template` / `cidaas_template_group` **do not** use this setting.
- Prefer the notification-srv resources for **new** infrastructure as code.

## Authentication and scopes

- The provider uses **client credentials** from environment variables `TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID` and `TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET` (non-interactive app).
- Template APIs typically require **`cidaas:templates_read`**, **`cidaas:templates_write`**, and **`cidaas:templates_delete`**.
- Service setup / provider config require **`cidaas:service_setups_read`**, **`cidaas:service_setups_write`**, **`cidaas:service_setups_delete`**, and **`cidaas:provider_config_write`**. Exact enforcement is on notification-srv and your tenant — grant these on the client used by Terraform.

## Use cases and building blocks

### Custom template type, then templates in a group

1. Define **`cidaas_notification_template_type`** with `category = "custom"`, `template_key`, `description`, and `communication_methods` (see [Casing](#api-casing-and-terraform-state) below).
2. Create **`cidaas_notification_template`** rows per `group_id`, `template_key`, `communication_method`, `locale`, and `message_format`.
3. Use **`depends_on`** if Terraform cannot infer order.

See [examples/resources/cidaas_notification_template_type/resource.tf](https://github.com/Cidaas/terraform-provider-cidaas/blob/master/examples/resources/cidaas_notification_template_type/resource.tf) and [examples/resources/cidaas_notification_template/resource.tf](https://github.com/Cidaas/terraform-provider-cidaas/blob/master/examples/resources/cidaas_notification_template/resource.tf).

### Create a template group and manage locales

1. **`cidaas_notifications_template_group`** — group metadata only (`group_id`, **`tg_type`**, `description`, `default_locale`, optional `comm_setting_*`). **`tg_type`**: use **`cidaas`** for platform groups; **`developer`** when the same `group_id` is used with **`cidaas_notification_template`** / custom template types; **`reminder`** for reminders. Create does **not** send `copy`; notification-srv may still seed locales from `default` (tenant-specific, e.g. `en`, `de`, `de-DE`, `en-US`). Migrating from `copy_*` on the group: [Migration: Template group locale copy](migration-notifications-template-group-locales.md).
2. **`cidaas_notifications_template_group_locale`** — one resource per locale: **Create** copies templates (`PUT` with `copy.locale[]`); **Read** checks `GET …/templatefilters`; **Destroy** bulk-deletes templates for that locale. Cannot delete the group's `default_locale` until you change it on the group resource.
3. Extra API locales not covered by locale resources are **not** auto-removed (import or destroy a locale resource, or clean up in Admin).

Discover source locale codes: `GET …/templategroups/{copy_from_group_id}/templatefilters` before defining `copy_from_locale`.

Configure **`comm_setting_*`** on the group when you need per-channel service setup ids; resolve ids with **`data.cidaas_notification_service_setups`** (active only) or **`data.cidaas_notification_service_setup`** / the managed resource (any status).

See [examples/resources/cidaas_notifications_template_group/resource.tf](https://github.com/Cidaas/terraform-provider-cidaas/blob/master/examples/resources/cidaas_notifications_template_group/resource.tf) and [examples/resources/cidaas_notifications_template_group_locale/resource.tf](https://github.com/Cidaas/terraform-provider-cidaas/blob/master/examples/resources/cidaas_notifications_template_group_locale/resource.tf).

### Service setup and provider config (IaC)

Terraform manages communication providers through **notification-srv** APIs (`/servicesetups`, `/providerconfigs`). The tenant is taken from your instance token — do **not** pass `saas_instance_id`.

| API | Purpose |
| --- | --- |
| `GET` / `POST` / `PATCH` / `DELETE` `/servicesetups` | Service setup CRUD |
| `GET` / `POST` `/providerconfigs` | Provider config upsert |
| Verify endpoint | **Not used by Terraform** — verify in service-desk |

**Install order**

1. **`cidaas_notification_service_setup`** — metadata; `status` is computed (`in-progress` until you verify). Optional `parent_service_setup_id` may be auto-filled when omitted.
2. **`cidaas_notification_provider_config`** — credentials via `config_data_wo` + `schemaData` JSON (wizard shape). The provider injects `configData.id` from `service_setup_id` when omitted.
3. **Manual verify** in service-desk (or API outside Terraform).
4. **`terraform plan`** — refresh shows `status = active` with no config change. Only then does **`data.cidaas_notification_service_setups`** include the setup (list is **active-only**).

**Provider config JSON (`schemaData` mode):**

```hcl
config_data_wo = jsonencode({
  commProvider = "custom-twilio-sms"
  commMethod   = "sms"
  schemaData = {
    accountSid = var.twilio_account_sid
    authToken  = var.twilio_auth_token
  }
})
config_data_wo_version = "1"
```

Required top-level keys: `commProvider`, `commMethod`, `schemaData`. Field names inside `schemaData` must match the provider schema (obtain from the Admin / service-desk wizard for that communication provider). Write-only arguments require Terraform **≥ 1.11**.

**Advanced:** set `config_data` to the full stored `configData` from `GET` (import/export); secrets are stored in state.

**Secret rotation:** increment `config_data_wo_version` only; re-verify manually if the provider requires it.

**Troubleshooting**

- **Validation errors on create** — check `commProvider` matches the service setup, and `schemaData` field names match the provider schema.
- **status stuck `in-progress`** — complete verification in service-desk, then refresh. List datasources will not show the setup until `active`.
- **Destroy provider config** — Terraform removes state only; delete the service setup (when not `active`) to remove remote credentials. Active setups must be deactivated outside Terraform before delete.
- **Destroy service setup returns 404** — treated as success (already removed remotely).

See [examples/resources/cidaas_notification_service_setup/resource.tf](https://github.com/Cidaas/terraform-provider-cidaas/blob/master/examples/resources/cidaas_notification_service_setup/resource.tf).

### Discover existing templates or groups (graph API)

- **`data.cidaas_notification_templates`** — POST `graph/templates/` with a JSON **`graph_filter`** body.
- **`data.cidaas_notification_template_groups`** — POST `graph/templategroups/` with a JSON **`graph_filter`** body.

See [examples/datasources/cidaas_notification_templates.tf](https://github.com/Cidaas/terraform-provider-cidaas/blob/master/examples/datasources/cidaas_notification_templates.tf) and [examples/datasources/cidaas_notification_template_groups.tf](https://github.com/Cidaas/terraform-provider-cidaas/blob/master/examples/datasources/cidaas_notification_template_groups.tf).

### System template types (category `cidaas`)

- **System** template types are **pre-provisioned**; you usually **`terraform import`** them and only change allowed fields (for example **`custom_attributes`**), depending on resource logic.
- You **cannot** create a new system template type from scratch via Terraform the same way as a custom type.

### Legacy templates-srv

- Use **`cidaas_template`** and **`cidaas_template_group`** only for existing configurations that still target templates-srv.
- Do not mix legacy resources with notification-srv paths for the same logical workflow without understanding both APIs.

## API casing and Terraform state

- **notification-srv** JSON uses **lowercase** communication methods (`email`, `sms`, `ivr`, `push`).
- For **`cidaas_notification_template_type`**, the provider accepts **case-insensitive** `communication_methods` in configuration and **normalizes** them to lowercase in plan/state so Terraform stays consistent with the API (no spurious diffs after apply).
- **`cidaas_notification_template`** already expects lowercase **`communication_method`** and **`message_format`** (`html`, `text`, `media`) to match the API.

## Known limitations

- **Admin-only or UI-only** flows may exist that are not exposed as Terraform resources; this provider only covers what is implemented in code against the public HTTP APIs.
- **Graph filters** are passed through as JSON strings; invalid filter shapes fail at runtime with API errors—validate against notification-srv graph filter rules.
- **Import** identifiers are resource-specific (for example template document **`id`** for `cidaas_notification_template`); see each resource’s documentation.
- **Destroy** behavior for system or protected objects follows API rules; some resources may warn instead of deleting.
- **Service setup list** (`data.cidaas_notification_service_setups`) is **active-only**; poll with GET-by-id / managed resource for `in-progress` setups.

## Related registry documentation

- Provider schema: **`base_url`**, **`notifications_context_path`**
- Resources: **`cidaas_notifications_template_group`**, **`cidaas_notifications_template_group_locale`**, **`cidaas_notification_template_type`**, **`cidaas_notification_template`**, **`cidaas_notification_service_setup`**, **`cidaas_notification_provider_config`**
- Data sources: **`cidaas_notification_templates`**, **`cidaas_notification_template_groups`**, **`cidaas_notification_service_setups`**, **`cidaas_notification_service_setup`**

Run `go generate ./...` after changing schemas so the [Terraform Registry](https://registry.terraform.io/providers/cidaas/cidaas/latest/docs) docs stay in sync.
