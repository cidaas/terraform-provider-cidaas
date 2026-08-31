# Example: cidaas_translations Resource (v4 Trustdesk)
#
# Configures localization key-value translations per locale ID (e.g. `en`, `de`, `fr`).

resource "cidaas_translations" "example" {
  locale_id = "fr"
  enabled   = true
  translations = {
    "login.title"  = "Bienvenue"
    "login.submit" = "Envoyer"
  }
}
