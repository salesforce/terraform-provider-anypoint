variable "root_org" {
  default = "xx1f55d6-213d-4f60-845c-207286484cd1"
}

variable "owner_id" {
  default = "18f23771-c78a-4be2-af8f-1bae66f43942"
}

resource "anypoint_bg" "bg" {
  name                                    = "TEST_BG_TF"
  parent_organization_id                  = var.root_org
  owner_id                                = var.owner_id
  entitlements_createsuborgs              = true
  entitlements_createenvironments         = true
  entitlements_globaldeployment           = true
  entitlements_vcoresproduction_assigned  = 0
  entitlements_vcoressandbox_assigned     = 0
  entitlements_vcoresdesign_assigned      = 0
  entitlements_staticips_assigned         = 0
  entitlements_vpcs_assigned              = 1
  entitlements_loadbalancer_assigned      = 0
  entitlements_vpns_assigned              = 1
}

resource "anypoint_env" "env" {
  org_id  = anypoint_bg.bg.id    # environment related business group
  name    = "DEV"                  # environment name
  type    = "sandbox"              # environment type : sandbox/production
}

resource "anypoint_rolegroup" "rg" {
  org_id          = var.root_org
  name            = "arolegroup_example"
  description     = "This a rolegroup example "
  external_names  = tolist(["VAL1", "VAL2"])
}