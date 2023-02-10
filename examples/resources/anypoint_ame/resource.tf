resource "anypoint_ame" "ame" {
  org_id = var.root_org
  env_id = var.env_id
  region_id = "us-east-1"
  exchange_id = "myExchangeId"
  encrypted = true
}
