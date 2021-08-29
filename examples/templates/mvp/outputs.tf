# output "csv" {
#   value = local.bgs_list
# }

# output "roles" {
#   value = local.data_roles_list
# } 

# output "teams1" {
#   //count = length(data.anypoint_teams.teams)

# #   org_id = anypoint_bg.bgs[tonumber(element(local.envs_list, count.index).bg_index)].id
# #   name = element(local.envs_list, count.index).name
#   //value = data.anypoint_teams.teams.teams
#   value = [for tr in data.anypoint_teams.teams.teams : tr.team_id if tr.team_name == "CAT"]
# #   for team in data.anypoint_teams.teams.teams : team.team_name 
# #    value = team.team_name
  
# } 