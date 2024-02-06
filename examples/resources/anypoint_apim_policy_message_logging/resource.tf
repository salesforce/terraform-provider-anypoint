resource "anypoint_apim_policy_message_logging" "policy01" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = anypoint_apim_mule4.api.id
  disabled = false
  asset_version = "2.0.1"
  configuration_data {
    logging_configuration {
      name = "configuration 01"
      message = "#[attributes.headers['id']]"
      conditional = "#[attributes.headers['id']==1]"
      category = "My_01_Prefix_"
      level = "INFO"
      first_section = true
      second_section = false
    }
    logging_configuration {
      name = "configuration 02"
      message = "#[attributes.headers['Authorization']]"
      level = "DEBUG"
      first_section = true
      second_section = false
    }
  }
}

resource "anypoint_apim_policy_message_logging" "policy02" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = anypoint_apim_mule4.api02.id
  disabled = false
  asset_version = "2.0.1"
  configuration_data {
    logging_configuration {
      name = "configuration 01"
      message = "#[attributes.headers['id']]"
      conditional = "#[attributes.headers['id']==1]"
      category = "My_01_Prefix_"
      level = "INFO"
      first_section = true
      second_section = false
    }
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