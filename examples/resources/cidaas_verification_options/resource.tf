resource "cidaas_suggest_verification_method" "web" {
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

resource "cidaas_verification_options" "web" {
  name        = "web-verification-options"
  description = "Verification options for web apps"

  verification_options = {
    setting                        = "ALWAYS"
    allowed_methods                = ["EMAIL", "SMS"]
    use_default_password_policy    = true
    suggest_verification_method_id = cidaas_suggest_verification_method.web.id
  }
}
