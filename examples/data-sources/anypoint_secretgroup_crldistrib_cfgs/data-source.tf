data "anypoint_secretgroup_crldistrib_cfgs" "cfg" {
  id     = "fc18a686-f3c3-4d50-8072-0de215814d25"
  sg_id  = var.sg_id
  org_id = var.root_org
  env_id = var.env_id
}