data "anypoint_apim_instance_upstreams" "upstreams" {
  id     = "19205930"
  org_id = var.root_org
  env_id = var.env_id
}