resource "anypoint_idp_saml" "example1" {
  org_id = var.root_org
  name = "SAML 2.0 provider"
  saml {
    issuer   = "http://idp.example.com/auth/realms/master"
    audience = "example1.anypoint.mulesoft.com"

    public_key = tolist(["MIICmzCCAYMCBgF+m6ogEzANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDDAZtYXN0ZXIwHhcNMjIwMTI3MTMxMDI0WhcNMzIwMTI3MTMxMjA0WjARMQ8wDQYDVQQDDAZtYXN0ZXIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCH91QuAKq0wzQjDExmWEqSNno/wnbNvZbMb33fcnl1gQ64LlY/AFnv2RJySQ2qm5qM1q5zGJc1jy/gxS/3Rp2iUQ+NBntgdAUg/4h64lh/sM76xoRv0T9zc8tvZn5PWkqJTgksACOnAGQaxCAHqKpLS6pEHJELvlUK6sTOJSAHB1KNd7ixOW6TLkFXXdQRiN1AkJG8shOiX/lb2Vnj3KwK3+/5JmRziuWiZHKR/2TeMuAD+8GJ8AWGrGbDkQe04kbSDYNWXKjZlqfPXYx+Vfrzyijun99f+WBQSGbyakgHTDszSMkcnHI0MniotCm4mEsMroUY16YIGIpQUpB+Ghg9AgMBAAEwDQYJKoZIhvcNAQELBQADggEBAF9WeaNB9nA+Ri03IU5slROnzgSB49FOAbcrtO4ml2p1UPdwb8X1QBJSPKwRMEbXQxGddq1HtOyyayL9Ii5ogMwz8uhxxLym2MlzMUb/KAbI7cJJrRcvwhCGqIyfe932VN4v5a3/FYIHbfmmo8CKDUQmybLB8+LQ+"])

    sp_initiated_sso_enabled          = true
    idp_initiated_sso_enabled         = true
    require_encrypted_saml_assertions = true
  }
  sp_sign_on_url  = "http://idp.example.com/auth/realms/master/protocol/saml"
  sp_sign_out_url = "http://idp.example.com/auth/realms/master/protocol/saml"
}

resource "anypoint_idp_saml" "example2" {
  org_id = var.root_org
  name = "SAML 2.0 provider"
  saml {
    issuer   = "http://idp.example.com/auth/realms/master"
    audience = "example1.anypoint.mulesoft.com"

    public_key = tolist(["MIICmzCCAYMCBgF+m6ogEzANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDDAZtYXN0ZXIwHhcNMjIwMTI3MTMxMDI0WhcNMzIwMTI3MTMxMjA0WjARMQ8wDQYDVQQDDAZtYXN0ZXIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCH91QuAKq0wzQjDExmWEqSNno/wnbNvZbMb33fcnl1gQ64LlY/AFnv2RJySQ2qm5qM1q5zGJc1jy/gxS/3Rp2iUQ+NBntgdAUg/4h64lh/sM76xoRv0T9zc8tvZn5PWkqJTgksACOnAGQaxCAHqKpLS6pEHJELvlUK6sTOJSAHB1KNd7ixOW6TLkFXXdQRiN1AkJG8shOiX/lb2Vnj3KwK3+/5JmRziuWiZHKR/2TeMuAD+8GJ8AWGrGbDkQe04kbSDYNWXKjZlqfPXYx+Vfrzyijun99f+WBQSGbyakgHTDszSMkcnHI0MniotCm4mEsMroUY16YIGIpQUpB+Ghg9AgMBAAEwDQYJKoZIhvcNAQELBQADggEBAF9WeaNB9nA+Ri03IU5slROnzgSB49FOAbcrtO4ml2p1UPdwb8X1QBJSPKwRMEbXQxGddq1HtOyyayL9Ii5ogMwz8uhxxLym2MlzMUb/KAbI7cJJrRcvwhCGqIyfe932VN4v5a3/FYIHbfmmo8CKDUQmybLB8+LQ+"])

    sp_initiated_sso_enabled          = true
    idp_initiated_sso_enabled         = true
    require_encrypted_saml_assertions = true

    claims_mapping_email_attribute     = "email1"
    claims_mapping_group_attribute     = "groups1"
    claims_mapping_lastname_attribute  = "lastname1"
    claims_mapping_username_attribute  = "username1"
    claims_mapping_firstname_attribute = "firstname1"
  }
  sp_sign_on_url  = "http://idp.example.com/auth/realms/master/protocol/saml"
  sp_sign_out_url = "http://idp.example.com/auth/realms/master/protocol/saml"
}