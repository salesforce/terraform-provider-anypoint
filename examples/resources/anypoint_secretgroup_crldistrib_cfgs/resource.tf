resource "anypoint_secretgroup_crldistrib_cfgs" "cfg" {
  org_id = var.root_org
  env_id = var.env_id
  sg_id = anypoint_secretgroup.sg.id
  name = "sg-crl-distrib-cfg-01"
  complete_crl_issuer_url = "http://crl.microsoft.com/pki/crl/products/microsoftrootcert.crl"
  frequency = 2
  distributor_certificate_path = anypoint_secretgroup_certificate.cert02.path
  ca_certificate_path = anypoint_secretgroup_certificate.cert02.path
}