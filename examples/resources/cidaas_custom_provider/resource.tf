resource "cidaas_custom_provider" "example" {
  provider_name          = "custom_oidc_provider"
  display_name           = "Custom OpenID Connect Provider"
  standard_type          = "OIDC"
  client_id              = "your-client-id"
  client_secret          = "your-client-secret"
  authorization_endpoint = "https://idp.example.com/oauth2/v1/authorize"
  token_endpoint         = "https://idp.example.com/oauth2/v1/token"
  userinfo_endpoint      = "https://idp.example.com/oauth2/v1/userinfo"
  logo_url               = "https://cdn.example.com/logo.png"
  domains                = ["example.com"]
  pkce                   = true
  userinfo_source        = "USERINFOENDPOINT"
  owner                  = "client"
}
