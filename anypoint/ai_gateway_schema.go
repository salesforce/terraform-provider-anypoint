package anypoint

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// aiGatewaySchema returns the full schema map for anypoint_ai_gateway.
// Sub-block schemas are split into helpers below to keep this readable.
func aiGatewaySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"last_updated": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The last time this resource has been updated locally.",
		},
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Composite identifier: {org_id}/{env_id}/{api_instance_id}/ai_gateway.",
		},
		"org_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The organization id where the api instance is defined.",
		},
		"env_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The environment id where the api instance is defined.",
		},
		"api_instance_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The API Manager instance id to attach the gateway policies to. Reference `anypoint_apim_mule4.<x>.id` or `anypoint_apim_flexgateway.<x>.id`. The underlying Exchange asset must be of LLMAsset type for the LLM-class policies (every policy in the composite except `agent-connection-telemetry` and `rate-limiting`) to apply — Anypoint enforces this at policy POST time and surfaces a 400 if the asset is not LLM-typed.",
		},

		"llm_provider":       {Type: schema.TypeList, Required: true, MaxItems: 1, Elem: &schema.Resource{Schema: aiGatewayLLMProviderSchema()}, Description: "The upstream LLM provider configuration. Resolves to the provider + transcoding policy pair (or provider-only for `bedrock` pass-through)."},
		"prompt_guard":       {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: aiGatewayPromptGuardSchema()}, Description: "Optional input filtering block. Pick one guard sub-block matching `type`."},
		"routing":            {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: aiGatewayRoutingSchema()}, Description: "Optional upstream-routing block."},
		"llm_rate_limit":     {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: aiGatewayLLMRateLimitSchema()}, Description: "Token-aware rate limit. Maps to `llm-token-rate-limit`. Mutually exclusive with `request_rate_limit`."},
		"request_rate_limit": {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: aiGatewayRequestRateLimitSchema()}, Description: "Generic request-count rate limit. Maps to `rate-limiting`. Mutually exclusive with `llm_rate_limit`."},
		"telemetry":          {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: aiGatewayTelemetrySchema()}, Description: "Optional `agent-connection-telemetry` block. Always applied at inbound order 1 when present."},

		"asset_versions": {
			Type:        schema.TypeMap,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Optional per-policy version override. Keys are snake_case-ified asset_ids (e.g. `bedrock_llm_provider_policy`). Falls through to the composite's known-good defaults when omitted.",
		},

		"applied_policies": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Computed list of policies the composite owns on the api_instance, with their resolved order and version. Order is per-injection-point.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"asset_id":        {Type: schema.TypeString, Computed: true},
					"id":              {Type: schema.TypeString, Computed: true},
					"order":           {Type: schema.TypeInt, Computed: true},
					"injection_point": {Type: schema.TypeString, Computed: true},
					"version":         {Type: schema.TypeString, Computed: true},
				},
			},
		},
		"external_policies_detected": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "True when the api_instance carries policies the composite does not recognise. Composite leaves them untouched on Update/Delete.",
		},
	}
}

func aiGatewayLLMProviderSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"type": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"bedrock", "bedrock-anthropic"}, false),
			Description:  "Provider family. `bedrock` calls Bedrock directly (no transcoding); `bedrock-anthropic` applies the `bedrock-anthropic-transcoding-policy` so OpenAI-shaped client requests are translated to Anthropic Claude requests. OpenAI and Gemini providers will land in a later release.",
		},
		"bedrock": {
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Description: "Configuration for the `bedrock-llm-provider-policy`. Required when `type` is `bedrock` or `bedrock-anthropic`.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"aws_region":                     {Type: schema.TypeString, Required: true, Description: "AWS region (`awsRegion`)."},
					"aws_access_key_id":              {Type: schema.TypeString, Optional: true, Sensitive: true, Description: "Static AWS Access Key ID (`awsAccessKeyId`). Provide either this or `aws_access_key_id_selector`."},
					"aws_access_key_id_selector":     {Type: schema.TypeString, Optional: true, Description: "DataWeave selector resolving the AWS Access Key ID at runtime (`awsAccessKeyIdSelector`)."},
					"aws_secret_access_key":          {Type: schema.TypeString, Optional: true, Sensitive: true, Description: "Static AWS Secret Access Key (`awsSecretAccessKey`)."},
					"aws_secret_access_key_selector": {Type: schema.TypeString, Optional: true, Description: "DataWeave selector for the AWS Secret Access Key (`awsSecretAccessKeySelector`)."},
					"aws_session_token":              {Type: schema.TypeString, Optional: true, Sensitive: true, Description: "Optional AWS Session Token for temporary credentials (`awsSessionToken`)."},
					"service_name":                   {Type: schema.TypeString, Optional: true, Default: "bedrock", Description: "AWS service name for signing requests (`serviceName`). Defaults to `bedrock`."},
					"timeout":                        {Type: schema.TypeInt, Optional: true, Default: 60000, Description: "Timeout in milliseconds for the Bedrock API call (`timeout`)."},
				},
			},
		},
	}
}

func aiGatewayPromptGuardSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"type": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"semantic-openai"}, false),
			Description:  "Guard family. `semantic-openai` applies `semantic-prompt-guard-policy-openai`. Other guard families (regex, semantic-huggingface, bedrock-guardrails) will land in a later release.",
		},
		"semantic_openai": {
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Description: "Configuration for the `semantic-prompt-guard-policy-openai`.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"openai_url":             {Type: schema.TypeString, Required: true, Description: "OpenAI embeddings endpoint URL (`openaiUrl`)."},
					"openai_api_key":         {Type: schema.TypeString, Required: true, Sensitive: true, Description: "OpenAI API key (`openaiApiKey`)."},
					"openai_embedding_model": {Type: schema.TypeString, Required: true, Description: "OpenAI embedding model identifier (`openaiEmbeddingModel`)."},
					"openai_provider":        {Type: schema.TypeString, Optional: true, Description: "Optional OpenAI provider override (`openaiProvider`)."},
					"threshold":              {Type: schema.TypeFloat, Required: true, Description: "Cosine-similarity threshold (`threshold`) above which a prompt is blocked."},
					"deny_topics": {
						Type:        schema.TypeList,
						Required:    true,
						Description: "List of denied topics (`denyTopics`). Each entry pairs a topic name with its pre-computed embedding JSON string.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name":       {Type: schema.TypeString, Required: true, Description: "Topic name (`name`), e.g. `Politics`, `Violence`."},
								"embeddings": {Type: schema.TypeString, Required: true, Sensitive: true, Description: "Pre-computed embedding vectors for the topic's example utterances, as a JSON string (`embeddings`)."},
							},
						},
					},
					"timeout":                {Type: schema.TypeInt, Optional: true, Description: "Timeout in milliseconds (`timeout`)."},
				},
			},
		},
	}
}

func aiGatewayRoutingSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"type": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"model-based"}, false),
			Description:  "Routing family. `model-based` applies `model-based-routing`. Semantic routing families will land in a later release.",
		},
		"model_based": {
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Description: "Configuration for the `model-based-routing` policy.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"supported_vendors": {
						Type:        schema.TypeList,
						Required:    true,
						Description: "List of supported provider vendors (`supportedVendors`). Routing decides upstream based on the inbound `model` field matching one of these.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"vendor":       {Type: schema.TypeString, Required: true, Description: "LLM vendor name (`vendor`), e.g. `Openai`, `Gemini`, `AzureOpenAI`."},
								"target_model": {Type: schema.TypeString, Optional: true, Description: "Optional target model from the vendor (`targetModel`), e.g. `gpt-4`, `gemini-pro`."},
							},
						},
					},
					"fallback":          {Type: schema.TypeString, Optional: true, Description: "Optional fallback target identifier (`fallback`) when no rule matches."},
				},
			},
		},
	}
}

func aiGatewayLLMRateLimitSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"maximum_tokens":              {Type: schema.TypeInt, Required: true, Description: "Maximum tokens per window (`maximumTokens`)."},
		"time_period_in_milliseconds": {Type: schema.TypeInt, Required: true, Description: "Window size in milliseconds (`timePeriodInMilliseconds`)."},
		"key_selector":                {Type: schema.TypeString, Required: true, Description: "DataWeave expression identifying the rate-limit key (`keySelector`)."},
	}
}

func aiGatewayRequestRateLimitSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"clusterizable":  {Type: schema.TypeBool, Optional: true, Default: false, Description: "Whether the rate limit applies cluster-wide (`clusterizable`)."},
		"expose_headers": {Type: schema.TypeBool, Optional: true, Default: false, Description: "Whether to expose rate-limit headers in responses (`exposeHeaders`)."},
		"key_selector":   {Type: schema.TypeString, Optional: true, Description: "Optional DataWeave expression identifying the rate-limit key (`keySelector`). When unset, a global counter is used."},
		"rate_limits": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "List of rate-limit windows. Each entry maps to one element of the policy's `rateLimits` array.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"maximum_requests":            {Type: schema.TypeInt, Required: true, Description: "Max requests per window (`maximumRequests`)."},
					"time_period_in_milliseconds": {Type: schema.TypeInt, Required: true, Description: "Window size in milliseconds (`timePeriodInMilliseconds`)."},
				},
			},
		},
	}
}

func aiGatewayTelemetrySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"enabled":         {Type: schema.TypeBool, Optional: true, Default: true, Description: "When `false`, the telemetry policy is not applied."},
		"source_agent_id": {Type: schema.TypeString, Optional: true, Description: "Identifier emitted by the telemetry policy (`sourceAgentId`)."},
	}
}
