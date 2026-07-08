###############################################################################
# anypoint_private_space_vpns (plural)
#
# Read every VPN connection attached to a private space. Useful for asserting
# that no rogue connection has been created out of band, or for fanning a
# downstream resource over all connections.
###############################################################################

data "anypoint_private_space_vpns" "all" {
  org_id           = var.root_org
  private_space_id = var.private_space_id
}

output "vpn_connection_ids" {
  value = [for c in data.anypoint_private_space_vpns.all.connections : c.connection_id]
}

output "vpn_connections_summary" {
  description = "name -> total vpn-member count per connection."
  value = {
    for c in data.anypoint_private_space_vpns.all.connections : c.name => length(c.vpns)
  }
}

# Example: pick a connection by display name and reference its first member's
# static routes downstream.
output "office_vpn_static_routes" {
  value = try(
    [for c in data.anypoint_private_space_vpns.all.connections : c if c.name == "office-vpn"][0].vpns[0].static_routes,
    null,
  )
}
