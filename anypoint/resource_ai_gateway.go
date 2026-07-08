package anypoint

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mulesoft-anypoint/anypoint-client-go/apim_policy"
)

func resourceAiGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAiGatewayCreate,
		ReadContext:   resourceAiGatewayRead,
		UpdateContext: resourceAiGatewayUpdate,
		DeleteContext: resourceAiGatewayDelete,
		CustomizeDiff: resourceAiGatewayCustomizeDiff,
		Description: `
		Creates and manages an AI Gateway: a curated stack of Anypoint policies applied to an existing API Manager instance
		so that single Terraform resource expresses a complete LLM proxy. Internally the composite resolves the declared
		blocks to the right mix of inbound + outbound policies on the MuleSoft public mule4 policies group, in a
		known-good execution order.

		The underlying Exchange asset bound to ` + "`api_instance_id`" + ` must be of LLMAsset type — every LLM-class policy
		in the catalog (everything except ` + "`agent-connection-telemetry`" + ` and ` + "`rate-limiting`" + `) declares
		` + "`assetTypes: ['llm']`" + ` and Anypoint rejects the POST otherwise.

		For surgical control of any single policy use the underlying ` + "`anypoint_apim_policy_*`" + ` resources directly —
		the composite leaves them alone.
		`,
		Schema: aiGatewaySchema(),
	}
}

// resourceAiGatewayCustomizeDiff validates block↔type invariants, the
// rate-limit mutex, and pre-computes the `applied_policies` list so `plan`
// shows the resolved stack before apply.
func resourceAiGatewayCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if _, llm := d.GetOk("llm_rate_limit"); llm {
		if _, req := d.GetOk("request_rate_limit"); req {
			return fmt.Errorf("anypoint_ai_gateway: `llm_rate_limit` and `request_rate_limit` are mutually exclusive — declare at most one")
		}
	}

	if err := validateAiGatewayProviderShape(d); err != nil {
		return err
	}
	if err := validateAiGatewayGuardShape(d); err != nil {
		return err
	}
	if err := validateAiGatewayRoutingShape(d); err != nil {
		return err
	}

	inbound, outbound := resolvePolicies(diffReadAdapter{d})
	planned := flattenResolvedAppliedPolicies(inbound, outbound)
	sortAppliedPoliciesForState(planned)

	// If nothing changed in the resolved stack, leave `applied_policies` as
	// computed-known-after-apply so we don't churn the plan.
	if d.HasChange("llm_provider") || d.HasChange("prompt_guard") || d.HasChange("routing") ||
		d.HasChange("llm_rate_limit") || d.HasChange("request_rate_limit") || d.HasChange("telemetry") ||
		d.HasChange("asset_versions") || d.Id() == "" {
		if err := d.SetNewComputed("applied_policies"); err != nil {
			return err
		}
		_ = planned // suppress unused — full plan render lands when Terraform refreshes computed
	}
	return nil
}

// diffReadAdapter adapts schema.ResourceDiff to aiGatewayResourceLike so the
// resolver can be shared with the CRUD paths.
type diffReadAdapter struct{ d *schema.ResourceDiff }

func (a diffReadAdapter) Get(key string) any { return a.d.Get(key) }
func (a diffReadAdapter) GetOk(key string) (any, bool) {
	v := a.d.Get(key)
	switch x := v.(type) {
	case nil:
		return v, false
	case string:
		return v, x != ""
	case []any:
		return v, len(x) > 0
	case map[string]any:
		return v, len(x) > 0
	case bool:
		return v, x
	case int:
		return v, x != 0
	}
	return v, v != nil
}

func validateAiGatewayProviderShape(d *schema.ResourceDiff) error {
	list, ok := d.Get("llm_provider").([]any)
	if !ok || len(list) == 0 {
		return nil // Required will catch absent block at plan
	}
	cfg, _ := list[0].(map[string]any)
	if cfg == nil {
		return nil
	}
	t, _ := cfg["type"].(string)
	switch t {
	case "bedrock", "bedrock-anthropic":
		if firstSubBlock(cfg, "bedrock") == nil {
			return fmt.Errorf("anypoint_ai_gateway: llm_provider.type = %q requires `bedrock { ... }` sub-block", t)
		}
	}
	return nil
}

