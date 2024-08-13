data "anypoint_fabrics_associations" "assoc" {
  fabrics_id = "YOUR_FABRICS_ID"
  org_id = var.root_org
}

output "associations" {
  value = data.anypoint_fabrics_associations.assoc.associations
}