terraform {
  required_providers {
    cloudhub = {
      //versions = ["0.2"]
      source = "anypoint.mulesoft.com/automation/cloudhub"
    }
  }
}

provider "cloudhub" {}


data "cloudhub_vpcs" "all" {}


output "all_vpcs" {
  value = data.cloudhub_vpcs.all
}
