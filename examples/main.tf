terraform {
  required_providers {
    cloudhub = {
      //versions = ["0.2"]
      source = "anypoint.mulesoft.com/automation/cloudhub"
    }
  }
}

provider "cloudhub" {
  client_id = "xxxxxxx"
  client_secret = "xxxxxx"
  org_id = "xxxxxx"
}


data "cloudhub_vpcs" "all" {}


output "all_vpcs" {
  value = data.cloudhub_vpcs.all
}
