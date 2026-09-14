# Deprecated for new designs: prefer cidaas_notifications_template_group on v4+.
resource "cidaas_template_group" "example" {
  group_id = "example_group"

  email_sender_config = {
    from_email = "noreply@example.com"
    from_name  = "Example"
  }
}
