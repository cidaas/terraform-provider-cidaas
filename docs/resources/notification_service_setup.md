---
page_title: "cidaas_notification_service_setup Resource - cidaas"
subcategory: ""
description: |-
  Manages a communication provider service setup via notification-srv.
---

# cidaas_notification_service_setup (Resource)

Manages a **communication provider service setup** via **notification-srv** (`POST` / `PATCH` / `DELETE` `/{notifications_context_path}/servicesetups`). The tenant comes from your instance token — do **not** pass `saas_instance_id`.

**`status`** is computed from `GET` and reflects manual verification in service-desk (`in-progress` → `active`). Terraform does **not** call verify.

Pair with **`cidaas_notification_provider_config`** for credentials.

**Scopes:** `cidaas:service_setups_read`, `cidaas:service_setups_write`, `cidaas:service_setups_delete`.

**Update:** only **`name`** and **`description`** are sent on `PATCH`.

**Destroy:** deletes the remote setup when allowed by the API. If the setup was already removed outside Terraform (HTTP **404**), destroy succeeds and clears state. The API blocks delete while **`status = active`** — deactivate outside Terraform first, then destroy.

**Import:** `terraform import cidaas_notification_service_setup.NAME <service_setup_id>`

## Example Usage

```terraform
resource "cidaas_notification_service_setup" "twilio_sms" {
  name                  = "Twilio SMS"
  service_id            = "custom-twilio-sms"
  communication_methods = ["sms"]
}

resource "cidaas_notification_provider_config" "twilio_sms" {
  service_setup_id = cidaas_notification_service_setup.twilio_sms.id

  config_data_wo = jsonencode({
    commProvider = "custom-twilio-sms"
    commMethod   = "sms"
    schemaData = {
      accountSid = var.twilio_account_sid
      authToken  = var.twilio_auth_token
    }
  })
  config_data_wo_version = "1"
}
```

After `terraform apply`, verify the provider in **service-desk**, then run `terraform plan` — `status` should become `active` with no config change. Until then, `data.cidaas_notification_service_setups` will not list this setup (active-only).

## Schema

### Required

- `communication_methods` (Set of String) Communication methods: `email`, `sms`, `ivr`, `push`. Changing forces replacement.
- `name` (String) Human-readable name.
- `service_id` (String) Service id (`serviceDescInfo.serviceId`). Changing forces replacement.

### Optional

- `description` (String) Optional description.
- `has_remote_templates` (Boolean) Whether templates are remote. Default: `false`.
- `parent_service_setup_id` (String) Optional + computed. Omit to let the platform auto-fill the parent when applicable; value is stored after create/refresh.
- `service_category` (String) Optional + computed. Default: `comm_prov`.

### Read-Only

- `id` (String) Service setup `_id`.
- `status` (String) Setup status from the API (`in-progress`, `active`, `inactive`).
