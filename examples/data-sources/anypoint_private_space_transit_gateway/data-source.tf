data "anypoint_private_space_transit_gateway" "tgw" {
  org_id             = var.root_org
  private_space_id   = "7f747999-a9bb-41d5-bfe8-d6b8cca68c62"
  transit_gateway_id = "tgw-017e20b9ce00c865c"
}

output "tgw_status" {
  value = data.anypoint_private_space_transit_gateway.tgw.status_attachment
}
