# The data types "SELECT", "MULTISELECT" and "RADIO" are configured as shown below

resource "cidaas_registration_field" "select" {
  data_type                                      = "RADIO"
  field_key                                      = "sample_select_field"
  field_type                                     = "CUSTOM"
  internal                                       = false
  required                                       = true
  read_only                                      = false
  is_group                                       = false
  unique                                         = false
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
      attributes = [
        {
          key   = "test_key"
          value = "test_value"
        }
      ]
    },
    {
      locale       = "de-DE"
      name         = "Beispielfeld"
      required_msg = "Dieses Feld ist erforderlich"
      attributes = [
        {
          key   = "test_key"
          value = "test_value"
        }
      ]
    }
  ]
}
