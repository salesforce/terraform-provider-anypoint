package anypoint

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mulesoft-anypoint/anypoint-client-go/apim_policy"
)

func resourceApimInstancePolicyCustom() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApimInstancePolicyCustomCreate,
		ReadContext:   resourceApimInstancePolicyCustomRead,
		UpdateContext: resourceApimInstancePolicyCustomUpdate,
		DeleteContext: resourceApimInstancePolicyCustomDelete,
		CustomizeDiff: resourceApimInstancePolicyCustomCustomizeDiff,
		Description: `
		Create and manage an API Policy of any type. The supplied ` + "`configuration_data`" + ` is
		validated against the policy template's published JSON Schema at plan time when
		the template exposes one, so configuration errors surface before ` + "`terraform apply`" + `.
		` + "`asset_version`" + ` changes are applied in place — no resource replacement.
		`,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The last time this resource has been updated locally.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy's unique id",
			},
			"apim_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The api manager instance id where the api instance is defined.",
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
				Description: "The environment id where api instance is defined.",
			},
			"audit": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The instance's auditing data",
			},
			"master_organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization id where the api instance is defined.",
			},
			"configuration_data": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The policy configuration data in json format",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsJSON),
			},
			"policy_template_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy template id",
			},
			"order": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The policy execution order. Lower values execute earlier. Leave unset to let Anypoint append at the end of the stack. Updating this value reorders the policy in place via PATCH. For outbound policies the value is applied via a follow-up PATCH after Create — the outbound POST endpoint does not accept `order` in its body.",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the policy is disabled.",
			},
			"pointcut_data": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The method & resource conditions",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method_regex": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "The list of HTTP methods",
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateDiagFunc: validation.ToDiagFunc(
									validation.StringInSlice(
										[]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD", "TRACE"},
										false,
									),
								),
							},
						},
						"uri_template_regex": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "URI template regex",
						},
					},
				},
			},
			"asset_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The policy template group id in anypoint exchange. Don't change unless mulesoft has renamed the policy group id.",
			},
			"asset_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The policy template id in anypoint exchange. Don't change unless mulesoft has renamed the policy asset id.",
			},
			"asset_version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The policy template version in Anypoint Exchange. Changing this value upgrades the policy in place via PATCH — no resource replacement. Validated against the new template's published JSON Schema at plan time.",
			},
			"injection_point": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "inbound",
				ValidateFunc: validation.StringInSlice([]string{"inbound", "outbound"}, false),
				Description:  "Where the policy is applied. 'inbound' (default) creates the policy via `POST .../policies`; 'outbound' creates it via `POST .../xapi/v1/.../policies/outbound-policies` and requires `upstream_id`. Required for credential-injection and LLM-provider policies that are outbound-only.",
			},
			"upstream_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Identifier of the upstream this outbound policy is bound to. Required when `injection_point = \"outbound\"` and must be left empty for inbound policies. Reference an `anypoint_api_instance_upstream.<x>.id`.",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// isOutboundPolicy reports whether the resource is configured to use the outbound endpoint family.
func isOutboundPolicy(d *schema.ResourceData) bool {
	return d.Get("injection_point").(string) == "outbound"
}

// resourceApimInstancePolicyCustomCustomizeDiff runs plan-time validation of
// `configuration_data` against the policy template's published JSON Schema and
// of the `injection_point` / `upstream_id` invariants. Legacy templates with
// no JSON Schema (mule3 / older mule4) are skipped — there is nothing to
// validate against.
func resourceApimInstancePolicyCustomCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, m any) error {
	if m == nil {
		return nil
	}
	pco, ok := m.(ProviderConfOutput)
	if !ok {
		return nil
	}

	injectionPoint, _ := d.Get("injection_point").(string)
	upstreamId, _ := d.Get("upstream_id").(string)
	if injectionPoint == "outbound" && upstreamId == "" && d.NewValueKnown("upstream_id") {
		return fmt.Errorf("upstream_id is required when injection_point = \"outbound\"")
	}
	if injectionPoint == "inbound" && upstreamId != "" {
		return fmt.Errorf("upstream_id must be empty when injection_point = \"inbound\"")
	}

	orgId, _ := d.Get("org_id").(string)
	assetGroupId, _ := d.Get("asset_group_id").(string)
	assetId, _ := d.Get("asset_id").(string)
	assetVersion, _ := d.Get("asset_version").(string)
	configJSON, _ := d.Get("configuration_data").(string)
	if orgId == "" || assetGroupId == "" || assetId == "" || assetVersion == "" {
		return nil
	}
	if !d.NewValueKnown("configuration_data") || !d.NewValueKnown("asset_version") {
		return nil
	}
	entry, diags := loadPolicyTemplateCached(ctx, &pco, orgId, assetGroupId, assetId, assetVersion)
	if diags.HasError() {
		return diagsToError(diags)
	}
	if entry == nil || entry.Schema == nil {
		return nil
	}
	assetCoord := assetGroupId + "/" + assetId + "/" + assetVersion
	if vdiags := validateConfigAgainstSchema(configJSON, entry.Schema, assetCoord); vdiags.HasError() {
		return diagsToError(vdiags)
	}
	return nil
}

func resourceApimInstancePolicyCustomCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	api := pco.apimpolicyclient.DefaultAPI
	var id int32
	var httpr *http.Response
	if isOutboundPolicy(d) {
		upstreamId := d.Get("upstream_id").(string)
		if upstreamId == "" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Missing upstream_id for outbound policy on api " + apimid,
				Detail:   "When injection_point = \"outbound\", upstream_id is required and must reference an existing api instance upstream id.",
			})
			return diags
		}
		apimIdInt, err := strconv.Atoi(apimid)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid apim_id for outbound policy",
				Detail:   "apim_id must be the numeric api instance id, got: " + apimid,
			})
			return diags
		}
		body, err := newApimPolicyCustomOutboundBody(d, int32(apimIdInt), upstreamId)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to parse policy configuration for api " + apimid,
				Detail:   err.Error(),
			})
			return diags
		}
		res, hr, err := api.PostApimOutboundPolicy(authctx, orgid, envid, apimid).ApimOutboundPolicyBody(*body).Execute()
		httpr = hr
		if err != nil {
			details := extractAPIErrorDetail(err, httpr)
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create custom outbound policy for api " + apimid,
				Detail:   details,
			})
			return diags
		}
		if len(res) == 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Empty response creating outbound policy for api " + apimid,
				Detail:   "API returned an empty policy array — expected at least one record per upstream id.",
			})
			return diags
		}
		id = res[0].GetId()
	} else {
		body, err := newApimPolicyCustomBody(d)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to parse policy configuration for api " + apimid,
				Detail:   err.Error(),
			})
			return diags
		}
		res, hr, err := api.PostApimPolicy(authctx, orgid, envid, apimid).ApimPolicyBody(*body).Execute()
		httpr = hr
		if err != nil {
			details := extractAPIErrorDetail(err, httpr)
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create custom policy for api " + apimid,
				Detail:   details,
			})
			return diags
		}
		id = res.GetId()
	}
	defer httpr.Body.Close()
	d.SetId(strconv.Itoa(int(id)))
	// Outbound POST body cannot carry `order` — patch it in afterwards when the user
	// declared one, so a single `apply` lands the resource at the desired position.
	if isOutboundPolicy(d) {
		if val, ok := d.GetOk("order"); ok {
			patchBody := map[string]any{"order": int32(val.(int))}
			_, phttpr, perr := api.PatchApimPolicy(authctx, orgid, envid, apimid, strconv.Itoa(int(id))).Body(patchBody).Execute()
			if perr != nil {
				details := extractAPIErrorDetail(perr, phttpr)
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to apply order to outbound custom policy " + strconv.Itoa(int(id)) + " for api " + apimid,
					Detail:   details,
				})
				return diags
			}
			if phttpr != nil {
				phttpr.Body.Close()
			}
		}
	}
	diags = append(diags, resourceApimInstancePolicyCustomRead(ctx, d, m)...)
	//in case disabled
	disabled := d.Get("disabled").(bool)
	if disabled {
		diags = append(diags, disableApimInstancePolicyCustom(ctx, d, m)...)
		diags = append(diags, resourceApimInstancePolicyCustomRead(ctx, d, m)...)
	}

	return diags
}

func resourceApimInstancePolicyCustomRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	injection := d.Get("injection_point").(string)
	if isComposedResourceId(id) {
		orgid, envid, apimid, injection, id, diags = decomposeApimPolicyCustomId(d)
	}
	if diags.HasError() {
		return diags
	}
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	//perform request — both inbound and outbound policies live on the singular inbound GET endpoint.
	api := pco.apimpolicyclient.DefaultAPI
	res, httpr, err := api.GetApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	_ = injection
	if err != nil {
		if httpr != nil && httpr.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read custom policy " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// process data
	data := flattenApimInstancePolicy(res)
	if cfg, err := flattenApimPolicyCustomCfg(d, res); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse configuration data of custom policy " + id + " for api " + apimid,
			Detail:   err.Error(),
		})
		return diags
	} else {
		data["configuration_data"] = cfg
	}
	if err := setApimInstancePolicyAttributesToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set api custom policy " + id + " details attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(id)
	d.Set("apim_id", apimid)
	d.Set("env_id", envid)
	d.Set("org_id", orgid)
	d.Set("injection_point", injection)
	return diags
}

func resourceApimInstancePolicyCustomUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	// Capture the planned `disabled` value up front — the mid-Update Read below
	// rewrites `disabled` in state with the API-returned value, which would flip
	// the branch in the toggle block at the bottom of this function.
	plannedDisabled := d.Get("disabled").(bool)
	//detect change
	if d.HasChanges("configuration_data", "pointcut_data", "asset_version", "order") {
		pco := m.(ProviderConfOutput)
		orgid := d.Get("org_id").(string)
		envid := d.Get("env_id").(string)
		apimid := d.Get("apim_id").(string)
		id := d.Get("id").(string)
		authctx := getApimPolicyAuthCtx(ctx, &pco)
		//prepare body
		body, err := newApimPolicyCustomPatchBody(d)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to parse policy configuration for api " + apimid,
				Detail:   err.Error(),
			})
			return diags
		}
		//perform request — outbound policies share the singular inbound PATCH endpoint.
		api := pco.apimpolicyclient.DefaultAPI
		_, httpr, err := api.PatchApimPolicy(authctx, orgid, envid, apimid, id).Body(body).Execute()
		if err != nil {
			details := extractAPIErrorDetail(err, httpr)
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update custom policy for api " + apimid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		diags = append(diags, resourceApimInstancePolicyCustomRead(ctx, d, m)...)
	}
	if d.HasChange("disabled") {
		if plannedDisabled {
			diags = append(diags, disableApimInstancePolicyCustom(ctx, d, m)...)
		} else {
			diags = append(diags, enableApimInstancePolicyCustom(ctx, d, m)...)
		}
		diags = append(diags, resourceApimInstancePolicyCustomRead(ctx, d, m)...)
	}

	return diags
}

func resourceApimInstancePolicyCustomDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	api := pco.apimpolicyclient.DefaultAPI
	httpr, err := api.DeleteApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete custom policy " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return diags
}

func enableApimInstancePolicyCustom(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	api := pco.apimpolicyclient.DefaultAPI
	_, httpr, err := api.EnableApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to enable custom policy " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	return diags
}

func disableApimInstancePolicyCustom(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	api := pco.apimpolicyclient.DefaultAPI
	_, httpr, err := api.DisableApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to disable custom policy " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	return diags
}

// flattenApimPolicyCustomCfg merges the API-returned `configurationData` over the
// declared `configuration_data` from prior state. Sensitive fields (e.g. clientId,
// clientSecret) are not returned by the GET endpoint, so falling back to the prior
// declared values prevents perpetual `+` drift on every plan.
func flattenApimPolicyCustomCfg(d *schema.ResourceData, policy *apim_policy.ApimPolicy) (string, error) {
	data := policy.GetConfigurationData()
	dst := make(map[string]any)
	if raw, ok := d.Get("configuration_data").(string); ok && raw != "" {
		if err := json.Unmarshal([]byte(raw), &dst); err != nil {
			return "", fmt.Errorf("configuration_data expected to be a valid JSON Object. %s", err.Error())
		}
	}
	maps.Copy(dst, data)
	b, err := json.Marshal(dst)
	if err != nil {
		return "", fmt.Errorf("unable to marshal merged configuration_data. %s", err.Error())
	}
	return string(b), nil
}

func newApimPolicyCustomOutboundBody(d *schema.ResourceData, apiVersionId int32, upstreamId string) (*apim_policy.ApimOutboundPolicyBody, error) {
	body := apim_policy.NewApimOutboundPolicyBodyWithDefaults()
	if val, ok := d.GetOk("configuration_data"); ok {
		var cfg map[string]any
		err := json.Unmarshal([]byte(val.(string)), &cfg)
		if err != nil {
			return nil, fmt.Errorf("configuration_data expected to be a valid JSON Object. %s", err.Error())
		}
		body.SetConfigurationData(cfg)
	}
	body.SetApiVersionId(apiVersionId)
	if val, ok := d.GetOk("asset_group_id"); ok {
		body.SetGroupId(val.(string))
	}
	if val, ok := d.GetOk("asset_id"); ok {
		body.SetAssetId(val.(string))
	}
	if val, ok := d.GetOk("asset_version"); ok {
		body.SetAssetVersion(val.(string))
	}
	body.SetUpstreamIds([]string{upstreamId})
	return body, nil
}

