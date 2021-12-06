resource "anypoint_rolegroup" "rg" {
  org_id = var.root_org
  name = "arolegroup_example"
  description = "This a rolegroup example "
  external_names = tolist(["VAL1", "VAL2"])
}