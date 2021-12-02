resource "anypoint_vpc" "avpc" {
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
