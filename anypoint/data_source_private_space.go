package anypoint

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/private_space"
)

var PRIVATE_SPACE_SCHEMA = map[string]*schema.Schema{
	"id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The unique identifier of the private space.",
	},
	"name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The name of the private space.",
	},
	"org_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ID of the organization to which the private space belongs.",
	},
	"root_org_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ID of the root organization to which the private space belongs.",
	},
	"status": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The current status of the private space.",
	},
	"status_message": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The status message of the private space.",
	},
	"region": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The region of the private space.",
	},
	"provisioning_status": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Provisioning status.",
	},
	"provisioning_message": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Provisioning message.",
	},
	"environments_type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The type of associated environments.",
	},
	"environments_business_groups": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Business groups associated with the associated environments.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"network_region": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The network region of the private space.",
	},
	"network_cidr_block": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The CIDR block of the network.",
	},
	"network_internal_dns_servers": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "List of DNS servers.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"network_internal_dns_special_domains": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Special DNS domains.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"network_inbound_static_ips": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Inbound static IPs.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"network_outbound_static_ips": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Outbound static IPs.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"network_dns_target": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The DNS target.",
	},
	"network_internal_dns_target": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The internal DNS target.",
	},
	"network_reserved_cidrs": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Reserved CIDR blocks.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"firewall_rules": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Firewall rules for the private space.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cidr_block": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The CIDR block for the firewall rule.",
				},
				"protocol": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The protocol for the firewall rule.",
				},
				"from_port": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The starting port for the firewall rule.",
				},
				"to_port": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The ending port for the firewall rule.",
				},
				"type": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The type of the firewall rule.",
				},
			},
		},
	},
	"enable_iam_role": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Indicates whether IAM roles are enabled for the private space.",
	},
	"enable_egress": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Indicates whether egress is enabled for the private space.",
	},
	"enable_network_isolation": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Indicates whether network isolation is enabled for the private space.",
	},
	"mule_app_deployment_count": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "The number of Mule deployments in the private space.",
	},
	"days_left_for_relaxed_quota": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "The number of days left for relaxed quota.",
	},
	"vpc_migration_in_progress": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Indicates whether VPC migration is in progress for the private space.",
	},
}

func dataSourcePrivateSpace() *schema.Resource {
	ps_schema := cloneSchema(PRIVATE_SPACE_SCHEMA)
	ps_schema["org_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The master organization id where the private space is defined.",
	}
	ps_schema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The private space id.",
	}

	return &schema.Resource{
		ReadContext: dataSourcePrivateSpaceRead,
		Description: `
		Reads a ` + "`" + `private space` + "`" + ` in your business group.
		`,
		Schema: ps_schema,
	}
}

func dataSourcePrivateSpaceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	id := d.Get("id").(string)
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	//request
	res, httpr, err := pco.privatespaceclient.DefaultAPI.GetPrivateSpace(authctx, orgid, id).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Get Private Space for org " + orgid + " and id " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	private_space := flattenPrivateSpaceData(res)
	//save in data source schema
	if err := setPrivateSpaceAttributesToResourceData(d, private_space); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set Private Space " + id + " in org " + orgid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(id)
	return diags
}

func flattenPrivateSpaceData(item *private_space.PrivateSpace) map[string]any {
	result := make(map[string]any)
	result["id"] = item.GetId()
	result["name"] = item.GetName()
	result["org_id"] = item.GetOrganizationId()
	result["root_org_id"] = item.GetRootOrganizationId()
	result["status"] = item.GetStatus()
	result["status_message"] = item.GetStatusMessage()
	result["region"] = item.GetRegion()
	if val, ok := item.GetProvisioningOk(); ok {
		result["provisioning_status"] = val.GetStatus()
		result["provisioning_message"] = val.GetMessage()
	}
	if val, ok := item.GetEnvironmentsOk(); ok {
		result["environments_type"] = val.GetType()
		result["environments_business_groups"] = val.GetBusinessGroups()
	}
	if val, ok := item.GetNetworkOk(); ok {
		result["network_region"] = val.GetRegion()
		result["network_cidr_block"] = val.GetCidrBlock()
		if dns, ok := val.GetInternalDnsOk(); ok {
			result["network_internal_dns_servers"] = dns.GetDnsServers()
			result["network_internal_dns_special_domains"] = dns.GetSpecialDomains()
		}
		result["network_inbound_static_ips"] = val.GetInboundStaticIps()
		result["network_outbound_static_ips"] = val.GetOutboundStaticIps()
		result["network_dns_target"] = val.GetDnsTarget()
		result["network_internal_dns_target"] = val.GetInternalDnsTarget()
		result["network_reserved_cidrs"] = val.GetReservedCidrs()
	}
	result["firewall_rules"] = flattenPrivateSpaceFirewallRules(item.GetFirewallRules())
	result["enable_iam_role"] = item.GetEnableIAMRole()
	result["enable_egress"] = item.GetEnableEgress()
	result["enable_network_isolation"] = item.GetEnableNetworkIsolation()
	result["mule_app_deployment_count"] = item.GetMuleAppDeploymentCount()
	result["days_left_for_relaxed_quota"] = item.GetDaysLeftForRelaxedQuota()
	result["vpc_migration_in_progress"] = item.GetVpcMigrationInProgress()

	return result
}

func flattenPrivateSpaceFirewallRules(rules []private_space.FirewallRule) []map[string]any {
	result := make([]map[string]any, len(rules))
	for i, rule := range rules {
		result[i] = flattenPrivateSpaceFirewallRule(&rule)
	}
	return result
}

func flattenPrivateSpaceFirewallRule(rule *private_space.FirewallRule) map[string]any {
	result := make(map[string]any)
	result["cidr_block"] = rule.GetCidrBlock()
	result["protocol"] = rule.GetProtocol()
	result["from_port"] = rule.GetFromPort()
	result["to_port"] = rule.GetToPort()
	result["type"] = rule.GetType()
	return result
}

/*
* Copies the given private space instance into the given resource data
 */
func setPrivateSpaceAttributesToResourceData(d *schema.ResourceData, data map[string]any) error {
	attributes := getPrivateSpaceAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set flex gateway target attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getPrivateSpaceAttributes() []string {
	return getSchemaKeys(PRIVATE_SPACE_SCHEMA)
}
