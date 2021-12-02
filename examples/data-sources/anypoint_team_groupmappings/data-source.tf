data "anypoint_team_groupmappings" "team_gmap" {
  org_id  = var.root_org
  team_id = "YOUR_TEAM_ID"
}

output "team_gmap" {
  value = data.anypoint_team_groupmappings.team_gmap
}