data "anypoint_fabrics_list" "all" {
  org_id = var.org_id
}

output "all" {
  value = data.anypoint_fabrics_list.all.list
}