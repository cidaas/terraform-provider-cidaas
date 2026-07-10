---
page_title: "cidaas_notification_provider_config Resource - cidaas"
subcategory: ""
description: |-
  Stores provider credentials for a service setup via notification-srv (proxied to mplace provconfs).
---

# cidaas_notification_provider_config (Resource)

Stores **provider credentials** for a service setup via **notification-srv** (`POST /providerconfigs`). notification-srv proxies to mplace admin **`provconfs`** (upsert by `_id` = `service_setup_id`).

**Recommended:** use `config_data_wo` with wizard-shaped JSON (`commProvider`, `commMethod`, `schemaData`) so mplace validates and maps fields server-side.

-> **Note:** Write-Only argument `config_data_wo` is available to use in place of `config_data`. Write-only arguments are supported in HashiCorp Terraform 1.11.0 and later. [Learn more](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments).

**Verification** is manual (service-desk). Refresh `cidaas_notification_service_setup.status` after verify.

**Destroy** removes the resource from Terraform state only; notification-srv has no DELETE for provider configs.

**Import:** `terraform import cidaas_notification_provider_config.NAME <service_setup_id>`

## Example Usage

```terraform
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

Rotate secrets by incrementing `config_data_wo_version` and updating `config_data_wo`. Re-verify manually in service-desk if required.

## Schema

### Required

- `service_setup_id` (String) Service setup `_id`. Changing forces replacement.

### Optional

- `config_data` (String, Sensitive) Full `configData` JSON (stored in state). Exactly one of `config_data` or `config_data_wo` is required.
- `config_data_wo` (String, Sensitive, [Write-only](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments)) Write-Only `configData` JSON. Not stored in plan or state. Must be set with `config_data_wo_version`.
- `config_data_wo_version` (String) Increment to push an updated `config_data_wo` to the API.

### Read-Only

- `id` (String) Same as `service_setup_id`.
