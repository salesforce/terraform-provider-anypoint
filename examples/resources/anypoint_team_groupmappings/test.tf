variable "root_org" {
  default = "xx1f55d6-213d-4f60-845c-207286484cd1"
}

variable "root_team" {
  default = "xx1f55d6-213d-4f60-er5c-4t3286484cd1"
}

resource "anypoint_team" "team" {
  org_id         = var.root_org                 # the business group id
  parent_team_id = var.root_team        # the root team id
  team_name      = "Terraform Provider Team"
  team_type      = "internal"
}