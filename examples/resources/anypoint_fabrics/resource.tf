resource "anypoint_fabrics" "fabrics" {
  org_id = var.root_org
  name = "terraform-eks-rtf"
  region = "us-east-1"
  vendor = "eks"
}
