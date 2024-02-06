resource "anypoint_secretgroup_tlscontext_mule" "mule" {
  org_id = var.root_org
  env_id = var.env_id
  sg_id = anypoint_secretgroup.sg.id
  name = "terraform-tls-context-mule"
  keystore_path = anypoint_secretgroup_keystore.jks.path
  truststore_path = anypoint_secretgroup_truststore.jks.path
  cipher_suites = tolist([])
  acceptable_tls_versions {
    tls_v1_dot1 = false
    tls_v1_dot2 = true
    tls_v1_dot3 = true
  }
  insecure = false
}
