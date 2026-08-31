resource "cidaas_scope" "sample" {
  security_level        = "CONFIDENTIAL"
  scope_key             = "terraform-sample-scope"
  required_user_consent = false
  group_name            = []
  localized_descriptions = [
    {
      title       = "Cidaas Scope Tunisia Title"
      locale      = "ar-TN"
      description = "This is scope in local ar-TN"
    },
    {
      title       = "Cidaas Scope German Title"
      locale      = "de-DE"
      description = "This is scope in local de-DE"
    },
    {
      title       = "Cidaas Scope India Title"
      locale      = "en-IN"
      description = "This is scope in local en-IN"
    }
  ]
}
