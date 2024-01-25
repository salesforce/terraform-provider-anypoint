package anypoint

import (
	"context"
	"io"
	"sort"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/apim_upstream"
)

func dataSourceApimInstanceUpstreams() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApimInstanceUpstreamsRead,
		Description: `
		Read an API Manager Instance Upstreams.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The API Instance's unique id",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the flex gateway instance is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the flex gateway instance is defined.",
			},
			"upstreams": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of existing upstreams in this particular api instance",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"audit": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The upstream's auditing data",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The upstream's id",
						},
						"label": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The upstream's label",
						},
						"uri": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The upstram URI",
						},
						"tls_context": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"secret_group_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The secret group id",
									},
									"tls_context_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The TLS context id in the given secret group",
									},
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The TLS context name",
									},
									"authorized": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "The TLS context authorization status",
									},
									"audit": {
										Type:        schema.TypeMap,
										Computed:    true,
										Description: "The auditing data for tls context",
									},
								},
							},
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Description: "The total number of available results",
				Computed:    true,
			},
		},
	}
}

func dataSourceApimInstanceUpstreamsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	id := d.Get("id").(string)

	authctx := getApimUpstreamAuthCtx(ctx, &pco)

	res, httpr, err := pco.apimupstreamclient.DefaultApi.GetApimInstanceUpstreams(authctx, orgid, envid, id).Execute()
	defer httpr.Body.Close()
	if err != nil && httpr.StatusCode >= 400 {
		var details string
		if httpr != nil {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get API manager instance " + id + " upstreams",
			Detail:   details,
		})
		return diags
	}

	upstreams := flattenApimUpstreamsResult(res.GetUpstreams())
	if err := d.Set("upstreams", upstreams); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set upstreams of apim instance " + id,
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("total", res.GetTotal()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number of apim instance upstreams for instance " + id,
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func flattenApimUpstreamsResult(upstreams []apim_upstream.UpstreamDetails) []interface{} {
	length := len(upstreams)
	if length > 0 {
		res := make([]interface{}, length)
		for i, upstream := range upstreams {
			res[i] = flattenApimUpstream(&upstream)
		}
		return res
	}
	return []interface{}{}
}

func flattenApimUpstream(upstream *apim_upstream.UpstreamDetails) map[string]interface{} {
	result := make(map[string]interface{})

	if val, ok := upstream.GetAuditOk(); ok {
		result["audit"] = flattenApimUpstreamAudit(val)
	}
	if val, ok := upstream.GetIdOk(); ok {
		result["id"] = *val
	}
	if val, ok := upstream.GetLabelOk(); ok {
		result["label"] = *val
	}
	if val, ok := upstream.GetUriOk(); ok {
		result["uri"] = *val
	}
	if tlscontext, ok := upstream.GetTlsContextOk(); ok {
		tlcres := make(map[string]interface{})
		if val, ok := tlscontext.GetAuditOk(); ok {
			tlcres["audit"] = flattenApimUpstreamAudit(val)
		}
		if val, ok := tlscontext.GetAuthorizedOk(); ok {
			tlcres["authorized"] = *val
		}
		if val, ok := tlscontext.GetNameOk(); ok {
			tlcres["name"] = *val
		}
		if val, ok := tlscontext.GetTlsContextIdOk(); ok && val != nil {
			tlcres["tls_context_id"] = *val
		}
		if val, ok := tlscontext.GetSecretGroupIdOk(); ok && val != nil {
			tlcres["secret_group_id"] = *val
		}
		result["tls_context"] = []interface{}{tlcres}
	}

	return result
}

func flattenApimUpstreamAudit(audit *apim_upstream.Audit) map[string]interface{} {
	result := make(map[string]interface{})
	if audit == nil {
		return result
	}
	if created, ok := audit.GetCreatedOk(); ok && created != nil {
		if val, ok := created.GetDateOk(); ok && val != nil {
			result["created"] = val.String()
		}
	}
	if updated, ok := audit.GetUpdatedOk(); ok && updated != nil {
		if val, ok := updated.GetDateOk(); ok && updated != nil {
			result["updated"] = val.String()
		}
	}
	return result
}

/*
 * Returns authentication context (includes authorization header)
 */
func getApimUpstreamAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, apim_upstream.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, apim_upstream.ContextServerIndex, pco.server_index)
}

// sorts list of upstreams by their creation date
func sortApimUpstreams(list []apim_upstream.UpstreamDetails) {
	sort.SliceStable(list, func(i, j int) bool {
		i_date := list[i].GetAudit().Created.GetDate()
		j_date := list[j].GetAudit().Created.GetDate()
		return i_date.Before(j_date)
	})
}
