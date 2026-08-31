# The password_echo system field confirms the password entered in the password field.
# Import the existing system field first, then manage match_with settings.
#
# terraform import cidaas_registration_field.password_echo password_echo

resource "cidaas_registration_field" "password_echo" {
  field_key  = "password_echo"
  field_type = "SYSTEM"
  data_type  = "PASSWORD"
  order      = 5

  internal  = false
  read_only = false
  required  = false
  claimable = true
  enabled   = true

  field_definition = {
    max_length = 20
    match_with = "password"
  }

  local_texts = [
    {
      locale         = "en-US"
      name           = "Confirm Password"
      max_length_msg = "Confirm Password cannot be more than 20 chars"
      match_with_msg = "Confirm Password Must Match with Password"
    },
    {
      locale         = "de-DE"
      name           = "Passwort bestätigen"
      max_length_msg = "Die Bestätigung des Passworts wird benötigt."
      required_msg   = "Die Bestätigung des Passworts wird benötigt."
      match_with_msg = "Die angegebenen Passwörter stimmen nicht überein"
    },
  ]
}
