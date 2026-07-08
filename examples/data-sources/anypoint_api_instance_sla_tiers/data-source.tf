data "anypoint_api_instance_sla_tiers" "all" {
  org_id = var.org_id
  env_id = var.env_id
  api_id = anypoint_apim_mule4.api.id

  params {
    limit  = 200
    offset = 0
  }
}
