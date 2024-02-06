#Client Id Enforcement Policy Example
resource "anypoint_apim_policy_custom" "policy_custom_01" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = anypoint_apim_mule4.api.id
  disabled = false
  asset_group_id="68ef9520-24e9-4cf2-b2f5-620025690913"
  asset_id="client-id-enforcement"
  asset_version = "1.3.2"
  configuration_data = jsonencode({
    credentialsOriginHasHttpBasicAuthenticationHeader = "customExpression"
    clientIdExpression = "#[attributes.headers['client_id']]"
    clientSecretExpression = "#[attributes.headers['client_secret']]"
  })
}

#Rate Limit Policy Example
resource "anypoint_apim_policy_custom" "policy_custom_02" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = anypoint_apim_mule4.api.id
  disabled = false
  asset_group_id="68ef9520-24e9-4cf2-b2f5-620025690913"
  asset_id="rate-limiting"
  asset_version = "1.4.0"

  configuration_data = jsonencode({
    keySelector= "#[attributes.queryParams['identifier']]"
    rateLimits = [
      { maximumRequests = 50
        timePeriodInMilliseconds = 300000
      },
      {
        maximumRequests = 10000
        timePeriodInMilliseconds = 3600000
      }
    ]
    exposeHeaders = true
    clusterizable = true
  })
  pointcut_data {
    method_regex = ["GET", "POST"]
    uri_template_regex = "/api/v1/.*"
  }
  pointcut_data {
    method_regex = ["PUT"]
    uri_template_regex = "/api/v1/.*"
  }
}


#Basic Auth Policy Example
resource "anypoint_apim_policy_custom" "policy_custom_03" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = anypoint_apim_mule4.api.id
  disabled = false
  asset_group_id="68ef9520-24e9-4cf2-b2f5-620025690913"
  asset_id="http-basic-authentication"
  asset_version = "1.3.1"

  configuration_data = jsonencode({
    username= "user"
    password = "mySupaDupaPasswordWithALotOfCharacters"
  })
}

#Message Logging Policy Example
resource "anypoint_apim_policy_custom" "policy_custom_04" {
  org_id = var.root_org
  env_id = var.env_id
  apim_id = anypoint_apim_mule4.api.id
  disabled = false
  asset_group_id="68ef9520-24e9-4cf2-b2f5-620025690913"
  asset_id="message-logging"
  asset_version = "2.0.1"

  configuration_data = jsonencode({
    loggingConfiguration = [
      {
        itemName = "Default configuration"
        itemData = {
          message = "#[attributes.headers['id']]"
          conditional = "#[attributes.headers['id']==1]"
          category = "My_Prefix"
          level = "INFO"
          firstSection = true
          secondSection = true
        }
      }
    ]
  })
}