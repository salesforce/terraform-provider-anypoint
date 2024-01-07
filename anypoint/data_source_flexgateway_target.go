package anypoint

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	flex_gateway "github.com/mulesoft-anypoint/anypoint-client-go/flex_gateway"
)

func dataSourceFlexGatewayTarget() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFlexGatewayTargetRead,
		Description: `
		Read all flex gateway targets.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The flex gateway target's unique id",
			},
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
			"replicas": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of replicas by status type",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the flex gateway replicas",
						},
						"count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of the flex gateway replicas",
						},
						"certificate_expiration_dates": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Certificate expiration dates for the given replicas",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
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
			"versions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of version numbers",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the version number",
			},
		},
	}
}

func dataSourceFlexGatewayTargetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	id := d.Get("id").(string)
	authctx := getFlexGatewayAuthCtx(ctx, &pco)

	res, httpr, err := pco.flexgatewayclient.DefaultApi.GetFlexGatewayTargetById(authctx, orgid, envid, id).Execute()
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
			Summary:  "Unable to get flex gateway target " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	data := flattenFlexGatewayTargetDetails(&res)
	if err := setFlexGatewayTargetAttributesToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set flex gateway target attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(string(res.GetId()))
	return diags
}

func flattenFlexGatewayTargetDetails(target *flex_gateway.FlexGatewayTargetDetails) map[string]interface{} {
	elem := make(map[string]interface{})
	if val, ok := target.GetNameOk(); ok && val != nil {
		elem["name"] = *val
	}
	if val, ok := target.GetStatusOk(); ok && val != nil {
		elem["status"] = *val
	}
	if val, ok := target.GetReplicasOk(); ok && val != nil {
		elem["replicas"] = flattenFlexGatewayTargetReplicas(*val)
	}
	if val, ok := target.GetTagsOk(); ok && val != nil {
		elem["tags"] = *val
	}
	if val, ok := target.GetLastUpdateOk(); ok && val != nil {
		elem["last_update"] = val.String()
	}
	if val, ok := target.GetVersionsOk(); ok && val != nil {
		elem["versions"] = *val
	}
	if val, ok := target.GetVersionOk(); ok && val != nil {
		elem["version"] = *val
	}

	return elem
}

func flattenFlexGatewayTargetReplicas(replicas []flex_gateway.FlexGatewayTargetDetailsReplicas) []map[string]interface{} {
	slice := make([]map[string]interface{}, len(replicas))
	for i, r := range replicas {
		elem := make(map[string]interface{})
		if val, ok := r.GetStatusOk(); ok && val != nil {
			elem["status"] = *val
		}
		if val, ok := r.GetCountOk(); ok && val != nil {
			elem["count"] = *val
		}
		if dates, ok := r.GetCertificateExpirationDatesOk(); ok && dates != nil {
			strdates := make([]string, len(*dates))
			for j, d := range *dates {
				strdates[j] = d.String()
			}
			elem["certificate_expiration_dates"] = strdates
		}
		slice[i] = elem
	}
	return slice
}

func setFlexGatewayTargetAttributesToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getFlexGatewayTargetAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set flex gateway target attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getFlexGatewayTargetAttributes() []string {
	attributes := [...]string{
		"name", "status", "replicas", "tags",
		"last_update", "versions", "version",
	}
	return attributes[:]
}
