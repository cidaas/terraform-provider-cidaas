resource "cidaas_user_setup" "sample" {
  name        = "developer-registration-setup"
  description = "Registration setup for developer clients"

  user_setup = {
    allowed_fields                    = ["email", "given_name", "family_name", "mobile_number"]
    required_fields                   = ["email", "given_name", "family_name"]
    allow_login_with                  = ["EMAIL", "MOBILE"]
    enable_deduplication              = true
    validate_phone_number             = true
    validate_email                    = true
    auto_activate_user                = false
    allow_disposable_email            = false
    accept_roles_in_the_registration  = false
    send_welcome_notification         = true
    birthdate_as_date                 = true
    communication_medium_verification = "email_verification_required"
    auto_confirm_communication_method = ["email"]
    verification_for_medium           = ["mobile_number"]

    operations_allowed_groups = [
      {
        group_id      = "7db543cd-810a-48d9-a78b-59bc65cda894"
        roles         = ["DEVELOPER"]
        default_roles = ["USER"]
      }
    ]
  }

  consent_refs = [
    "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"
  ]
}
