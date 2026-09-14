resource "cidaas_app" "sample_app" {
  client_name      = "sample_app"
  client_type      = "SINGLE_PAGE"
  company_name     = "Widas"
  company_address  = "01"
  company_website  = "https://example.com"
  client_id        = "sample_app_client_id"
  client_secret    = "sample_app_client_secret"
  allowed_scopes   = ["openid"]
  redirect_uris = [
    "https://example.com/callback"
  ]
  allowed_logout_urls = [
    "https://example.com/logout"
  ]
}
