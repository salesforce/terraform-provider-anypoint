data "anypoint_ame" "ame_list" {
  org_id = var.root_org
  env_id = var.env_id
  region_id = "us-east-1"

  params {
    offset = 2
    limit = 10
  }
}