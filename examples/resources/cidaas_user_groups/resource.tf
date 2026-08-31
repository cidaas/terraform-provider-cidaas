# In the below examples, 'parent-user-group' is the top-level group, and its group_id is passed as parent_id in the 'child-user-group' resource.

resource "cidaas_user_groups" "parent-user-group" {
  group_type                     = "test_terraform"
  group_id                       = "sample-group-id"
  group_name                     = "sample-group-name"
  logo_url                       = "https://cidaas.de/logo"
  description                    = "sample parent user groups description"
  custom_fields                  = {}
  make_first_user_admin          = true
  member_profile_visibility      = "full"
  none_member_profile_visibility = "public"
}


resource "cidaas_user_groups" "child-user-group" {
  group_type  = "test_terraform"
  group_id    = "sample-child-group-id-sub"
  group_name  = "sample-child-group-name"
  logo_url    = "https://cidaas.de/logo"
  description = "sample child user groups description"
  custom_fields = {
    first_name  = "cidaas"
    family_name = "widaas"
  }
  parent_id = cidaas_user_groups.sample.group_id
}
