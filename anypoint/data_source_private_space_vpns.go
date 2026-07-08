package anypoint

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePrivateSpaceVpns() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePrivateSpaceVpnsRead,
		Description: "Lists every VPN connection attached to a CloudHub 2.0 private space. Each " +
			"entry has the same shape as the singular `anypoint_private_space_vpn` data source. " +
			"Use to assert that no rogue connections exist or to fan a downstream resource over " +
			"every connection.",
		Schema: map[string]*schema.Schema{
			"org_id":           {Type: schema.TypeString, Required: true, Description: "The organization id where the private space lives."},
			"private_space_id": {Type: schema.TypeString, Required: true, Description: "The private space id."},
			"connections": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_id": {Type: schema.TypeString, Computed: true},
						"name":          {Type: schema.TypeString, Computed: true},
						"vpns": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name":                  {Type: schema.TypeString, Computed: true},
									"local_asn":             {Type: schema.TypeString, Computed: true},
									"remote_asn":            {Type: schema.TypeString, Computed: true},
									"remote_ip_address":     {Type: schema.TypeString, Computed: true},
									"static_routes":         {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
									"startup_action":        {Type: schema.TypeString, Computed: true},
									"vpn_id":                {Type: schema.TypeString, Computed: true},
									"connection_id":         {Type: schema.TypeString, Computed: true},
									"connection_name":       {Type: schema.TypeString, Computed: true},
									"vpn_connection_status": {Type: schema.TypeString, Computed: true},
									"vpn_tunnels": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"psk":             {Type: schema.TypeString, Computed: true, Sensitive: true},
												"ptp_cidr":        {Type: schema.TypeString, Computed: true},
												"is_logs_enabled": {Type: schema.TypeBool, Computed: true},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourcePrivateSpaceVpnsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	psid := d.Get("private_space_id").(string)
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	list, httpr, err := pco.privatespaceclient.DefaultAPI.GetPrivateSpaceVpnConnections(authctx, orgid, psid).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to List vpn connections for private space " + psid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	out := make([]map[string]any, 0, len(list))
	for _, c := range list {
		out = append(out, map[string]any{
			"connection_id": c.GetId(),
			"name":          c.GetName(),
			"vpns":          flattenPrivateSpaceVpns(c.GetVpns(), nil),
		})
	}
	if err := d.Set("connections", out); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set connections attribute",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(ComposeResourceId([]string{orgid, psid, "vpns"}))
	return diags
}
