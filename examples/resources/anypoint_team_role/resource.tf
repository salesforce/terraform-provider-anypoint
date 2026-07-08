// anypoint_team_role manages a single role grant on a team. It is additive and
// composes well with other modules that grant additional roles to the same team.
// Do NOT mix with anypoint_team_roles on the same team — they will fight.

resource "anypoint_team_role" "org_admin" {
  org_id  = var.root_org
  team_id = anypoint_team.it_admin.id
  role_id = "00000000-0000-0000-0000-000000000000" # Organization Administrator (look up via anypoint_roles)

  context_params = {
    org = anypoint_bg.app_x.id
  }
}

// Environment-scoped role — both org and envId are required.
resource "anypoint_team_role" "env_admin" {
  org_id  = var.root_org
  team_id = anypoint_team.it_admin.id
  role_id = "11111111-1111-1111-1111-111111111111" # Environment Administrator

  context_params = {
    org   = anypoint_bg.app_x.id
    envId = anypoint_env.prod.id
  }
}
