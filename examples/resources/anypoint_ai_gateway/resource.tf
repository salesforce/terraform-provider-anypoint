# Minimal: bedrock-anthropic LLM gateway
resource "anypoint_ai_gateway" "minimal" {
  org_id          = var.org_id
  env_id          = var.env_id
  api_instance_id = anypoint_apim_mule4.llm_api.id

  llm_provider {
    type = "bedrock-anthropic"
    bedrock {
      aws_region            = "us-east-1"
      aws_access_key_id     = var.aws_access_key_id
      aws_secret_access_key = var.aws_secret_access_key
    }
  }
}

# Full-featured: prompt guard + model-based routing + telemetry + LLM-token rate limit
resource "anypoint_ai_gateway" "prod" {
  org_id          = var.org_id
  env_id          = var.env_id
  api_instance_id = anypoint_apim_mule4.llm_api.id

  llm_provider {
    type = "bedrock-anthropic"
    bedrock {
      aws_region            = "us-east-1"
      aws_access_key_id     = var.aws_access_key_id
      aws_secret_access_key = var.aws_secret_access_key
      timeout               = 60000
    }
  }

  prompt_guard {
    type = "semantic-openai"
    semantic_openai {
      openai_url             = "https://api.openai.com/v1"
      openai_api_key         = var.openai_api_key
      openai_embedding_model = "text-embedding-3-small"
      threshold              = 0.85
      deny_topics {
        name       = "Politics"
        embeddings = var.politics_embeddings_json
      }
      deny_topics {
        name       = "Violence"
        embeddings = var.violence_embeddings_json
      }
    }
  }

  routing {
    type = "model-based"
    model_based {
      supported_vendors {
        vendor       = "Openai"
        target_model = "gpt-4"
      }
      supported_vendors {
        vendor = "AzureOpenAI"
      }
    }
  }

  llm_rate_limit {
    maximum_tokens              = 50000
    time_period_in_milliseconds = 60000
    key_selector                = "#[attributes.headers['client-id']]"
  }

  telemetry {
    enabled         = true
    source_agent_id = "prod-gateway"
  }

  asset_versions = {
    llm_proxy_core                       = "1.0.1"
    bedrock_anthropic_transcoding_policy = "1.0.0"
    bedrock_llm_provider_policy          = "1.0.3"
    semantic_prompt_guard_policy_openai  = "1.0.1"
    model_based_routing                  = "1.0.1"
    llm_token_rate_limit                 = "1.0.1"
    agent_connection_telemetry           = "1.0.0"
  }
}

output "ai_gateway_applied_policies" {
  value = anypoint_ai_gateway.prod.applied_policies
}

output "ai_gateway_external_drift" {
  value = anypoint_ai_gateway.prod.external_policies_detected
}
