package anypoint

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpn "github.com/mulesoft-consulting/anypoint-client-go/vpn"
)

func dataSourceVPN() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPNRead,
		Description: `
		Reads a specific ` + "`" + `vpn` + "`" + ` in the businessgroup and vpc
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "VPN id",
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"remote_asn": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"remote_ip_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tunnel_configs": {
				Type:     schema.TypeList,
				Required: true,
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
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vpn_connection_status": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"local_asn": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"vpn_tunnels": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accepted_route_count": {
							Type:     schema.TypeInt,
							Required: true,
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
			},
			"update_available": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func dataSourceVPNRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	fmt.Println("hero1")
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	vpcid := d.Get("vpc_id").(string)
	orgid := d.Get("org_id").(string)
	vpnid := d.Get("id").(string)
	authctx := getVPNAuthCtx(ctx, &pco)

	//request specific VPN
	res, httpr, err := pco.vpnclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdIpsecVpnIdGet(authctx, orgid, vpcid, vpnid).Execute()
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
			Summary:  "Unable to Get VPN",
			Detail:   details,
		})
		return diags
	}
	//process data
	vpninstance := flattenVPNData(&res)
	//save in data source schema
	if err := setVPNCoreAttributesToResourceData(d, vpninstance); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set VPN",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(vpcid)

	return diags
}

/*
* Transforms a vpn.VpnGet object to the dataSourceVPC schema
* Easily said: Transforms library API response object to a schema object
* @param vpcitem *vpc.Vpc the vpc struct
* @return the vpc mapped struct
 */
//Runtime error: invalid memory address or nil pointer dereference is created here
func flattenVPNData(vpnItem *vpn.VpnGet) map[string]interface{} {
	// Tried to get the json from the vpnItem this way with vpnitemValueReference as the method parameter
	// var jsonFromVpnGet []byte
	// var diags diag.Diagnostics
	// vpnItem := *vpnitemValueReference
	// jsonFromVpnGet, error := vpnItem.MarshalJSON()
	// if vpnitemValueReference != nil {
	// 	if error != nil {
	// 		var details string
	// 		details = error.Error()
	// 		diags := append(diags, diag.Diagnostic{
	// 			Severity: diag.Error,
	// 			Summary:  "Unable to marshall the vpnItem",
	// 			Detail:   details,
	// 		})
	// 		fmt.Println("These are the diags: ")
	// 		fmt.Println(diags)
	// 	} else {
	// 		fmt.Println("This is the jsonFromVpnGet: ")
	// 		fmt.Println(jsonFromVpnGet)
	// 	}
	// }
	item := make(map[string]interface{})
	item["id"] = vpnItem.GetId()
	item["update_available"] = vpnItem.GetUpdateAvailable()
	item["name"] = vpnItem.GetName()
	item["remote_asn"] = vpnItem.GetSpec().RemoteAsn
	item["remote_ip_address"] = vpnItem.GetSpec().RemoteIpAddress
	item["remote_networks"] = vpnItem.GetSpec().RemoteNetworks
	item["vpn_connection_status"] = vpnItem.GetState().VpnConnectionStatus
	item["created_at"] = vpnItem.GetState().CreatedAt
	item["local_asn"] = vpnItem.GetState().LocalAsn
	item["failed_reason"] = vpnItem.GetState().FailedReason

	//Create the TunnelConfigs - this works
	tcs := make([]interface{}, len(*vpnItem.GetSpec().TunnelConfigs))
	for i, tc := range *vpnItem.GetSpec().TunnelConfigs {
		jsonTc := make(map[string]interface{})
		jsonTc["psk"] = tc.GetPsk()
		jsonTc["ptp_cidr"] = tc.GetPtpCidr()
		jsonTc["rekey_margin_in_seconds"] = tc.GetRekeyMarginInSeconds()
		jsonTc["rekey_fuzz"] = tc.GetRekeyFuzz()
		tcs[i] = jsonTc
	}
	item["tunnel_configs"] = tcs

	//Create the VpnTunnels
	//GetState doesn't work. .RemoteAsn in state also doesn't work.
	//Next line creates: Runtime error: invalid memory address or nil pointer dereference
	vpnts := make([]interface{}, len(*vpnItem.GetState().VpnTunnels))
	for i, vpnt := range *vpnItem.GetState().VpnTunnels {
		jsonVpnt := make(map[string]interface{})
		jsonVpnt["accepted_route_count"] = vpnt.GetAcceptedRouteCount()
		jsonVpnt["last_status_change"] = vpnt.GetLastStatusChange()
		jsonVpnt["local_external_ip_address"] = vpnt.GetLocalExternalIpAddress()
		jsonVpnt["local_ptp_ip_address"] = vpnt.GetLocalPtpIpAddress()
		jsonVpnt["remote_ptp_ip_address"] = vpnt.GetRemotePtpIpAddress()
		jsonVpnt["psk"] = vpnt.GetPsk()
		jsonVpnt["status"] = vpnt.GetStatus()
		jsonVpnt["status_message"] = vpnt.GetStatusMessage()
		vpnts[i] = jsonVpnt
	}
	item["vpn_tunnels"] = vpnts
	return item
}

/*
* Copies the given vpn instance into the given resource data
* @param d *schema.ResourceData the resource data schema
* @param vpnitem map[string]interface{} the vpn instance
 */
func setVPNCoreAttributesToResourceData(d *schema.ResourceData, vpnitem map[string]interface{}) error {
	attributes := getVPNCoreAttributes()
	if vpnitem != nil {
		for _, attr := range attributes {
			if err := d.Set(attr, vpnitem[attr]); err != nil {
				return fmt.Errorf("unable to set VPN attribute %s\n details: %s", attr, err)
			}
		}
	}
	return nil
}

func getVPNCoreAttributes() []string {
	attributes := [...]string{
		"org_id", "vpc_id", "name", "remote_asn", "remote_ip_address",
		"tunnel_configs", "remote_networks", "vpn_connection_status",
		"created_at", "local_asn", "vpn_tunnels", "failedReason", "updateAvailable",
	}
	return attributes[:]
}
