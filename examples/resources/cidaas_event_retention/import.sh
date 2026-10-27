# cidaas_event_retention is a tenant singleton; the import ID is always "event_retention".
# Import adopts every configurable event type with its current daysToLive.

terraform import cidaas_event_retention.sample event_retention
