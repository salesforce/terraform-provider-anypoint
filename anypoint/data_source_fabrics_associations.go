package anypoint

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rtf "github.com/mulesoft-anypoint/anypoint-client-go/rtf"
)

func dataSourceFabricsAssociations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFabricsAssociationsRead,
		Description: `
		Reads all ` + "`" + `Runtime Fabrics'` + "`" + ` available in your org.
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
			"associations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique id of the fabrics instance in the platform.",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The organization id associated with fabrics.",
						},
						"env_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The environment associated with fabrics.",
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

func dataSourceFabricsAssociationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	fabricsId := d.Get("fabrics_id").(string)
	authctx := getFabricsAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.rtfclient.DefaultApi.GetFabricsAssociations(authctx, orgid, fabricsId).Execute()
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
			Summary:  "Unable to get fabrics associations",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	list := flattenFabricsAssociationsData(res)
	//save in data source schema
	if err := d.Set("associations", list); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set fabrics associations",
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("total", len(list)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number fabrics associations",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenFabricsAssociationsData(associations []rtf.FabricsAssociationsInner) []interface{} {
	if len(associations) == 0 {
		return make([]interface{}, 0)
	}
	res := make([]interface{}, len(associations))
	for i, association := range associations {
		res[i] = flattenFabricsAssociationData(&association)
	}
	return res
}

func flattenFabricsAssociationData(association *rtf.FabricsAssociationsInner) map[string]interface{} {
	mappedItem := make(map[string]interface{})

	if val, ok := association.GetIdOk(); ok {
		mappedItem["id"] = *val
	}
	if val, ok := association.GetOrganizationIdOk(); ok {
		mappedItem["org_id"] = *val
	}
	if val, ok := association.GetEnvironmentIdOk(); ok {
		mappedItem["env_id"] = *val
	}

	return mappedItem
}
