# Look up the latest version of an Exchange asset
data "anypoint_exchange_asset" "latest_petstore" {
  org_id   = var.org_id
  asset_id = "petstore"
}

# Look up a specific version
data "anypoint_exchange_asset" "v1_petstore" {
  org_id   = var.org_id
  asset_id = "petstore"
  version  = "1.0.0"
}

output "petstore_status" {
  value = data.anypoint_exchange_asset.latest_petstore.status
}
