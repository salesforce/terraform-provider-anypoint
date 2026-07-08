package anypoint

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// policyTemplateEntry is what we cache per (group, asset, version) — the parsed
// configuration JSON Schema (nil when the template carries no schema).
type policyTemplateEntry struct {
	Schema *jsonschema.Schema
}

var policyTemplateCache sync.Map // key: "{group}/{asset}/{version}" → *policyTemplateEntry

func policyTemplateCacheKey(group, asset, version string) string {
	return group + "/" + asset + "/" + version
}

// loadPolicyTemplateCached fetches a single policy template via xapi/v1 and
// returns the cached entry. Cache lives for the lifetime of the provider
// process. Returns an entry with a nil Schema when the template carries no
// JSON Schema (legacy mule3 / older mule4 policies expose a property list
// instead of a JSON Schema object).
func loadPolicyTemplateCached(
	ctx context.Context,
	pco *ProviderConfOutput,
	orgId, assetGroupId, assetId, assetVersion string,
) (*policyTemplateEntry, diag.Diagnostics) {
	key := policyTemplateCacheKey(assetGroupId, assetId, assetVersion)
	if v, ok := policyTemplateCache.Load(key); ok {
		return v.(*policyTemplateEntry), nil
	}

	authctx := getApimPolicyAuthCtx(ctx, pco)
	res, httpr, err := pco.apimpolicyclient.DefaultAPI.
		GetOrgExchangePolicyTemplateDetails(authctx, orgId, assetGroupId, assetId, assetVersion).
		SplitModel(true).
		Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		return nil, diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to fetch policy template %s/%s/%s for plan-time validation", assetGroupId, assetId, assetVersion),
			Detail:   details,
		}}
	}
	if httpr != nil && httpr.Body != nil {
		defer httpr.Body.Close()
	}

	entry := &policyTemplateEntry{}
	if cfg := res.GetConfiguration(); cfg != nil {
		// Configuration is `interface{}` — could be a JSON Schema object (modern
		// split-asset policies) or an array of PolicyConfiguration items (legacy
		// policies). We only compile when it is an object.
		if _, ok := cfg.(map[string]any); ok {
			b, mErr := json.Marshal(cfg)
			if mErr != nil {
				return nil, diag.Diagnostics{{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Unable to marshal policy template %s/%s/%s configuration schema", assetGroupId, assetId, assetVersion),
					Detail:   mErr.Error(),
				}}
			}
			compiler := jsonschema.NewCompiler()
			compiler.Draft = jsonschema.Draft2019
			if cErr := compiler.AddResource("config.json", strings.NewReader(string(b))); cErr != nil {
				return nil, diag.Diagnostics{{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Unable to load policy template %s/%s/%s configuration schema", assetGroupId, assetId, assetVersion),
					Detail:   cErr.Error(),
				}}
			}
			schema, cErr := compiler.Compile("config.json")
			if cErr != nil {
				return nil, diag.Diagnostics{{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Unable to compile policy template %s/%s/%s configuration schema", assetGroupId, assetId, assetVersion),
					Detail:   cErr.Error(),
				}}
			}
			entry.Schema = schema
		}
	}

	policyTemplateCache.Store(key, entry)
	return entry, nil
}

// validateConfigAgainstSchema parses the configuration_data JSON string and
// validates it against the compiled schema. Returns a diag.Diagnostics with
// one entry per validation failure when invalid.
func validateConfigAgainstSchema(configJSON string, schema *jsonschema.Schema, assetCoord string) diag.Diagnostics {
	if schema == nil {
		// Legacy policy template with no schema — nothing to validate at plan time.
		return nil
	}
	if strings.TrimSpace(configJSON) == "" {
		return nil
	}
	var parsed any
	if err := json.Unmarshal([]byte(configJSON), &parsed); err != nil {
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Invalid configuration_data for %s: not valid JSON", assetCoord),
			Detail:   err.Error(),
		}}
	}
	if err := schema.Validate(parsed); err != nil {
		return flattenJSONSchemaError(err, assetCoord)
	}
	return nil
}

// flattenJSONSchemaError walks a santhosh-tekuri ValidationError tree and
// emits one diag.Error per leaf so users see a per-field message instead of
// a single dense wrapped error.
func flattenJSONSchemaError(err error, assetCoord string) diag.Diagnostics {
	verr, ok := err.(*jsonschema.ValidationError)
	if !ok {
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "configuration_data failed schema validation for " + assetCoord,
			Detail:   err.Error(),
		}}
	}
	var diags diag.Diagnostics
	for _, leaf := range collectLeafErrors(verr) {
		path := leaf.InstanceLocation
		if path == "" {
			path = "/"
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("configuration_data invalid for %s at %s", assetCoord, path),
			Detail:   leaf.Message,
		})
	}
	if len(diags) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "configuration_data failed schema validation for " + assetCoord,
			Detail:   verr.Error(),
		})
	}
	return diags
}

func collectLeafErrors(v *jsonschema.ValidationError) []*jsonschema.ValidationError {
	if v == nil {
		return nil
	}
	if len(v.Causes) == 0 {
		return []*jsonschema.ValidationError{v}
	}
	var out []*jsonschema.ValidationError
	for _, c := range v.Causes {
		out = append(out, collectLeafErrors(c)...)
	}
	return out
}

// diagsToError flattens a diag.Diagnostics into a plain error for use inside
// CustomizeDiff (which only accepts error). Concatenates summary + detail per
// diag with "; " separator. Drops non-Error severities.
func diagsToError(diags diag.Diagnostics) error {
	if len(diags) == 0 {
		return nil
	}
	var parts []string
	for _, d := range diags {
		if d.Severity != diag.Error {
			continue
		}
		if d.Detail != "" {
			parts = append(parts, fmt.Sprintf("%s: %s", d.Summary, d.Detail))
		} else {
			parts = append(parts, d.Summary)
		}
	}
	if len(parts) == 0 {
		return nil
	}
	return fmt.Errorf("%s", strings.Join(parts, "; "))
}
