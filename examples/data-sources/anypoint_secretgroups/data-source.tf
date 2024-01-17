data "anypoint_secretgroups" "list" {
  org_id = var.org_id
  env_id = var.env_id
}