package anypoint

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iancoleman/strcase"
	"github.com/mulesoft-anypoint/anypoint-client-go/apim_policy"
)

func dataSourceApimInstancePolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApimInstancePolicyRead,
		Description: `
		Read an API Manager Instance Policy.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The policy's unique id",
			},
			"apim_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The api manager instance id where the api instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the api instance is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
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
				Computed:    true,
				Description: "The policy configuration data",
				Elem: &schema.Schema{
					Type: schema.TypeMap,
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
				Computed:    true,
				Description: "Whether the policy is disabled.",
			},
			"pointcut_data": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The Method & resource conditions",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method_regex": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The list of HTTP methods",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"uri_template_regex": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URI template regex",
						},
					},
				},
			},
			"asset_group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "policy exchange asset group id.",
			},
			"asset_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "policy exchange asset id.",
			},
			"asset_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "policy exchange asset version.",
			},
		},
	}
}

func dataSourceApimInstancePolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
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
			Summary:  "Unable to get policy " + id + " for api " + apimid,
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
			Summary:  "Unable to set api policy details attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(id)
	return diags
}

func flattenApimInstancePolicy(policy *apim_policy.ApimPolicy) map[string]interface{} {
	result := make(map[string]interface{})
	if val, ok := policy.GetIdOk(); ok {
		result["id"] = strconv.Itoa(int(*val))
	}
	if val, ok := policy.GetAuditOk(); ok {
		result["audit"] = flattenApimInstancePolicyAudit(val)
	}
	if val, ok := policy.GetMasterOrganizationIdOk(); ok {
		result["master_organization_id"] = *val
	}
	if val, ok := policy.GetOrderOk(); ok {
		result["order"] = int(*val)
	}
	if val, ok := policy.GetDisabledOk(); ok {
		result["disabled"] = *val
	}
	if val, ok := policy.GetPointcutDataOk(); ok {
		result["pointcut_data"] = flattenApimInstancePolicyPointcutData(val)
	}
	if val, ok := policy.GetGroupIdOk(); ok {
		result["asset_group_id"] = *val
	}
	if val, ok := policy.GetAssetIdOk(); ok {
		result["asset_id"] = *val
	}
	if val, ok := policy.GetAssetVersionOk(); ok {
		result["asset_version"] = *val
	}
	if val, ok := policy.GetPolicyTemplateIdOk(); ok {
		result["policy_template_id"] = *val
	}
	if val, ok := policy.GetConfigurationDataOk(); ok {
		result["configuration_data"] = []interface{}{flattenApimInstancePolicyConfData(val)}
	}
	return result
}

func flattenApimInstancePolicyAudit(audit *apim_policy.Audit) map[string]interface{} {
	result := make(map[string]interface{})
	if audit == nil {
		return result
	}
	if created, ok := audit.GetCreatedOk(); ok && created != nil {
		if val, ok := created.GetDateOk(); ok && val != nil {
			result["created"] = val
		}
	}
	if updated, ok := audit.GetUpdatedOk(); ok && updated != nil {
		if val, ok := updated.GetDateOk(); ok && updated != nil {
			result["updated"] = val
		}
	}
	return result
}

func flattenApimInstancePolicyConfData(conf map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range conf {
		m, ok := v.(map[string]interface{})
		if ok {
			b, _ := json.Marshal(m)
			result[strcase.ToSnake(k)] = string(b)
			continue
		}
		t, ok := v.([]interface{})
		if ok {
			b, _ := json.Marshal(t)
			result[strcase.ToSnake(k)] = string(b)
			continue
		}
		result[strcase.ToSnake(k)] = ConvPrimtiveInterface2String(v)
	}
	return result
}

func setApimInstancePolicyAttributesToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getApimInstancePolicyAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set api manager instance policy attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getApimInstancePolicyAttributes() []string {
	attributes := [...]string{
		"audit", "master_organization_id", "configuration_data",
		"order", "disabled", "policy_template_id", "asset_group_id",
		"asset_id", "asset_version", "pointcut_data",
	}
	return attributes[:]
}

func flattenApimInstancePolicyPointcutData(collection []apim_policy.PointcutDataItem) []interface{} {
	slice := make([]interface{}, len(collection))
	for i, item := range collection {
		data := make(map[string]interface{})
		if val, ok := item.GetMethodRegexOk(); ok {
			data["method_regex"] = strings.Split(*val, "|")
		}
		if val, ok := item.GetUriTemplateRegexOk(); ok {
			data["uri_template_regex"] = *val
		}
		slice[i] = data
	}
	return slice
}

/*
 * Returns authentication context (includes authorization header)
 */
func getApimPolicyAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, apim_policy.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, apim_policy.ContextServerIndex, pco.server_index)
}
