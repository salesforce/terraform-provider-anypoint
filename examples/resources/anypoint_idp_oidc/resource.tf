resource "anypoint_idp_oidc" "example1" {
  org_id = var.root_org
  name = "openid connect provider"
  oidc_provider {
    authorize_url = "http://idp.example.com/auth/realms/master/protocol/openid-connect/auth"
    token_url     = "http://idp.example.com/auth/realms/master/protocol/openid-connect/token"
    userinfo_url  = "http://idp.example.com/auth/realms/master/protocol/openid-connect/userinfo"

    client_registration_url = "http://idp.example.com/auth/realms/master/clients-registrations/openid-connect"

    issuer = "http://idp.example.com/auth/realms/master"

    allow_untrusted_certificates = true
  }
}


resource "anypoint_idp_oidc" "example2" {
  org_id = var.root_org
  name = "openid connect provider 2"
  oidc_provider {
    authorize_url = "http://idp.example.com/auth/realms/master/protocol/openid-connect/auth"
    token_url     = "http://idp.example.com/auth/realms/master/protocol/openid-connect/token"
    userinfo_url  = "http://idp.example.com/auth/realms/master/protocol/openid-connect/userinfo"

    issuer = "http://idp.example.com/auth/realms/master"

    client_credentials_id     = "anypoint-oidc"
    client_credentials_secret = "63b376f8-3ece-44f6-869c-33fe9022fdc4"

    allow_untrusted_certificates = true
  }
}