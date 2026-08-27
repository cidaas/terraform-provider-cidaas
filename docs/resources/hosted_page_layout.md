Manages a hosted page layout via `/hostedpages-srv/hosted-page-layouts`. Links branding (`layout`) to a hosted page group and theme. `group_id` is computed from the access token (not configurable). Requires `cidaas:hosted_pages_*` and `cidaas:themes_*` scopes plus admin group claims for writes.## Example Usage

```terraform
resource "cidaas_hosted_page_group" "example" {
  name           = "example-hpgroup"
  default_locale = "en"
  hosted_pages = [
    {
      hosted_page_id = "login"
      locale         = "en"
      url            = "https://your-tenant.cidaas.eu/login"
    }
  ]
}

resource "cidaas_theme" "example" {
  filename    = "example-theme.css"
  css_content = "body { background-color: #f5f5f5; }"
}

resource "cidaas_hosted_page_layout" "example" {
  description = "Example branding layout"

  layout = {
    hosted_page_group = cidaas_hosted_page_group.example.name
    theme             = cidaas_theme.example.filename
    primary_color     = "#1a0dab"
    accent_color      = "#0066cc"
    content_align     = "CENTER"
    media_type        = "IMAGE"
  }

  resources = {
    "default-hosted-pages-webapp" = {
      translation_set = "default"
      theme           = cidaas_theme.example.filename
    }
  }
}
```