package anypoint

import (
	"context"
	"io"
	"maps"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secretgroup_tlscontext "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_tlscontext"
)

func dataSourceSecretGroupTlsContexts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretGroupTlsContextsRead,
		Description: `
		Query all or part of available tls-contexts for a given secret-group, organization and environment.
		`,
		Schema: map[string]*schema.Schema{
			"sg_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The secret-group id where the tls-context instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the tls-context instance is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the tls-context instance is defined.",
			},
			"tlscontexts": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List tls-contexts result of the query",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the tls-context",
						},
						"expiration_date": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The expiration date of the tls-context",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The specific type of the tls-context",
						},
						"path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The path of the tls-context",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The id of the tls-context",
						},
					},
				},
			},
		},
	}
}

func dataSourceSecretGroupTlsContextsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	authctx := getSgTlsContextAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.sgtlscontextclient.DefaultApi.GetSecretGroupTlsContexts(authctx, orgid, envid, sgid).Execute()
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
			Summary:  "Unable to get tls-contexts for secret-group " + sgid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	// process response
	data := flattenSgTlsContextSummaryCollection(res)
	if err := d.Set("tlscontexts", data); err != nil {
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

func flattenSgTlsContextSummaryCollection(collection []secretgroup_tlscontext.TlsContextSummary) []interface{} {
	length := len(collection)
	if length > 0 {
		res := make([]interface{}, length)
		for i, tls := range collection {
			res[i] = flattenSgTlsContextSummary(&tls)
		}
		return res
	}
	return make([]interface{}, 0)
}

func flattenSgTlsContextSummary(tls *secretgroup_tlscontext.TlsContextSummary) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := tls.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := tls.GetExpirationDateOk(); ok {
		item["expiration_date"] = *val
	}
	if val, ok := tls.GetTypeOk(); ok {
		item["type"] = *val
	}
	if val, ok := tls.GetMetaOk(); ok {
		maps.Copy(item, flattenSgTlsContextMeta(val))
	}
	return item
}

func flattenSgTlsContextMeta(meta *secretgroup_tlscontext.Meta) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := meta.GetIdOk(); ok {
		item["id"] = *val
	}
	if val, ok := meta.GetPathOk(); ok {
		item["path"] = *val
	}
	return item
}

/*
 * Returns authentication context (includes authorization header)
 */
func getSgTlsContextAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, secretgroup_tlscontext.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, secretgroup_tlscontext.ContextServerIndex, pco.server_index)
}
