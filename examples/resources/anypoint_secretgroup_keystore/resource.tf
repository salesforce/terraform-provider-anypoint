resource "anypoint_secretgroup_keystore" "keystore" {
  org_id = var.root_org
  env_id = var.env_id
  sg_id = anypoint_secretgroup.sg.id
  name = "terraform-keystore-example-01"
  type = "PEM"
  key = "${path.module}/keys/myserver.local.key"
  certificate = "${path.module}/keys/myserver.local.crt"
}

resource "anypoint_secretgroup_keystore" "jks" {
  org_id = var.root_org
  env_id = var.env_id
  sg_id = anypoint_secretgroup.sg.id
  name = "terraform-keystore-example-jks-01"
  type = "JKS"
  keystore = "${path.module}/keys/my-terraform.keystore"
  algorithm = "PKIX"
  alias = "terraform_example"
  store_passphrase = "123456"
  key_passphrase = "123456"
}