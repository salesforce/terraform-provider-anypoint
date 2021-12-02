resource "anypoint_user" "user" {
  org_id = var.root_org
  username = "my_unique_username01"
  first_name = "terraform"
  last_name = "provider"
  email = "terraform@provider.com"
  phone_number = "0756224452"
  password = "my_super_secret_pwd"
}