package anypoint

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mulesoft-anypoint/anypoint-client-go/apim_policy"
)

func resourceApimInstancePolicyClientIdEnf() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApimInstancePolicyClientIdEnfCreate,
		ReadContext:   resourceApimInstancePolicyClientIdEnfRead,
		UpdateContext: resourceApimInstancePolicyClientIdEnfUpdate,
		DeleteContext: resourceApimInstancePolicyClientIdEnfDelete,
		Description: `
		Create and manage an API Policy of type client-id-enforcement.
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
						"credentials_origin_has_http_basic_authentication_header": {
							Type:     schema.TypeString,
							Required: true,
							Description: `
							Whether to use custom header or to use basic authentication header.
							Values can be either "httpBasicAuthenticationHeader" or "customExpression".
							In the case of using "httpBasicAuthenticationHeader", you don't need to supply client_id_expression or client_secret_expression.
							`,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice(
									[]string{
										"httpBasicAuthenticationHeader",
										"customExpression",
									},
									false,
								),
							),
						},
						"client_id_expression": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The client id header location",
						},
						"client_secret_expression": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The client secret header location",
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
				Computed:    true,
				Description: "The policy order.",
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
							MinItems:    1,
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
				Default:     "client-id-enforcement",
				Description: "The policy template id in anypoint exchange. Don't change unless mulesoft has renamed the policy asset id.",
			},
			"asset_version": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "1.3.2",
				Description: "the policy template version in anypoint exchange.",
			},
		},
		CustomizeDiff: func(ctx context.Context, rd *schema.ResourceDiff, i interface{}) error {
			return validateClientIdEnfCfg(rd)
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceApimInstancePolicyClientIdEnfCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	//prepare body
	body := newApimPolicyClientIdEnfBody(d)
	//perform request
	res, httpr, err := pco.apimpolicyclient.DefaultApi.PostApimPolicy(authctx, orgid, envid, apimid).ApimPolicyBody(*body).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create policy client-id-enforcement for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	id := res.GetId()
	d.SetId(strconv.Itoa(int(id)))
	diags = append(diags, resourceApimInstancePolicyClientIdEnfRead(ctx, d, m)...)
	// in case disabled
	disabled := d.Get("disabled").(bool)
	if disabled {
		diags = append(diags, disableApimInstancePolicyClientIdEnf(ctx, d, m)...)
		diags = append(diags, resourceApimInstancePolicyClientIdEnfRead(ctx, d, m)...)
	}

	return diags
}

func resourceApimInstancePolicyClientIdEnfRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	if isComposedResourceId(id) {
		orgid, envid, apimid, id = decomposeApimPolicyClientIdEnfId(d)
	}
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.apimpolicyclient.DefaultApi.GetApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read policy client-id-enforcement " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// process data
	data := flattenApimInstancePolicy(res)
	if err := setApimInstancePolicyAttributesToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set api policy client-id-enforcement details attributes",
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

func resourceApimInstancePolicyClientIdEnfUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	//detect change
	if d.HasChanges("configuration_data", "pointcut_data") {
		pco := m.(ProviderConfOutput)
		orgid := d.Get("org_id").(string)
		envid := d.Get("env_id").(string)
		apimid := d.Get("apim_id").(string)
		id := d.Get("id").(string)
		authctx := getApimPolicyAuthCtx(ctx, &pco)
		//prepare body
		body := newApimPolicyClientIdEnfPatchBody(d)
		//perform request
		_, httpr, err := pco.apimpolicyclient.DefaultApi.PatchApimPolicy(authctx, orgid, envid, apimid, id).Body(body).Execute()
		if err != nil {
			var details string
			if httpr != nil && httpr.StatusCode >= 400 {
				defer httpr.Body.Close()
				b, _ := io.ReadAll(httpr.Body)
				details = string(b)
			} else {
				details = err.Error()
			}
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update policy client-id-enforcement for api " + apimid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		diags = append(diags, resourceApimInstancePolicyClientIdEnfRead(ctx, d, m)...)
	}
	if d.HasChange("disabled") {
		disabled := d.Get("disabled").(bool)
		if disabled {
			diags = append(diags, disableApimInstancePolicyClientIdEnf(ctx, d, m)...)
		} else {
			diags = append(diags, enableApimInstancePolicyClientIdEnf(ctx, d, m)...)
		}
		diags = append(diags, resourceApimInstancePolicyClientIdEnfRead(ctx, d, m)...)
	}

	return diags
}

func resourceApimInstancePolicyClientIdEnfDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	httpr, err := pco.apimpolicyclient.DefaultApi.DeleteApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete policy client-id-enforcement " + id + " for api " + apimid,
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

func enableApimInstancePolicyClientIdEnf(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	_, httpr, err := pco.apimpolicyclient.DefaultApi.EnableApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to enable policy client-id-enforcement " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	return diags
}

func disableApimInstancePolicyClientIdEnf(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	_, httpr, err := pco.apimpolicyclient.DefaultApi.DisableApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to disable policy client-id-enforcement " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	return diags
}

func newApimPolicyClientIdEnfBody(d *schema.ResourceData) *apim_policy.ApimPolicyBody {
	body := apim_policy.NewApimPolicyBodyWithDefaults()
	if val, ok := d.GetOk("configuration_data"); ok {
		l := val.([]interface{})
		cfg := l[0].(map[string]interface{})
		data := newApimPolicyClientIdEnfCfg(cfg)
		body.SetConfigurationData(data)
	}
	if val, ok := d.GetOk("pointcut_data"); ok {
		body.SetPointcutData(newApimPolicyClientIdEnfPointcutDataBody(val.([]interface{})))
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
	return body
}

func newApimPolicyClientIdEnfPatchBody(d *schema.ResourceData) map[string]interface{} {
	body := make(map[string]interface{})
	if val, ok := d.GetOk("configuration_data"); ok {
		l := val.([]interface{})
		cfg := l[0].(map[string]interface{})
		data := newApimPolicyClientIdEnfCfg(cfg)
		body["configurationData"] = data
	}
	if val, ok := d.GetOk("pointcut_data"); ok {
		collection := newApimPolicyClientIdEnfPointcutDataBody(val.([]interface{}))
		slice := make([]map[string]interface{}, len(collection))
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
	return body
}

func newApimPolicyClientIdEnfCfg(input map[string]interface{}) map[string]interface{} {
	body := make(map[string]interface{})
	mode := input["credentials_origin_has_http_basic_authentication_header"].(string)
	body["credentialsOriginHasHttpBasicAuthenticationHeader"] = mode
	if mode == "httpBasicAuthenticationHeader" {
		body["clientIdExpression"] = "#[attributes.headers['client_id']]"
	} else {
		body["clientIdExpression"] = input["client_id_expression"]
		body["clientSecretExpression"] = input["client_secret_expression"]
	}
	return body
}

func newApimPolicyClientIdEnfPointcutDataBody(collection []interface{}) []apim_policy.PointcutDataItem {
	slice := make([]apim_policy.PointcutDataItem, len(collection))
	for i, item := range collection {
		data := item.(map[string]interface{})
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

func validateClientIdEnfCfg(d *schema.ResourceDiff) error {
	c := d.Get("configuration_data")
	l := c.([]interface{})
	cfg := l[0].(map[string]interface{})
	mode := cfg["credentials_origin_has_http_basic_authentication_header"].(string)

	if _, ok := cfg["client_id_expression"]; !ok && mode == "customExpression" {
		return fmt.Errorf("client_id_expression is required in \"customExpression\" mode")
	}

	if _, ok := cfg["client_secret_expression"]; !ok && mode == "customExpression" {
		return fmt.Errorf("client_secret_expression is required in \"customExpression\" mode")
	}
	return nil
}

func decomposeApimPolicyClientIdEnfId(d *schema.ResourceData) (string, string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2], s[3]
}
