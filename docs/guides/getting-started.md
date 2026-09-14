---
page_title: "Getting Started with the cidaas Provider"
---

# Getting Started

This guide walks through the minimum steps to manage a cidaas tenant with Terraform.

## 1. Prerequisites

- Terraform >= 1.0
- A cidaas **v3** or **v4 (Trustdesk)** tenant
- A non-interactive OAuth client (`client_id` / `client_secret`) with the scopes required by the resources you will manage

## 2. Configure credentials

```bash
export TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID="your-client-id"
export TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET="your-client-secret"
export CIDAAS_VERSION="4.x"
```

See [Authentication](authentication.md) for details.

## 3. Minimal configuration

```hcl
terraform {
  required_providers {
    cidaas = {
      source  = "Cidaas/cidaas"
      version = ">= 4.0.0"
    }
  }
}

provider "cidaas" {
  base_url       = "https://your-tenant.cidaas.eu"
  cidaas_version = "4.x"
}
```

## 4. Add a resource

Start with a shared resource such as a scope:

```hcl
resource "cidaas_scope" "sample" {
  security_level        = "CONFIDENTIAL"
  scope_key             = "terraform-sample-scope"
  required_user_consent = false
}
```

Copy-paste HCL and full attribute docs are on each resource page. Start from the [provider overview](../index.md) or [cidaas_scope](../resources/scope.md).

## 5. Plan and apply

```bash
terraform init
terraform plan
terraform apply
```

## Next steps

- [Authentication](authentication.md)
- [Provider configuration](version-targeting.md)
- [Cidaas v3 to v4 migration guide](v3-to-v4-migration.md)
- [Resource dependency order](resource-dependencies.md)
