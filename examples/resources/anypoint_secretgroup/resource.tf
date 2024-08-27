resource "anypoint_secretgroup" "sg" {
  org_id = var.root_org
  env_id = var.env_id
  name = "sg-example"
  downloadable = true
}