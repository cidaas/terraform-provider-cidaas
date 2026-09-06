# Example: cidaas_hosted_page_layout Resource (v4 Trustdesk)
#
# This resource configures hosted page layouts, linking a hosted_page_group
# with a theme CSS file, custom branding colors, and translation sets.

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
