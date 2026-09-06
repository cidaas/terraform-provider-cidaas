resource "cidaas_federation_provider" "example" {
  provider_name          = "federated_oauth2_provider"
  display_name           = "Federated OAuth2 Provider (v4)"
  standard_type          = "OAUTH2"
  client_id              = "v4-client-id"
  client_secret          = "v4-client-secret"
  authorization_endpoint = "https://auth.example.com/oauth/authorize"
  token_endpoint         = "https://auth.example.com/oauth/token"
  userinfo_endpoint      = "https://auth.example.com/oauth/userinfo"
  logo_url               = "https://cdn.example.com/logo.svg"
  domains                = ["example.com"]
  owner                  = "client"
}
