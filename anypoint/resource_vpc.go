package anypoint

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
			"orgid": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
	orgid := d.Get("orgid").(string)

	authctx := getVPCAuthCtx(&pco)

	body := newVPCBody(d)

	//request vpc creation
	res, httpr, err := pco.vpcclient.DefaultApi.OrganizationsOrgIdVpcsPost(authctx, orgid).VpcCore(*body).Execute()
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
	orgid := d.Get("orgid").(string)

	authctx := getVPCAuthCtx(&pco)

	res, httpr, err := pco.vpcclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdGet(authctx, orgid, vpcid).Execute()
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
	orgid := d.Get("orgid").(string)

	authctx := getVPCAuthCtx(&pco)

	if d.HasChanges(getVPCCoreAttributes()...) {
		body := newVPCBody(d)
		//request vpc creation
		_, httpr, err := pco.vpcclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdPut(authctx, orgid, vpcid).VpcCore(*body).Execute()
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
	orgid := d.Get("orgid").(string)

	authctx := getVPCAuthCtx(&pco)

	httpr, err := pco.vpcclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdDelete(authctx, orgid, vpcid).Execute()
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
 */
func newVPCBody(d *schema.ResourceData) *vpc.VpcCore {
	body := vpc.NewVpcCoreWithDefaults()

	body.SetName(d.Get("name").(string))
	body.SetRegion(d.Get("region").(string))
	body.SetCidrBlock(d.Get("cidr_block").(string))
	body.SetIsDefault(d.Get("is_default").(bool))
	body.SetOwnerId(d.Get("owner_id").(string))

	//preparing shared with list
	sw := d.Get("shared_with").([]interface{})
	shared_with := make([]string, len(sw))
	for index, e := range sw {
		shared_with[index] = e.(string)
	}
	body.SetSharedWith(shared_with)

	//preparing associated environments list
	aes := d.Get("associated_environments").([]interface{})
	associated_environments := make([]string, len(aes))
	for index, ae := range aes {
		associated_environments[index] = ae.(string)
	}
	body.SetAssociatedEnvironments(associated_environments)

	//preparing internal_dns structure
	idss := d.Get("internal_dns_servers").([]interface{})
	dns_servers := make([]string, len(idss))
	for index, dns_server := range idss {
		dns_servers[index] = dns_server.(string)
	}
	idsds := d.Get("internal_dns_special_domains").([]interface{})
	special_domains := make([]string, len(idsds))
	for index, special_domain := range idsds {
		special_domains[index] = special_domain.(string)
	}
	body.SetInternalDns(*vpc.NewInternalDns(dns_servers, special_domains))

	//preparing firewall rules
	orules := d.Get("firewall_rules").([]interface{})
	frules := make([]vpc.FirewallRule, len(orules))
	for index, rule := range orules {
		frules[index] = *vpc.NewFirewallRule(rule.(map[string]interface{})["cidr_block"].(string), rule.(map[string]interface{})["from_port"].(int32), rule.(map[string]interface{})["protocol"].(string), rule.(map[string]interface{})["to_port"].(int32))
	}
	body.SetFirewallRules(frules)

	//preparing vpc routes
	oroutes := d.Get("vpc_routes").([]interface{})
	vpcroutes := make([]vpc.VpcRoute, len(orules))
	for index, route := range oroutes {
		vpcroutes[index] = *vpc.NewVpcRoute(route.(map[string]interface{})["cidr"].(string), route.(map[string]interface{})["next_hop"].(string))
	}
	body.SetVpcRoutes(vpcroutes)

	return body
}

/*
 * Returns authentication context (includes authorization header)
 */
func getVPCAuthCtx(pco *ProviderConfOutput) context.Context {
	ctxbckgrnd := context.Background()
	return context.WithValue(ctxbckgrnd, vpc.ContextAccessToken, pco.access_token)
}
