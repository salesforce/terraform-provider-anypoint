resource "anypoint_team" "team" {
  org_id = var.root_org
  # only one params block is used.
  params {
    ancestor_team_id = ""      # team_id that must appear in the team's ancestor_team_ids.
    parent_team_id   = ""      # team_id of the immediate parent of the team to return.
    team_id          = ""      # id of the team to return.
    team_type        = ""      # return only teams that are of this type
    search           = ""      # A search string to use for case-insensitive partial matches on team name
    offset           = 0       # The number of records to omit from the response.
    limit            = 200     # Maximum records to retrieve per request. 
    sort             = ""      # The field to sort on.
    ascending        = true    # Whether to sort ascending or descending.
  }
}