func validateAiGatewayGuardShape(d *schema.ResourceDiff) error {
	list, ok := d.Get("prompt_guard").([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	cfg, _ := list[0].(map[string]any)
	if cfg == nil {
		return nil
	}
	t, _ := cfg["type"].(string)
	switch t {
	case "semantic-openai":
		if firstSubBlock(cfg, "semantic_openai") == nil {
			return fmt.Errorf("anypoint_ai_gateway: prompt_guard.type = %q requires `semantic_openai { ... }` sub-block", t)
		}
	}
	return nil
}

func validateAiGatewayRoutingShape(d *schema.ResourceDiff) error {
	list, ok := d.Get("routing").([]any)
	if !ok || len(list) == 0 {
		return nil
	}
	cfg, _ := list[0].(map[string]any)
	if cfg == nil {
		return nil
	}
	t, _ := cfg["type"].(string)
	switch t {
	case "model-based":
		if firstSubBlock(cfg, "model_based") == nil {
			return fmt.Errorf("anypoint_ai_gateway: routing.type = %q requires `model_based { ... }` sub-block", t)
		}
	}
	return nil
}

// resourceAiGatewayCreate applies the composite. Strategy:
//  1. Pre-check the api_instance has no composite-owned policies (single-composite-per-apim rule).
//  2. Resolve inbound + outbound policy lists.
//  3. POST inbound policies in declared order; on mid-stream failure, delete partials.
//  4. POST outbound policies (transcoding before provider).
//  5. SetId composite id.
//  6. Read back to populate computed fields.
func resourceAiGatewayCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgID := d.Get("org_id").(string)
	envID := d.Get("env_id").(string)
	apimID := d.Get("api_instance_id").(string)

	if pre := aiGatewayPreCheckSingleComposite(ctx, &pco, orgID, envID, apimID); pre.HasError() {
		return pre
	}

	upstreamID, upDiags := aiGatewayDiscoverUpstream(ctx, &pco, orgID, envID, apimID)
	if upDiags.HasError() {
		return upDiags
	}

	apimIDInt, err := strconv.Atoi(apimID)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid api_instance_id for ai_gateway",
			Detail:   "api_instance_id must be the numeric api manager instance id, got: " + apimID,
		})
	}

	inbound, outbound := resolvePolicies(d)
	created := make([]createdPolicyRef, 0, len(inbound)+len(outbound))

	api := pco.apimpolicyclient.DefaultAPI
	authctx := getApimPolicyAuthCtx(ctx, &pco)

	for _, p := range inbound {
		id, cDiags := postInboundPolicy(api, authctx, orgID, envID, apimID, p)
		if cDiags.HasError() {
			rollbackAiGatewayPolicies(api, authctx, orgID, envID, apimID, created)
			return cDiags
		}
		created = append(created, createdPolicyRef{ID: id, AssetID: p.AssetID, InjectionPoint: p.InjectionPoint, Version: p.Version, Order: p.Order})
	}

	for i, p := range outbound {
		id, cDiags := postOutboundPolicy(api, authctx, orgID, envID, apimID, int32(apimIDInt), upstreamID, p)
		if cDiags.HasError() {
			rollbackAiGatewayPolicies(api, authctx, orgID, envID, apimID, created)
			return cDiags
		}
		created = append(created, createdPolicyRef{ID: id, AssetID: p.AssetID, InjectionPoint: p.InjectionPoint, Version: p.Version, Order: i + 1})
	}

	d.SetId(fmt.Sprintf("%s/%s/%s/ai_gateway", orgID, envID, apimID))
	if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
		diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set last_updated", Detail: err.Error()})
	}
	diags = append(diags, resourceAiGatewayRead(ctx, d, m)...)
	return diags
}

func resourceAiGatewayRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	orgID, envID, apimID, idDiags := decomposeAiGatewayID(d)
	if idDiags.HasError() {
		return idDiags
	}

	authctx := getApimPolicyAuthCtx(ctx, &pco)
	api := pco.apimpolicyclient.DefaultAPI

	res, httpr, err := api.GetApimPolicies(authctx, orgID, envID, apimID).FullInfo(false).Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to list policies for api " + apimID,
			Detail:   extractAPIErrorDetail(err, httpr),
		})
	}
	defer httpr.Body.Close()

	external := false
	owned := make([]apim_policy.ApimPolicy, 0)
	if res != nil && res.ArrayOfApimPolicy != nil {
		for _, p := range *res.ArrayOfApimPolicy {
			assetID, _ := p.GetAssetIdOk()
			if assetID != nil && aiGatewayCatalogContains(*assetID) {
				owned = append(owned, p)
			} else if assetID != nil {
				external = true
			}
		}
	}

	if len(owned) == 0 {
		// Composite was created against an api_instance that no longer carries
		// any catalog policy. Treat as gone — the user can re-apply or remove.
		d.SetId("")
		return nil
	}

	applied := make([]any, 0, len(owned))
	for _, p := range owned {
		entry := map[string]any{
			"asset_id":        p.GetAssetId(),
			"id":              strconv.Itoa(int(p.GetId())),
			"order":           int(p.GetOrder()),
			"injection_point": aiGatewayCatalogInjection(p.GetAssetId()),
			"version":         p.GetAssetVersion(),
		}
		applied = append(applied, entry)
	}
	sortAppliedPoliciesForState(applied)

	if err := d.Set("applied_policies", applied); err != nil {
		diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set applied_policies", Detail: err.Error()})
	}
	if err := d.Set("external_policies_detected", external); err != nil {
		diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set external_policies_detected", Detail: err.Error()})
	}
	if external {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "External policies detected on api instance " + apimID,
			Detail:   "Policies the ai_gateway composite does not recognise are present on the api_instance. They will be left untouched by composite Update/Delete. Manage them via the corresponding `anypoint_apim_policy_*` resources.",
		})
	}

	d.Set("org_id", orgID)
	d.Set("env_id", envID)
	d.Set("api_instance_id", apimID)
	return diags
}

func resourceAiGatewayUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	if !d.HasChanges("llm_provider", "prompt_guard", "routing", "llm_rate_limit", "request_rate_limit", "telemetry", "asset_versions") {
		// last_updated only — refresh state and exit.
		if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
			diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set last_updated", Detail: err.Error()})
		}
		return append(diags, resourceAiGatewayRead(ctx, d, m)...)
	}

	orgID, envID, apimID, idDiags := decomposeAiGatewayID(d)
	if idDiags.HasError() {
		return idDiags
	}

	authctx := getApimPolicyAuthCtx(ctx, &pco)
	api := pco.apimpolicyclient.DefaultAPI

	// List current owned policies so we can diff against desired.
	current, currentDiags := aiGatewayListOwnedPolicies(ctx, &pco, orgID, envID, apimID)
	if currentDiags.HasError() {
		return currentDiags
	}

	desiredInbound, desiredOutbound := resolvePolicies(d)
	desired := make(map[string]resolvedPolicy, len(desiredInbound)+len(desiredOutbound))
	for _, p := range desiredInbound {
		desired[p.AssetID] = p
	}
	for _, p := range desiredOutbound {
		desired[p.AssetID] = p
	}

	have := make(map[string]apim_policy.ApimPolicy, len(current))
	for _, p := range current {
		have[p.GetAssetId()] = p
	}

	apimIDInt, err := strconv.Atoi(apimID)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid api_instance_id for ai_gateway update",
			Detail:   apimID,
		})
	}
	upstreamID, upDiags := aiGatewayDiscoverUpstream(ctx, &pco, orgID, envID, apimID)
	if upDiags.HasError() {
		return upDiags
	}

	// Apply additions first (avoid no-provider window during apply).
	for assetID, p := range desired {
		if _, found := have[assetID]; found {
			continue
		}
		if p.InjectionPoint == AiGatewayInjectionInbound {
			if _, cDiags := postInboundPolicy(api, authctx, orgID, envID, apimID, p); cDiags.HasError() {
				return cDiags
			}
		} else {
			if _, cDiags := postOutboundPolicy(api, authctx, orgID, envID, apimID, int32(apimIDInt), upstreamID, p); cDiags.HasError() {
				return cDiags
			}
		}
	}

	// PATCH changed (config / version / order).
	for assetID, p := range desired {
		existing, found := have[assetID]
		if !found {
			continue
		}
		needsPatch := existing.GetAssetVersion() != p.Version || int(existing.GetOrder()) != p.Order
		// We always PATCH config since we can't cheaply diff the masked sensitive fields server-side.
		if !needsPatch {
			needsPatch = true
		}
		patchBody := buildPolicyPatchBody(p)
		policyID := strconv.Itoa(int(existing.GetId()))
		_, httpr, perr := api.PatchApimPolicy(authctx, orgID, envID, apimID, policyID).Body(patchBody).Execute()
		if perr != nil {
			return append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update policy " + assetID + " on api " + apimID,
				Detail:   extractAPIErrorDetail(perr, httpr),
			})
		}
		if httpr != nil {
			httpr.Body.Close()
		}
	}

	// Delete removed.
	for assetID, existing := range have {
		if _, keep := desired[assetID]; keep {
			continue
		}
		policyID := strconv.Itoa(int(existing.GetId()))
		httpr, derr := api.DeleteApimPolicy(authctx, orgID, envID, apimID, policyID).Execute()
		if derr != nil {
			return append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to remove policy " + assetID + " from api " + apimID,
				Detail:   extractAPIErrorDetail(derr, httpr),
			})
		}
		if httpr != nil {
			httpr.Body.Close()
		}
	}

	if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
		diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set last_updated", Detail: err.Error()})
	}
	return append(diags, resourceAiGatewayRead(ctx, d, m)...)
}

func resourceAiGatewayDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	orgID, envID, apimID, idDiags := decomposeAiGatewayID(d)
	if idDiags.HasError() {
		return idDiags
	}

	owned, ownedDiags := aiGatewayListOwnedPolicies(ctx, &pco, orgID, envID, apimID)
	if ownedDiags.HasError() {
		return ownedDiags
	}

	authctx := getApimPolicyAuthCtx(ctx, &pco)
	api := pco.apimpolicyclient.DefaultAPI

	for _, p := range owned {
		policyID := strconv.Itoa(int(p.GetId()))
		httpr, derr := api.DeleteApimPolicy(authctx, orgID, envID, apimID, policyID).Execute()
		if derr != nil {
			// Best-effort: log + continue so a stuck policy doesn't strand the
			// rest of the stack.
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Unable to delete composite-owned policy " + p.GetAssetId() + " on api " + apimID,
				Detail:   extractAPIErrorDetail(derr, httpr),
			})
		}
		if httpr != nil {
			httpr.Body.Close()
		}
	}

	d.SetId("")
	return diags
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

type createdPolicyRef struct {
	ID             string
	AssetID        string
	InjectionPoint string
	Version        string
	Order          int
}

func aiGatewayCatalogInjection(assetID string) string {
	if e, ok := aiGatewayPolicyCatalog[assetID]; ok {
		return e.InjectionPoint
	}
	return ""
}

func decomposeAiGatewayID(d *schema.ResourceData) (orgID, envID, apimID string, diags diag.Diagnostics) {
	id := d.Id()
	if id == "" {
		// Fall back to declared fields (Create has not yet set the composite id when CustomizeDiff fires).
		return d.Get("org_id").(string), d.Get("env_id").(string), d.Get("api_instance_id").(string), nil
	}
	parts := strings.Split(id, "/")
	if len(parts) != 4 || parts[3] != "ai_gateway" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid ai_gateway composite id",
			Detail:   "Expected {org_id}/{env_id}/{api_instance_id}/ai_gateway, got: " + id,
		})
		return "", "", "", diags
	}
	return parts[0], parts[1], parts[2], nil
}

func aiGatewayPreCheckSingleComposite(ctx context.Context, pco *ProviderConfOutput, orgID, envID, apimID string) diag.Diagnostics {
	var diags diag.Diagnostics
	authctx := getApimPolicyAuthCtx(ctx, pco)
	res, httpr, err := pco.apimpolicyclient.DefaultAPI.GetApimPolicies(authctx, orgID, envID, apimID).FullInfo(false).Execute()
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to list policies for api " + apimID + " (pre-check)",
			Detail:   extractAPIErrorDetail(err, httpr),
		})
	}
	defer httpr.Body.Close()
	if res == nil || res.ArrayOfApimPolicy == nil {
		return nil
	}
	for _, p := range *res.ArrayOfApimPolicy {
		assetID, _ := p.GetAssetIdOk()
		if assetID != nil && aiGatewayCatalogContains(*assetID) {
			return append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "ai_gateway composite cannot adopt api instance " + apimID,
				Detail:   "Catalog policy " + *assetID + " is already attached. Remove it (or another `anypoint_ai_gateway` managing this instance) before creating this resource.",
			})
		}
	}
	return nil
}

func aiGatewayListOwnedPolicies(ctx context.Context, pco *ProviderConfOutput, orgID, envID, apimID string) ([]apim_policy.ApimPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	authctx := getApimPolicyAuthCtx(ctx, pco)
	res, httpr, err := pco.apimpolicyclient.DefaultAPI.GetApimPolicies(authctx, orgID, envID, apimID).FullInfo(false).Execute()
	if err != nil {
		return nil, append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to list policies for api " + apimID,
			Detail:   extractAPIErrorDetail(err, httpr),
		})
	}
	defer httpr.Body.Close()
	out := make([]apim_policy.ApimPolicy, 0)
	if res != nil && res.ArrayOfApimPolicy != nil {
		for _, p := range *res.ArrayOfApimPolicy {
			if assetID, _ := p.GetAssetIdOk(); assetID != nil && aiGatewayCatalogContains(*assetID) {
				out = append(out, p)
			}
		}
	}
	return out, nil
}

