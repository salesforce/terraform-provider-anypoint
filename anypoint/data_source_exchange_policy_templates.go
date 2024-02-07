package anypoint

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/apim_policy"
)

func dataSourceExchangePolicyTemplates() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceExchangePolicyTemplatesRead,
		Description: `
		Query all or part of exchange policy templates.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id.",
			},
			"params": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The search parameters. Should only provide one occurrence of the block.",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"env_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The environment id",
						},
						"split_model": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to include asset split model",
						},
						"latest": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to include only latest versions",
						},
						"api_instance_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Include only templates used for given api instance id",
						},
						"include_configuration": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to include configuration",
						},
						"automated_only": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to include automated policies only",
						},
					},
				},
			},
			"templates": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The result of the query",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The exchange policy template id.",
						},
						"audit": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The exchange policy template auditing data.",
						},
						"group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy template group id in exchange.",
						},
						"asset_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy template asset id.",
						},
						"version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy template version.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy template name.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy template description.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy template type.",
						},
						"is_ootb": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the policy template is out of the box.",
						},
						"stage": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy template stage.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy template status.",
						},
						"yaml_md5": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The yaml file checksum for data integrity.",
						},
						"jar_md5": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The jar file checksum for data integrity.",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The orgnization id.",
						},
						"min_mule_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The minimum mule version to use the policy.",
						},
						"supported_policies_versions": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The supported policies versions.",
						},
						"category": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The category of the policy template.",
						},
						"violation_category": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The violation category of the policy template.",
						},
						"resource_level_supported": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the policy template supports resource level.",
						},
						"encryption_supported": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the policy template supports encryption.",
						},
						"standalone": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the policy template is standalone.",
						},
						"required_characteristics": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of required characteristics.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"identity_management_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of identity management.",
						},
						"provided_characteristics": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of provided characteristics.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"raml_snippet": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy template snippet in RAML.",
						},
						"raml_v1_snippet": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy template snippet in RAML v1.",
						},
						"oas_v2_snippet": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy template snippet in OAS v2.",
						},
						"oas_v3_snippet": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy template snippet in OAS v3.",
						},
						"applicable": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the template snippet is applicable.",
						},
						"configuration": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The policy template list of property configurations.",
							Elem: &schema.Resource{
								Schema: EXCHANGE_POLICY_TEMPLATE_CONFIG,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceExchangePolicyTemplatesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("params").(*schema.Set)
	orgid := d.Get("org_id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	req := pco.apimpolicyclient.DefaultApi.GetOrgExchangePolicyTemplates(authctx, orgid)
	req, errDiags := parseExchangePolicyTemplatesSearchOpts(req, searchOpts)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}
	//execut request
	res, httpr, err := req.Execute()
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
			Summary:  "Unable to get exchange policy templates for org " + orgid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	data := flattenExchangePolicyTemplatesResult(res)
	if err := d.Set("templates", data); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set exchange policy templates for org " + orgid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func parseExchangePolicyTemplatesSearchOpts(req apim_policy.DefaultApiGetOrgExchangePolicyTemplatesRequest, params *schema.Set) (apim_policy.DefaultApiGetOrgExchangePolicyTemplatesRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if params.Len() == 0 {
		return req, diags
	}
	opts := params.List()[0]
	for k, v := range opts.(map[string]interface{}) {
		if k == "env_id" {
			req = req.EnvironmentId(v.(string))
			continue
		}
		if k == "split_model" {
			req = req.SplitModel(v.(bool))
			continue
		}
		if k == "latest" {
			req = req.Latest(v.(bool))
			continue
		}
		if k == "api_instance_id" {
			req = req.ApiInstanceId(v.(string))
			continue
		}
		if k == "include_configuration" {
			req = req.IncludeConfiguration(v.(bool))
			continue
		}
		if k == "automated_only" {
			req = req.AutomatedOnly(v.(bool))
			continue
		}
	}
	return req, diags
}

func flattenExchangePolicyTemplatesResult(collection []apim_policy.ExchangePolicyTemplate) []interface{} {
	slice := make([]interface{}, len(collection))
	for i, template := range collection {
		slice[i] = flattenExchangePolicyTemplate(&template)
	}
	return slice
}
