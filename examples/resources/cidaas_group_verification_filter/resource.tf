resource "cidaas_group_type" "sample" {
  group_type    = "sample_group_type"
  role_mode     = "Classic"
  allowed_roles = ["USER", "ADMIN"]
}

resource "cidaas_user_groups" "sample" {
  group_type = cidaas_group_type.sample.group_type
  group_id   = "sample_group_id"
  group_name = "Sample User Group"
}

resource "cidaas_group_verification_filter" "sample" {
  description     = "Filter verifying user group and role access"
  match_condition = "OR"

  filters {
    group_id = "CIDAAS_ADMINS"

    role_filter {
      roles           = ["ADMIN"]
      match_condition = "OR"
    }
  }

  filters {
    group_type = "CIDAAS_ADMINS_GROUPTYPE"
  }
}
