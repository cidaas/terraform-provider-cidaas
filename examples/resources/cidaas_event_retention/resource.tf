# Example: cidaas_event_retention Resource (v4 Trustdesk)
#
# Sets how many days changelog / user activity events are kept, per event type.
# Requires the event retention feature (custom_reporting) on the tenant.
# Only listed event types are managed; removing one (or destroying the resource)
# restores its default daysToLive. See the computed `catalog` for configurable
# event types and their min/max bounds.

resource "cidaas_event_retention" "sample" {
  retention = {
    ACCOUNT_MODIFIED = 1000 # bounds 365..1095, default 730
    ACCOUNT_DELETED  = 45   # bounds 7..60, default 30
  }
}

output "event_retention_catalog" {
  value = cidaas_event_retention.sample.catalog
}
