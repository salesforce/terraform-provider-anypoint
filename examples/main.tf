terraform {
  required_providers {
    cloudhub = {
      //versions = ["0.2"]
      source = "anypoint.mulesoft.com/automation/cloudhub"
    }
  }
}

provider "cloudhub" {
  client_id = "7627f157dac94390951b4d804d218289"
  client_secret = "75Db67bE3DFf420A9e038adAff3CEFBd"
  org_id = "a9e7fe3f-c09f-4b05-9b2f-c786e009ce94"
}


data "cloudhub_vpcs" "all" {}


output "all_vpcs" {
  value = data.cloudhub_vpcs.all
}
