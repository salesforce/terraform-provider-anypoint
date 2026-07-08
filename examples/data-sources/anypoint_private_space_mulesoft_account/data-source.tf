# Returns the region-bound MuleSoft AWS account id for the given private space.
# Wire this into an aws_ram_principal_association so MuleSoft can attach the
# transit gateway through your RAM share.

data "anypoint_private_space_mulesoft_account" "this" {
  org_id           = var.root_org
  private_space_id = "7f747999-a9bb-41d5-bfe8-d6b8cca68c62"
}

output "mulesoft_aws_account_id" {
  value = data.anypoint_private_space_mulesoft_account.this.account_id
}
