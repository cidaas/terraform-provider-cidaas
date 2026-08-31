![Logo](https://raw.githubusercontent.com/cidaas/terraform-provider-cidaas/master/logo.jpg)

# Terraform Provider for cidaas (v3 & v4)

The Terraform provider for **cidaas** enables interaction with cidaas instances allowing infrastructure-as-code management and CRUD operations for applications, user configurations, scopes, roles, registration fields, hosted pages, and security settings across both **cidaas v3** and **cidaas v4 (Trustdesk)**.

---

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- Go >= 1.25 (for building from source)
- A cidaas **v3** or **v4** tenant and Non-Interactive OAuth client credentials (`client_id` and `client_secret`) with appropriate scope permissions.

---

## Getting Started & Authentication

To authenticate and authorize Terraform operations with cidaas and select the target tenant version (`v3.x` / `3.x` or `v4.x` / `4.x`), set your credentials and `CIDAAS_VERSION` via environment variables or CI/CD pipeline variables:

### Linux / macOS:
```bash
export TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID="your-client-id"
export TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET="your-client-secret"
export CIDAAS_VERSION="3.x" # or "v3.x", "4.x", "v4.x"
```

### Windows (PowerShell):
```powershell
$env:TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID="your-client-id"
$env:TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET="your-client-secret"
$env:CIDAAS_VERSION="3.x" # or "v3.x", "4.x", "v4.x"
```

### GitLab CI / CD Pipeline Variables:
You can define pipeline variables under **Settings $\rightarrow$ CI/CD $\rightarrow$ Variables**:
- `TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID`: `your-client-id`
- `TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET`: `your-client-secret`
- `CIDAAS_VERSION`: `3.x` (or `v3.x`, `4.x`, `v4.x`)

---

## Target Version Configuration (`CIDAAS_VERSION`)

You can specify the target cidaas major version (`3.x` / `v3.x` or `4.x` / `v4.x`) via environment variables, CI/CD pipeline variables, or directly in `provider.tf`.

### Precedence Hierarchy (Highest to Lowest):
1. **HCL Provider Attribute**: `cidaas_version` in `provider "cidaas"` block (`provider.tf`)
2. **Environment Variable**: `TERRAFORM_PROVIDER_CIDAAS_VERSION`
3. **Environment Variable**: `CIDAAS_VERSION`

> [!IMPORTANT]
> `cidaas_version` is **required**. If it is not set in `provider.tf` or via environment/CI variables, the provider halts with an explicit error.

Accepted version inputs are normalized automatically (`3`, `v3`, `3.x`, `v3.x` $\rightarrow$ `3.x`; `4`, `v4`, `4.x`, `v4.x` $\rightarrow$ `4.x`).

### Example Terraform Configuration (`provider.tf`):
```hcl
terraform {
  required_providers {
    cidaas = {
      source  = "Cidaas/cidaas"
      version = ">= 4.0.0-alpha.1"
    }
  }
}

provider "cidaas" {
  base_url       = "https://your-tenant.cidaas.de"
  cidaas_version = "3.x" # Optional: "3.x", "v3.x", "4.x", or "v4.x"
}
```

---

## Resource Inventory

### Trustdesk (v4 Only)
| Resource | API Endpoint | Description |
|----------|--------------|-------------|
| `cidaas_app_configuration` | `/app-srv/apps` | App configuration (appv3) |
| `cidaas_theme` | `/hostedpages-srv/themes` | Custom hosted page themes |
| `cidaas_translations` | `/hostedpages-srv/translations` | Hosted page localization strings |
| `cidaas_hosted_page_group` | `/hostedpages-srv/hpgroup` | Hosted page groups |
| `cidaas_hosted_page_layout` | `/hostedpages-srv/hosted-page-layouts` | Hosted page layouts |
| `cidaas_user_setup` | `/user-srv/usersetup` | Tenant user registration setup |
| `cidaas_suggest_verification_method` | `/verification-actions-srv/suggest-verification-configs` | Suggested verification methods |
| `cidaas_verification_options` | `/verification-actions-srv/verification-options` | Verification options |
| `cidaas_app` | *Deprecated stub* | Legacy appv1 resource (deprecated in favor of `cidaas_app_configuration`) |

### Shared Resources (v3 & v4)
| Resource | Description |
|----------|-------------|
| `cidaas_registration_field` | Registration field definitions & regex validators |
| `cidaas_role` | Tenant roles |
| `cidaas_user_groups` | User groups |
| `cidaas_scope` | OAuth scopes |
| `cidaas_scope_group` | Scope groups |
| `cidaas_password_policy` | Password policies |
| `cidaas_security_settings` | Tenant security settings |
| `cidaas_template` | Notification templates |
| `cidaas_notification_service_setup` | Notification service setup |

See [docs/](docs/) for detailed attribute schemas and documentation for each resource, and [examples/resources/](examples/resources/) for sample Terraform HCL configurations.

---

## Local Development

```bash
# Build the binary locally
make build

# Install locally into ~/.terraform.d/plugins
make install

# Run unit tests
make test

# Run acceptance tests against v3 or v4 tenant
CIDAAS_VERSION=3.x TF_ACC=1 BASE_URL=https://... make testacc
CIDAAS_VERSION=4.x TF_ACC=1 BASE_URL=https://... make testacc
```

---

## CI / CD Pipeline

The GitLab CI pipeline (`.gitlab-ci.yml`) runs parallel acceptance test matrices:
- `acceptance_test_v3`: `CIDAAS_VERSION=3.x`
- `acceptance_test_v4`: `CIDAAS_VERSION=4.x`

Both invoke `make test-ci`.

---

## License

See [LICENSE](LICENSE).

---

Crafted with ❤️ by the **cidaas Platform Team**.

