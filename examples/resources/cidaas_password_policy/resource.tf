resource "cidaas_password_policy" "sample" {
  policy_name = "sample_terraform_policy"
  password_policy = {
    block_compromised = false,
    deny_usage_count  = 3,
    strength_regexes = [
      "^(?=.*[A-Za-z])(?!.*\\s).{6,15}$"
    ],
    change_enforcement = {
      expiration_in_days         = 90
      notify_user_before_in_days = 7
    }
  }
}
