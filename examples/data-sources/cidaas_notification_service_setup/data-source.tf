resource "cidaas_notification_service_setup" "twilio_sms" {
  name                  = "Twilio SMS"
  service_id            = "twilio-sms"
  communication_methods = ["sms"]
}

data "cidaas_notification_service_setup" "example" {
  service_setup_id = cidaas_notification_service_setup.twilio_sms.id
}
