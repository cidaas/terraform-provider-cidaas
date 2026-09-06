resource "cidaas_notification_service_setup" "twilio_sms" {
  name                  = "Twilio SMS"
  service_id            = "twilio-sms"
  communication_methods = ["sms"]
  description           = "Twilio SMS communication provider setup"
}
