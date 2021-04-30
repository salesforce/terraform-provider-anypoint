package anypoint

import (
	"context"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	vpc "github.com/mulesoft-consulting/cloudhub-client-go/vpc"
)

func dataSourceVPCs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPCsRead,
		Schema: map[string]*schema.Schema{
			"vpcs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
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
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"internal_dns_special_domains": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"is_default": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"associated_environments": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"owner_id": {
							Type:     schema.TypeString,
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
										Required: true,
									},
									"protocol": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"from_port": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"to_port": {
										Type:     schema.TypeInt,
										Optional: true,
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
										Optional: true,
									},
									"cidr": {
										Type:     schema.TypeString,
										Optional: true,
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

func dataSourceVPCsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("orgid").(string)
	if orgid == "" {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Organization ID (orgid) is required",
			Detail:   "Organization ID (orgid) must be provided",
		})
		return diags
	}
	authctx := getVPCAuthCtx(&pco)

	//request vpcs
	res, httpr, err := pco.vpcclient.DefaultApi.OrganizationsOrgIdVpcsGet(authctx, orgid).Execute()

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
			Summary:  "Unable to Get VPCs",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	//process data
	data := res.GetData()
	vpcs := flattenVPCsData(&data)
	//save in data source schema
	if err := d.Set("vpcs", vpcs); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set VPCs",
			Detail:   "Unable to set VPCs in resource schema",
		})
		return diags
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenVPCsData(vpcs *[]vpc.Vpc) []interface{} {
	if vpcs != nil && len(*vpcs) > 0 {
		res := make([]interface{}, len(*vpcs))

		for i, vpcitem := range *vpcs {
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

			res[i] = item
		}
		return res
	}

	return make([]interface{}, 0)
}
