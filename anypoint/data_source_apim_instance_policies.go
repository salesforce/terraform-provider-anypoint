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

func dataSourceApimInstancePolicies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApimInstancePoliciesRead,
		Description: `
		Read all API Manager Instance Policies.
		`,
		Schema: map[string]*schema.Schema{
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
			"policies": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of policies result of the query",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy id.",
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
				},
			},
		},
	}
}

func dataSourceApimInstancePoliciesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.apimpolicyclient.DefaultApi.GetApimPolicies(authctx, orgid, envid, apimid).FullInfo(false).Execute()
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
			Summary:  "Unable to get policies for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	data := flattenApimInstancePolicies(*res.ArrayOfApimPolicy)
	if err := d.Set("policies", data); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set policies for api instance " + apimid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func flattenApimInstancePolicies(collection []apim_policy.ApimPolicy) []interface{} {
	slice := make([]interface{}, len(collection))
	for i, policy := range collection {
		slice[i] = flattenApimInstancePolicy(&policy)
	}
	return slice
}
