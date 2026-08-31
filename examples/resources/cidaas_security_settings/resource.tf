# Requires OAuth scopes: cidaas:fds_settings_read, cidaas:fds_settings_write

resource "cidaas_security_settings" "example" {
  blocking_setting = {
    enabled = true
  }

  repeated_login_blocking_mechanism = {
    blocking_duration_in_min     = 10
    blocked_count                = 4
    blocked_count_unknown_device = 2
  }

  rule_configuration = {
    repeated_login_blocking_mechanism_enabled = true
  }
}
