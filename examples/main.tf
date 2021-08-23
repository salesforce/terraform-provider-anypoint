terraform {
  required_providers {
    anypoint = {
      //versions = ["0.2"]
      source = "anypoint.mulesoft.com/automation/anypoint"
    }
  }
}

variable client_id {
  type        = string
  default     = ""
  description = "the client_id of the anypoint connected app"
}

variable client_secret {
  type        = string
  default     = ""
  description = "the client_secret of the anypoint connected app"
}

variable username {
  type        = string
  default     = ""
  description = "the username of the anypoint user"
}

variable password {
  type        = string
  default     = ""
  description = "the password of the anypoint user"
}

variable org_id {
  type        = string
  default     = ""
  description = "the anypoint organization id"
}

variable env_id {
  type        = string
  default     = "3d5af0fe-ba6b-4c52-8b5e-ed9da2fa8167"
  description = "the anypoint environment id"
}


provider "anypoint" {
  username = var.username
  password = var.password
}

# resource "anypoint_env" "env" {
#   name = "my-ENV1"
#   type = "sandbox"
#   org_id = var.org_id
# }

# output "env" {
#   value = anypoint_env.env
# }

resource "anypoint_mq" "amq" {
  defaultttl = 604800000
  defaultlockttl = 120000
  type = "queue"
  encrypted = true
  org_id = var.org_id
  env_id = var.env_id
  //env_id = "3d5af0fe-ba6b-4c52-8b5e-ed9da2fa8167"
  region_id = "eu-west-2"
  queue_id = "tq-test6"
}

output "amq" {
  value = anypoint_mq.amq
}


 /* resource "anypoint_bg" "bg" {
  name = "my BG"
  parentorganizationid = var.org_id
  ownerid = "18f23771-c78a-4be2-af8f-1bae66f43942"
  entitlements_createsuborgs = true
  entitlements_createenvironments = false
  entitlements_globaldeployment = false
  entitlements_vcoresproduction_assigned = 0.1
  entitlements_vcoressandbox_assigned = 0.2
  entitlements_vcoresdesign_assigned = 0.1
  entitlements_staticips_assigned = 0
  entitlements_vpcs_assigned = 1
  entitlements_loadbalancer_assigned = 0
  entitlements_vpns_assigned = 0
}

output "bg" {
  value = anypoint_bg.bg
} */

# data "anypoint_vpcs" "all" {
#   orgid = var.org_id
# }



# locals {
#   csv_data = file("${path.module}/csv/vpcs.csv")
#   list_separator = ";"

#   vpc_instances = csvdecode(local.csv_data)
# }

# resource "anypoint_vpc" "example" {
#   count = length(local.vpc_instances)

#   name = element(local.vpc_instances, count.index).name
#   region = element(local.vpc_instances, count.index).region
#   owner_id = element(local.vpc_instances, count.index).owner_id
#   cidr_block = element(local.vpc_instances, count.index).cidr_block
#   internal_dns_servers = compact(split(local.list_separator, element(local.vpc_instances, count.index).internal_dns_servers))  
#   internal_dns_special_domains = compact(split(local.list_separator, element(local.vpc_instances, count.index).internal_dns_special_domains)) 
#   is_default = element(local.vpc_instances, count.index).is_default
#   associated_environments = compact(split(local.list_separator, element(local.vpc_instances, count.index).associated_environments))
#   shared_with = compact(split(local.list_separator, element(local.vpc_instances, count.index).shared_with))

# }


# output "name" {
#   value = anypoint_vpc.example[*].name
# }

# output "region" {
#   value = anypoint_vpc.example[*].region
# }

# output "cidr_block" {
#   value = anypoint_vpc.example[*].cidr_block
# }






#data "anypoint_vpcs" "all" {}

# resource "anypoint_vpc" "avpc" {
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
#   value = anypoint_vpc.avpc.region
# }

# output "id" {
#   value = anypoint_vpc.avpc.id
# }

# output "name" {
#   value = anypoint_vpc.avpc.name
# }

# output "cidrblock" {
#   value = anypoint_vpc.avpc.cidr_block
# }
