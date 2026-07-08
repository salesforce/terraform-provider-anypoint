package anypoint

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	apim_tier "github.com/mulesoft-anypoint/anypoint-client-go/apim_tier"
)

func dataSourceApiInstanceSlaTiers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApiInstanceSlaTiersRead,
		Description: "Reads all SLA tiers for a given API Manager instance.",
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the resource is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the resource is defined.",
			},
			"api_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The API Manager instance id.",
			},
			"params": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Optional pagination parameters.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"limit": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     200,
							Description: "Maximum number of tiers to return.",
						},
						"offset": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Offset of the first tier to return.",
						},
					},
				},
			},
			"tiers": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of SLA tiers.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The numeric SLA tier id.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The SLA tier name.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The SLA tier description.",
						},
						"auto_approve": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether contract requests are auto-approved.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The tier status.",
						},
						"api_version_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The API version id.",
						},
						"limits": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Rate limit definitions.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"visible": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"time_period_in_milliseconds": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"maximum_requests": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"application_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"master_organization_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"organization_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of tiers returned by the API.",
			},
		},
	}
}

func dataSourceApiInstanceSlaTiersRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apiid := d.Get("api_id").(string)

	limit := int32(200)
	offset := int32(0)
	if v, ok := d.GetOk("params"); ok {
		params := v.(*schema.Set).List()
		if len(params) > 0 {
			p := params[0].(map[string]any)
			limit = int32(p["limit"].(int))
			offset = int32(p["offset"].(int))
		}
	}

	authctx := getApimTierAuthCtx(ctx, &pco)

	res, httpr, err := pco.apimtierclient.DefaultAPI.GetApiInstanceTiers(authctx, orgid, envid, apiid).
		Limit(limit).Offset(offset).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to list SLA tiers for api " + apiid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(ComposeResourceId([]string{orgid, envid, apiid}))
	d.Set("total", int(res.GetTotal()))

	tiers := res.GetTiers()
	if err := d.Set("tiers", flattenApimTierList(tiers)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set tiers",
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func flattenApimTierList(tiers []apim_tier.SlaTier) []map[string]any {
	result := make([]map[string]any, len(tiers))
	for i, t := range tiers {
		result[i] = map[string]any{
			"id":                     int(t.GetId()),
			"name":                   t.GetName(),
			"description":            t.GetDescription(),
			"auto_approve":           t.GetAutoApprove(),
			"status":                 t.GetStatus(),
			"api_version_id":         t.GetApiVersionId(),
			"application_count":      int(t.GetApplicationCount()),
			"master_organization_id": t.GetMasterOrganizationId(),
			"organization_id":        t.GetOrganizationId(),
			"limits":                 flattenApimTierLimits(t.GetLimits()),
		}
	}
	return result
}
