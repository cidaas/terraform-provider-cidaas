resource "cidaas_registration_field" "date" {
  data_type                                      = "DATE"
  field_key                                      = "sample_date_field"
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
  field_definition = {
    min_date          = "1899-12-31T18:38:50Z"
    max_date          = "2024-06-28T18:30:00Z"
    initial_date      = "2024-05-31T18:30:00Z"
    initial_date_view = "multi-year"
  }
}
