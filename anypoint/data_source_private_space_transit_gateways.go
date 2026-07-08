package anypoint

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	private_space "github.com/mulesoft-anypoint/anypoint-client-go/private_space"
)

func dataSourcePrivateSpaceTransitGateways() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePrivateSpaceTransitGatewaysRead,
		Description: "Lists transit gateway attachments on a private space.",
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"private_space_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"transit_gateways": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transit_gateway_id":     {Type: schema.TypeString, Computed: true},
						"name":                   {Type: schema.TypeString, Computed: true},
						"account_id":             {Type: schema.TypeString, Computed: true},
						"region":                 {Type: schema.TypeString, Computed: true},
						"space_name":             {Type: schema.TypeString, Computed: true},
						"resource_share_id":      {Type: schema.TypeString, Computed: true},
						"resource_share_account": {Type: schema.TypeString, Computed: true},
						"status_gateway":         {Type: schema.TypeString, Computed: true},
						"status_attachment":      {Type: schema.TypeString, Computed: true},
						"tgw_resource_url":       {Type: schema.TypeString, Computed: true},
						"routes": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourcePrivateSpaceTransitGatewaysRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	psid := d.Get("private_space_id").(string)
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	list, httpr, err := pco.privatespaceclient.DefaultAPI.GetPrivateSpaceTransitGateways(authctx, orgid, psid).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to List transit gateways for private space " + psid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	if err := d.Set("transit_gateways", flattenTransitGatewayList(list)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set transit gateways list",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func flattenTransitGatewayList(list []private_space.TransitGateway) []map[string]any {
	out := make([]map[string]any, 0, len(list))
	for i := range list {
		tgw := &list[i]
		spec := tgw.GetSpec()
		share := spec.GetResourceShare()
		status := tgw.GetStatus()
		out = append(out, map[string]any{
			"transit_gateway_id":     tgw.GetId(),
			"name":                   tgw.GetName(),
			"account_id":             tgw.GetAccountId(),
			"region":                 spec.GetRegion(),
			"space_name":             spec.GetSpaceName(),
			"resource_share_id":      share.GetId(),
			"resource_share_account": share.GetAccount(),
			"status_gateway":         status.GetGateway(),
			"status_attachment":      status.GetAttachment(),
			"tgw_resource_url":       status.GetTgwResource(),
			"routes":                 status.GetRoutes(),
		})
	}
	return out
}
