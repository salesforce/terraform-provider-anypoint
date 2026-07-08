###############################################################################
# anypoint_private_space_vpn
#
# Manages a VPN connection on a CloudHub 2.0 private space. A connection groups
# one or more VPN members (each a customer endpoint pair) and is the same
# object the Anypoint UI exposes under Private Space → Connectivity → VPN.
#
# Constraints enforced at plan time (no destructive action is taken if any of
# these are violated):
#
#   * every member must be either static (`static_routes` set) or BGP
#     (`remote_asn` set) — not both, not neither;
#   * all members of a given connection must share the same `local_asn`;
#   * all members of a given connection must use the same mode (all static
#     or all BGP — mixed connections are rejected by the API);
#   * `startup_action` is a per-member field that applies to every tunnel
#     of that member ("start" = Anypoint initiates the tunnel = the
#     "Automatic Tunnel Initiation" checkbox ON in the UI; "add" = OFF).
#
# Immutable fields (changing them on an existing member forces destroy +
# recreate of the whole connection): `local_asn`, `remote_asn`,
# `remote_ip_address`, `psk`, `ptp_cidr`. Other fields are PATCHed in place.
###############################################################################

# --- Optional reference data sources ----------------------------------------

# The Anypoint-side public IPs that customers must allow on their firewall.
data "anypoint_private_space" "this" {
  id     = var.private_space_id
  org_id = var.root_org
}

# Vendor / model / firmware compatibility matrix maintained by Anypoint.
data "anypoint_private_space_supported_vpn_configs" "this" {
  org_id = var.root_org
}

# --- Sensitive material ------------------------------------------------------

# 64-char PSK per tunnel. Anypoint rejects PSKs starting with '0', so we force
# a leading letter via interpolation. Treat the resulting `local.ps_vpn_psk`
# values as secrets — they only ever appear in terraform state and in the
# PATCH request body.
resource "random_password" "psk" {
  count   = 2
  length  = 63
  special = false
  upper   = false
}

locals {
  psk = [for p in random_password.psk : "a${p.result}"]
}

# --- Variant A: single static-routes VPN ------------------------------------
#
# Use this shape when the customer side has no BGP-capable router and the
# routes that need to traverse the tunnel are fixed CIDRs.

resource "anypoint_private_space_vpn" "static_single" {
  org_id           = var.root_org
  private_space_id = var.private_space_id
  name             = "office-vpn"

  vpns {
    name              = "primary"
    local_asn         = "64512"
    remote_ip_address = var.customer_public_ip
    static_routes     = ["10.2.0.0/20", "10.8.0.0/20"]
    startup_action    = "start"

    vpn_tunnels {
      psk      = local.psk[0]
      ptp_cidr = "169.254.10.0/30"
    }
    vpn_tunnels {
      psk      = local.psk[1]
      ptp_cidr = "169.254.12.0/30"
    }
  }
}

# --- Variant B: BGP VPN with two members ------------------------------------
#
# Multi-member connection. Each member is a separate customer device (or VPN
# concentrator) pairing with Anypoint. Both members MUST share the same
# `local_asn` and the same mode (BGP here). Append new members at the tail of
# the list to trigger an in-place POST /vpns; remove from the tail to trigger
# an in-place DELETE /vpns/{id}. Mid-list reordering is rejected with a clear
# diagnostic.

# resource "anypoint_private_space_vpn" "bgp_multi" {
#   org_id           = var.root_org
#   private_space_id = var.private_space_id
#   name             = "dc-vpn"
#
#   vpns {
#     name              = "dc-east"
#     local_asn         = "64512"
#     remote_asn        = "65001"
#     remote_ip_address = "203.0.113.10"
#     startup_action    = "start"
#     vpn_tunnels {
#       psk      = local.psk[0]
#       ptp_cidr = "169.254.20.0/30"
#     }
#     vpn_tunnels {
#       psk      = local.psk[1]
#       ptp_cidr = "169.254.22.0/30"
#     }
#   }
#
#   vpns {
#     name              = "dc-west"
#     local_asn         = "64512" # MUST match the local_asn of every other member
#     remote_asn        = "65002"
#     remote_ip_address = "203.0.113.20"
#     startup_action    = "start"
#     vpn_tunnels {
#       psk      = local.psk[0]
#       ptp_cidr = "169.254.24.0/30"
#     }
#     vpn_tunnels {
#       psk      = local.psk[1]
#       ptp_cidr = "169.254.26.0/30"
#     }
#   }
# }

# --- Known upstream limitation: startup_action drift ------------------------
#
# Anypoint's GET endpoint reports `startup_action = "start"` for every tunnel
# even after a successful PATCH to `"add"` (the PATCH itself takes effect —
# its response body confirms the new value). Terraform will therefore plan a
# perpetual `start -> add` diff on every refresh if you set `startup_action`
# to `"add"`. Until the platform fixes this, opt out of the diff with:
#
#   lifecycle {
#     ignore_changes = [vpns]
#   }
#
# Applied to the resource it suppresses ALL post-create vpns drift; use only
# if you accept manual reconciliation of any further member changes.

# --- Outputs ----------------------------------------------------------------

output "anypoint_outbound_static_ips" {
  description = "Whitelist these Anypoint IPs on your customer firewall."
  value       = data.anypoint_private_space.this.outbound_static_ips
}

output "static_vpn_id" {
  value = anypoint_private_space_vpn.static_single.connection_id
}

output "supported_vpn_vendors" {
  value = [for v in data.anypoint_private_space_supported_vpn_configs.this.vendors : v.name]
}
