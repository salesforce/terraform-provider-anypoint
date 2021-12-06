data "anypoint_roles" "roles" {
  params {
    name             = ""     # search by the name of a role
    description      = ""     # search by the description of a role 
    include_internal = ""     # to include internal roles in results
    search           = ""     # a search string to use for partial matches of role names
    offset           = 0      # pagination parameter to start returning results from this position of matches. default 0
    limit            = 150    # pagination parameter for how many results to return. default 200
    ascending        = true   # sort order for filtering. default true
  }
}

output "roles" {
  value = data.anypoint_roles.roles
}