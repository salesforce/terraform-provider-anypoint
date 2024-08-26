data "anypoint_app_deployments_v2" "apps" {
  org_id = var.root_org
  env_id = var.env_id
}
