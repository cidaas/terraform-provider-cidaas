# Look up a federation provider by its unique provider_name ...
data "cidaas_federation_provider" "google" {
  provider_name = "default_google"
}

# ... or by id.
data "cidaas_federation_provider" "by_id" {
  id = "3f453233-92d4-475b-b10e-813fbd476c50"
}

output "google_client_id" {
  value = data.cidaas_federation_provider.google.oauth2.client_id
}
