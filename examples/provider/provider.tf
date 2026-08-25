terraform {
  required_providers {
    cidaas = {
      source  = "Cidaas/cidaas"
      version = ">= 4.0.0-alpha.1"
    }
  }
}

provider "cidaas" {
  # Authenticate with:
  #   TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID
  #   TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET
  base_url = "https://your-tenant.cidaas.eu"
}
