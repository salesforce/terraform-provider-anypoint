package anypoint

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/rtf"
)

var FabricsHealthStatusDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"healthy": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "True if the component is healthy",
		},
		"updated_at": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"probes": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Probes collected for this health check. Only applicable for Appliance probes.",
		},
		"failed_probes": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Probe failures attributing to the result of this health check.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"reason": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"last_transition_at": {
						Type:     schema.TypeInt,
						Computed: true,
					},
				},
			},
		},
	},
}

func dataSourceFabricsHealth() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFabricsHealthRead,
		Description: `
		Reads ` + "`" + `Runtime Fabrics'` + "`" + ` health and monitoring metrics.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Description: "The business group id",
				Required:    true,
			},
			"fabrics_id": {
				Type:        schema.TypeString,
				Description: "The runtime fabrics id",
				Required:    true,
			},
			"cluster_monitoring": {
				Type:        schema.TypeList,
				Description: "The ability to monitor and report the status of the Runtime Fabric cluster.",
				Computed:    true,
				Elem:        FabricsHealthStatusDefinition,
			},
			"manage_deployments": {
				Type:        schema.TypeList,
				Description: "The ability to create, update, or delete application deployments in this Runtime Fabric.",
				Computed:    true,
				Elem:        FabricsHealthStatusDefinition,
			},
			"load_balancing": {
				Type:        schema.TypeList,
				Description: "The ability to accept inbound requests and load-balance across different replicas of application instances.",
				Computed:    true,
				Elem:        FabricsHealthStatusDefinition,
			},
			"anypoint_monitoring": {
				Type:        schema.TypeList,
				Description: "The ability to see metrics and logs in Anypoint Monitoring.",
				Computed:    true,
				Elem:        FabricsHealthStatusDefinition,
			},
			"external_log_forwarding": {
				Type:        schema.TypeList,
				Description: "The ability to forward application logs to an external provider.",
				Computed:    true,
				Elem:        FabricsHealthStatusDefinition,
			},
			"appliance": {
				Type:        schema.TypeList,
				Description: "Detailed status of the appliance, when applicable.",
				Computed:    true,
				Elem:        FabricsHealthStatusDefinition,
			},
			"infrastructure": {
				Type:        schema.TypeList,
				Description: "Detailed status of the infrastructure supporting the Runtime Fabric cluster.",
				Computed:    true,
				Elem:        FabricsHealthStatusDefinition,
			},
			"persistent_gateway": {
				Type:        schema.TypeList,
				Description: "Detailed status of the persistent gateway for Runtime Fabric cluster.",
				Computed:    true,
				Elem:        FabricsHealthStatusDefinition,
			},
		},
	}
}

func dataSourceFabricsHealthRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	fabricsid := d.Get("fabrics_id").(string)
	authctx := getFabricsAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.rtfclient.DefaultApi.GetFabricsHealth(authctx, orgid, fabricsid).Execute()
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
			Summary:  "Unable to get fabrics health metrics",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	data := flattenFabricsHealthData(res)
	//save in data source schema
	if err := setFabricsHealthResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set fabrics health attributes",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenFabricsHealthData(data *rtf.FabricsHealth) map[string]interface{} {
	mappedItem := make(map[string]interface{})

	if val, ok := data.GetClusterMonitoringOk(); ok {
		mappedItem["cluster_monitoring"] = []interface{}{flattenFabricsHealthStatusData(val)}
	} else {
		mappedItem["cluster_monitoring"] = []interface{}{}
	}
	if val, ok := data.GetManageDeploymentsOk(); ok {
		mappedItem["manage_deployments"] = []interface{}{flattenFabricsHealthStatusData(val)}
	} else {
		mappedItem["manage_deployments"] = []interface{}{}
	}
	if val, ok := data.GetLoadBalancingOk(); ok {
		mappedItem["load_balancing"] = []interface{}{flattenFabricsHealthStatusData(val)}
	} else {
		mappedItem["load_balancing"] = []interface{}{}
	}
	if val, ok := data.GetAnypointMonitoringOk(); ok {
		mappedItem["anypoint_monitoring"] = []interface{}{flattenFabricsHealthStatusData(val)}
	} else {
		mappedItem["anypoint_monitoring"] = []interface{}{}
	}
	if val, ok := data.GetExternalLogForwardingOk(); ok {
		mappedItem["external_log_forwarding"] = []interface{}{flattenFabricsHealthStatusData(val)}
	} else {
		mappedItem["external_log_forwarding"] = []interface{}{}
	}
	if val, ok := data.GetApplianceOk(); ok {
		mappedItem["appliance"] = []interface{}{flattenFabricsHealthStatusData(val)}
	} else {
		mappedItem["appliance"] = []interface{}{}
	}
	if val, ok := data.GetInfrastructureOk(); ok {
		mappedItem["infrastructure"] = []interface{}{flattenFabricsHealthStatusData(val)}
	} else {
		mappedItem["infrastructure"] = []interface{}{}
	}
	if val, ok := data.GetPersistentGatewayOk(); ok {
		mappedItem["persistent_gateway"] = []interface{}{flattenFabricsHealthStatusData(val)}
	} else {
		mappedItem["persistent_gateway"] = []interface{}{}
	}

	return mappedItem
}

func flattenFabricsHealthStatusData(data *rtf.FabricsHealthStatus) map[string]interface{} {
	mappedItem := make(map[string]interface{})
	if val, ok := data.GetHealthyOk(); ok {
		mappedItem["healthy"] = *val
	}
	if val, ok := data.GetProbesOk(); ok {
		mappedItem["probes"] = *val
	}
	if val, ok := data.GetUpdatedAtOk(); ok {
		mappedItem["updated_at"] = *val
	}
	if val, ok := data.GetFailedProbesOk(); ok {
		list := make([]interface{}, len(val))
		for i, fhsfb := range val {
			list[i] = flattenFabricsHealthStatusFailedProbesInner(&fhsfb)
		}
		mappedItem["failed_probes"] = list
	}
	return mappedItem
}

func flattenFabricsHealthStatusFailedProbesInner(data *rtf.FabricsHealthStatusFailedProbesInner) map[string]interface{} {
	mappedItem := make(map[string]interface{})
	if val, ok := data.GetNameOk(); ok {
		mappedItem["name"] = *val
	}
	if val, ok := data.GetReasonOk(); ok {
		mappedItem["reason"] = *val
	}
	if val, ok := data.GetLastTransitionAtOk(); ok {
		mappedItem["last_transition_at"] = *val
	}
	return mappedItem
}

func setFabricsHealthResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getFabricsHealthAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set fabrics helm repo attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getFabricsHealthAttributes() []string {
	attributes := [...]string{
		"cluster_monitoring", "manage_deployments", "load_balancing",
		"anypoint_monitoring", "external_log_forwarding", "appliance",
		"infrastructure", "persistent_gateway",
	}
	return attributes[:]
}
