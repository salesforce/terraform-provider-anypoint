data "anypoint_exchange_policy_template" "policy" {
  org_id = var.root_org
  group_id = "68ef9520-24e9-4cf2-b2f5-620025690913"
  id = "rate-limiting"
  version = "1.4.0"
}