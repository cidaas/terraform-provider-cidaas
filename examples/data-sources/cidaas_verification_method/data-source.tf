# Read one verification method (system or pluggable) by its verification_type.
data "cidaas_verification_method" "totp" {
  verification_type = "TOTP"
}

output "totp_active" {
  value = data.cidaas_verification_method.totp.active
}
