package anypoint

// MuleSoft public mule4 policies Exchange group. All AI Gateway catalog assets
// live under this group. Probed 2026-06-04.
const AiGatewayPolicyGroupID = "68ef9520-24e9-4cf2-b2f5-620025690913"

// Catalog asset IDs.
const (
	AssetLLMProxyCore                       = "llm-proxy-core"
	AssetBedrockProvider                    = "bedrock-llm-provider-policy"
	AssetBedrockAnthropicTranscoding        = "bedrock-anthropic-transcoding-policy"
	AssetGeminiProvider                     = "gemini-llm-provider-policy"
	AssetGeminiTranscoding                  = "gemini-transcoding-policy"
	AssetOpenAITranscoding                  = "openai-transcoding-policy"
	AssetRegexPromptGuard                   = "regex-prompt-guard-policy"
	AssetSemanticPromptGuardOpenAI          = "semantic-prompt-guard-policy-openai"
	AssetSemanticPromptGuardHuggingface     = "semantic-prompt-guard-policy-huggingface"
	AssetBedrockGuardrails                  = "bedrock-guardrails-policy"
	AssetModelBasedRouting                  = "model-based-routing"
	AssetSemanticRoutingOpenAI              = "semantic-routing-policy-openai"
	AssetSemanticRoutingHuggingface         = "semantic-routing-policy-huggingface"
	AssetLLMTokenRateLimit                  = "llm-token-rate-limit"
	AssetRateLimiting                       = "rate-limiting"
	AssetAgentConnectionTelemetry           = "agent-connection-telemetry"
)

const (
	AiGatewayInjectionInbound  = "inbound"
	AiGatewayInjectionOutbound = "outbound"
)

type aiGatewayPolicyEntry struct {
	AssetID        string
	InjectionPoint string
}

// aiGatewayPolicyCatalog enumerates every policy the composite knows how to
// apply. Lookup by asset_id is used during Read to decide whether a policy on
// the api_instance belongs to the composite.
var aiGatewayPolicyCatalog = map[string]aiGatewayPolicyEntry{
	AssetLLMProxyCore:                   {AssetID: AssetLLMProxyCore, InjectionPoint: AiGatewayInjectionInbound},
	AssetBedrockProvider:                {AssetID: AssetBedrockProvider, InjectionPoint: AiGatewayInjectionOutbound},
	AssetBedrockAnthropicTranscoding:    {AssetID: AssetBedrockAnthropicTranscoding, InjectionPoint: AiGatewayInjectionOutbound},
	AssetGeminiProvider:                 {AssetID: AssetGeminiProvider, InjectionPoint: AiGatewayInjectionOutbound},
	AssetGeminiTranscoding:              {AssetID: AssetGeminiTranscoding, InjectionPoint: AiGatewayInjectionOutbound},
	AssetOpenAITranscoding:              {AssetID: AssetOpenAITranscoding, InjectionPoint: AiGatewayInjectionOutbound},
	AssetRegexPromptGuard:               {AssetID: AssetRegexPromptGuard, InjectionPoint: AiGatewayInjectionInbound},
	AssetSemanticPromptGuardOpenAI:      {AssetID: AssetSemanticPromptGuardOpenAI, InjectionPoint: AiGatewayInjectionInbound},
	AssetSemanticPromptGuardHuggingface: {AssetID: AssetSemanticPromptGuardHuggingface, InjectionPoint: AiGatewayInjectionInbound},
	AssetBedrockGuardrails:              {AssetID: AssetBedrockGuardrails, InjectionPoint: AiGatewayInjectionInbound},
	AssetModelBasedRouting:              {AssetID: AssetModelBasedRouting, InjectionPoint: AiGatewayInjectionInbound},
	AssetSemanticRoutingOpenAI:          {AssetID: AssetSemanticRoutingOpenAI, InjectionPoint: AiGatewayInjectionInbound},
	AssetSemanticRoutingHuggingface:     {AssetID: AssetSemanticRoutingHuggingface, InjectionPoint: AiGatewayInjectionInbound},
	AssetLLMTokenRateLimit:              {AssetID: AssetLLMTokenRateLimit, InjectionPoint: AiGatewayInjectionInbound},
	AssetRateLimiting:                   {AssetID: AssetRateLimiting, InjectionPoint: AiGatewayInjectionInbound},
	AssetAgentConnectionTelemetry:       {AssetID: AssetAgentConnectionTelemetry, InjectionPoint: AiGatewayInjectionInbound},
}

// aiGatewayDefaultVersions pins the version of each catalog policy applied by
// the composite when the user does not override via `asset_versions`. Snapshot
// taken 2026-06-04; bump per release after smoke-testing newer template
// versions.
var aiGatewayDefaultVersions = map[string]string{
	AssetLLMProxyCore:                   "1.0.1",
	AssetBedrockProvider:                "1.0.3",
	AssetBedrockAnthropicTranscoding:    "1.0.0",
	AssetGeminiProvider:                 "1.0.2",
	AssetGeminiTranscoding:              "1.0.0",
	AssetOpenAITranscoding:              "1.0.2",
	AssetRegexPromptGuard:               "1.0.0",
	AssetSemanticPromptGuardOpenAI:      "1.0.1",
	AssetSemanticPromptGuardHuggingface: "1.0.0",
	AssetBedrockGuardrails:              "1.0.0",
	AssetModelBasedRouting:              "1.0.1",
	AssetSemanticRoutingOpenAI:          "1.0.1",
	AssetSemanticRoutingHuggingface:     "1.0.1",
	AssetLLMTokenRateLimit:              "1.0.1",
	AssetRateLimiting:                   "1.4.1",
	AssetAgentConnectionTelemetry:       "1.0.0",
}

// aiGatewayCatalogContains reports whether the asset_id is one the composite
// recognizes. Used during Read to separate composite-owned from external
// policies on the api_instance.
func aiGatewayCatalogContains(assetID string) bool {
	_, ok := aiGatewayPolicyCatalog[assetID]
	return ok
}

// aiGatewayPolicyVersion resolves the version to apply for a catalog asset.
// Caller may pass an override map (from the `asset_versions` schema block);
// nil override falls through to the default.
func aiGatewayPolicyVersion(assetID string, overrides map[string]string) string {
	if overrides != nil {
		if v, ok := overrides[assetID]; ok && v != "" {
			return v
		}
	}
	return aiGatewayDefaultVersions[assetID]
}
