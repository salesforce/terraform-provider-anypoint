data "anypoint_apim" "instances" {
  org_id = var.root_org
  env_id = var.env_id

  params {
    query = "flex-backend-app"
    sort = "createdDate"
  }
}