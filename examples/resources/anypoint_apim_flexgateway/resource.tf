resource "anypoint_apim_flexgateway" "fg" {
  org_id = var.root_org
  env_id = var.env_id
  asset_group_id = var.root_org
  asset_id = "flex-backend-app-test"
  asset_version = "1.0.0"
  deployment_target_id = "c33dac89-4ca6-4951-9ad5-19ace129029e"
  deployment_expected_status = "deployed"
  deployment_overwrite = true
  deployment_type = "HY"
  instance_label  = "my flex instance"
  endpoint_uri = "http://consumer.url"
  routing {
    label = "my-route01"
    upstreams {
      label = "upstream03"
      weight = 70
    }
    upstreams {
      label = "upstream01"
      weight = 30
    }
    rules {
      methods = [ "POST", "GET" ]
      host = "http://.*example\\.com"
      path = "/api/.*"
      headers = {
        "x-example-header" = ".*"
        "x-correlation-id" = "[\\W\\d\\-]+"
      }
    }
  }
  upstreams {
    label = "upstream01"
    uri = "http://google.com"
    tls_context {
      secret_group_id = "39731075-0521-47aa-82b2-d9745f2ac2eb"
      tls_context_id = "a20282b6-3708-4f2a-93b2-18fdd7d9fa34"
    }
  }
  upstreams {
    label = "upstream03"
    uri = "http://helloworld.com"
  }
}