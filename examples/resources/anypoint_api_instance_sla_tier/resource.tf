resource "anypoint_api_instance_sla_tier" "gold" {
  org_id       = var.org_id
  env_id       = var.env_id
  api_id       = anypoint_apim_mule4.api.id
  name         = "Gold"
  description  = "High throughput tier"
  auto_approve = false
  status       = "ACTIVE"

  limits {
    visible                    = true
    time_period_in_milliseconds = 60000
    maximum_requests           = 1000
  }

  limits {
    visible                    = false
    time_period_in_milliseconds = 3600000
    maximum_requests           = 10000
  }
}
