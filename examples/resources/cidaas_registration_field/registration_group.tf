resource "cidaas_registration_field" "registration_group" {
  data_type                                      = "TEXT"
  field_key                                      = "sample_group"
  field_type                                     = "CUSTOM"
  internal                                       = true
  required                                       = false
  read_only                                      = true
  is_group                                       = true
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
      locale = "en-US"
      name   = "Sample Field"
    },
    {
      locale = "de-DE"
      name   = "Beispielfeld"
    }
  ]
}
