resource "anypoint_apim_policy_client_id_enforcement" "policy01" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = anypoint_apim_mule4.api01.id
  disabled = false
  asset_version = "1.3.2"
  configuration_data {
    credentials_origin_has_http_basic_authentication_header = "customExpression"
    client_id_expression = "#[attributes.headers['client_id']]"
    client_secret_expression = "#[attributes.headers['client_secret']]"
  }
}


resource "anypoint_apim_policy_client_id_enforcement" "policy02" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = anypoint_apim_mule4.api02.id
  disabled = false
  asset_version = "1.3.2"
  configuration_data {
    credentials_origin_has_http_basic_authentication_header = "httpBasicAuthenticationHeader"
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