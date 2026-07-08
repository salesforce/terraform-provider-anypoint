data "anypoint_private_space_iam_roles" "role" {
  org_id = var.root_org
  private_space_id = "326a2e9e-fb92-4518-b61e-ce1e6de08b39"
}

output "private_space_iam_roles" {
  value = data.anypoint_private_space_iam_roles.role
}