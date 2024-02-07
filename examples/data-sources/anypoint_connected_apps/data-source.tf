data "anypoint_connected_apps" "clients" {
    org_id = var.root_org
    params {
        search = "my-app"
        limit = 10
    }
}