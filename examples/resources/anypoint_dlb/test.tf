variable "root_org" {
  default = "xx1f55d6-213d-4f60-845c-207286484cd1"
}

variable "owner_id" {
  default = "18f23771-c78a-4be2-af8f-1bae66f43942"
}

resource "anypoint_bg" "bg" {
  name = "TEST_BG_TF"
  parent_organization_id = var.root_org
  owner_id = var.owner_id
  entitlements_createsuborgs = true
  entitlements_createenvironments = true
  entitlements_globaldeployment = true
  entitlements_vcoresproduction_assigned = 0
  entitlements_vcoressandbox_assigned = 0
  entitlements_vcoresdesign_assigned = 0
  entitlements_staticips_assigned = 0
  entitlements_vpcs_assigned = 1
  entitlements_loadbalancer_assigned = 0
  entitlements_vpns_assigned = 1
}

resource "anypoint_vpc" "vpc" {
  org_id = anypoint_bg.bg.id
  name = "myAwesomeVPC"
  region = "us-east-2"
  owner_id = var.owner_id
  cidr_block = "192.168.0.0/24"
  internal_dns_servers = []
  internal_dns_special_domains = []
  is_default = true
  associated_environments = []
  shared_with = []
  firewall_rules {
    cidr_block = "0.0.0.0/0"
    from_port = 8081
    protocol = "tcp"
    to_port = 8082
  }
  firewall_rules {
      cidr_block = "10.0.0.0/20"
      from_port = 8091
      protocol = "tcp"
      to_port = 8092
  }
  vpc_routes {
    cidr = "10.0.0.0/20"
    next_hop = "Local"
  }
  vpc_routes{
    cidr = "0.0.0.0/0"
    next_hop = "Internet Gateway"
  }
}
