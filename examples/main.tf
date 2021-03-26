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


locals {
  csv_data = file("${path.module}/csv/vpcs.csv")
  list_separator = ";"

  vpc_instances = csvdecode(local.csv_data)
}

resource "cloudhub_vpc" "example" {
  count = length(local.vpc_instances)

  name = element(local.vpc_instances, count.index).name
  region = element(local.vpc_instances, count.index).region
  owner_id = element(local.vpc_instances, count.index).owner_id
  cidr_block = element(local.vpc_instances, count.index).cidr_block
  internal_dns_servers = compact(split(local.list_separator, element(local.vpc_instances, count.index).internal_dns_servers))  
  internal_dns_special_domains = compact(split(local.list_separator, element(local.vpc_instances, count.index).internal_dns_special_domains)) 
  is_default = element(local.vpc_instances, count.index).is_default
  associated_environments = compact(split(local.list_separator, element(local.vpc_instances, count.index).associated_environments))
  shared_with = compact(split(local.list_separator, element(local.vpc_instances, count.index).shared_with))

}


output "name" {
  value = cloudhub_vpc.example[*].name
}

output "region" {
  value = cloudhub_vpc.example[*].region
}

output "cidr_block" {
  value = cloudhub_vpc.example[*].cidr_block
}






#data "cloudhub_vpcs" "all" {}

# resource "cloudhub_vpc" "avpc" {
#   name = "myAwesomeVPC"
#   region = "us-east-2"
#   owner_id = ""
#   cidr_block = "192.168.0.0/24"
#   internal_dns_servers = []
#   internal_dns_special_domains = []
#   is_default = true
#   associated_environments = []
#   shared_with = []
#   //firewall_rules = []
#   //vpc_routes = []
# }




# output "region" {
#   value = cloudhub_vpc.avpc.region
# }

# output "id" {
#   value = cloudhub_vpc.avpc.id
# }

# output "name" {
#   value = cloudhub_vpc.avpc.name
# }

# output "cidrblock" {
#   value = cloudhub_vpc.avpc.cidr_block
# }
