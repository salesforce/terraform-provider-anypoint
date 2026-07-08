# Search Exchange assets by type
data "anypoint_exchange_assets" "all_rest_apis" {
  types           = "rest-api"
  organization_id = var.org_id
  limit           = 50
  sort            = "name"
  ascending       = "true"
}

output "rest_api_names" {
  value = [for a in data.anypoint_exchange_assets.all_rest_apis.assets : a.name]
}

# Filter by name substring
data "anypoint_exchange_assets" "payments_apis" {
  types  = "rest-api"
  search = "payment"
}
