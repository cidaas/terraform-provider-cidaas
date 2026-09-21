variable "cidaas_client_id" {
  type        = string
  description = "cidaas client ID for the Terraform provider (or set TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID)."
}

variable "cidaas_client_secret" {
  type        = string
  sensitive   = true
  description = "cidaas client secret for the Terraform provider (or set TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET)."
}
