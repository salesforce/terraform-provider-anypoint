variable "root_org" {
  default = "xx1f55d6-213d-4f60-845c-207286484cd1"
}

variable "root_team" {
  default = "xx1f55d6-ccc1-123e-845c-207286484cd1"
}

resource "anypoint_user" "user" {
  org_id = var.root_org
  username = "my_unique_username01"
  first_name = "terraform"
  last_name = "provider"
  email = "terraform@provider.com"
  phone_number = "0756224452"
  password = "my_super_secret_pwd"
}

resource "anypoint_team" "team" {
  org_id         = var.root_org                 # the business group id
  parent_team_id = var.root_team        # the root team id
  team_name      = "Terraform Provider Team"
  team_type      = "internal"
}