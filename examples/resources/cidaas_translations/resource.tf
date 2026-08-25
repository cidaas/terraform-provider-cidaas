resource "cidaas_translations" "example" {
  locale_id = "fr"
  enabled   = true
  translations = {
    "login.title"  = "Bienvenue"
    "login.submit" = "Envoyer"
  }
}
