# Track drift between a declared Exchange asset version and the latest version
# published on Anypoint Exchange. Works for any Exchange asset — apim policy
# templates, custom policy templates, API specs, custom assets.

# Example 1: drift report on a single MuleSoft-published policy template.
data "anypoint_exchange_asset_version_drift" "obo" {
  org_id           = var.root_org
  asset_group_id   = "68ef9520-24e9-4cf2-b2f5-620025690913" # MuleSoft global org id
  asset_id         = "credential-injection-oauth2-obo"
  declared_version = "1.1.0"
}

output "obo_drift_report" {
  value = {
    latest    = data.anypoint_exchange_asset_version_drift.obo.latest_version
    severity  = data.anypoint_exchange_asset_version_drift.obo.drift_severity
    outdated  = data.anypoint_exchange_asset_version_drift.obo.is_outdated
    known     = data.anypoint_exchange_asset_version_drift.obo.is_known
    available = data.anypoint_exchange_asset_version_drift.obo.available_versions
  }
}

# Example 2: chain a drift report directly off an existing resource.
# Convenient for CI gates — fails the build when any managed policy goes stale.
data "anypoint_exchange_asset_version_drift" "managed" {
  org_id           = var.root_org
  asset_group_id   = anypoint_apim_policy_custom.obo.asset_group_id
  asset_id         = anypoint_apim_policy_custom.obo.asset_id
  declared_version = anypoint_apim_policy_custom.obo.asset_version
}

# Example 3: fan-out drift report across N policies.
locals {
  policies = {
    obo = {
      group_id = "68ef9520-24e9-4cf2-b2f5-620025690913"
      asset_id = "credential-injection-oauth2-obo"
      version  = "1.0.0"
    }
    client_id_enf = {
      group_id = "68ef9520-24e9-4cf2-b2f5-620025690913"
      asset_id = "client-id-enforcement"
      version  = "1.3.0"
    }
  }
}

data "anypoint_exchange_asset_version_drift" "fleet" {
  for_each         = local.policies
  org_id           = var.root_org
  asset_group_id   = each.value.group_id
  asset_id         = each.value.asset_id
  declared_version = each.value.version
}

output "fleet_outdated" {
  value = {
    for k, v in data.anypoint_exchange_asset_version_drift.fleet :
    k => v.latest_version if v.is_outdated
  }
}
