# System method (EMAIL, SMS, TOTP, ...): it already exists on the tenant. Terraform adopts it and manages
# active, locale_text and encryption_config. Destroying the resource only removes it from state.
resource "cidaas_verification_method" "email" {
  verification_type = "EMAIL"
  active            = true

  # Optional. The API replaces the whole list; when omitted, the current texts are kept.
  locale_text = [
    { locale = "en", name = "Email", description = "Verify with a code sent by email" },
    { locale = "de", name = "E-Mail", description = "Bestätigung mit einem Code per E-Mail" },
  ]
}

# Disable a system method.
resource "cidaas_verification_method" "pattern" {
  verification_type = "PATTERN"
  active            = false
}

# System method with an existing encryption configuration on the tenant: only enabled and mode can change.
resource "cidaas_verification_method" "sms" {
  verification_type = "SMS"

  encryption_config = {
    enabled = true
    mode    = "REQUIRED"
  }
}

# Pluggable method (owner CLIENT): created and deleted by Terraform.
resource "cidaas_verification_method" "custom_otp" {
  verification_type = "CUSTOM_OTP"
  active            = true

  url       = "https://mfa.example.com/verify"
  auth_type = "Bearer"
  cidaas_auth_details = {
    client_id = "11111111-2222-3333-4444-555555555555"
  }

  factor     = "POSSESSION"
  category   = "code"
  icon       = "https://cdn.example.com/custom-otp.svg"
  amr_values = ["custom_otp"]

  locale_text = [
    { locale = "en", name = "Custom OTP", description = "One-time code from the custom MFA service" },
  ]
}
