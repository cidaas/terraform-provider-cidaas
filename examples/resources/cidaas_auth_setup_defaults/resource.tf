# Tenant Default Authentication Setup (Trustdesk / v4)
# Maps to GET/PUT {base_url}/app-srv/apps/auth-setup-defaults
#
# Requires:
#   - cidaas_version = "4.x"
#   - OAuth scopes: cidaas:apps_read, cidaas:apps_write
#   - Admin / app-manager roles for Trustdesk Default Authentication Setup
#
# Singleton: id is always "default". Destroy removes state only (does not delete tenant defaults).
# Per-app overrides: cidaas_app_configuration.authentication_setup (login_spi is app-only; net_id is tenant-only).

resource "cidaas_auth_setup_defaults" "tenant" {
  name                            = "default"
  description                     = "Trustdesk default authentication setup"
  auto_login_after_register       = false
  register_with_login_information = false
  enable_password_less_auth       = true
  allow_user_level_multi_provider = true
  social_business_ids             = false
  net_id                          = false
}
