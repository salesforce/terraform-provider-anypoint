resource "anypoint_team_groupmappings" "team_gmap" {
  org_id  = var.root_org
  team_id = anypoint_team.team.id

  groupmappings {
    external_group_name = "gr_name01"     #the group name in the IDP side
    provider_id         = "pr01"          #the identity provider id
    membership_type     = "maintainer"    #enum : member or maintainer
  }

  groupmappings {
    external_group_name = "gr_name_pr02_01"     #the group name in the IDP side
    provider_id         = "pr02"                #the identity provider id
    membership_type     = "member"              #enum : member or maintainer
  }
}