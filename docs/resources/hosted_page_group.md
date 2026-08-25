Manages a hosted page group via `POST/GET/DELETE /hostedpages-srv/hpgroup`. Requires `cidaas:hosted_pages_write`, `cidaas:hosted_pages_read`, `cidaas:hosted_pages_delete`.## Example Usage

```terraform
resource "cidaas_hosted_page_group" "example" {
  name           = "external-customer-group"
  default_locale = "en"
  group_owner    = "client"

  hosted_pages = [
    {
      hosted_page_id = "login"
      locale         = "en"
      url            = "https://example.com/login"
    },
    {
      hosted_page_id = "register"
      locale         = "en"
      url            = "https://example.com/register"
    }
  ]
}
```