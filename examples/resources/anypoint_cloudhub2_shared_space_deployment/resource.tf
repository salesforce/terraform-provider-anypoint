resource "anypoint_cloudhub2_shared_space_deployment" "deployment" {
  org_id = var.root_org
  env_id = var.env_id
  name   = "your-awesome-app"
  application {
    desired_state = "STARTED"
    vcores = 0.1
    object_store_v2_enabled = true
    ref {
      group_id    = var.root_org
      artifact_id = "your-awesome-app-artifact"
      version     = "1.0.0"
      packaging   = "jar"
    }
    configuration {
      mule_agent_app_props_service {
        properties = {
          props1 = "value"
          props2 = "value"
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
    target_id = "cloudhub-us-east-1"
    replicas = 1
    deployment_settings {
      clustered = false
      jvm_args = ""
      update_strategy = "rolling"
      disable_am_log_forwarding = true
      persistent_object_store = true
      disable_external_log_forwarding = true
      generate_default_public_url = true
      runtime {
        version = "4.7.0:20e-java8"
      }
      http {
        inbound_last_mile_security = true
      }
    }
  }
}