resource "anypoint_team" "team" {
  org_id         = var.root_org                 # the business group id
  parent_team_id = var.root_team        # the root team id
  team_name      = "Terraform Provider Team"
  team_type      = "internal"
}

output "team" {
  value = anypoint_team.team
}
