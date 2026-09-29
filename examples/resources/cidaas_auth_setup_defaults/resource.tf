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
