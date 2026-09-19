---
page_title: "Authenticating the cidaas Provider"
---

# Authentication

The provider authenticates with a non-interactive OAuth client against your cidaas tenant.

## Required credentials

Credentials can be configured directly in the `provider "cidaas"` block (`client_id` and `client_secret`) or set via environment variables.

| HCL Attribute | Environment Variable Fallback | Purpose | Sensitive |
|---------------|-------------------------------|---------|-----------|
| `client_id` | `TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID` | OAuth client ID | No |
| `client_secret` | `TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET` | OAuth client secret | Yes (`Sensitive: true`) |

Provider HCL block attributes take precedence over environment variables if specified. Note that `client_secret` is marked `Sensitive: true` in the provider schema to prevent leaking in logs and state diffs.

## Version selection (also required)

| Source | Precedence |
|--------|------------|
| `cidaas_version` in the `provider` block | Highest |
| `TERRAFORM_PROVIDER_CIDAAS_VERSION` | |
| `CIDAAS_VERSION` | Lowest among these |

One of the above must be set. Accepted values normalize to `3.x` or `4.x` (for example `3`, `v3`, `3.x`, `4.x`, `v4.x`).

## Least privilege

Grant only the scopes listed on each resource documentation page (for example `cidaas:scopes_read` / `cidaas:scopes_write` for `cidaas_scope`). Using an admin client with all scopes works for evaluation but is not recommended for production pipelines.

## Example

### HCL Configuration (Recommended)

```hcl
provider "cidaas" {
  base_url      = "https://your-tenant.cidaas.eu"
  cidaas_version = "4.x"
  client_id     = var.cidaas_client_id
  client_secret = var.cidaas_client_secret
}
```

### Environment Variable Configuration (Fallback)

```bash
export TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID="..."
export TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET="..."
```