func newApimPolicyCustomBody(d *schema.ResourceData) (*apim_policy.ApimPolicyBody, error) {
	body := apim_policy.NewApimPolicyBody()
	if val, ok := d.GetOk("configuration_data"); ok {
		var cfg map[string]any
		err := json.Unmarshal([]byte(val.(string)), &cfg)
		if err != nil {
			return nil, fmt.Errorf("configuration_data expected to be a valid JSON Object. %s", err.Error())
		}
		body.SetConfigurationData(cfg)
	}
	if val, ok := d.GetOk("pointcut_data"); ok {
		body.SetPointcutData(newApimPolicyCustomPointcutDataBody(val.([]any)))
	}
	if val, ok := d.GetOk("asset_group_id"); ok {
		body.SetGroupId(val.(string))
	}
	if val, ok := d.GetOk("asset_id"); ok {
		body.SetAssetId(val.(string))
	}
	if val, ok := d.GetOk("asset_version"); ok {
		body.SetAssetVersion(val.(string))
	}
	if val, ok := d.GetOk("order"); ok {
		body.SetOrder(int32(val.(int)))
	}
	return body, nil
}

func newApimPolicyCustomPatchBody(d *schema.ResourceData) (map[string]any, error) {
	body := make(map[string]any)
	if val, ok := d.GetOk("configuration_data"); ok {
		var cfg map[string]any
		err := json.Unmarshal([]byte(val.(string)), &cfg)
		if err != nil {
			return nil, fmt.Errorf("configuration_data expected to be a valid JSON Object. %s", err.Error())
		}
		body["configurationData"] = cfg
	}
	if val, ok := d.GetOk("pointcut_data"); ok {
		collection := newApimPolicyClientIdEnfPointcutDataBody(val.([]any))
		slice := make([]map[string]any, len(collection))
		for i, item := range collection {
			m, _ := item.ToMap()
			slice[i] = m
		}
		body["pointcutData"] = slice
	} else {
		body["pointcutData"] = nil
	}
	if val, ok := d.GetOk("asset_group_id"); ok {
		body["groupId"] = val
	}
	if val, ok := d.GetOk("asset_id"); ok {
		body["assetId"] = val
	}
	if val, ok := d.GetOk("asset_version"); ok {
		body["assetVersion"] = val
	}
	if val, ok := d.GetOk("order"); ok {
		body["order"] = int32(val.(int))
	}
	return body, nil
}

func newApimPolicyCustomPointcutDataBody(collection []any) []apim_policy.PointcutDataItem {
	slice := make([]apim_policy.PointcutDataItem, len(collection))
	for i, item := range collection {
		data := item.(map[string]any)
		body := apim_policy.NewPointcutDataItem()
		if val, ok := data["method_regex"]; ok && val != nil {
			set := val.(*schema.Set)
			body.SetMethodRegex(JoinStringInterfaceSlice(set.List(), "|"))
		}
		if val, ok := data["uri_template_regex"]; ok {
			body.SetUriTemplateRegex(val.(string))
		}
		slice[i] = *body
	}
	return slice
}

// decomposeApimPolicyCustomId accepts two composite-id shapes:
//   - 4-segment: ORG_ID/ENV_ID/APIM_ID/POLICY_ID — legacy, always inbound
//   - 5-segment: ORG_ID/ENV_ID/APIM_ID/INJECTION_POINT/POLICY_ID — required for outbound
//
// Returns (orgId, envId, apimId, injectionPoint, policyId, diags).
func decomposeApimPolicyCustomId(d *schema.ResourceData) (string, string, string, string, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	s := DecomposeResourceId(d.Id())
	switch len(s) {
	case 4:
		return s[0], s[1], s[2], "inbound", s[3], diags
	case 5:
		injection := s[3]
		if injection != "inbound" && injection != "outbound" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid APIM Policy Custom ID format",
				Detail:   fmt.Sprintf("Expected INJECTION_POINT to be 'inbound' or 'outbound', got %q in %s", injection, d.Id()),
			})
			return "", "", "", "", "", diags
		}
		return s[0], s[1], s[2], injection, s[4], diags
	default:
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid APIM Policy Custom ID format",
			Detail:   fmt.Sprintf("Expected ORG_ID/ENV_ID/APIM_ID/POLICY_ID or ORG_ID/ENV_ID/APIM_ID/INJECTION_POINT/POLICY_ID, got %s", d.Id()),
		})
		return "", "", "", "", "", diags
	}
}
