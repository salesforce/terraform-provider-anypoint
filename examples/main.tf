terraform {
  required_providers {
    cloudhub = {
      //versions = ["0.2"]
      source = "anypoint.mulesoft.com/automation/cloudhub"
    }
  }
}

variable client_id {
  type        = string
  default     = ""
  description = "the client_id of the cloudhub connected app"
}

variable client_secret {
  type        = string
  default     = ""
  description = "the client_secret of the cloudhub connected app"
}

variable org_id {
  type        = string
  default     = ""
  description = "the cloudhub organization id"
}


provider "cloudhub" {
  client_id = var.client_id
  client_secret = var.client_secret
  org_id = var.org_id
}


# resource "cloudhub_vpc" "avpc" {
#   name = "myAwesomeVPC"
#   region = "us-east-1"
#   owner_id = ""
#   cidr_block = "10.0.0.0/20"
#   internal_dns_servers = []
#   internal_dns_special_domains = []
#   is_default = false
#   associated_environments = []
#   shared_with = []
#   //firewall_rules = []
#   //vpc_routes = []
# }

data "cloudhub_vpcs" "all" {}


output "created_vpcs" {
  #value = cloudhub_vpc.avpc.id
  value = data.cloudhub_vpcs.all
}
