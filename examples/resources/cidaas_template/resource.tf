// custom template example
resource "cidaas_template" "custom-template-1" {
  locale        = "en-IN"
  template_key  = "TERRAFORM_TEMPLATE"
  template_type = "EMAIL"
  content       = "Indian sample content"
  subject       = "Email custom template subject with Indian English locale"
}

// custom template example with same template_key as custom-template-1 but different template_type and locale
resource "cidaas_template" "custom-template-2" {
  locale        = "de-DE"
  template_key  = "TERRAFORM_TEMPLATE"
  template_type = "SMS"
  content       = "Sample SMS template content in German English"
}

// custom template example with same template_key and template_type as custom-template-2 but different locale
resource "cidaas_template" "custom-template-3" {
  locale        = "en-US"
  template_key  = "TERRAFORM_TEMPLATE"
  template_type = "SMS"
  content       = "Sample SMS template content in US English"
}


// System templates are created by setting the flag is_system_template to true.
// By default, this value is false and creates custom templates when applied.
// Validation checks are applied, and suggestions are provided in error messages to assist users in creating system templates.
// System templates cannot be imported using the standard Terraform import command.
// Instead, users must create a configuration that matches the existing system template and run terraform apply.

// Example of a system template for the template group "sample_group":
resource "cidaas_template" "system-template-1" {
  locale             = "en-US"
  template_key       = "VERIFY_USER"
  template_type      = "SMS"
  content            = "Hi {{name}}, here is the {{code}} to verify the user"
  is_system_template = true
  group_id           = "sample_group"
  processing_type    = "GENERAL"
  verification_type  = "SMS"
  usage_type         = "VERIFICATION_CONFIGURATION"
}

// Example of a  system template for the system default template_group "default"
resource "cidaas_template" "system-template-2" {
  locale             = "en-US"
  template_key       = "NOTIFY_COMMUNICATION_CHANGE"
  template_type      = "SMS"
  content            = "Your mobile number changed in {{account_name}}-account to {{communication_medium_value}}."
  is_system_template = true
  group_id           = "default"
  processing_type    = "GENERAL"
  usage_type         = "GENERAL"
}
