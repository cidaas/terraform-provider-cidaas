# Example: cidaas_theme Resource (v4 Trustdesk)
#
# Manages custom CSS themes uploaded to Cidaas Trustdesk.
# The filename must end with `.css`.

resource "cidaas_theme" "example" {
  filename    = "custom-style-v1.css"
  css_content = <<-EOT
  body {
    font-family: sans-serif;
    background-color: #f5f5f5;
  }
  EOT
}
