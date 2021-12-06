data "anypoint_rolegroups" "result" {
  org_id = var.root_org   # business group id
}

output "rgs" {
  value = data.anypoint_rolegroups.result
}