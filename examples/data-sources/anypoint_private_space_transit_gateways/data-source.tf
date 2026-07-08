data "anypoint_private_space_transit_gateways" "all" {
  org_id           = var.root_org
  private_space_id = "7f747999-a9bb-41d5-bfe8-d6b8cca68c62"
}

output "tgw_ids" {
  value = [for t in data.anypoint_private_space_transit_gateways.all.transit_gateways : t.transit_gateway_id]
}
