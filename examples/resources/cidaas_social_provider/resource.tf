variable "google_client_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "cidaas_social_provider" "sample" {
  name          = "Sample Social Provider"
  provider_name = "google"
  enabled       = true
  client_id     = "8d789b3d-b312"

  # Write-Only: not stored in plan or state. Increment client_secret_wo_version to trigger an update.
  client_secret_wo         = var.google_client_secret
  client_secret_wo_version = "1"

  # Alternative: client_secret = "96ae-ea2e8d8e6708" (stored in the state file).

  scopes                   = ["profile", "email"]
  enabled_for_admin_portal = true
  claims = {
    required_claims = {
      user_info = ["name"]
      id_token  = ["phone_number"]
    }
    optional_claims = {
      user_info = ["website"]
      id_token  = ["street_address"]
    }
  }
  userinfo_fields = [
    {
      inner_key       = "sample_custom_field"
      external_key    = "external_sample_cf"
      is_custom_field = true
      is_system_field = false
    },
    {
      inner_key       = "sample_system_field"
      external_key    = "external_sample_sf"
      is_custom_field = false
      is_system_field = true
    }
  ]
}
