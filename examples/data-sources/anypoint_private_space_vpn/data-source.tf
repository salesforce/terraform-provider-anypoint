###############################################################################
# anypoint_private_space_vpn (singular)
#
# Read one VPN connection by its connection_id. Use this when you need to
# reference an existing VPN connection (created out of band or by another
# stack) from inside your terraform configuration.
###############################################################################

data "anypoint_private_space_vpn" "this" {
  org_id           = var.root_org
  private_space_id = var.private_space_id
  connection_id    = var.connection_id
}

# Connection-level fields.
output "vpn_connection_name" {
  value = data.anypoint_private_space_vpn.this.name
}

# Per-member fields. Each member exposes connection_name, local_asn,
# remote_asn, remote_ip_address, static_routes, vpn_id, vpn_connection_status
# and the nested vpn_tunnels block (psk is masked).
output "vpn_member_ids" {
  value = [for v in data.anypoint_private_space_vpn.this.vpns : v.vpn_id]
}

output "vpn_first_member_status" {
  description = "Lifecycle status of the first VPN member (unavailable / pending / available)."
  value       = try(data.anypoint_private_space_vpn.this.vpns[0].vpn_connection_status, null)
}
