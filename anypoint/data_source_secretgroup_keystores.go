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
	"github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_keystore"
)

func dataSourceSecretGroupKeystores() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretGroupKeystoresRead,
		Description: `
		Query all or part of available keystores for a given secret-group, organization and environment.
		`,
		Schema: map[string]*schema.Schema{
			"sg_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The secret-group id where the keystore instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the keystore instance is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the keystore instance is defined.",
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
			"keystores": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List keystores result of the query",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the keystore",
						},
						"expiration_date": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The expiration date of the keystore",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The specific type of the keystore",
						},
						"path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The path of the keystore",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The id of the keystore",
						},
					},
				},
			},
		},
	}
}

func dataSourceSecretGroupKeystoresRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("params").(*schema.Set)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	authctx := getSgKeystoreAuthCtx(ctx, &pco)
	//prepare request
	req := pco.sgkeystoreclient.DefaultApi.GetSecretGroupKeystores(authctx, orgid, envid, sgid)
	req, errDiags := parseSgKeystoreSearchOpts(req, searchOpts)
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
			Summary:  "Unable to get keystores for secret-group " + sgid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// process response
	data := flattenSgKeystoresSummaryCollection(res)
	if err := d.Set("keystores", data); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set keystores for secret-group " + sgid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

/*
Parses the secret-group keystore search options in order to check if the required search parameters are set correctly.
Appends the parameters to the given request
*/
func parseSgKeystoreSearchOpts(req secretgroup_keystore.DefaultApiGetSecretGroupKeystoresRequest, params *schema.Set) (secretgroup_keystore.DefaultApiGetSecretGroupKeystoresRequest, diag.Diagnostics) {
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

func flattenSgKeystoresSummaryCollection(collection []secretgroup_keystore.KeystoreSummary) []interface{} {
	length := len(collection)
	if length > 0 {
		res := make([]interface{}, length)
		for i, keystore := range collection {
			res[i] = flattenSgKeystoreSummary(&keystore)
		}
		return res
	}
	return make([]interface{}, 0)
}

func flattenSgKeystoreSummary(ksummary *secretgroup_keystore.KeystoreSummary) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := ksummary.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := ksummary.GetExpirationDateOk(); ok {
		item["expiration_date"] = *val
	}
	if val, ok := ksummary.GetTypeOk(); ok {
		item["type"] = *val
	}
	if val, ok := ksummary.GetMetaOk(); ok {
		maps.Copy(item, flattenSgKeystoreMeta(val))
	}
	return item
}
