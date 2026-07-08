# Reads the static route table of the private space network (destination + target).
# Targets include "Local", "IGW", "TGW" (with connection metadata), etc.

data "anypoint_private_space_routes" "rt" {
  org_id           = var.root_org
  private_space_id = "7f747999-a9bb-41d5-bfe8-d6b8cca68c62"
}

output "ps_routes" {
  value = data.anypoint_private_space_routes.rt.routes
}
