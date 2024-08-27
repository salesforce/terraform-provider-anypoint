resource "anypoint_fabrics_associations" "assoc" {
  org_id = var.root_org
  fabrics_id = "4c641268-3917-45b0-acb8-f7cb0c0318ab"

  # Associate a specific environment in a specific org
  associations {
    env_id = "7074fcee-9b23-4ab6-97e8-5de5f4aef17d"
    org_id = "aa1f00d6-213d-4f60-845b-207286484bd1"
  }

  # Associate all sandbox environments for all orgs
  associations {
    env_id = "sandbox"
    org_id = "all"
  }

  # Associate all production environments for all orgs
  associations {
    env_id = "production"
    org_id = "all"
  }

  # Associate all sandbox environments for a specific org
  associations {
    env_id = "sandbox"
    org_id = "aa1f00d6-213d-4f60-845b-207286484bd1"
  }

  # Associate all environments for all orgs
  associations {
    env_id = "all"
    org_id = "all"
  }
}
