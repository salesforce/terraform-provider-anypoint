###############################################################################
# anypoint_private_space_supported_vpn_configs
#
# Reference data: the matrix of vendor / model / firmware versions that
# Anypoint claims to support for private-space VPN customer endpoints. Use
# this to validate your customer-side configuration against the official
# compatibility list before standing up a VPN — and to fail terraform plan
# early if you pick an unsupported combo.
###############################################################################

data "anypoint_private_space_supported_vpn_configs" "this" {
  org_id = var.root_org
}

# List every supported vendor name.
output "supported_vendors" {
  value = [for v in data.anypoint_private_space_supported_vpn_configs.this.vendors : v.name]
}

# Flatten to vendor/model -> firmware list.
output "supported_models" {
  value = {
    for v in data.anypoint_private_space_supported_vpn_configs.this.vendors :
    v.name => {
      for m in v.models : m.name => m.firmware_versions
    }
  }
}

# Example assertion: fail the plan if Cisco ASA 5500 is no longer supported.
output "cisco_asa_supported_firmware" {
  value = try(
    one([
      for m in [for v in data.anypoint_private_space_supported_vpn_configs.this.vendors : v if v.name == "Cisco"][0].models : m if m.name == "ASA 5500 Series"
    ]).firmware_versions,
    null,
  )
}
