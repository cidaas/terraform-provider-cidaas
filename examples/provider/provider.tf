terraform {
  required_providers {
    cidaas = {
      source  = "Cidaas/cidaas"
      version = ">= 4.0.0-alpha.1"
    }
  }
}

provider "cidaas" {
  # Authenticate via client_id and client_secret, or environment variables:
  #   TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID
  #   TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET
  # Target version (required): HCL or TERRAFORM_PROVIDER_CIDAAS_VERSION / CIDAAS_VERSION
  base_url       = "https://your-tenant.cidaas.eu"
  cidaas_version = "4.x"
  client_id      = var.cidaas_client_id
  client_secret  = var.cidaas_client_secret
}
