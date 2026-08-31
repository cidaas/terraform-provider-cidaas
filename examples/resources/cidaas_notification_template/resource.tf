# Example: one notification template row (notification-srv).
# Create a matching cidaas_notification_template_type first, then reference its template_key here.
resource "cidaas_notification_template" "welcome_email_en" {
  group_id             = "default"
  template_key         = "WELCOME_USER"
  communication_method = "email"
  locale               = "en"
  message_format       = "html"
  owner                = "client"
  description          = "Welcome email template managed by Terraform (example)."
  subject              = "Welcome"
  content              = "<p>Welcome to our service.</p>"
  enabled              = true
}
