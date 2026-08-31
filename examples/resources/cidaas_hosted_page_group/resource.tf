# Example: cidaas_hosted_page_group Resource (v4 Trustdesk)
#
# This resource manages hosted page groups in Cidaas Trustdesk (v4).
# It defines mappings of hosted page types (login, register) by locale
# allowing you to assign custom URLs per language and group.

resource "cidaas_hosted_page_group" "example" {
  name           = "external-customer-group"
  default_locale = "en"
  group_owner    = "client"

  hosted_pages = [
    {
      hosted_page_id = "login"
      locale         = "en"
      url            = "https://example.com/login"
    },
    {
      hosted_page_id = "register"
      locale         = "en"
      url            = "https://example.com/register"
    }
  ]
}
