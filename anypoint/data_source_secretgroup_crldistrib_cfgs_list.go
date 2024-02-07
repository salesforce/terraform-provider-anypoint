package anypoint

import (
	"context"
	"io"
	"maps"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secretgroup_crl_distributor_configs "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_crl_distributor_configs"
)

func dataSourceSecretGroupCrlDistribCfgsList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretGroupCrlDistribCfgsListRead,
		Description: `
		Query all or part of available crl-distributor-configs for a given secret-group, organization and environment.
		`,
		Schema: map[string]*schema.Schema{
			"sg_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The secret-group id where the crl-distributor-configs instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the crl-distributor-configs instance is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the crl-distributor-configs instance is defined.",
			},
			"list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List crl-distributor-configs result of the query",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the crl-distributor-configs",
						},
						"expiration_date": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The expiration date of the crl-distributor-configs",
						},
						"path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The path of the crl-distributor-configs",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The id of the crl-distributor-configs",
						},
					},
				},
			},
		},
	}
}

func dataSourceSecretGroupCrlDistribCfgsListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	authctx := getSgCrlDistribCfgsAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.sgcrldistribcfgsclient.DefaultApi.GetSecretGroupCrlDistribCfgsList(authctx, orgid, envid, sgid).Execute()
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
			Summary:  "Unable to get crl-distributor-configss for secret-group " + sgid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// process response
	data := flattenSgCrlDistribCfgsSummaryCollection(res)
	if err := d.Set("list", data); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set tls-contexts for secret-group " + sgid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenSgCrlDistribCfgsSummaryCollection(collection []secretgroup_crl_distributor_configs.CrlDistribCfgSummary) []interface{} {
	length := len(collection)
	if length > 0 {
		res := make([]interface{}, length)
		for i, cdc := range collection {
			res[i] = flattenSgCrlDistribCfgsSummary(&cdc)
		}
		return res
	}
	return make([]interface{}, 0)
}

func flattenSgCrlDistribCfgsSummary(cdc *secretgroup_crl_distributor_configs.CrlDistribCfgSummary) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := cdc.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := cdc.GetExpirationDateOk(); ok {
		item["expiration_date"] = *val
	}
	if val, ok := cdc.GetMetaOk(); ok {
		maps.Copy(item, flattenSgCrlDistribCfgsMeta(val))
	}
	return item
}

func flattenSgCrlDistribCfgsMeta(meta *secretgroup_crl_distributor_configs.Meta) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := meta.GetIdOk(); ok {
		item["id"] = *val
	}
	if val, ok := meta.GetPathOk(); ok {
		item["path"] = *val
	}
	return item
}
