variable "root_org" {
  default = "xx1f55d6-213d-4f60-845c-207286484cd1"
}

variable "username" {
  default = "my_unique_username"
}

resource "anypoint_rolegroup" "rg" {
  org_id = var.root_org
  name = "arolegroup_example"
  description = "This a rolegroup example "
  external_names = tolist(["VAL1", "VAL2"])
}

resource "anypoint_user" "user" {
  org_id = var.root_org
  username = var.username
  first_name = "terraform"
  last_name = "provider"
  email = "terraform@provider.com"
  phone_number = "0756224452"
  password = "my_super_secret_pwd"
}