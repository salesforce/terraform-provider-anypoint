data "anypoint_dlb" "dlb" {
  org_id = var.root_org                 # The Business Group Id
  vpc_id = "vpc-0c87748024561f029"      # The VPC id
  id = "61ad0a3a0322c72f8129bbca"       # The DLB id
}

