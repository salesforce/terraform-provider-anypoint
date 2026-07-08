package anypoint

import (
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resolvedPolicy is the in-memory representation of a single policy the
// composite wants to apply. Inbound entries carry their explicit order;
// outbound entries are sequenced by their position in the resolved list.
type resolvedPolicy struct {
	AssetID        string
	Version        string
	InjectionPoint string
	Order          int            // inbound order; outbound entries use position-in-slice
	Config         map[string]any // configurationData payload (camelCase keys per Anypoint API)
}

// aiGatewayResourceLike abstracts schema.ResourceData and schema.ResourceDiff
// so the resolver can be exercised from both CRUD and CustomizeDiff.
type aiGatewayResourceLike interface {
	Get(key string) any
	GetOk(key string) (any, bool)
}

// resolvePolicies converts the schema state into the ordered list of policies
// the composite will POST. Returns inbound + outbound slices separately so the
// caller can sequence the API calls correctly.
//
// Inbound order rule (skipped slots collapse):
//
//	1. agent-connection-telemetry
//	2. {prompt guard}
//	3. {routing}
//	4. {llm_rate_limit | request_rate_limit}
//	5. llm-proxy-core         (ALWAYS applies)
//
// Outbound order rule (POST sequence; outbound policies do not honour the
// `order` field — see PR #149 outbound follow-up PATCH note):
//
//	1. {transcoding}          (only for `bedrock-anthropic`)
//	2. {provider}
func resolvePolicies(d aiGatewayResourceLike) (inbound, outbound []resolvedPolicy) {
	versions := aiGatewayAssetVersionsMap(d)

	// Inbound stack — order is composite-pinned.
	if tel := aiGatewayTelemetryPolicy(d, versions); tel != nil {
		tel.Order = 1
		inbound = append(inbound, *tel)
	}
	if g := aiGatewayGuardPolicy(d, versions); g != nil {
		g.Order = 2
		inbound = append(inbound, *g)
	}
	if r := aiGatewayRoutingPolicy(d, versions); r != nil {
		r.Order = 3
		inbound = append(inbound, *r)
	}
	if rl := aiGatewayRateLimitPolicy(d, versions); rl != nil {
		rl.Order = 4
		inbound = append(inbound, *rl)
	}
	inbound = append(inbound, resolvedPolicy{
		AssetID:        AssetLLMProxyCore,
		Version:        aiGatewayPolicyVersion(AssetLLMProxyCore, versions),
		InjectionPoint: AiGatewayInjectionInbound,
		Order:          5,
		Config:         map[string]any{},
	})

	// Outbound stack — POST sequence implies execution order.
	outbound = aiGatewayOutboundPolicies(d, versions)
	return inbound, outbound
}

func aiGatewayAssetVersionsMap(d aiGatewayResourceLike) map[string]string {
	raw, ok := d.GetOk("asset_versions")
	if !ok {
		return nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok && s != "" {
			// Schema keys are snake_case; catalog asset IDs are kebab-case.
			out[snakeToKebab(k)] = s
		}
	}
	return out
}

func snakeToKebab(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '_' {
			b[i] = '-'
		} else {
			b[i] = s[i]
		}
	}
	return string(b)
}

