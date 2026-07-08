package anypoint

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mulesoft-anypoint/anypoint-client-go/apim_policy"
)

func resourceApimInstancePolicyMessageLogging() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApimInstancePolicyMessageLoggingCreate,
		ReadContext:   resourceApimInstancePolicyMessageLoggingRead,
		UpdateContext: resourceApimInstancePolicyMessageLoggingUpdate,
		DeleteContext: resourceApimInstancePolicyMessageLoggingDelete,
		Description: `
		Create and manage an API Policy of type message-logging.
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
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "The policy configuration data",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"logging_configuration": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "The list of logging configurations",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The configuration name",
									},
									"message": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "DataWeave Expression for extracting information from the message to log. e.g. #[attributes.headers['id']]",
										ValidateDiagFunc: validation.ToDiagFunc(
											validation.StringMatch(
												regexp.MustCompile(`^\s*#\[.+\]\s*$`),
												"field value should be a dataweave expression",
											),
										),
									},
									"conditional": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "DataWeave Expression to filter which messages to log. e.g. #[attributes.headers['id']==1]",
										ValidateDiagFunc: validation.ToDiagFunc(
											validation.StringMatch(
												regexp.MustCompile(`^\s*#\[.+\]\s*$`),
												"field value should be a dataweave expression",
											),
										),
									},
									"category": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Prefix in the log sentence.",
									},
									"level": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "INFO",
										Description: "Logging level, possible values: INFO, WARN, ERROR or DEBUG",
										ValidateDiagFunc: validation.ToDiagFunc(
											validation.StringInSlice(
												[]string{"INFO", "WARN", "ERROR", "DEBUG"},
												false,
											),
										),
									},
									"first_section": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     true,
										Description: "Log before calling the API",
									},
									"second_section": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Logging after calling the API",
									},
								},
							},
						},
					},
				},
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
				Description: "The policy execution order. Lower values execute earlier. Leave unset to let Anypoint append at the end of the stack. Updating this value reorders the policy in place via PATCH.",
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
				Description: "The Method & resource conditions",
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
				Optional:    true,
				ForceNew:    true,
				Default:     "68ef9520-24e9-4cf2-b2f5-620025690913",
				Description: "The policy template group id in anypoint exchange. Don't change unless mulesoft has renamed the policy group id.",
			},
			"asset_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "message-logging",
				Description: "The policy template id in anypoint exchange. Don't change unless mulesoft has renamed the policy asset id.",
			},
			"asset_version": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "2.0.1",
				Description: "the policy template version in anypoint exchange.",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceApimInstancePolicyMessageLoggingCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	//prepare body
	body := newApimPolicyMessageLoggingBody(d)
	//perform request
	res, httpr, err := pco.apimpolicyclient.DefaultAPI.PostApimPolicy(authctx, orgid, envid, apimid).ApimPolicyBody(*body).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create policy message logging for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	id := res.GetId()
	d.SetId(strconv.Itoa(int(id)))
	diags = append(diags, resourceApimInstancePolicyMessageLoggingRead(ctx, d, m)...)
	//in case disabled
	disabled := d.Get("disabled").(bool)
	if disabled {
		diags = append(diags, disableApimInstancePolicyMessageLogging(ctx, d, m)...)
		diags = append(diags, resourceApimInstancePolicyMessageLoggingRead(ctx, d, m)...)
	}

	return diags
}

func resourceApimInstancePolicyMessageLoggingRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	if isComposedResourceId(id) {
		orgid, envid, apimid, id, diags = decomposeApimPolicyMessageLoggingId(d)
	}
	if diags.HasError() {
		return diags
	}
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.apimpolicyclient.DefaultAPI.GetApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read policy message logging " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// process data
	data := flattenApimInstancePolicy(res)
	data["configuration_data"] = []any{flattenApimPolicyMessageLoggingCfg(d, res)}
	if err := setApimInstancePolicyAttributesToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set api policy message logging details attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(id)
	d.Set("apim_id", apimid)
	d.Set("env_id", envid)
	d.Set("org_id", orgid)
	return diags
}

func resourceApimInstancePolicyMessageLoggingUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	// Capture the planned `disabled` value up front — the mid-Update Read below
	// rewrites `disabled` in state with the API-returned value, which would flip
	// the branch in the toggle block at the bottom of this function.
	plannedDisabled := d.Get("disabled").(bool)
	//detect change
	if d.HasChanges("configuration_data", "pointcut_data", "order") {
		pco := m.(ProviderConfOutput)
		orgid := d.Get("org_id").(string)
		envid := d.Get("env_id").(string)
		apimid := d.Get("apim_id").(string)
		id := d.Get("id").(string)
		authctx := getApimPolicyAuthCtx(ctx, &pco)
		//prepare body
		body := newApimPolicyMessageLoggingPatchBody(d)
		//perform request
		_, httpr, err := pco.apimpolicyclient.DefaultAPI.PatchApimPolicy(authctx, orgid, envid, apimid, id).Body(body).Execute()
		if err != nil {
			details := extractAPIErrorDetail(err, httpr)
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update policy message logging for api " + apimid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		diags = append(diags, resourceApimInstancePolicyMessageLoggingRead(ctx, d, m)...)
	}
	if d.HasChange("disabled") {
		if plannedDisabled {
			diags = append(diags, disableApimInstancePolicyMessageLogging(ctx, d, m)...)
		} else {
			diags = append(diags, enableApimInstancePolicyMessageLogging(ctx, d, m)...)
		}
		diags = append(diags, resourceApimInstancePolicyMessageLoggingRead(ctx, d, m)...)
	}

	return diags
}

func resourceApimInstancePolicyMessageLoggingDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	httpr, err := pco.apimpolicyclient.DefaultAPI.DeleteApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete policy message logging " + id + " for api " + apimid,
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

func enableApimInstancePolicyMessageLogging(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	_, httpr, err := pco.apimpolicyclient.DefaultAPI.EnableApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to enable policy message logging " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	return diags
}

func disableApimInstancePolicyMessageLogging(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	_, httpr, err := pco.apimpolicyclient.DefaultAPI.DisableApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to disable policy message logging " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	return diags
}

func flattenApimPolicyMessageLoggingCfg(_ *schema.ResourceData, policy *apim_policy.ApimPolicy) map[string]any {
	data := make(map[string]any)
	cfg := policy.GetConfigurationData()
	logging_cfg := cfg["loggingConfiguration"].([]any)
	lcfg_result := make([]map[string]any, len(logging_cfg))
	for i, item := range logging_cfg {
		lcfg := item.(map[string]any)
		item_data := lcfg["itemData"].(map[string]any)
		c := make(map[string]any)
		c["name"] = lcfg["itemName"]
		c["message"] = item_data["message"]
		if val, ok := item_data["conditional"]; ok {
			c["conditional"] = val
		}
		if val, ok := item_data["category"]; ok {
			c["category"] = val
		}
		c["level"] = item_data["level"]
		c["first_section"] = item_data["firstSection"]
		c["second_section"] = item_data["secondSection"]
		lcfg_result[i] = c
	}
	data["logging_configuration"] = lcfg_result
	return data
}

func newApimPolicyMessageLoggingBody(d *schema.ResourceData) *apim_policy.ApimPolicyBody {
	body := apim_policy.NewApimPolicyBody()
	if val, ok := d.GetOk("configuration_data"); ok {
		l := val.([]any)
		cfg := l[0].(map[string]any)
		data := newApimPolicyMessageLoggingCfg(cfg)
		body.SetConfigurationData(data)
	}
	if val, ok := d.GetOk("pointcut_data"); ok {
		body.SetPointcutData(newApimPolicyMessageLoggingPointcutDataBody(val.([]any)))
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
	return body
}

func newApimPolicyMessageLoggingPatchBody(d *schema.ResourceData) map[string]any {
	body := make(map[string]any)
	if val, ok := d.GetOk("configuration_data"); ok {
		l := val.([]any)
		cfg := l[0].(map[string]any)
		data := newApimPolicyMessageLoggingCfg(cfg)
		body["configurationData"] = data
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
	return body
}

func newApimPolicyMessageLoggingCfg(input map[string]any) map[string]any {
	body := make(map[string]any)
	logging_cfg := input["logging_configuration"].([]any)
	lcfg := make([]map[string]any, len(logging_cfg))
	for i, item := range logging_cfg {
		cfg := item.(map[string]any)
		//item data configuration
		d := make(map[string]any)
		d["message"] = cfg["message"]
		if val, ok := cfg["conditional"]; ok && val != nil && val.(string) != "" {
			d["conditional"] = val
		}
		if val, ok := cfg["category"]; ok && val != nil && val.(string) != "" {
			d["category"] = val
		}
		d["level"] = cfg["level"]
		d["firstSection"] = cfg["first_section"]
		d["secondSection"] = cfg["second_section"]
		//item
		c := make(map[string]any)
		c["itemName"] = cfg["name"]
		c["itemData"] = d
		lcfg[i] = c
	}
	body["loggingConfiguration"] = lcfg
	return body
}

func newApimPolicyMessageLoggingPointcutDataBody(collection []any) []apim_policy.PointcutDataItem {
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

func decomposeApimPolicyMessageLoggingId(d *schema.ResourceData) (string, string, string, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	s := DecomposeResourceId(d.Id())
	if len(s) != 4 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid APIM Policy Message Logging ID format",
			Detail:   fmt.Sprintf("Expected ORG_ID/ENV_ID/APIM_ID/INSTANCE_ID, got %s", d.Id()),
		})
		return "", "", "", "", diags
	}
	return s[0], s[1], s[2], s[3], diags
}
