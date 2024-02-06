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