---
page_title: "cidaas_notification_service_setup Resource - cidaas"
subcategory: ""
description: |-
  Manages a communication provider service setup via notification-srv (proxied to mplace-srv).
---

# cidaas_notification_service_setup (Resource)

Manages a **communication provider service setup** via **notification-srv** (`/servicesetups`). notification-srv proxies to **mplace-srv**; the tenant comes from your instance token — do **not** pass `saas_instance_id`.

**`status`** is computed from `GET` and reflects manual verification in service-desk (`in-progress` → `active`). Terraform does **not** call verify.

Pair with **`cidaas_notification_provider_config`** for credentials.

**Import:** `terraform import cidaas_notification_service_setup.NAME <service_setup_id>`

## Example Usage

```terraform
resource "cidaas_notification_service_setup" "twilio_sms" {
  name                  = "Twilio SMS"
  service_id            = "twilio-sms"
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

After `terraform apply`, verify the provider in **service-desk**, then run `terraform plan` — `status` should become `active` with no config change.

## Schema

### Required

- `communication_methods` (Set of String) Communication methods: `email`, `sms`, `ivr`, `push`.
- `name` (String) Human-readable name.
- `service_id` (String) Service id (`serviceDescInfo.serviceId`). Changing forces replacement.

### Optional

- `description` (String) Optional description.
- `has_remote_templates` (Boolean) Whether templates are remote. Default: `false`.
- `parent_service_setup_id` (String) Optional parent service setup id.
- `service_category` (String) Service category. Default: `comm_prov`.

### Read-Only

- `id` (String) Service setup `_id`.
- `status` (String) Setup status from the API (`in-progress`, `active`, `inactive`).
