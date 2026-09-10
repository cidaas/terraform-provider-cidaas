---
page_title: "Authenticating the cidaas Provider"
---

# Authentication

The provider authenticates with a non-interactive OAuth client against your cidaas tenant.

## Required credentials

| Variable | Purpose |
|----------|---------|
| `TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID` | OAuth client ID |
| `TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET` | OAuth client secret |

Both must be set in the environment (or CI/CD variables). They are **not** provider HCL attributes.

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

```hcl
provider "cidaas" {
  base_url       = "https://your-tenant.cidaas.eu"
  cidaas_version = "4.x"
}
```

```bash
export TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID="..."
export TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET="..."
```
