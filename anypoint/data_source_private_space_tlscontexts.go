package anypoint

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/private_space_tlscontext"
)

func dataSourcePrivateSpaceTlsContexts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePrivateSpaceTlsContextsRead,
		Description: `
		Reads all ` + "`" + `private space tls contexts` + "`" + ` in your private space.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"private_space_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tls_contexts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: PRIVATE_SPACE_TLSCONTEXT_SCHEMA,
				},
			},
		},
	}
}

func dataSourcePrivateSpaceTlsContextsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	private_space_id := d.Get("private_space_id").(string)
	authctx := getPrivateSpaceTlsContextAuthCtx(ctx, &pco)
	//request
	res, httpr, err := pco.privatespacetlscontextclient.DefaultApi.GetTlsContexts(authctx, orgid, private_space_id).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Get Private Space TLS Contexts for org " + orgid + " and private space " + private_space_id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	private_space_tls_contexts := flattenPrivateSpaceTlsContextsData(res)
	//save in data source schema
	if err := d.Set("tls_contexts", private_space_tls_contexts); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set Private Space TLS Contexts for org " + orgid + " and private space " + private_space_id,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func flattenPrivateSpaceTlsContextsData(item []private_space_tlscontext.TlsContext) []map[string]any {
	result := make([]map[string]any, len(item))
	for i, item := range item {
		result[i] = flattenPrivateSpaceTlsContextData(&item)
	}
	return result
}
