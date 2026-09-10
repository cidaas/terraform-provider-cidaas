resource "cidaas_group_type" "example" {
  group_type    = "terraform_example"
  role_mode     = "allowed_roles"
  description   = "Example group type managed by Terraform"
  allowed_roles = ["USER", "ADMIN"]
}
