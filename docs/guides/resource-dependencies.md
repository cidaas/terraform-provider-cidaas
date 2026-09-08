---
page_title: "Resource Dependency Order"
---

# Resource dependency order

Many cidaas resources reference IDs created by other resources. Express those links with Terraform references so apply order is correct.

## Typical Trustdesk (v4) flow

1. **Access building blocks** — `cidaas_role`, `cidaas_scope` / `cidaas_scope_group`, `cidaas_group_type`, `cidaas_user_groups`
2. **Application** — `cidaas_app_configuration` (often references scopes)
3. **Hosted pages** — `cidaas_hosted_page_group` → `cidaas_hosted_page_layout` / `cidaas_theme` / `cidaas_translations`
4. **Identity providers** — `cidaas_social_provider`, `cidaas_custom_provider`, `cidaas_federation_provider`
5. **Consent & registration** — `cidaas_consent` → `cidaas_consent_version` / groups; `cidaas_registration_field`
6. **Notifications** — template types/groups → templates; `cidaas_notification_service_setup`
7. **Webhooks & verification** — `cidaas_webhook`, verification resources

## Example references

```hcl
resource "cidaas_scope" "profile" {
  scope_name        = "profile_example"
  scope_description = "Profile scope"
}

resource "cidaas_hosted_page_group" "default" {
  # group fields per resource docs
}

resource "cidaas_hosted_page_layout" "login" {
  # reference the hosted page group id from the resource above
  # hosted_page_group_id = cidaas_hosted_page_group.default.id
}
```

Exact attribute names differ per resource — use each resource’s schema page under [docs/resources](../resources/).

## Version notes

`cidaas_app`, `cidaas_hosted_page`, `cidaas_social_provider`, and `cidaas_custom_provider` support both `3.x` and `4.x` target versions. On `4.x` (Trustdesk) environments, native resources such as `cidaas_app_configuration` and `cidaas_hosted_page_layout` / `cidaas_theme` / `cidaas_translations` offer enhanced functionality, while legacy resources remain fully operational. Shared core resources (scopes, roles, group types, notification service setups, webhooks) follow identical dependency principles across both version targets.
