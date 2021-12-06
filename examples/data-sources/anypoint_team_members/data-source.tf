resource "anypoint_team_members" "team_members" {
  org_id  = var.root_id
  team_id = "YOUR_TEAM_ID"
  # only one params block is supported
  params {
    membership_type = ""    #Include the group access mappings that grant the provided membership type By default, all group access mappings are returned
    identity_type   = ""    #A search string to use for case-insensitive partial matches on external group name
    member_ids      = ""    #Include the members of the team that have ids in this list
    search          = ""    #Maximum records to retrieve per request.
    offset          = 0     #The number of records to omit from the response.
    limit           = 200   #Maximum records to retrieve per request.
    sort            = ""    #The field to sort on
    ascending       = true  #Whether to sort ascending or descending
  }
}