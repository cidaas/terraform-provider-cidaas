variable "twilio_account_sid" {
  type      = string
  sensitive = true
  ephemeral = true
}

variable "twilio_auth_token" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "cidaas_notification_service_setup" "twilio_sms" {
  name                  = "Twilio SMS"
  service_id            = "twilio-sms"
  communication_methods = ["sms"]
  description           = "Twilio SMS communication provider setup"
}

resource "cidaas_notification_provider_config" "twilio_sms" {
  service_setup_id = cidaas_notification_service_setup.twilio_sms.id

  # Write-Only: not stored in plan or state. Increment config_data_wo_version to rotate secrets.
  config_data_wo = jsonencode({
    commProvider = "custom-twilio-sms"
    commMethod   = "sms"
    schemaData = {
      accountSid = var.twilio_account_sid
      authToken  = var.twilio_auth_token
    }
  })
  config_data_wo_version = "1"
}
