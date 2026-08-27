Manages a Suggest Verification method via `verification-actions-srv/suggest-verification-configs`. Referenced by `cidaas_verification_options.verification_options.suggest_verification_method_id`. Requires scopes `cidaas:verification_read`, `cidaas:verification_write`, `cidaas:verification_delete`.## Example Usage

```terraform
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
```