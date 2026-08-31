resource "cidaas_registration_field" "consent" {
  data_type                                      = "CONSENT"
  field_key                                      = "sample_consent_field"
  field_type                                     = "CUSTOM"
  internal                                       = false
  required                                       = true
  read_only                                      = false
  is_group                                       = false
  unique                                         = false
  overwrite_with_null_value_from_social_provider = true
  is_searchable                                  = true
  enabled                                        = true
  claimable                                      = true
  order                                          = 1
  parent_group_id                                = "DEFAULT"
  scopes                                         = ["profile"]
  local_texts = [
    {
      locale       = "en-US"
      name         = "sample_consent_field"
      required_msg = "The field is required"
      consent_label = {
        label      = "test",
        label_text = "test label text"
      }
    }
  ]
  consent_refs = ["3ea83b3f-de34-4bea-aebd-bcaac1fa1be5"]
}
