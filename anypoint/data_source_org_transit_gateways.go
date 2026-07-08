package anypoint

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOrgTransitGateways() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOrgTransitGatewaysRead,
		Description: "Lists transit gateways across an organization, optionally filtered by region and private space id.",
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "AWS region filter.",
			},
			"private_space_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Private space id filter.",
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

func dataSourceOrgTransitGatewaysRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	req := pco.privatespaceclient.DefaultAPI.GetOrgTransitGateways(authctx, orgid)
	if v, ok := d.GetOk("region"); ok {
		req = req.Region(v.(string))
	}
	if v, ok := d.GetOk("private_space_id"); ok {
		req = req.PrivateSpaceId(v.(string))
	}
	list, httpr, err := req.Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to List transit gateways for org " + orgid,
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
