terraform {
  required_providers {
    cidaas = {
      source  = "hashicorp.com/Cidaas/cidaas"
      version = "3.0.0"
    }
  }
}

provider "cidaas" {
  base_url      = "https://cidaas.de"
  client_id     = var.cidaas_client_id
  client_secret = var.cidaas_client_secret
}
