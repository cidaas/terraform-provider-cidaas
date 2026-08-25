Uploads a custom CSS theme to hostedpages-srv (`POST /hostedpages-srv/themes`). Requires scopes `cidaas:themes_write`, `cidaas:themes_read`, `cidaas:themes_delete`.## Example Usage

```terraform
resource "cidaas_theme" "example" {
  filename    = "custom-style-v1.css"
  css_content = <<-EOT
  body {
    font-family: sans-serif;
    background-color: #f5f5f5;
  }
  EOT
}
```