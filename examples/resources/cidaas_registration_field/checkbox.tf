resource "cidaas_registration_field" "checkbox" {
  data_type                                      = "CHECKBOX"
  field_key                                      = "sample_checkbox_field"
  field_type                                     = "CUSTOM"
  internal                                       = true
  required                                       = true
  read_only                                      = true
  is_group                                       = false
  unique                                         = true
  overwrite_with_null_value_from_social_provider = false
  is_searchable                                  = true
  enabled                                        = true
  claimable                                      = true
  order                                          = 1
  parent_group_id                                = "DEFAULT"
  scopes                                         = ["profile"]
  local_texts = [
    {
      locale       = "en-US"
      name         = "Sample Field"
      required_msg = "The field is required"
    },
    {
      locale       = "de-DE"
      name         = "Beispielfeld"
      required_msg = "Dieses Feld ist erforderlich"
    }
  ]
}
