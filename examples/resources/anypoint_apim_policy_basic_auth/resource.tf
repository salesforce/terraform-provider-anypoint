resource "anypoint_apim_policy_basic_auth" "policy01" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = anypoint_apim_mule4.api01.id
  disabled = false
  asset_version = "1.3.1"
  configuration_data {
    username = "myusername"
    password = "NotaRealPassword;)"
  }
}


resource "anypoint_apim_policy_basic_auth" "policy02" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = anypoint_apim_mule4.api02.id
  disabled = false
  asset_version = "1.3.1"
  configuration_data {
    username = "myOtherusername"
    password = "NotaRealPassword;)"
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