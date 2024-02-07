resource "anypoint_team_roles" "roles" {
  org_id = var.root_org
  team_id = anypoint_team.team.id

  # you can check the role data-source to get roles dynamically
  roles {
    role_id = "42ea6892-f95c-4d1b-ab48-687b1f6632fc"    # Access Controls Admin
    context_params = {
      org = anypoint_bg.bg.id           # the business group to which the role applies
      envId = anypoint_env.env.id       # if the role spans environments, the environment id
    }
  }
}