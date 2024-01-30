resource "anypoint_apim_flexgateway" "fg" {
  org_id = var.root_org
  env_id = "7074fcdd-9b23-4ab3-97c8-5db5f4adf17d"
  asset_group_id = var.root_org
  asset_id = "flex-backend-app-test"
  asset_version = "1.0.0"
  deployment_target_id = data.anypoint_flexgateway_target.target.id
  deployment_target_name = data.anypoint_flexgateway_target.target.name
  deployment_gateway_version = data.anypoint_flexgateway_target.target.version
  deployment_expected_status = "deployed"
  deployment_overwrite = true
  deployment_type = "HY"
  instance_label  = "my terraform flex instance"
  endpoint_proxy_uri = "http://consumer.url/hello/world/2"
  routing {
    label = "my-route01"
    upstreams {
      label = "upstream01"
      weight = 100
    }
    rules {
      methods = [ "POST", "GET" ]
      host = ".*"
      path = "/.*"
      headers = {
        "x-example-header" = ".*"
        "x-correlation-id" = ".*"
      }
    }
  }
  upstreams {
    label = "upstream01"
    uri = "http://192.168.1.166:3000"
  }
}