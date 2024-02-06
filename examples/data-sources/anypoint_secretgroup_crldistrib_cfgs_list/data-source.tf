data "anypoint_secretgroup_crldistrib_cfgs_list" "list" {
  sg_id  = var.sg_id
  org_id = var.root_org
  env_id = var.env_id
}