data "anypoint_flexgateway_registration_token" "token" {
  org_id = var.org_id
  env_id = var.env_id
}