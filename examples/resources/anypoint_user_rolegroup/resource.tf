resource "anypoint_user_rolegroup" "user_rolegroup" {
  org_id = var.root_org
  user_id = anypoint_user.user.id
  rolegroup_id = anypoint_user_rolegroup.rg.id
}