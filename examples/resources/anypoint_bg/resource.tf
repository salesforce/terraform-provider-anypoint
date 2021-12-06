resource "anypoint_bg" "bg" {
  name = "YOUR_BG_NAME"
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