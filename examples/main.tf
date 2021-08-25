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

variable root_team {
  type = string
  default = "99c41e16-1075-40ae-8c8b-d722a8256f81"
}


provider "anypoint" {
  username = var.username
  password = var.password
}


# resource "anypoint_bg" "bg" {
#   name = "my BG"
#   parentorganizationid = var.org_id
#   ownerid = "18f23771-c78a-4be2-af8f-1bae66f43942"
#   entitlements_createsuborgs = true
#   entitlements_createenvironments = false
#   entitlements_globaldeployment = false
#   entitlements_vcoresproduction_assigned = 0.1
#   entitlements_vcoressandbox_assigned = 0.2
#   entitlements_vcoresdesign_assigned = 0.1
#   entitlements_staticips_assigned = 0
#   entitlements_vpcs_assigned = 1
#   entitlements_loadbalancer_assigned = 0
#   entitlements_vpns_assigned = 0
# }

# output "bg" {
#   value = anypoint_bg.bg
# }

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



# resource "anypoint_team" "ateam" {
#   parent_team_id = "99c41e16-1075-40ae-8c8b-d722a8256f81"
#   team_name = "Terraform_Test_2"
#   team_type = "internal"
# }


# data "anypoint_teams" "teams" {
#   org_id = var.org_id
# }

resource "anypoint_team" "ateam" {
  org_id = var.org_id
  parent_team_id = var.root_team
  team_name = "my awesome terraform team"
}

output "teams" {
  value = resource.anypoint_team.ateam
}

resource "anypoint_team_roles" "troles" {
  org_id = var.org_id
  team_id = resource.anypoint_team.ateam.id
  roles {
    role_id = "42ea6892-f95c-4d1b-ab48-687b1f6632fc"
    context_params = {
      org = var.org_id
    }
  }
  
  roles {
    role_id = "3ef05c89-62bc-41b6-8339-3ee994f70c10"
    context_params = {
      org = var.org_id
    }
  }

  roles {
    role_id = "f14b0d23-a267-4014-9563-29d46a26295b"
    context_params = {
      org = var.org_id
      envId = "08d36da7-2232-4eb1-b74f-29abf9fe8559"
    }
  }
  
}

resource "anypoint_team_member" "member" {
  org_id = var.org_id
  team_id = resource.anypoint_team.ateam.id
  user_id = "18f23771-c78a-4be2-af8f-1bae66f43942"
}

# data "anypoint_team_roles" "troles" {
#   org_id = var.org_id
#   team_id = "9363290f-e795-4be6-9ca9-9efc1229b8f2"
# }

output "troles" {
  value = resource.anypoint_team_roles.troles
}

