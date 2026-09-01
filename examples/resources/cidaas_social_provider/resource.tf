resource "cidaas_social_provider" "google" {
  provider_name            = "google"
  name                     = "Google Login"
  client_id                = "google-client-id.apps.googleusercontent.com"
  client_secret            = "google-client-secret"
  enabled                  = true
  enabled_for_admin_portal = false
  scopes                   = ["openid", "email", "profile"]
  owner                    = "client"
}
