package anypoint

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/anypoint-client-go/dlb"
)

func dataSourceDLB() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDLBRead,
		Description: `
		Reads a specific ` + "`" + `dedicated load balancer` + "`" + ` instance.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
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
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_config": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"ip_addresses": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ip_whitelist": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"http_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_ssl_endpoint": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ssl_endpoints": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"private_key_digest": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_key_label": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_key_digest": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_key_cn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_key_label": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"verify_client_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mappings": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"input_uri": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"app_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"app_uri": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"upstream_protocol": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"static_ips_disabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"workers": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_cipher_suite": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"keep_url_encoding": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tlsv1": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"upstream_tlsv12": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"proxy_read_timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ip_addresses_info": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"static_ip": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"double_static_ips": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceDLBRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	dlbid := d.Get("id").(string)
	orgid := d.Get("org_id").(string)
	vpcid := d.Get("vpc_id").(string)

	authctx := getDLBAuthCtx(ctx, &pco)

	//request dlb
	res, httpr, err := pco.dlbclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdLoadbalancersDlbIdGet(authctx, orgid, vpcid, dlbid).Execute()
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
			Summary:  "Unable to Get DLB " + dlbid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	dlbinstance := flattenDLBData(&res)
	//save in data source schema
	if err := setDLBAttributesToResourceData(d, dlbinstance); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set DLB " + dlbid,
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(dlbid)

	return diags
}

/*
 * Copies the given dlb instance into the given resource data
 */
func setDLBAttributesToResourceData(d *schema.ResourceData, dlbitem map[string]interface{}) error {
	attributes := getDLBAttributes()
	if dlbitem != nil {
		for _, attr := range attributes {
			if err := d.Set(attr, dlbitem[attr]); err != nil {
				return fmt.Errorf("unable to set DLB attribute %s\n details: %s", attr, err)
			}
		}
	}
	return nil
}

/*
 * Transforms a dlb.Dlb object to the dataSourceDLB schema
 */
func flattenDLBData(dlbitem *dlb.Dlb) map[string]interface{} {
	if dlbitem != nil {
		item := make(map[string]interface{})
		item["id"] = dlbitem.GetId()
		item["name"] = dlbitem.GetName()
		item["vpc_id"] = dlbitem.GetVpcId()
		item["domain"] = dlbitem.GetDomain()
		item["state"] = dlbitem.GetState()
		item["deployment_id"] = dlbitem.GetDeploymentId()
		instance_config := dlbitem.GetInstanceConfig()
		item["instance_config"] = map[string]string{
			"image_name": instance_config.GetImageName(),
		}
		item["ip_addresses"] = dlbitem.GetIpAddresses()
		item["ip_whitelist"] = dlbitem.GetIpWhitelist()
		item["http_mode"] = dlbitem.GetHttpMode()
		item["default_ssl_endpoint"] = dlbitem.GetDefaultSslEndpoint()
		ssl_endpoints := make([]interface{}, len(dlbitem.GetSslEndpoints()))
		for j, ssl_endpoint := range dlbitem.GetSslEndpoints() {
			s := make(map[string]interface{})
			s["private_key_digest"] = ssl_endpoint.GetPrivateKeyDigest()
			s["public_key_label"] = ssl_endpoint.GetPublicKeyLabel()
			s["public_key_digest"] = ssl_endpoint.GetPublicKeyDigest()
			s["public_key_cn"] = ssl_endpoint.GetPublicKeyCN()
			s["private_key_label"] = ssl_endpoint.GetPrivateKeyLabel()
			s["verify_client_mode"] = ssl_endpoint.GetVerifyClientMode()
			mappings := make([]interface{}, len(ssl_endpoint.GetMappings()))
			for k, mapping := range ssl_endpoint.GetMappings() {
				m := make(map[string]interface{})
				m["input_uri"] = mapping.GetInputUri()
				m["app_name"] = mapping.GetAppName()
				m["app_uri"] = mapping.GetAppUri()
				m["upstream_protocol"] = mapping.GetUpstreamProtocol()
				mappings[k] = m
			}
			s["mappings"] = mappings
			ssl_endpoints[j] = s
		}
		item["ssl_endpoints"] = ssl_endpoints
		item["static_ips_disabled"] = dlbitem.GetStaticIPsDisabled()
		item["workers"] = dlbitem.GetWorkers()
		item["default_cipher_suite"] = dlbitem.GetDefaultCipherSuite()
		item["keep_url_encoding"] = dlbitem.GetKeepUrlEncoding()
		item["tlsv1"] = dlbitem.GetTlsv1()
		item["upstream_tlsv12"] = dlbitem.GetUpstreamTlsv12()
		item["proxy_read_timeout"] = dlbitem.GetProxyReadTimeout()
		ip_addresses_info := make([]interface{}, len(dlbitem.GetIpAddressesInfo()))
		for j, ip_address_info := range dlbitem.GetIpAddressesInfo() {
			info := make(map[string]interface{})
			info["ip"] = ip_address_info.GetIp()
			info["status"] = ip_address_info.GetStatus()
			info["static_ip"] = ip_address_info.GetStaticIp()
			ip_addresses_info[j] = info
		}
		item["ip_addresses_info"] = ip_addresses_info
		item["double_static_ips"] = dlbitem.GetDoubleStaticIps()
		return item
	}

	return nil
}

func getDLBCoreAttributes() []string {
	attributes := [...]string{
		"name", "state", "ip_whitelist", "http_mode", "default_ssl_endpoint",
		"tlsv1", "ssl_endpoints",
	}
	return attributes[:]
}

func getDLBAttributes() []string {
	attributes := [...]string{
		"id", "vpc_id", "name", "domain", "state", "deployment_id", "instance_config",
		"ip_addresses", "ip_whitelist", "http_mode", "default_ssl_endpoint",
		"ssl_endpoints", "static_ips_disabled", "workers", "default_cipher_suite",
		"keep_url_encoding", "tlsv1", "upstream_tlsv12", "proxy_read_timeout",
		"ip_addresses_info", "double_static_ips",
	}
	return attributes[:]
}
