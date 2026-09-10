# Example: cidaas_app_configuration Resource (v4 Trustdesk)
#
# Configures client applications in Cidaas v4.
# Supports NON_INTERACTIVE (M2M), SINGLE_PAGE_WEBAPP, REGULAR_WEB, and NATIVE_MOBILE client types.

resource "cidaas_app_configuration" "example" {
  client_name = "terraform-example-m2m"
  client_type = "NON_INTERACTIVE"
  enabled     = true

  grant_types = ["client_credentials"]

  ownership_details = {
    company_name    = "Example Corp"
    company_address = "1 Example Way"
    company_website = "https://example.com"
  }

  scopes = {
    allowed_scopes = ["openid", "profile"]
    default_scopes = ["openid"]
  }
}
