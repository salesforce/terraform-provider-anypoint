data "anypoint_secretgroup_keystore" "keystore" {
  id = "f6081c3f-b0e6-41be-b3f6-1965faac0119"
  sg_id = var.sg_id
  org_id = var.org_id
  env_id = var.env_id
}