data "anypoint_team_role" "org_admin" {
  org_id  = var.root_org
  team_id = "ID_OF_THE_TEAM"
  role_id = "00000000-0000-0000-0000-000000000000"

  context_params = {
    org = "ID_OF_THE_TARGET_BG"
    # envId = "ID_OF_THE_ENV"   # for environment-scoped roles only
  }
}

output "team_role_name" {
  value = data.anypoint_team_role.org_admin.name
}
