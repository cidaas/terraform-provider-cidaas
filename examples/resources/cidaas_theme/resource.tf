resource "cidaas_theme" "example" {
  filename    = "custom-style-v1.css"
  css_content = <<-EOT
  body {
    font-family: sans-serif;
    background-color: #f5f5f5;
  }
  EOT
}
