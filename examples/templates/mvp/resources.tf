# resource "anypoint_bg" "bgs" {
#   count = length(local.bgs_list)

#   name = element(local.bgs_list, count.index).name
#   parentorganizationid = element(local.bgs_list, count.index).parent_org_id
#   ownerid = element(local.bgs_list, count.index).ownerid
#   entitlements_createsuborgs = element(local.bgs_list, count.index).create_suborgs
#   entitlements_createenvironments = element(local.bgs_list, count.index).create_env
#   entitlements_globaldeployment = element(local.bgs_list, count.index).global_deployment
#   entitlements_vcoresproduction_assigned = element(local.bgs_list, count.index).vcores_prod
#   entitlements_vcoressandbox_assigned = element(local.bgs_list, count.index).vcores_sandbox
#   entitlements_vcoresdesign_assigned = element(local.bgs_list, count.index).vcores_design
#   entitlements_staticips_assigned = element(local.bgs_list, count.index).static_ips
#   entitlements_vpcs_assigned = element(local.bgs_list, count.index).vpcs
#   entitlements_loadbalancer_assigned = element(local.bgs_list, count.index).lbs
#   entitlements_vpns_assigned = element(local.bgs_list, count.index).vpns
# }

# resource "anypoint_bg" "bgs" {
#   for_each = { for inst in local.bgs_list : inst.name => inst }

#   name = each.value.name
#   parentorganizationid = each.value.parent_org_id
#   ownerid = each.value.ownerid
#   entitlements_createsuborgs = each.value.create_suborgs
#   entitlements_createenvironments = each.value.create_env
#   entitlements_globaldeployment = each.value.global_deployment
#   entitlements_vcoresproduction_assigned = each.value.vcores_prod
#   entitlements_vcoressandbox_assigned = each.value.vcores_sandbox
#   entitlements_vcoresdesign_assigned = each.value.vcores_design
#   entitlements_staticips_assigned = each.value.static_ips
#   entitlements_vpcs_assigned = each.value.vpcs
#   entitlements_loadbalancer_assigned = each.value.lbs
#   entitlements_vpns_assigned = each.value.vpns
# }

# resource "anypoint_env" "envs" {
#   count = length(local.envs_list)

#   org_id = anypoint_bg.bgs[element(local.envs_list, count.index).bg_name].id
#   name = element(local.envs_list, count.index).name
#   type = element(local.envs_list, count.index).type
# }

resource "anypoint_team" "parent_teams" {
  for_each = { for inst in local.teams_lvl1_list : inst.name => inst }
  # for_each = { for inst in local.teams_lvl1_list : inst.name => {
  #   p_parent_team_id  = try([for tr in data.anypoint_teams.teams.teams : tr.team_id if tr.team_name == inst.parent_team][0],"")
  #   name = inst.name
  #   type = inst.type
  #   } 
  # }


  //name = each.value.name
  org_id = var.root_org
  //sparent_team_id = ""
  //parent_team_id = each.value.parent_team == var.team_name ? data.anypoint_teams.teams[var.team_name] : var.root_team
  //parent_team_id = coalesce(each.value.parent_team, data.anypoint_teams.teams[var.team_name].id)
  # j = {for tr in data.anypoint_teams.teams.teams : team.team_name => tr
  #   sparent_team_id = tr.team_name == var.team_name ? tr.team_id : ""
  # }

  //local.s_parent_team_id  = [for tr in data.anypoint_teams.teams.teams : tr.team_id if tr.team_name == each.value.parent_team][0]

  //parent_team_id = each.value.parent_team == var.team_name ? data.anypoint_teams.teams[var.team_name] : var.root_team
  //parent_team_id = coalesce(each.value.p_parent_team_id,anypoint_team.teams[each.value.parent_team].team_id)
  parent_team_id = try([for tr in data.anypoint_teams.teams.teams : tr.team_id if tr.team_name == each.value.parent_team][0],[for tr in data.anypoint_teams.teams.teams : tr.team_id if tr.team_name == var.team_name][0])
  team_name = each.value.name
  team_type = each.value.type
}

