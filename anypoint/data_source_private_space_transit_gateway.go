package anypoint

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	private_space "github.com/mulesoft-anypoint/anypoint-client-go/private_space"
)

func dataSourcePrivateSpaceTransitGateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePrivateSpaceTransitGatewayRead,
		Description: "Reads one transit gateway attachment on a private space, by transit gateway id.",
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"private_space_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"transit_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The AWS transit gateway id (tgw-...).",
			},
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
	}
}

func dataSourcePrivateSpaceTransitGatewayRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	psid := d.Get("private_space_id").(string)
	tgwid := d.Get("transit_gateway_id").(string)
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	list, httpr, err := pco.privatespaceclient.DefaultAPI.GetPrivateSpaceTransitGateways(authctx, orgid, psid).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Read transit gateway " + tgwid + " for private space " + psid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	var found *private_space.TransitGateway
	for i := range list {
		if list[i].GetId() == tgwid {
			found = &list[i]
			break
		}
	}
	if found == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Transit gateway " + tgwid + " not found in private space " + psid,
		})
		return diags
	}
	if err := setPrivateSpaceTransitGatewayAttrs(d, found); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set transit gateway attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(ComposeResourceId([]string{orgid, psid, tgwid}))
	return diags
}
