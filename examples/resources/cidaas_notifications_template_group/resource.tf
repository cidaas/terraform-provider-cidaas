# Example: notification-srv template group (metadata only).
# Locales: use cidaas_notifications_template_group_locale (see ../cidaas_notifications_template_group_locale/resource.tf).
resource "cidaas_notifications_template_group" "example_dev" {
  group_id       = "example_dev_group"
  tg_type        = "developer"
  description    = "Example developer template group created with Terraform (min 10 chars)."
  default_locale = "en"
}
