package anypoint

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	vpc "github.com/mulesoft-consulting/cloudhub-client-go/vpc"
)

func dataSourceVPC() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPCRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"orgid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cidr_block": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"internal_dns_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"internal_dns_special_domains": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"associated_environments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"owner_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"shared_with": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"firewall_rules": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr_block": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"from_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"to_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"vpc_routes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"next_hop": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceVPCRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	vpcid := d.Get("id").(string)
	orgid := d.Get("orgid").(string)

	if vpcid == "" || orgid == "" {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "VPC id (id) and Organization ID (orgid) are required",
			Detail:   "VPC id (id) and Organization ID (orgid) must be provided",
		})
		return diags
	}

	authctx := getVPCAuthCtx(&pco)

	//request vpcs
	res, httpr, err := pco.vpcclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdGet(authctx, orgid, vpcid).Execute()
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
			Summary:  "Unable to Get VPC",
			Detail:   details,
		})
		return diags
	}
	//process data
	vpcinstance := flattenVPCData(&res)
	//save in data source schema
	if err := setVPCCoreAttributesToResourceData(d, vpcinstance); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set VPC",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(vpcid)

	return diags
}

/*
* Copies the given vpc instance into the given resource data
* @param d *schema.ResourceData the resource data schema
* @param vpcitem map[string]interface{} the vpc instance
 */
func setVPCCoreAttributesToResourceData(d *schema.ResourceData, vpcitem map[string]interface{}) error {
	attributes := getVPCCoreAttributes()
	if vpcitem != nil {
		for _, attr := range attributes {
			if err := d.Set(attr, vpcitem[attr]); err != nil {
				return fmt.Errorf("unable to set VPC attribute %s\n details: %s", attr, err)
			}
		}
	}
	return nil
}

/*
* Transforms a vpc.Vpc object to the dataSourceVPC schema
* @param vpcitem *vpc.Vpc the vpc struct
* @return the vpc mapped struct
 */
func flattenVPCData(vpcitem *vpc.Vpc) map[string]interface{} {
	if vpcitem != nil {
		item := make(map[string]interface{})

		item["id"] = vpcitem.GetId()
		item["name"] = vpcitem.GetName()
		item["region"] = vpcitem.GetRegion()
		item["cidr_block"] = vpcitem.GetCidrBlock()
		item["internal_dns_servers"] = vpcitem.GetInternalDns().DnsServers
		item["internal_dns_special_domains"] = vpcitem.GetInternalDns().SpecialDomains
		item["is_default"] = vpcitem.GetIsDefault()
		item["associated_environments"] = vpcitem.GetAssociatedEnvironments()
		item["owner_id"] = vpcitem.GetOwnerId()
		item["shared_with"] = vpcitem.GetSharedWith()

		frules := make([]interface{}, len(vpcitem.GetFirewallRules()))
		for j, frule := range vpcitem.GetFirewallRules() {
			r := make(map[string]interface{})
			r["cidr_block"] = frule.GetCidrBlock()
			r["protocol"] = frule.GetProtocol()
			r["from_port"] = frule.GetFromPort()
			r["to_port"] = frule.GetToPort()
			frules[j] = r
		}
		item["firewall_rules"] = frules

		vpcroutes := make([]interface{}, len(vpcitem.GetVpcRoutes()))
		for j, vpcroute := range vpcitem.GetVpcRoutes() {
			r := make(map[string]interface{})
			r["next_hop"] = vpcroute.GetNextHop()
			r["cidr"] = vpcroute.GetCIDR()
			vpcroutes[j] = r
		}
		item["vpc_routes"] = vpcroutes
		return item
	}

	return nil
}

func getVPCCoreAttributes() []string {
	attributes := [...]string{
		"name", "region", "cidr_block", "internal_dns_servers", "internal_dns_special_domains",
		"is_default", "associated_environments", "owner_id", "shared_with", "firewall_rules", "vpc_routes",
	}
	return attributes[:]
}
