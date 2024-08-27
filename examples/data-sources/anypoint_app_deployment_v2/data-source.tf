data "anypoint_app_deployment_v2" "app" {
  id = "de32fc9d-6b25-4d6f-bd5e-cac32272b2f7"
  org_id = var.root_org
  env_id = var.env_id
}
