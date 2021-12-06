resource "anypoint_team_roles" "team_roles" {
  org_id  = var.root_org
  team_id = "ID_OF_THE_TEAM"
  # only one params block is supported
  params {
    role_id   = ""    #return only role assignments containing one of the supplied role_ids
    search    = ""    #A search string to use for case-insensitive partial matches on role name
    offset    = 0     # The number of records to omit from the response.
    limit     = 200   # Maximum records to retrieve per request. 
  }
}

output "team_roles" {
  value = data.anypoint_team_roles.team_roles
}