// aiGatewayDiscoverUpstream returns the first upstream id on the api_instance.
// Outbound policies must be bound to an upstream id per the Anypoint API
// contract; LLM gateways typically run on a single-upstream api_instance, so
// the first one is the right choice.
func aiGatewayDiscoverUpstream(ctx context.Context, pco *ProviderConfOutput, orgID, envID, apimID string) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	authctx := getApimUpstreamAuthCtx(ctx, pco)
	res, httpr, err := pco.apimupstreamclient.DefaultApi.GetApimInstanceUpstreams(authctx, orgID, envID, apimID).Execute()
	if err != nil {
		return "", append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to list upstreams for api " + apimID,
			Detail:   extractAPIErrorDetail(err, httpr),
		})
	}
	defer httpr.Body.Close()
	for _, u := range res.GetUpstreams() {
		if id, ok := u.GetIdOk(); ok && id != nil && *id != "" {
			return *id, nil
		}
	}
	return "", append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "No upstream on api " + apimID,
		Detail:   "ai_gateway requires the api_instance to have at least one upstream so outbound policies can be bound. Create one via the Anypoint UI or `anypoint_api_instance_upstream` before declaring this resource.",
	})
}

func postInboundPolicy(api *apim_policy.DefaultAPIService, authctx context.Context, orgID, envID, apimID string, p resolvedPolicy) (string, diag.Diagnostics) {
	body := apim_policy.NewApimPolicyBody()
	body.SetGroupId(AiGatewayPolicyGroupID)
	body.SetAssetId(p.AssetID)
	body.SetAssetVersion(p.Version)
	body.SetConfigurationData(p.Config)
	if p.Order > 0 {
		body.SetOrder(int32(p.Order))
	}
	res, httpr, err := api.PostApimPolicy(authctx, orgID, envID, apimID).ApimPolicyBody(*body).Execute()
	if err != nil {
		return "", diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "Unable to apply ai_gateway policy " + p.AssetID + " on api " + apimID,
			Detail:   extractAPIErrorDetail(err, httpr),
		}}
	}
	if httpr != nil {
		httpr.Body.Close()
	}
	return strconv.Itoa(int(res.GetId())), nil
}

func postOutboundPolicy(api *apim_policy.DefaultAPIService, authctx context.Context, orgID, envID, apimID string, apiVersionID int32, upstreamID string, p resolvedPolicy) (string, diag.Diagnostics) {
	body := apim_policy.NewApimOutboundPolicyBodyWithDefaults()
	body.SetGroupId(AiGatewayPolicyGroupID)
	body.SetAssetId(p.AssetID)
	body.SetAssetVersion(p.Version)
	body.SetConfigurationData(p.Config)
	body.SetApiVersionId(apiVersionID)
	body.SetUpstreamIds([]string{upstreamID})
	res, httpr, err := api.PostApimOutboundPolicy(authctx, orgID, envID, apimID).ApimOutboundPolicyBody(*body).Execute()
	if err != nil {
		return "", diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "Unable to apply ai_gateway outbound policy " + p.AssetID + " on api " + apimID,
			Detail:   extractAPIErrorDetail(err, httpr),
		}}
	}
	if httpr != nil {
		httpr.Body.Close()
	}
	if len(res) == 0 {
		return "", diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "Empty response posting outbound policy " + p.AssetID + " on api " + apimID,
			Detail:   "API returned no policy id — expected one per upstream.",
		}}
	}
	return strconv.Itoa(int(res[0].GetId())), nil
}

func buildPolicyPatchBody(p resolvedPolicy) map[string]any {
	body := map[string]any{
		"configurationData": p.Config,
		"groupId":           AiGatewayPolicyGroupID,
		"assetId":           p.AssetID,
		"assetVersion":      p.Version,
	}
	if p.Order > 0 {
		body["order"] = int32(p.Order)
	}
	return body
}

// rollbackAiGatewayPolicies makes a best effort to delete partial policies on
// Create failure so the api_instance is not left half-attached.
func rollbackAiGatewayPolicies(api *apim_policy.DefaultAPIService, authctx context.Context, orgID, envID, apimID string, created []createdPolicyRef) {
	// Reverse order — outbound first, then inbound by latest first.
	sort.SliceStable(created, func(i, j int) bool { return i > j })
	for _, c := range created {
		httpr, _ := api.DeleteApimPolicy(authctx, orgID, envID, apimID, c.ID).Execute()
		if httpr != nil {
			httpr.Body.Close()
		}
	}
}
