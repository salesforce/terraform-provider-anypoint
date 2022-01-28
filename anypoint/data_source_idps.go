package anypoint

import (
	"context"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	idp "github.com/mulesoft-consulting/anypoint-client-go/idp"
)

func dataSourceIDPs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIDPsRead,
		Description: `
		Reads all ` + "`" + `identity providers` + "`" + ` in your business group.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The business group id",
			},
			"idps": {
				Type:        schema.TypeList,
				Description: "List of providers for the given org",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The provider id",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The business group id",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name fo the provider",
						},
						"type": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The type of the provider. Contains the name (saml or oidc) and the description",
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

func dataSourceIDPsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	authctx := getIDPAuthCtx(ctx, &pco)

	//request env
	res, httpr, err := pco.idpclient.DefaultApi.OrganizationsOrgIdIdentityProvidersGet(authctx, orgid).Execute()
	defer httpr.Body.Close()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Get IDPs for org " + orgid,
			Detail:   details,
		})
		return diags
	}
	//process data
	idps := flattenIDPsData(res.GetData())
	//save in data source schema
	if err := d.Set("idps", idps); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set IDPs for org " + orgid,
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("total", res.GetTotal()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number IDPs for org " + orgid,
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

/*
* Transforms a list of idp summaries object to the dataSourceIDPs schema
 */
func flattenIDPsData(idps []idp.IdpSummary) []interface{} {
	result := make([]interface{}, len(idps))
	for i, idpItem := range idps {
		item := make(map[string]interface{})
		item["provider_id"] = idpItem.GetProviderId()
		item["org_id"] = idpItem.GetOrgId()
		item["name"] = idpItem.GetName()
		t := idpItem.GetType()
		tmp := make(map[string]string)
		tmp["description"] = t.GetDescription()
		tmp["name"] = t.GetName()
		item["type"] = tmp
		result[i] = item
	}
	return result
}
