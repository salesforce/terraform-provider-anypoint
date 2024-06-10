resource "anypoint_apim_mule4" "api" {
  org_id = var.root_org
  env_id = var.env_id
  asset_group_id = var.root_org
  asset_id = "mule-app-test"
  asset_version = "1.0.0"
  instance_label  = "my mule4 instance"
  description = "my description"
  endpoint_uri = "http://consumer.url"
}