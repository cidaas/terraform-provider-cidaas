resource "cidaas_registration_field" "text" {
  data_type                                      = "TEXT"
  field_key                                      = "sample_text_field"
  field_type                                     = "CUSTOM"  // CUSTOM and SYSTEM, SYSTEM can not be created but modified
  internal                                       = true      // Default: false
  required                                       = true      // Default: false
  read_only                                      = true      // Default: false
  is_group                                       = false     // Default: false
  unique                                         = true      // Default: false
  overwrite_with_null_value_from_social_provider = false     // Default: true
  is_searchable                                  = true      // Default: true
  enabled                                        = true      // Default: true
  claimable                                      = true      // Default: true
  order                                          = 1         // Default: 1
  parent_group_id                                = "DEFAULT" // Default: DEFAULT
  scopes                                         = ["profile"]
  local_texts = [
    {
      locale         = "en-US"
      name           = "Sample Field"
      required_msg   = "The field is required"
      max_length_msg = "Maximum 99 chars allowed"
      min_length_msg = "Minimum 99 chars allowed"
    },
    {
      locale         = "de-DE"
      name           = "Beispielfeld"
      required_msg   = "Dieses Feld ist erforderlich"
      max_length_msg = "DE maximum 99 chars allowed"
      min_length_msg = "DE minimum 10 chars allowed"
    }
  ]
  field_definition = {
    regex = "^.{10,100}$"
  }
}
