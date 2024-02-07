resource "anypoint_apim_policy_rate_limiting" "policy01" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = anypoint_apim_mule4.api.id
  disabled = false
  asset_version = "1.4.0"
  configuration_data {
    key_selector = "#[attributes.queryParams['identifier']]"
    rate_limits {
      maximum_requests = 100
      time_period_in_milliseconds = 1000
    }
    rate_limits {
      maximum_requests = 1000
      time_period_in_milliseconds = 1000000
    }
    expose_headers = false
    clusterizable = true
  }
}


resource "anypoint_apim_policy_rate_limiting" "policy02" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = anypoint_apim_mule4.api02.id
  disabled = false
  asset_version = "1.4.0"
  configuration_data {
    key_selector = "#[attributes.queryParams['identifier']]"
    rate_limits {
      maximum_requests = 100
      time_period_in_milliseconds = 1000
    }
    expose_headers = true
    clusterizable = true
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