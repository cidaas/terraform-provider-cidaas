# Example: cidaas_consent_group Resource (v3.x)
#
# Categorizes related consent items under a common group identifier.
# Note for v4.x: Set cidaas_version = "3.x" in the provider block to manage v3 consents.

resource "cidaas_consent_group" "sample" {
  group_name  = "sample_consent_group"
  description = "sample description"
}
