# This is a sample configuration for setting up a webhook with multiple authentication options.
# The available authentication types include apikey_config, totp_config, and cidaas_auth_config.

# When the auth_type is set to "APIKEY", the apikey_config is required, while totp_config and cidaas_auth_config are optional.
# These optional configurations can be removed if not needed. However, by including them, you can easily switch the auth_type 
# to other options by simply updating the auth_type value without needing to modify other parts of the configuration.

resource "cidaas_webhook" "sample_webhook" {
  auth_type = "APIKEY"
  url       = "https://cidaas.de/webhook-srv/webhook"
  events = [
    "ACCOUNT_MODIFIED"
  ]
  apikey_config = {
    key         = "api-key"
    placeholder = "test-apikey-placeholder"
    placement   = "query"
  }
  totp_config = {
    key         = "totp-key"
    placeholder = "test-totp-placeholder"
    placement   = "query"
  }
  cidaas_auth_config = {
    client_id = "ce90d6ba-9a5a-49b6-9a50-b8db759e9b90"
  }
}
