Manages custom locale translations via hostedpages-srv (`/hostedpages-srv/translations`). API JSON field is `translation` (map). Requires `cidaas:hosted_pages_*` scopes.## Example Usage

```terraform
resource "cidaas_translations" "example" {
  locale_id = "fr"
  enabled   = true
  translations = {
    "login.title"  = "Bienvenue"
    "login.submit" = "Envoyer"
  }
}
```