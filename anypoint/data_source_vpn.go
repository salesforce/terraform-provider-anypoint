package anypoint

import (
	"context"
	"errors"
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
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	vpcid := d.Get("vpc_id").(string)
	orgid := d.Get("org_id").(string)
	vpnid := d.Id()
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
	vpninstance, err := flattenVPNData(&res)
	if err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse VPN data",
			Detail:   err.Error(),
		})
		return diags
	}
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

/**
Transforms a vpn.VpnGet object to the dataSourceVPC schema
Easily said: Transforms library API response object to a schema object
@param vpcitem *vpc.Vpc the vpc struct
@return the vpc mapped struct
*/
func flattenVPNData(vpnItem *vpn.VpnGet) (map[string]interface{}, error) {
	item := make(map[string]interface{})
	var vpnState vpn.State
	if val, ok := vpnItem.GetStateOk(); ok {
		vpnState = *val
	} else {
		return nil, errors.New("couldn't parse vpn state field")
	}
	var vpnSpec vpn.Spec
	if val, ok := vpnItem.GetSpecOk(); ok {
		vpnSpec = *val
	} else {
		return nil, errors.New("couldn't parse vpn spec field")
	}

	item["id"] = vpnItem.GetId()
	item["name"] = vpnItem.GetName()
	item["update_available"] = vpnItem.GetUpdateAvailable()
	item["remote_asn"] = *vpnSpec.RemoteAsn
	item["remote_ip_address"] = vpnSpec.RemoteIpAddress
	item["remote_networks"] = vpnSpec.RemoteNetworks
	item["vpn_connection_status"] = vpnState.VpnConnectionStatus

	vpnTunnelConfig := *vpnSpec.TunnelConfigs
	//Create the TunnelConfigs - this works
	tcs := make([]map[string]interface{}, len(vpnTunnelConfig))
	for i, tc := range vpnTunnelConfig {
		jsonTc := make(map[string]interface{})
		jsonTc["psk"] = tc.GetPsk()
		jsonTc["ptp_cidr"] = tc.GetPtpCidr()
		jsonTc["rekey_margin_in_seconds"] = tc.GetRekeyMarginInSeconds()
		jsonTc["rekey_fuzz"] = tc.GetRekeyFuzz()
		tcs[i] = jsonTc
	}
	item["tunnel_configs"] = tcs

	// The list of tunnels may not exist when the vpn is on error
	if val, ok := vpnState.GetVpnTunnelsOk(); ok {
		vpnTunnels := *val
		vpnts := make([]map[string]interface{}, len(vpnTunnels))
		for i, vpnt := range vpnTunnels {
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
	}
	if val, ok := vpnState.GetFailedReasonOk(); ok { //may not exist if vpn is on error
		item["failed_reason"] = *val
	}
	if val, ok := vpnState.GetCreatedAtOk(); ok { //may not exist if vpn is on error
		item["created_at"] = *val
	}
	if val, ok := vpnState.GetLocalAsnOk(); ok { //may not exist if vpn is on error
		item["local_asn"] = *val
	}

	return item, nil
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
		"name", "remote_asn", "remote_ip_address",
		"tunnel_configs", "remote_networks", "vpn_connection_status",
		"created_at", "local_asn", "vpn_tunnels", "failed_reason", "update_available",
	}
	return attributes[:]
}
