resource "anypoint_secretgroup_tlscontext_securityfabric" "sf" {
  org_id = var.root_org
  env_id = var.env_id
  sg_id = anypoint_secretgroup.sg.id
  name = "terraform-tlscontext-sf"
  keystore_path = anypoint_secretgroup_keystore.jks.path
  truststore_path = anypoint_secretgroup_truststore.jks.path

  acceptable_tls_versions {
    tls_v1_dot1 = false
    tls_v1_dot2 = true
    tls_v1_dot3 = true
  }
  enable_mutual_authentication = true
  acceptable_cipher_suites {
    aes256_sha256 = true
    dhe_rsa_aes256_sha256 = true
  }
  mutual_authentication {
    cert_checking_strength = "Lax"
    verification_depth = 2
    authentication_overrides {
      allow_self_signed = true
    }
  }
}
