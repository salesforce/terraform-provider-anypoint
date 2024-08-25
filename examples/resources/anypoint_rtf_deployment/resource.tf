resource "anypoint_rtf_deployment" "deployment" {
  org_id = var.root_org
  env_id = var.env_id
  name   = "your-awesome-app"
  application {
    desired_state = "STARTED"
    ref {
      group_id    = var.root_org
      artifact_id = "your-artifact-id"
      version     = "1.0.2"
      packaging   = "jar"
    }
    configuration {
      mule_agent_app_props_service {
        properties = {
          props1 = "value01"
          props2 = "value02"
        }
        secure_properties = {
          secure_props1 = "secret_value"
        }
      }
      mule_agent_logging_service {
        scope_logging_configurations {
          scope     = "mule.package"
          log_level = "DEBUG"
        }
      }
    }
  }

  target {
    provider = "MC"
    target_id = var.fabrics_id
    replicas = 1
    deployment_settings {
      clustered = false
      enforce_deploying_replicas_across_nodes = false
      http {
        inbound_public_url = "http://private.example.net/(.+),http://another.example.net/(.+)"
        inbound_last_mile_security = true
        inbound_forward_ssl_session = false
      }
      jvm_args = ""
      update_strategy = "rolling"
      disable_am_log_forwarding = false
      persistent_object_store = false
      disable_external_log_forwarding = false
      generate_default_public_url = false
      runtime {
        version = "4.7.0:20e-java8"
      }
      resources {
        cpu_reserved = "100m"
        cpu_limit = "1000m"
        memory_reserved = "1000Mi"
        memory_limit = "1000Mi"
      }
    }
  }
}