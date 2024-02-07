
resource "anypoint_secretgroup_certificate" "cert" {
  org_id = var.root_org
  env_id = var.env_id
  sg_id = anypoint_secretgroup.sg.id
  name = "terraform-cert-example"
  type = "PEM"
  certificate = "${path.module}/keys/myserver.local.crt"
}