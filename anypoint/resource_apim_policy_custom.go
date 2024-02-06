package anypoint

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
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
		Description: `
		Create and manage an API Policy of any type.
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
				ForceNew:    true,
				Description: "the policy template version in anypoint exchange.",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceApimInstancePolicyCustomCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	//prepare body
	body, err := newApimPolicyCustomBody(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse policy configuration for api " + apimid,
			Detail:   err.Error(),
		})
		return diags
	}
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
			Summary:  "Unable to create custom policy for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	id := res.GetId()
	d.SetId(strconv.Itoa(int(id)))
	diags = append(diags, resourceApimInstancePolicyCustomRead(ctx, d, m)...)
	//in case disabled
	disabled := d.Get("disabled").(bool)
	if disabled {
		diags = append(diags, disableApimInstancePolicyCustom(ctx, d, m)...)
		diags = append(diags, resourceApimInstancePolicyCustomRead(ctx, d, m)...)
	}

	return diags
}

func resourceApimInstancePolicyCustomRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	if isComposedResourceId(id) {
		orgid, envid, apimid, id = decomposeApimPolicyCustomId(d)
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
	return diags
}

func resourceApimInstancePolicyCustomUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		body, err := newApimPolicyCustomPatchBody(d)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to parse policy configuration for api " + apimid,
				Detail:   err.Error(),
			})
			return diags
		}
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
				Summary:  "Unable to update custom policy for api " + apimid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		diags = append(diags, resourceApimInstancePolicyCustomRead(ctx, d, m)...)
	}
	if d.HasChange("disabled") {
		disabled := d.Get("disabled").(bool)
		if disabled {
			diags = append(diags, disableApimInstancePolicyCustom(ctx, d, m)...)
		} else {
			diags = append(diags, enableApimInstancePolicyCustom(ctx, d, m)...)
		}
		diags = append(diags, resourceApimInstancePolicyCustomRead(ctx, d, m)...)
	}

	return diags
}

func resourceApimInstancePolicyCustomDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func enableApimInstancePolicyCustom(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
			Summary:  "Unable to enable custom policy " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	return diags
}

func disableApimInstancePolicyCustom(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
			Summary:  "Unable to disable custom policy " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	return diags
}

func flattenApimPolicyCustomCfg(d *schema.ResourceData, policy *apim_policy.ApimPolicy) (string, error) {
	data := policy.GetConfigurationData()
	var dst map[string]interface{}
	err := json.Unmarshal([]byte(d.Get("configuration_data").(string)), &dst)
	if err != nil {
		return "", fmt.Errorf("configuration_data expected to be a valid JSON Object. %s", err.Error())
	}
	maps.Copy(dst, data)
	b, _ := json.Marshal(data)
	return string(b), nil
}

func newApimPolicyCustomBody(d *schema.ResourceData) (*apim_policy.ApimPolicyBody, error) {
	body := apim_policy.NewApimPolicyBody()
	if val, ok := d.GetOk("configuration_data"); ok {
		var cfg map[string]interface{}
		err := json.Unmarshal([]byte(val.(string)), &cfg)
		if err != nil {
			return nil, fmt.Errorf("configuration_data expected to be a valid JSON Object. %s", err.Error())
		}
		body.SetConfigurationData(cfg)
	}
	if val, ok := d.GetOk("pointcut_data"); ok {
		body.SetPointcutData(newApimPolicyCustomPointcutDataBody(val.([]interface{})))
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
	return body, nil
}

func newApimPolicyCustomPatchBody(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{})
	if val, ok := d.GetOk("configuration_data"); ok {
		var cfg map[string]interface{}
		err := json.Unmarshal([]byte(val.(string)), &cfg)
		if err != nil {
			return nil, fmt.Errorf("configuration_data expected to be a valid JSON Object. %s", err.Error())
		}
		body["configurationData"] = cfg
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
	return body, nil
}

func newApimPolicyCustomPointcutDataBody(collection []interface{}) []apim_policy.PointcutDataItem {
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

func decomposeApimPolicyCustomId(d *schema.ResourceData) (string, string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2], s[3]
}
