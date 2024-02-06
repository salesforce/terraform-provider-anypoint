package anypoint

import (
	"context"
	"io"
	"maps"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	secretgroup_truststore "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_truststore"
)

func dataSourceSecretGroupTruststores() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretGroupTruststoresRead,
		Description: `
		Query all or part of available truststore for a given secret-group, organization and environment.
		`,
		Schema: map[string]*schema.Schema{
			"sg_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The secret-group id where the truststore instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the truststore instance is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the truststore instance is defined.",
			},
			"params": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The search parameters. Should only provide one occurrence of the block.",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Filter the elements on the response to be of a specific type from {PEM, JKS, JCEKS, PKCS12}",
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice(
									[]string{"PEM", "JKS", "JCEKS", "PKCS12"},
									false,
								),
							),
						},
					},
				},
			},
			"truststores": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List truststores result of the query",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the truststore",
						},
						"expiration_date": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The expiration date of the truststore",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The specific type of the truststore",
						},
						"path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The path of the truststore",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The id of the truststore",
						},
					},
				},
			},
		},
	}
}

func dataSourceSecretGroupTruststoresRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("params").(*schema.Set)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	authctx := getSgTruststoreAuthCtx(ctx, &pco)
	//prepare request
	req := pco.sgtruststoreclient.DefaultApi.GetSecretGroupTruststores(authctx, orgid, envid, sgid)
	req, errDiags := parseSgTruststoreSearchOpts(req, searchOpts)
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
			Summary:  "Unable to get truststores for secret-group " + sgid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// process response
	data := flattenSgTruststoresSummaryCollection(res)
	if err := d.Set("truststores", data); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set truststores for secret-group " + sgid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenSgTruststoresSummaryCollection(collection []secretgroup_truststore.TruststoreSummary) []interface{} {
	if len(collection) > 0 {
		res := make([]interface{}, len(collection))
		for i, store := range collection {
			res[i] = flattenSgTruststoreSummary(&store)
		}
		return res
	}
	return make([]interface{}, 0)
}

func flattenSgTruststoreSummary(store *secretgroup_truststore.TruststoreSummary) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := store.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := store.GetExpirationDateOk(); ok {
		item["expiration_date"] = *val
	}
	if val, ok := store.GetTypeOk(); ok {
		item["type"] = *val
	}
	if val, ok := store.GetMetaOk(); ok {
		maps.Copy(item, flattenSgTruststoreMeta(val))
	}
	return item
}

func flattenSgTruststoreMeta(meta *secretgroup_truststore.Meta) map[string]interface{} {
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
Parses the secret-group truststore search options in order to check if the required search parameters are set correctly.
Appends the parameters to the given request
*/
func parseSgTruststoreSearchOpts(req secretgroup_truststore.DefaultApiGetSecretGroupTruststoresRequest, params *schema.Set) (secretgroup_truststore.DefaultApiGetSecretGroupTruststoresRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if params.Len() == 0 {
		return req, diags
	}
	opts := params.List()[0]
	for k, v := range opts.(map[string]interface{}) {
		if k == "type" {
			req = req.Type_(v.(string))
			continue
		}
	}
	return req, diags
}
