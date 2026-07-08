package anypoint

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePrivateSpaceVpn() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePrivateSpaceVpnRead,
		Description: "Reads a single CloudHub 2.0 private-space VPN connection by its connection id. " +
			"Returns every member of the connection along with their tunnels. Note that `psk` is " +
			"never returned by the Anypoint API (always empty) and `startup_action` reflects the " +
			"API's stale GET value (see the `anypoint_private_space_vpn` resource for context).",
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the private space lives.",
			},
			"private_space_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The private space id.",
			},
			"connection_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The VPN connection id (GUID).",
			},
			"name": {Type: schema.TypeString, Computed: true, Description: "Display name of the VPN connection."},
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
	}
}

func dataSourcePrivateSpaceVpnRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	psid := d.Get("private_space_id").(string)
	cid := d.Get("connection_id").(string)
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	res, httpr, err := pco.privatespaceclient.DefaultAPI.GetPrivateSpaceVpnConnection(authctx, orgid, psid, cid).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Read vpn connection " + cid + " for private space " + psid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	if err := setPrivateSpaceVpnConnectionAttrs(d, res); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set vpn connection attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(ComposeResourceId([]string{orgid, psid, cid}))
	return diags
}
