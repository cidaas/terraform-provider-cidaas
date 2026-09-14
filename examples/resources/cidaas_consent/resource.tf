# Example: cidaas_consent Resource (v3.x)
#
# Manages consent items associated with a cidaas_consent_group.
# Note for v4.x: Set cidaas_version = "3.x" in the provider block to manage v3 consents.

resource "cidaas_consent_group" "sample" {
  group_name  = "customer_terms"
  description = "Customer Terms of Service Group"
}

resource "cidaas_consent" "sample" {
  consent_group_id = cidaas_consent_group.sample.id
  name             = "sample_consent"
  enabled          = true
}
