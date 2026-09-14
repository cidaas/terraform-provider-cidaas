---
page_title: "Resource Dependency Order"
---

# Resource dependency order

Many cidaas resources reference IDs created by other resources. Express those links with Terraform references so apply order is correct.

## Typical Trustdesk (v4) flow

1. **Access building blocks** — `cidaas_role`, `cidaas_scope` / `cidaas_scope_group`, `cidaas_group_type`, `cidaas_user_groups`
2. **Application** — `cidaas_app_configuration` (often references scopes)
3. **Hosted pages** — `cidaas_hosted_page` → `cidaas_hosted_page_layout` / `cidaas_theme` / `cidaas_translations`
4. **Identity providers** — `cidaas_federation_provider`
5. **Consent & registration** — `cidaas_consent` → `cidaas_consent_version` / groups; `cidaas_registration_field`
6. **Notifications** — template types/groups → templates; `cidaas_notification_service_setup`
7. **Webhooks & verification** — `cidaas_webhook`, verification resources

## Example references

```hcl
resource "cidaas_scope" "profile" {
  scope_key             = "profile_example"
  required_user_consent = false
}

resource "cidaas_hosted_page" "default" {
  hosted_page_group_name = "default"
  default_locale         = "en-US"
}

resource "cidaas_hosted_page_layout" "login" {
  # reference the hosted page group name from the resource above
  # layout = { hosted_page_group = cidaas_hosted_page.default.hosted_page_group_name }
}
```

Exact attribute names differ per resource — use each resource’s schema page (for example [cidaas_scope](../resources/scope.md) or the [provider overview](../index.md)).

## Dependency Best Practices

Ensure dependent resources reference primary resources using standard Terraform HCL attributes (e.g. `cidaas_hosted_page.default.hosted_page_group_name` or `cidaas_scope.profile.scope_name`) so that implicit dependencies are naturally created.
