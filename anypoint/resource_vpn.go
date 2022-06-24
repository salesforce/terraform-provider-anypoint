package anypoint

import (
	"context"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	vpn "github.com/mulesoft-consulting/anypoint-client-go/vpn"
)

func resourceVPN() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPNCreate,
		ReadContext:   resourceVPNRead,
		DeleteContext: resourceVPNDelete,
		// UpdateContext: resourceVPNUpdate,
		Description: `
		Creates a ` + "`" + `vpn` + "`" + `component.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"remote_asn": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"remote_ip_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tunnel_configs": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"psk": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ptp_cidr": {
							Type:     schema.TypeString,
							Required: true,
						},
						"rekey_margin_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"rekey_fuzz": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"remote_networks": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vpn_connection_status": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"local_asn": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"vpn_tunnels": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accepted_route_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"last_status_change": {
							Type:     schema.TypeString,
							Required: true,
						},
						"local_external_ip_address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"local_ptp_ip_address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"remote_ptp_ip_address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"psk": {
							Type:     schema.TypeString,
							Required: true,
						},
						"status": {
							Type:     schema.TypeString,
							Required: true,
						},
						"status_message": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"failed_reason": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"update_available": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPNCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	vpcid := d.Get("vpc_id").(string)
	authctx := getVPNAuthCtx(ctx, &pco)

	body := newVPNBody(d)

	//request vpn creation
	res, httpr, err := pco.vpnclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdIpsecPost(authctx, orgid, vpcid).VpnPostReqBody(*body).Execute()
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
			Summary:  "Unable to Create VPN",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(res.GetId())

	resourceVPNRead(ctx, d, m)

	return diags
}

func resourceVPNRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	vpcid := d.Get("vpc_id").(string)
	vpnid := d.Id()

	authctx := getVPNAuthCtx(ctx, &pco)

	res, httpr, err := pco.vpnclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdIpsecVpnIdGet(authctx, orgid, vpcid, vpnid).Execute()
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
			Summary:  "Unable to Get VPN",
			Detail:   details,
		})
		return diags
	}
	//process data
	//FLATTENVPNDATA DOESNT WORK CORRECTLY YET, throws error: Runtime error: invalid memory address or nil pointer dereference
	vpcinstance := flattenVPNData(&res)
	//save in data source schema
	//setVPNCoreAttributesToResourceData hasn't been tested yet
	if err := setVPNCoreAttributesToResourceData(d, vpcinstance); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set VPC",
			Detail:   err.Error(),
		})
		return diags
	}
	return diags
}

//Method hasn't been tested yet.
func resourceVPNDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	vpcid := d.Get("vpc_id").(string)
	vpnid := d.Id()

	authctx := getVPNAuthCtx(ctx, &pco)

	httpr, err := pco.vpnclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdIpsecVpnIdDelete(authctx, orgid, vpcid, vpnid).Execute()
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
			Summary:  "Unable to Delete VPN",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

/*
 * Creates a new VPN Requestbody struct from the resource data schema
 */
//Method works
func newVPNBody(d *schema.ResourceData) *vpn.VpnPostReqBody {
	body := vpn.NewVpnPostReqBodyWithDefaults()
	body.SetName(d.Get("name").(string))
	body.SetRemoteAsn(int32(d.Get("remote_asn").(int)))
	body.SetRemoteIpAddress(d.Get("remote_ip_address").(string))

	//preparing remote_networks
	rn := d.Get("remote_networks").([]interface{})
	remote_networks := make([]string, len(rn))
	for index, e := range rn {
		remote_networks[index] = e.(string)
	}
	body.SetRemoteNetworks(remote_networks)
	//preparing tunnel_configs
	tc := d.Get("tunnel_configs").([]interface{})
	tunnel_configs := make([]vpn.TunnelConfig, len(tc))
	for index, tunnel_config := range tc {
		tunnel_configs[index] = *vpn.NewTunnelConfig(tunnel_config.(map[string]interface{})["psk"].(string), tunnel_config.(map[string]interface{})["ptp_cidr"].(string))
	}
	body.SetTunnelConfigs(tunnel_configs)

	return body
}

/*
 * Returns authentication context (includes authorization header)
 */
func getVPNAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, vpn.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, vpn.ContextServerIndex, pco.server_index)
}
