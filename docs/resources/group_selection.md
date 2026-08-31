Resource for managing group selection configurations in cidaas v4.x (Trustdesk).## Example Usage

```terraform
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

resource "cidaas_group_selection" "sample" {
  name                             = "sample_group_selection"
  description                      = "Group selection options for login"
  is_group_login_selection_enabled = true
  always_show_group_selection      = false
  selectable_groups                = [cidaas_user_groups.sample.group_id]
  selectable_group_types           = [cidaas_group_type.sample.group_type]
}
```