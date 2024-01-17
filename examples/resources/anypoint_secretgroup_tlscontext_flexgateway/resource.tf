resource "anypoint_secretgroup_tlscontext_flexgateway" "fg" {
  org_id = var.root_org
  env_id = var.env_id
  sg_id = var.secretgroup_id
  name = "context03"
  keystore_path = "keystores/ed570161-78cc-493f-9707-0f9ade7d8a1a"
  truststore_path = "truststores/a5a97893-acc2-4f50-bd2d-2ffbff304b90"
  min_tls_version = "TLSv1.3"
  max_tls_version = "TLSv1.3"
  alpn_protocols = tolist(["h2","http/1.1"])
  inbound_settings {
    enable_client_cert_validation = true
  }
  outbound_settings {
    skip_server_cert_validation = true
  }
  cipher_suites = tolist([])
}
