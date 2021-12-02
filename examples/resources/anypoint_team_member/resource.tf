resource "anypoint_team_member" "team_member" {
  org_id = var.root_org
  team_id = anypoint_team.team.id
  user_id = anypoint_user.user.id
}
