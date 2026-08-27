resource "cidaas_suggest_verification_method" "sample" {
  name        = "web-suggest-methods"
  description = "Standard suggest flow for interactive web apps"

  suggest_verification_method = {
    mandatory_config = {
      methods = ["EMAIL"]
      range   = "ONEOF"
    }
    optional_config = {
      methods = ["TOTP"]
    }
    skip_duration_in_days = 7
  }
}
