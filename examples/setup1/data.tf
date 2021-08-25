locals {
  csv_folder = "${path.module}/csv"
  bg_csv_data = file("${local.csv_folder}/bgs.csv")
  env_csv_data = file("${local.csv_folder}/envs.csv")
  users_csv_data = file("${local.csv_folder}/users.csv")
  teams_lvl1_csv_data = file("${local.csv_folder}/teams_lvl1.csv")
  teams_lvl1_roles_csv_data = file("${local.csv_folder}/teams_lvl1_roles.csv")
  teams_lvl1_members_csv_data = file("${local.csv_folder}/teams_lvl1_members.csv")
  teams_lvl2_csv_data = file("${local.csv_folder}/teams_lvl2.csv")
  teams_lvl2_roles_csv_data = file("${local.csv_folder}/teams_lvl2_roles.csv")
  teams_lvl2_members_csv_data = file("${local.csv_folder}/teams_lvl2_members.csv")


  bgs_list = csvdecode(local.bg_csv_data)
  envs_list = csvdecode(local.env_csv_data)
  users_list = csvdecode(local.users_csv_data)
  
  teams_lvl1_list = csvdecode(local.teams_lvl1_csv_data)
  teams_lvl1_roles_list = csvdecode(local.teams_lvl1_roles_csv_data)
  teams_lvl1_members_list = csvdecode(local.teams_lvl1_members_csv_data)
  
  teams_lvl2_list = csvdecode(local.teams_lvl2_csv_data)
  teams_lvl2_roles_list = csvdecode(local.teams_lvl2_roles_csv_data)
  teams_lvl2_members_list = csvdecode(local.teams_lvl2_members_csv_data)

}

# data "anypoint_vpcs" "all" {
#   orgid = var.org_id
# }

# data "anypoint_vpcs" "all" {}


# data "anypoint_team_roles" "troles" {
#   org_id = var.org_id
#   team_id = "9363290f-e795-4be6-9ca9-9efc1229b8f2"
# }


# data "anypoint_teams" "teams" {
#   org_id = var.org_id
# }