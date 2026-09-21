variable "custom_provider_client_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "cidaas_custom_provider" "sample" {
  standard_type          = "OAUTH2"
  authorization_endpoint = "https://cidaas.de/authz-srv/authz"
  token_endpoint         = "https://cidaas.de/token-srv/token"
  provider_name          = "terraform-sample"
  display_name           = "Terraform"
  logo_url               = "https://cidaas.de/logo"
  userinfo_endpoint      = "https://cidaas.de/users-srv/userinfo"
  scope_display_label    = "terraform sample scope display name"
  client_id              = "acb-4a6b-9777-8a64abe6af"

  # Write-Only: not stored in plan or state. Increment client_secret_wo_version to trigger an update.
  client_secret_wo         = var.custom_provider_client_secret
  client_secret_wo_version = "1"

  # Alternative: client_secret = "zcb-4a6b-9777-8a64abe6ay" (stored in the state file).

  domains = ["cidaas.de", "cidaas.org"]

  scopes = [
    {
      recommended = true
      required    = true
      scope_name  = "email"
    },
    {
      recommended = true
      required    = true
      scope_name  = "openid"
    },
  ]

  userinfo_fields = {
    family_name        = { "ext_field_key" = "cp_family_name" }
    address            = { "ext_field_key" = "cp_address" }
    birthdate          = { "ext_field_key" = "01-01-2000" }
    email              = { "ext_field_key" = "cp@cidaas.de" }
    email_verified     = { "ext_field_key" = "email_verified", "default" = false }
    gender             = { "ext_field_key" = "male" }
    nickname           = { "ext_field_key" = "nickname" }
    given_name         = { "ext_field_key" = "cp_given_name" }
    locale             = { "ext_field_key" = "cp_locale" }
    middle_name        = { "ext_field_key" = "cp_middle_name" }
    mobile_number      = { "ext_field_key" = "100000000" }
    phone_number       = { "ext_field_key" = "10000000" }
    picture            = { "ext_field_key" = "https://cidaas.de/image.jpg" }
    preferred_username = { "ext_field_key" = "cp_preferred_username" }
    profile            = { "ext_field_key" = "cp_profile" }
    updated_at         = { "ext_field_key" = "01-01-01" }
    website            = { "ext_field_key" = "https://cidaas.de" }
    zoneinfo           = { "ext_field_key" = "cp_zone_info" }
    custom_fields = {
      zipcode         = "123456"
      alternate_phone = "1234567890"
    }
  }
}
