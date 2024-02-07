data "anypoint_apim_instance_policies" "policies" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = "19250669"
}