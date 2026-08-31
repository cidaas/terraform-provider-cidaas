# One locale per resource: copy on create, bulk-delete templates on destroy.
# Discover source locale codes on copy_from_group_id via GET …/templategroups/{id}/templatefilters.
resource "cidaas_notifications_template_group" "example_dev" {
  group_id       = "example_dev_group"
  tg_type        = "developer"
  description    = "Example developer template group created with Terraform (min 10 chars)."
  default_locale = "en"
}

resource "cidaas_notifications_template_group_locale" "en" {
  group_id           = cidaas_notifications_template_group.example_dev.group_id
  locale             = "en"
  copy_from_group_id = "default"
  copy_from_locale   = "en"
}

resource "cidaas_notifications_template_group_locale" "de" {
  group_id           = cidaas_notifications_template_group.example_dev.group_id
  locale             = "de"
  copy_from_group_id = "default"
  copy_from_locale   = "en"
}

resource "cidaas_notifications_template_group_locale" "en_us" {
  group_id           = cidaas_notifications_template_group.example_dev.group_id
  locale             = "en-US"
  copy_from_group_id = "default"
  copy_from_locale   = "en"
}
