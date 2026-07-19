---
page_title: "cidaas_notification_service_setup Data Source - cidaas"
subcategory: ""
description: |-
  Reads a single service setup from notification-srv GET /servicesetups/{id} (any status).
---

# cidaas_notification_service_setup (Data Source)

Reads a single **service setup** from notification-srv `GET /{notifications_context_path}/servicesetups/{id}`.

Unlike **`data.cidaas_notification_service_setups`** (active-only list), this lookup returns the setup in **any** status — including **`in-progress`** — so you can poll `status` after create or before wiring template-group `comm_setting_*`.

**Scopes:** `cidaas:service_setups_read`.

## Example Usage

```terraform
data "cidaas_notification_service_setup" "twilio" {
  service_setup_id = cidaas_notification_service_setup.twilio_sms.id
}

output "twilio_status" {
  value = data.cidaas_notification_service_setup.twilio.status
}
```

## Schema

### Required

- `service_setup_id` (String) Service setup `_id` to read.

### Read-Only

- `description` (String) Description.
- `has_remote_templates` (Boolean) Whether templates are remote.
- `id` (String) Datasource instance id (UUID).
- `name` (String) Human-readable name.
- `parent_service_setup_id` (String) Parent setup id when applicable.
- `service_category` (String) Service category from `serviceDescInfo`.
- `service_id` (String) Service id from `serviceDescInfo`.
- `status` (String) Setup status (`in-progress`, `active`, `inactive`, …).
