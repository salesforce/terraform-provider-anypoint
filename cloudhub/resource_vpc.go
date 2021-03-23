package cloudhub

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	vpc "github.com/mulesoft-consulting/cloudhub-client-go/vpc"
)

func resourceVPC() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCCreate,
		ReadContext:   resourceVPCRead,
		UpdateContext: resourceVPCUpdate,
		DeleteContext: resourceVPCDelete,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cidr_block": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				Default:  false,
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
				Optional: true,
				ForceNew: true,
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
							Required: true,
						},
						"from_port": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"to_port": {
							Type:     schema.TypeInt,
							Required: true,
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
							Required: true,
						},
						"cidr": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceVPCCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	body := newVPCBody(d)

	//request vpc creation
	res, httpr, err := pco.vpcclient.DefaultApi.OrganizationsOrgIdVpcsPost(pco.authctx, pco.org_id).VpcCore(*body).Execute()
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
			Summary:  "Unable to Create VPC",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(res.GetId())

	resourceVPCRead(ctx, d, m)

	return diags
}

func resourceVPCRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	vpcid := d.Id()

	res, httpr, err := pco.vpcclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdGet(pco.authctx, pco.org_id, vpcid).Execute()
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

	return diags
}

func resourceVPCUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	vpcid := d.Id()

	if d.HasChanges(getVPCCoreAttributes()...) {
		body := newVPCBody(d)
		//request vpc creation
		_, httpr, err := pco.vpcclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdPut(pco.authctx, pco.org_id, vpcid).VpcCore(*body).Execute()
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
				Summary:  "Unable to Update VPC",
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceVPCRead(ctx, d, m)
}

func resourceVPCDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	vpcid := d.Id()

	httpr, err := pco.vpcclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdDelete(pco.authctx, pco.org_id, vpcid).Execute()
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
			Summary:  "Unable to Delete VPC",
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
* Creates a new VPC Core Struct from the resource data schema
* @param d *schema.ResourceData
* @return vpcCore object
 */
func newVPCBody(d *schema.ResourceData) *vpc.VpcCore {
	body := vpc.NewVpcCoreWithDefaults()

	body.SetName(d.Get("name").(string))
	body.SetRegion(d.Get("region").(string))
	body.SetCidrBlock(d.Get("cidr_block").(string))
	body.SetInternalDns(*vpc.NewInternalDns(d.Get("internal_dns_servers").([]string), d.Get("internal_dns_special_domains").([]string)))
	body.SetIsDefault(d.Get("is_default").(bool))
	body.SetAssociatedEnvironments(d.Get("associated_environments").([]string))
	body.SetOwnerId(d.Get("owner_id").(string))
	body.SetSharedWith(d.Get("shared_with").([]string))

	orules := d.Get("firewall_rules").([]map[string]interface{})
	frules := make([]vpc.FirewallRule, len(orules))
	for index, rule := range orules {
		frules[index] = *vpc.NewFirewallRule(rule["cidr_block"].(string), rule["from_port"].(int32), rule["protocol"].(string), rule["to_port"].(int32))
	}
	body.SetFirewallRules(frules)

	oroutes := d.Get("vpc_routes").([]map[string]interface{})
	vpcroutes := make([]vpc.VpcRoute, len(orules))
	for index, route := range oroutes {
		vpcroutes[index] = *vpc.NewVpcRoute(route["cidr"].(string), route["next_hop"].(string))
	}
	body.SetVpcRoutes(vpcroutes)

	return body
}
