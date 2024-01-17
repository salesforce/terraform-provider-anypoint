data "anypoint_secretgroup_keystore" "list" {
  sg_id = "39731075-0521-47aa-82b2-d9745f2ac2eb"
  org_id = var.org_id
  env_id = var.env_id
}