data "anypoint_rolegroup" "result" {
  id = "ROLEGROUP_ID"
}

output "rg" {
  value = data.anypoint_rolegroup.result
}