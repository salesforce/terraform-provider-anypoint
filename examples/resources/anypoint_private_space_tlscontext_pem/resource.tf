
################################################################################
# Private Space
################################################################################
resource "anypoint_private_space" "my_ps" {
  org_id = var.root_org
  region = "eu-west-1"
  name = "cat-private-space"
  environments_business_groups = [ "all" ]
  environments_type = "sandbox"
  network_cidr_block = "10.0.0.0/16"
  network_region = "eu-west-1"
  dynamic "firewall_rules" {
    for_each = [
      {
        cidr_block = "0.0.0.0/0"
        from_port  = 80
        protocol   = "tcp"
        to_port    = 80
        type       = "inbound"
      },
      {
        cidr_block = "0.0.0.0/0"
        from_port  = 443
        protocol   = "tcp"
        to_port    = 443
        type       = "inbound"
      },
      {
        cidr_block = "0.0.0.0/0"
        from_port  = 0
        protocol   = "tcp"
        to_port    = 65535
        type       = "outbound"
      },
      {
        cidr_block = "0.0.0.0/0"
        from_port  = 443
        protocol   = "tcp"
        to_port    = 443
        type       = "outbound"
      }
    ]
    content {
      cidr_block = firewall_rules.value.cidr_block
      from_port  = firewall_rules.value.from_port
      protocol   = firewall_rules.value.protocol
      to_port    = firewall_rules.value.to_port
      type       = firewall_rules.value.type
    }
  }
}

################################################################################
# Wait for the Private Space to be fully provisioned
# (average provisioning time is ~15-20 minutes)
################################################################################
resource "time_sleep" "private_space_ready" {
  depends_on      = [anypoint_private_space.my_ps]
  create_duration = "20m"   # adjust if your org typically provisions faster/slower
}

################################################################################
# TLS Context – PEM
################################################################################
resource "anypoint_private_space_tlscontext_pem" "my_tlscontext" {
  depends_on            = [time_sleep.private_space_ready]

  org_id                = var.root_org
  name                  = "my-tlscontext"
  private_space_id      = anypoint_private_space.my_ps.id

  certificate           = file("${path.module}/keys/myserver.local.crt")
  key                   = file("${path.module}/keys/myserver.local.key")
  key_passphrase        = ""

  certificate_file_name = "myserver.local.crt"
  key_file_name         = "myserver.local.key"
}