resource "anypoint_team" "child_teams" {
  for_each = { for inst in local.teams_lvl2_list : inst.name => inst }
  org_id = var.root_org
  parent_team_id = try(anypoint_team.parent_teams[each.value.parent_team].team_id,[for tr in data.anypoint_teams.teams.teams : tr.team_id if tr.team_name == var.team_name][0])
  team_name = each.value.name
  team_type = each.value.type
}

# resource "anypoint_env" "envs" {
#   count = length(local.envs_list)

#   org_id = anypoint_bg.bgs[tonumber(element(local.envs_list, count.index).bg_index)].id
#   name = element(local.envs_list, count.index).name
#   type = element(local.envs_list, count.index).type
# }


# resource "anypoint_user" "users" {
#   count = length(local.users_list)

#   org_id = var.root_org
#   username = element(local.users_list, count.index).username
#   first_name = element(local.users_list, count.index).firstname
#   last_name = element(local.users_list, count.index).lastname
#   email = element(local.users_list, count.index).email
#   phone_number = element(local.users_list, count.index).phone
#   password = element(local.users_list, count.index).pwd
# }


# resource "anypoint_team" "lvl1_teams" {
#   count = length(local.teams_lvl1_list)

#   org_id = var.root_org
#   parent_team_id = var.root_team
#   team_name = element(local.teams_lvl1_list, count.index).name
#   team_type = element(local.teams_lvl1_list, count.index).type
# }

# resource "anypoint_team" "lvl2_teams" {
#   count = length(local.teams_lvl2_list)

#   org_id = var.root_org
#   parent_team_id = anypoint_team.lvl1_teams[tonumber(element(local.teams_lvl2_list, count.index).parent_team_index)].id
#   team_name = element(local.teams_lvl2_list, count.index).name
#   team_type = element(local.teams_lvl2_list, count.index).type
# }


# resource "anypoint_team_roles" "lvl1_teams_roles" {
#   count = length(local.teams_lvl1_list)

#   org_id = var.root_org
#   team_id = anypoint_team.lvl1_teams[count.index].id
  
#   dynamic "roles" {
#     for_each = [
#       for role in local.teams_lvl1_roles_list : role
#       if tonumber(role.team_index) == count.index
#     ]
#     content {
#       role_id = element([
#         for iter in local.data_roles_list : iter.role_id
#         if iter.name == roles.value.name
#       ], 0)
#       context_params = {
#         org = tonumber(roles.value["context_org_index"]) == -1 ? var.root_org : anypoint_bg.bgs[tonumber(roles.value["context_org_index"])].id
#         envId = length(roles.value["context_env_index"]) > 0 ? anypoint_env.envs[tonumber(roles.value["context_env_index"])].id : null
#       }
#     }
#   }
# }
# resource "anypoint_team_roles" "lvl2_teams_roles" {
#   count = length(local.teams_lvl2_list)

#   org_id = var.root_org
#   team_id = anypoint_team.lvl2_teams[count.index].id
  
#   dynamic "roles" {
#     for_each = [
#       for role in local.teams_lvl2_roles_list : role
#       if tonumber(role.team_index) == count.index
#     ]
#     content {
#       role_id = element([
#         for iter in local.data_roles_list : iter.role_id
#         if iter.name == roles.value.name
#       ], 0)
#       context_params = {
#         org = tonumber(roles.value["context_org_index"]) == -1 ? var.root_org : anypoint_bg.bgs[tonumber(roles.value["context_org_index"])].id
#         envId = length(roles.value["context_env_index"]) > 0 ? anypoint_env.envs[tonumber(roles.value["context_env_index"])].id : null
#       }
#     }
#   }
# }


# resource "anypoint_team_member" "lvl1_teams_members" {
#   count = length(local.teams_lvl1_members_list)

#   org_id = var.root_org
#   team_id = anypoint_team.lvl1_teams[tonumber(element(local.teams_lvl1_members_list, count.index).team_index)].id
#   user_id = anypoint_user.users[tonumber(element(local.teams_lvl1_members_list, count.index).user_index)].id
# }
# resource "anypoint_team_member" "lvl2_teams_members" {
#   count = length(local.teams_lvl2_members_list)

#   org_id = var.root_org
#   team_id = anypoint_team.lvl2_teams[tonumber(element(local.teams_lvl2_members_list, count.index).team_index)].id
#   user_id = anypoint_user.users[tonumber(element(local.teams_lvl2_members_list, count.index).user_index)].id
# }
