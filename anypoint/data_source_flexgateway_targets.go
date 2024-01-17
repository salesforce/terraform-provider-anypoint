package anypoint

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	flexgateway "github.com/mulesoft-anypoint/anypoint-client-go/flexgateway"
)

func dataSourceFlexGatewayTargets() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFlexGatewayTargetsRead,
		Description: `
		Read all flex gateway targets.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the flex gateway targets are defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the flex gateway targets are defined.",
			},
			"targets": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of flex gateway targets",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"organization_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The organization of the flex gateway target",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The id of this flex gateway target",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the flex gateway target",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the flex gateway target",
						},
						"replicas_connected": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of connected replicas",
						},
						"replicas_disconnected": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of disconnected replicas",
						},
						"tags": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of tags",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"last_update": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last update date-time",
						},
					},
				},
			},
		},
	}
}

func dataSourceFlexGatewayTargetsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	authctx := getFlexGatewayAuthCtx(ctx, &pco)
	//exec request
	res, httpr, err := pco.flexgatewayclient.DefaultApi.GetFlexGatewayTargets(authctx, orgid, envid).Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get flex gateway targets for org " + orgid + " and env " + envid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	data := flattenFlexGatewayTargets(res)
	if err := d.Set("targets", data); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set targets of flex gateway for org " + orgid + " and env " + envid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenFlexGatewayTargets(targets []flexgateway.FlexGatewayTargetSummary) []map[string]interface{} {
	slice := make([]map[string]interface{}, len(targets))
	for i, target := range targets {
		slice[i] = flattenFlexGatewayTargetSummary(&target)
	}

	return slice
}

func flattenFlexGatewayTargetSummary(target *flexgateway.FlexGatewayTargetSummary) map[string]interface{} {
	elem := make(map[string]interface{})
	if val, ok := target.GetIdOk(); ok && val != nil {
		elem["id"] = *val
	}
	if val, ok := target.GetOrganizationIdOk(); ok && val != nil {
		elem["organization_id"] = *val
	}
	if val, ok := target.GetNameOk(); ok && val != nil {
		elem["name"] = *val
	}
	if val, ok := target.GetStatusOk(); ok && val != nil {
		elem["status"] = *val
	}
	if replicas, ok := target.GetReplicasOk(); ok && replicas != nil {
		if val, ok := replicas.GetCONNECTEDOk(); ok && val != nil {
			elem["replicas_connected"] = int(*val)
		}
		if val, ok := replicas.GetDISCONNECTEDOk(); ok && val != nil {
			elem["replicas_disconnected"] = int(*val)
		}
	}
	if val, ok := target.GetTagsOk(); ok && val != nil {
		elem["tags"] = val
	}
	if val, ok := target.GetLastUpdateOk(); ok && val != nil {
		elem["last_update"] = val.String()
	}
	return elem
}
