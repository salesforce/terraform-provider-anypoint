# Lists transit gateways across the whole organization.
# Optional filters: region, private_space_id.

data "anypoint_transit_gateways" "all" {
  org_id = var.root_org
}

data "anypoint_transit_gateways" "us_east_2" {
  org_id = var.root_org
  region = "us-east-2"
}

output "all_tgws" {
  value = data.anypoint_transit_gateways.all.transit_gateways
}
