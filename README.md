# Terraform Provider for cidaas (v4)

Greenfield Terraform provider for **cidaas v4 (Trustdesk)**. Hosted pages Phase 1 + Phase 2 layout (#2415).

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- Go >= 1.25 (for building from source)
- A cidaas **v4** tenant and OAuth client with hosted pages / themes scopes

## Authentication

```bash
export TERRAFORM_PROVIDER_CIDAAS_CLIENT_ID="…"
export TERRAFORM_PROVIDER_CIDAAS_CLIENT_SECRET="…"
```

Provider block:

```hcl
provider "cidaas" {
  base_url = "https://your-tenant.cidaas.eu"
}
```

## Resources

| Resource | API |
|----------|-----|
| `cidaas_theme` | `/hostedpages-srv/themes` |
| `cidaas_translations` | `/hostedpages-srv/translations` |
| `cidaas_hosted_page_group` | `/hostedpages-srv/hpgroup` |
| `cidaas_hosted_page_layout` | `/hostedpages-srv/hosted-page-layouts` |

See [docs/](docs/) and [CHANGELOG.md](CHANGELOG.md).

## Local development

```bash
make build
make install   # installs 4.0.0-alpha.1 into ~/.terraform.d/plugins
make test      # unit tests
TF_ACC=1 BASE_URL=https://… make testacc
```

## CI

`acceptance_test` in [`.gitlab-ci.yml`](.gitlab-ci.yml) exports `CI_ID` / `CI_SECRET` / `BASE_URL` and runs `make test-ci`.

## License

See [LICENSE](LICENSE).
