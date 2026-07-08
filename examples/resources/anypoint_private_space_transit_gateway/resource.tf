# Attach an AWS Transit Gateway to an Anypoint CloudHub 2.0 private space.
#
# Prereqs (AWS side):
#   1. An aws_ec2_transit_gateway in the same region as the private space.
#   2. An aws_ram_resource_share with allow_external_principals = true.
#   3. The MuleSoft AWS account whitelisted as principal on that share.
#   4. The transit gateway associated to the share.
#
# Use the anypoint_private_space_mulesoft_account data source to discover the
# MuleSoft AWS account id for the target private space — it is region-bound,
# so do NOT hardcode it.

data "anypoint_private_space_mulesoft_account" "this" {
  org_id           = var.root_org
  private_space_id = var.private_space_id
}

resource "aws_ec2_transit_gateway" "this" {
  description                    = "anypoint-tgw"
  amazon_side_asn                = 64512
  auto_accept_shared_attachments = "enable"

  tags = {
    Name = "anypoint-tgw"
  }
}

resource "aws_ram_resource_share" "this" {
  name                      = "anypoint-tgw-share"
  allow_external_principals = true
}

resource "aws_ram_principal_association" "mulesoft" {
  principal          = data.anypoint_private_space_mulesoft_account.this.account_id
  resource_share_arn = aws_ram_resource_share.this.arn
}

resource "aws_ram_resource_association" "tgw" {
  resource_arn       = aws_ec2_transit_gateway.this.arn
  resource_share_arn = aws_ram_resource_share.this.arn
}

data "aws_caller_identity" "current" {}

resource "anypoint_private_space_transit_gateway" "this" {
  org_id           = var.root_org
  private_space_id = var.private_space_id
  name             = "my-tgw"

  # Trailing UUID of the RAM share ARN (NOT the full ARN).
  resource_share_id = element(
    split("/", aws_ram_resource_share.this.arn),
    length(split("/", aws_ram_resource_share.this.arn)) - 1,
  )
  resource_share_account = data.aws_caller_identity.current.account_id

  # Static routes pushed into the private space route table.
  # Must NOT equal or be more specific than the private space CIDR.
  routes = ["10.4.0.0/18"]

  depends_on = [
    aws_ram_resource_association.tgw,
    aws_ram_principal_association.mulesoft,
  ]
}
