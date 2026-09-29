# Example: cidaas_app_configuration Resource (v4 Trustdesk)
#
# Configures client applications in Cidaas v4.
# Supports NON_INTERACTIVE (M2M), SINGLE_PAGE_WEBAPP, REGULAR_WEB, and NATIVE_MOBILE client types.
#
# Requires OAuth scopes: cidaas:apps_read, cidaas:apps_write, cidaas:apps_delete
# Provider: cidaas_version = "4.x"

resource "cidaas_app_configuration" "example" {
  client_name = "terraform-example-spa"
  client_type = "SINGLE_PAGE"
  enabled     = true

  grant_types    = ["authorization_code", "refresh_token"]
  response_types = ["code"]

  ownership_details = {
    company_name    = "Example Corp"
    company_address = "1 Example Way"
    company_website = "https://example.com"
  }

  scopes = {
    allowed_scopes = ["openid", "profile"]
    default_scopes = ["openid"]
  }

  # Per-app overrides. Omit a bool to inherit cidaas_auth_setup_defaults (tenant).
  # net_id is tenant-only (cidaas_auth_setup_defaults); login_spi is app-only.
  authentication_setup = {
    auto_login_after_register       = false
    register_with_login_information = false
    enable_password_less_auth       = true
    allow_user_level_multi_provider = true
    social_business_ids             = false
    login_spi = {
      enable_login_spi = false
    }
  }
}
