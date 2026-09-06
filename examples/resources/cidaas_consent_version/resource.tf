# Example: cidaas_consent_version Resource (v3.x)
#
# Manages versioned consent content and locale mappings (SCOPES or URL mode).
# Note for v4.x: Set cidaas_version = "3.x" in the provider block to manage v3 consents.

resource "cidaas_consent_group" "sample" {
  group_name  = "sample_consent_group"
  description = "sample description"
}

resource "cidaas_consent" "sample" {
  consent_group_id = cidaas_consent_group.sample.id
  name             = "sample_consent"
  enabled          = true
}

resource "cidaas_consent_version" "v1" {
  version         = 1
  consent_id      = cidaas_consent.sample.id
  consent_type    = "SCOPES"
  scopes          = ["openid", "profile"]
  required_fields = ["name"]
  consent_locales = [
    {
      content = "Consent version in German"
      locale  = "de"
    },
    {
      content = "Consent version in English"
      locale  = "en"
    }
  ]
}
