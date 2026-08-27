# Deprecated legacy app resource (v4 provider).

The `cidaas_app` resource is **deprecated** on cidaas v4 (Trustdesk). Use [`cidaas_app_configuration`](app_configuration.md) instead.

This provider registers `cidaas_app` only so Terraform shows a deprecation warning when legacy HCL is present. Create/update/read/delete operations fail with a migration message — migrate state to `cidaas_app_configuration`.

## Migration

1. Replace `resource "cidaas_app"` with `resource "cidaas_app_configuration"` in your HCL.
2. Map fields to the appv3 schema (see [app_configuration.md](app_configuration.md)).
3. Remove legacy state: `terraform state rm cidaas_app.<name>`.
4. Apply the new resource (or import by `client_id` if the app already exists).
