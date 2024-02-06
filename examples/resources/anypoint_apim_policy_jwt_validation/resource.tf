resource "anypoint_apim_policy_jwt_validation" "policy01" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = anypoint_apim_mule4.api.id
  disabled = true
  asset_version = "1.3.2"
  configuration_data {
    jwt_origin          = "httpBearerAuthenticationHeader"
    signing_method      = "rsa"
    signing_key_length  = 512
    jwt_key_origin      = "text"
    text_key            = "your-(256|384|512)-bit-secret"
  }
  pointcut_data {
    method_regex = ["GET", "POST"]
    uri_template_regex = "/api/v1/.*"
  }
  pointcut_data {
    method_regex = ["PUT"]
    uri_template_regex = "/api/v1/.*"
  }
}

resource "anypoint_apim_policy_jwt_validation" "policy02" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = anypoint_apim_mule4.api.id
  disabled = true
  asset_version = "1.3.2"
  configuration_data {
    jwt_origin = "httpBearerAuthenticationHeader"
    signing_method = "rsa"
    signing_key_length = 512
    jwt_key_origin = "jwks"
    jwks_url = "http://your-jwks-service.example:80/base/path"
    jwks_service_time_to_live = 60
    jwks_service_connection_timeout = 1000
  }
}