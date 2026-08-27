Manages appv3 application configuration on cidaas v4 (Trustdesk) via `app-srv/apps`. Extdep content (user setup, hosted pages, verification options) is referenced by ID — create those resources separately and wire their IDs here.

Requires OAuth scopes `cidaas:apps_read`, `cidaas:apps_write`, and `cidaas:apps_delete`.

Replaces the deprecated [`cidaas_app`](app.md) resource.

## Example Usage

```terraform
resource "cidaas_app_configuration" "m2m" {
  client_name = "terraform-m2m-client"
  client_type = "NON_INTERACTIVE"
  enabled     = true

  grant_types = ["client_credentials"]

  ownership_details = {
    company_name    = "Example Corp"
    company_address = "1 Example Way"
    company_website = "https://example.com"
  }

  scopes = {
    allowed_scopes = ["openid", "profile"]
    default_scopes = ["openid"]
  }
}
```

## Import

Import by OAuth `client_id`:

```shell
terraform import cidaas_app_configuration.example <client_id>
```
