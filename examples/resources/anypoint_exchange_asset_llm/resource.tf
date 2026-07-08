# Minimal: OpenAI-backed LLM asset
resource "anypoint_exchange_asset_llm" "openai" {
  org_id   = var.org_id
  group_id = var.org_id
  asset_id = "my-openai-llm"
  version  = "1.0.0"

  name     = "My OpenAI LLM"
  platform = "openai"
}

# Bedrock-backed LLM asset with description + tags
resource "anypoint_exchange_asset_llm" "bedrock" {
  org_id   = var.org_id
  group_id = var.org_id
  asset_id = "my-bedrock-llm"
  version  = "1.0.0"

  name        = "My Bedrock LLM"
  description = "Bedrock-fronted LLM API for the AI Gateway composite."
  platform    = "bedrock"
  tags        = ["llm", "bedrock", "prod"]
}

# Wire into the AI Gateway composite end-to-end:
# the apim_mule4 instance must be backed by this LLM asset, otherwise
# Anypoint rejects llm-proxy-core POST with PolicyValidationError.
resource "anypoint_apim_mule4" "llm_api" {
  org_id           = var.org_id
  env_id           = var.env_id
  asset_group_id   = anypoint_exchange_asset_llm.bedrock.group_id
  asset_id         = anypoint_exchange_asset_llm.bedrock.asset_id
  asset_version    = anypoint_exchange_asset_llm.bedrock.version
  instance_label   = "ai-gw-bedrock"
  endpoint_uri     = "https://example.invalid/llm"
  endpoint_isCloudHub = false
}

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
    }
  }
}
