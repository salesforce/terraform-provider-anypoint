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

  role_names_list = distinct( concat([ for role in local.teams_lvl1_roles_list : role.name ], [ for role in local.teams_lvl2_roles_list : role.name ]) )

  #flattened result from roles data source
  data_roles_list = flatten([for iter in data.anypoint_roles.roles : iter.roles])
}


data "anypoint_roles" "roles" {
  count = length(local.role_names_list)
  params {
    search = element(local.role_names_list, count.index)
  }
}