resource "anypoint_secretgroup_truststore" "truststore" {
  org_id = var.root_org
  env_id = var.env_id
  sg_id = anypoint_secretgroup.sg.id
  name = "terraform-truststore-example-02"
  type = "PEM"
  truststore = "${path.module}/keys/myserver.local.crt"
}

resource "anypoint_secretgroup_truststore" "jks" {
  org_id = var.root_org
  env_id = var.env_id
  sg_id = anypoint_secretgroup.sg.id
  name = "terraform-truststore-example-jks-01"
  type = "JKS"
  truststore = "${path.module}/keys/my-terraform.keystore"
  algorithm = "PKIX"
  store_passphrase = "123456"
}