// aiGatewayTelemetryPolicy returns the telemetry policy iff the block is
// declared with `enabled = true` (the default when the block is declared).
func aiGatewayTelemetryPolicy(d aiGatewayResourceLike, versions map[string]string) *resolvedPolicy {
	blocks, ok := d.GetOk("telemetry")
	if !ok {
		return nil
	}
	list, ok := blocks.([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	cfg, _ := list[0].(map[string]any)
	if cfg == nil {
		return nil
	}
	enabled, _ := cfg["enabled"].(bool)
	if !enabled {
		return nil
	}
	conf := map[string]any{}
	if v, ok := cfg["source_agent_id"].(string); ok && v != "" {
		conf["sourceAgentId"] = v
	}
	return &resolvedPolicy{
		AssetID:        AssetAgentConnectionTelemetry,
		Version:        aiGatewayPolicyVersion(AssetAgentConnectionTelemetry, versions),
		InjectionPoint: AiGatewayInjectionInbound,
		Config:         conf,
	}
}

func aiGatewayGuardPolicy(d aiGatewayResourceLike, versions map[string]string) *resolvedPolicy {
	blocks, ok := d.GetOk("prompt_guard")
	if !ok {
		return nil
	}
	list, ok := blocks.([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	cfg, _ := list[0].(map[string]any)
	if cfg == nil {
		return nil
	}
	switch cfg["type"].(string) {
	case "semantic-openai":
		sub := firstSubBlock(cfg, "semantic_openai")
		if sub == nil {
			return nil
		}
		conf := map[string]any{}
		copyIfStringSet(sub, conf, "openai_url", "openaiUrl")
		copyIfStringSet(sub, conf, "openai_api_key", "openaiApiKey")
		copyIfStringSet(sub, conf, "openai_embedding_model", "openaiEmbeddingModel")
		copyIfStringSet(sub, conf, "openai_provider", "openaiProvider")
		if v, ok := sub["threshold"].(float64); ok {
			conf["threshold"] = v
		}
		if v, ok := sub["deny_topics"].([]any); ok && len(v) > 0 {
			topics := make([]map[string]any, 0, len(v))
			for _, t := range v {
				m, ok := t.(map[string]any)
				if !ok {
					continue
				}
				entry := map[string]any{}
				if s, _ := m["name"].(string); s != "" {
					entry["name"] = s
				}
				if s, _ := m["embeddings"].(string); s != "" {
					entry["embeddings"] = s
				}
				topics = append(topics, entry)
			}
			conf["denyTopics"] = topics
		}
		if v, ok := sub["timeout"].(int); ok && v > 0 {
			conf["timeout"] = v
		}
		return &resolvedPolicy{
			AssetID:        AssetSemanticPromptGuardOpenAI,
			Version:        aiGatewayPolicyVersion(AssetSemanticPromptGuardOpenAI, versions),
			InjectionPoint: AiGatewayInjectionInbound,
			Config:         conf,
		}
	}
	return nil
}

func aiGatewayRoutingPolicy(d aiGatewayResourceLike, versions map[string]string) *resolvedPolicy {
	blocks, ok := d.GetOk("routing")
	if !ok {
		return nil
	}
	list, ok := blocks.([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	cfg, _ := list[0].(map[string]any)
	if cfg == nil {
		return nil
	}
	switch cfg["type"].(string) {
	case "model-based":
		sub := firstSubBlock(cfg, "model_based")
		if sub == nil {
			return nil
		}
		conf := map[string]any{}
		if v, ok := sub["supported_vendors"].([]any); ok && len(v) > 0 {
			vendors := make([]map[string]any, 0, len(v))
			for _, e := range v {
				m, ok := e.(map[string]any)
				if !ok {
					continue
				}
				entry := map[string]any{}
				if s, _ := m["vendor"].(string); s != "" {
					entry["vendor"] = s
				}
				if s, _ := m["target_model"].(string); s != "" {
					entry["targetModel"] = s
				}
				vendors = append(vendors, entry)
			}
			conf["supportedVendors"] = vendors
		}
		copyIfStringSet(sub, conf, "fallback", "fallback")
		return &resolvedPolicy{
			AssetID:        AssetModelBasedRouting,
			Version:        aiGatewayPolicyVersion(AssetModelBasedRouting, versions),
			InjectionPoint: AiGatewayInjectionInbound,
			Config:         conf,
		}
	}
	return nil
}

func aiGatewayRateLimitPolicy(d aiGatewayResourceLike, versions map[string]string) *resolvedPolicy {
	if blocks, ok := d.GetOk("llm_rate_limit"); ok {
		if list, _ := blocks.([]any); len(list) > 0 {
			cfg, _ := list[0].(map[string]any)
			conf := map[string]any{}
			if v, ok := cfg["maximum_tokens"].(int); ok {
				conf["maximumTokens"] = v
			}
			if v, ok := cfg["time_period_in_milliseconds"].(int); ok {
				conf["timePeriodInMilliseconds"] = v
			}
			copyIfStringSet(cfg, conf, "key_selector", "keySelector")
			return &resolvedPolicy{
				AssetID:        AssetLLMTokenRateLimit,
				Version:        aiGatewayPolicyVersion(AssetLLMTokenRateLimit, versions),
				InjectionPoint: AiGatewayInjectionInbound,
				Config:         conf,
			}
		}
	}
	if blocks, ok := d.GetOk("request_rate_limit"); ok {
		if list, _ := blocks.([]any); len(list) > 0 {
			cfg, _ := list[0].(map[string]any)
			conf := map[string]any{}
			if v, ok := cfg["clusterizable"].(bool); ok {
				conf["clusterizable"] = v
			}
			if v, ok := cfg["expose_headers"].(bool); ok {
				conf["exposeHeaders"] = v
			}
			copyIfStringSet(cfg, conf, "key_selector", "keySelector")
			if v, ok := cfg["rate_limits"].([]any); ok {
				windows := make([]map[string]any, 0, len(v))
				for _, e := range v {
					ew, _ := e.(map[string]any)
					if ew == nil {
						continue
					}
					w := map[string]any{}
					if x, ok := ew["maximum_requests"].(int); ok {
						w["maximumRequests"] = x
					}
					if x, ok := ew["time_period_in_milliseconds"].(int); ok {
						w["timePeriodInMilliseconds"] = x
					}
					windows = append(windows, w)
				}
				conf["rateLimits"] = windows
			}
			return &resolvedPolicy{
				AssetID:        AssetRateLimiting,
				Version:        aiGatewayPolicyVersion(AssetRateLimiting, versions),
				InjectionPoint: AiGatewayInjectionInbound,
				Config:         conf,
			}
		}
	}
	return nil
}

// aiGatewayOutboundPolicies resolves the outbound stack from `llm_provider`.
// Transcoding policy is emitted first so its POST precedes the provider POST —
// outbound execution order is determined by POST sequence.
func aiGatewayOutboundPolicies(d aiGatewayResourceLike, versions map[string]string) []resolvedPolicy {
	blocks, ok := d.GetOk("llm_provider")
	if !ok {
		return nil
	}
	list, ok := blocks.([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	cfg, _ := list[0].(map[string]any)
	if cfg == nil {
		return nil
	}
	providerType, _ := cfg["type"].(string)
	bedrockSub := firstSubBlock(cfg, "bedrock")
	if bedrockSub == nil {
		return nil
	}
	out := make([]resolvedPolicy, 0, 2)
	if providerType == "bedrock-anthropic" {
		out = append(out, resolvedPolicy{
			AssetID:        AssetBedrockAnthropicTranscoding,
			Version:        aiGatewayPolicyVersion(AssetBedrockAnthropicTranscoding, versions),
			InjectionPoint: AiGatewayInjectionOutbound,
			Order:          len(out) + 1,
			Config:         map[string]any{},
		})
	}
	bedrockConfig := map[string]any{}
	copyIfStringSet(bedrockSub, bedrockConfig, "aws_region", "awsRegion")
	copyIfStringSet(bedrockSub, bedrockConfig, "aws_access_key_id", "awsAccessKeyId")
	copyIfStringSet(bedrockSub, bedrockConfig, "aws_access_key_id_selector", "awsAccessKeyIdSelector")
	copyIfStringSet(bedrockSub, bedrockConfig, "aws_secret_access_key", "awsSecretAccessKey")
	copyIfStringSet(bedrockSub, bedrockConfig, "aws_secret_access_key_selector", "awsSecretAccessKeySelector")
	copyIfStringSet(bedrockSub, bedrockConfig, "aws_session_token", "awsSessionToken")
	copyIfStringSet(bedrockSub, bedrockConfig, "service_name", "serviceName")
	if v, ok := bedrockSub["timeout"].(int); ok && v > 0 {
		bedrockConfig["timeout"] = v
	}
	out = append(out, resolvedPolicy{
		AssetID:        AssetBedrockProvider,
		Version:        aiGatewayPolicyVersion(AssetBedrockProvider, versions),
		InjectionPoint: AiGatewayInjectionOutbound,
		Order:          len(out) + 1,
		Config:         bedrockConfig,
	})
	return out
}

func firstSubBlock(parent map[string]any, key string) map[string]any {
	v, ok := parent[key]
	if !ok {
		return nil
	}
	list, ok := v.([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	m, _ := list[0].(map[string]any)
	return m
}

func copyIfStringSet(src, dst map[string]any, srcKey, dstKey string) {
	if v, ok := src[srcKey].(string); ok && v != "" {
		dst[dstKey] = v
	}
}

// flattenResolvedAppliedPolicies renders the resolver output into the shape
// the computed `applied_policies` schema expects. Used in CustomizeDiff so the
// plan output can show the resolved stack before apply.
func flattenResolvedAppliedPolicies(inbound, outbound []resolvedPolicy) []any {
	out := make([]any, 0, len(inbound)+len(outbound))
	for _, p := range inbound {
		out = append(out, map[string]any{
			"asset_id":        p.AssetID,
			"id":              "",
			"order":           p.Order,
			"injection_point": p.InjectionPoint,
			"version":         p.Version,
		})
	}
	for _, p := range outbound {
		out = append(out, map[string]any{
			"asset_id":        p.AssetID,
			"id":              "",
			"order":           p.Order,
			"injection_point": p.InjectionPoint,
			"version":         p.Version,
		})
	}
	return out
}

// sortAppliedPoliciesForState orders the list deterministically (inbound first
// by order, then outbound by order) so plan diffs are stable.
func sortAppliedPoliciesForState(list []any) {
	sort.SliceStable(list, func(i, j int) bool {
		a, _ := list[i].(map[string]any)
		b, _ := list[j].(map[string]any)
		if a == nil || b == nil {
			return false
		}
		ai, _ := a["injection_point"].(string)
		bi, _ := b["injection_point"].(string)
		if ai != bi {
			return ai == AiGatewayInjectionInbound // inbound first
		}
		ao, _ := a["order"].(int)
		bo, _ := b["order"].(int)
		return ao < bo
	})
}

// Silence unused-warnings until Phase D consumes these symbols. The resolver
// is the load-bearing module the rest of the file depends on, so the symbols
// stay exported within the package.
var _ = func() *schema.Resource { return nil }
