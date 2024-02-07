data "anypoint_secretgroup_tlscontext_flexgateway" "fg" {
  id     = var.id
  sg_id  = var.secretgroup_id
  org_id = var.org_id
  env_id = var.env_id
}