data "anypoint_secretgroup" "secretgroup" {
  id     = "your_secretgroup_id"
  org_id = var.org_id
  env_id = var.env_id
}