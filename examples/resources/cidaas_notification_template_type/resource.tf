// Example 1: System Template Type (pre-provisioned, only custom_attributes can be modified)
// System template types are managed by cidaas and cannot be created via Terraform.
// You can import existing system template types and only modify their custom_attributes.
resource "cidaas_notification_template_type" "verify_user_system" {
  template_key = "VERIFY_USER" # System template type (pre-provisioned by cidaas)

  description = "Verification template type for user verification" # Read-only, but required for import

  communication_methods = ["email", "sms"] # Read-only for system types; lowercase matches notification-srv JSON

  // Only custom_attributes can be modified for system template types
  custom_attributes = {
    "customFields.company_name"     = "allowed"
    "customFields.insurance_number" = "required"
  }
}

// Example 2: Custom Template Type (fully manageable via Terraform)
resource "cidaas_notification_template_type" "custom_notification" {
  template_key = "CUSTOM_NOTIFICATION"
  category     = "custom"
  description  = "Custom notification template type for special notifications"

  communication_methods = ["email", "sms", "push"]
  processing_types      = ["CODE", "LINK", "GENERAL"]
  usage_types           = ["GENERAL", "VERIFICATION_CONFIGURATION"]
  verification_types    = ["EMAIL", "SMS"]

  deactivatable = true

  system_attributes = {
    "code"        = "required"
    "name"        = "required"
    "email"       = "allowed"
    "expiry_time" = "allowed"
  }

  custom_attributes = {
    "company_name"  = "required"
    "support_email" = "allowed"
    "customer_id"   = "allowed"
  }

  context_attributes = {
    "locale"   = "required"
    "timezone" = "allowed"
  }

  template_group_ids = ["default", "custom_group"]
  msg_formats        = ["html", "text"]
}

// Example 3: Minimal Custom Template Type
resource "cidaas_notification_template_type" "simple_notification" {
  template_key = "SIMPLE_NOTIFICATION"
  category     = "custom"
  description  = "A simple custom template type with minimal configuration"

  communication_methods = ["email"]
}
