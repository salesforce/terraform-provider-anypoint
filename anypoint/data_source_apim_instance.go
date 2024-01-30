package anypoint

import (
	"context"
	"fmt"
	"io"
	"maps"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/apim"
)

func dataSourceApimInstance() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApimInstanceRead,
		Description: `
		Read an API Manager Instance of any type.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Instance's unique id",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the flex gateway instance is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the flex gateway instance is defined.",
			},
			"audit": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The instance's auditing data",
			},
			"master_organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The root business group id",
			},
			"instance_label": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Instance's label",
			},
			"asset_group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The API specification's business group id",
			},
			"asset_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The API specification's asset id in exchange",
			},
			"asset_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The API specification's version number in exchange",
			},
			"product_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's asset major version number ",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the instance",
			},
			"tags": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of tags",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"order": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The order of this instance in the API Manager instances list",
			},
			"provider_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The client identity provider's id to use for this instance",
			},
			"deprecated": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the instance is deprecated",
			},
			"last_active_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date of last activity for this instance",
			},
			"endpoint_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The endpoint URI of this instance API",
			},
			"is_public": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "If this API is Public",
			},
			"technology": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of API Manager instance. Always equals to 'flexGateway'",
			},
			"endpoint_audit": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The instance's endpoint auditing data",
			},
			"endpoint_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The instance's endpoint id",
			},
			"endpoint_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The endpoint's specification type",
			},
			"endpoint_api_gateway_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint's api gateway version",
			},
			"endpoint_proxy_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint's Proxy URI",
			},
			"endpoint_proxy_registration_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint's Proxy registration URI",
			},
			"endpoint_last_active_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint's last active date",
			},
			"endpoint_deployment_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint's deployment type",
			},
			"endpoint_tls_inbound_context": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"secret_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The secret group id",
						},
						"tls_context_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The TLS context id in the given secret group",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The TLS context name",
						},
						"authorized": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The TLS context authorization status",
						},
						"audit": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The auditing data for tls context",
						},
					},
				},
			},
			"endpoint_api_version_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The API Manager Instance id",
			},
			"deployment_audit_created_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's deployment auditing creation date",
			},
			"deployment_audit_updated_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's deployment auditing update date",
			},
			"deployment_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The instance's deployment id",
			},
			"deployment_application_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's deployment application id",
			},
			"deployment_application_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's deployment application name ",
			},
			"deployment_gateway_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's deployment gateway version",
			},
			"deployment_environment_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's deployment environment name",
			},
			"deployment_environment_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's deployment environment id",
			},
			"deployment_target_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's deployment flex gateway target id",
			},
			"deployment_target_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's deployment flex gateway target name",
			},
			"deployment_updated_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's deployment update date",
			},
			"deployment_type": {
				Type:        schema.TypeString,
				Description: "The instance's deployment update date",
				Computed:    true,
			},
			"deployment_expected_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's deployment expected status",
			},
			"deployment_api_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The API Manager Instance id",
			},
			"routing": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The instance's routing mapping",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The route label",
						},
						"upstreams": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Upstreams for this particular route",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The id of the upstream",
									},
									"weight": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The weight of this particular upstream. All upstreams for a single route should add up to 100.",
									},
								},
							},
						},
						"rules": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Define methods, hosts and other settings for this route",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"methods": {
										Type:        schema.TypeSet,
										Computed:    true,
										Description: "List of request methods that should be matched for this route. Supported values are GET, PUT, DELETE, OPTIONS, POST, PATCH, HEAD, TRACE, CONNECT",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"host": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "A regular expression to match the request host",
									},
									"path": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "A regular expression to match the request path",
									},
									"headers": {
										Type:        schema.TypeMap,
										Computed:    true,
										Description: "Map of header names and values (regular expressions) that must be present in the request",
									},
								},
							},
						},
					},
				},
			},
			"upstreams": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of existing upstreams in this particular api instance",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"audit": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The upstream's auditing data",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The upstream's id",
						},
						"label": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The upstream's label",
						},
						"uri": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The upstream URI",
						},
						"tls_context": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"secret_group_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The secret group id",
									},
									"tls_context_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The TLS context id in the given secret group",
									},
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The TLS context name",
									},
									"authorized": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "The TLS context authorization status",
									},
									"audit": {
										Type:        schema.TypeMap,
										Computed:    true,
										Description: "The auditing data for tls context",
									},
								},
							},
						},
					},
				},
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The API Instance status",
			},
			"autodiscovery_instance_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's discovery name",
			},
		},
	}
}

func dataSourceApimInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	id := d.Get("id").(string)
	authctx := getApimAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.apimclient.DefaultApi.GetApimInstanceDetails(authctx, orgid, envid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get API manager instance",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// process data
	data := flattenApimInstanceDetails(res)
	if err := setApimInstanceDetailsAttributesToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set API Manager instance details attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	//read all upstreams
	diags = append(diags, readApimInstanceUpstreamsOnly(ctx, d, m)...)
	d.SetId(string(res.GetId()))
	return diags
}

// Read all upstreams for a the api manager instance
func readApimInstanceUpstreamsOnly(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	id := d.Get("id").(string)
	if isComposedResourceId(id) {
		orgid, envid, id = decomposeApimFlexGatewayId(d)
	}

	authctx := getApimUpstreamAuthCtx(ctx, &pco)
	res, httpr, err := pco.apimupstreamclient.DefaultApi.GetApimInstanceUpstreams(authctx, orgid, envid, id).Execute()
	defer httpr.Body.Close()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get API manager instance " + id + " upstreams",
			Detail:   details,
		})
		return diags
	}
	list := res.GetUpstreams()
	sortApimUpstreams(list)
	data := flattenApimUpstreamsResult(list)
	if err := d.Set("upstreams", data); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set upstreams of API manager instance " + id,
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

// flattens api manager instance details
func flattenApimInstanceDetails(details *apim.ApimInstanceDetails) map[string]interface{} {
	if details == nil {
		return nil
	}
	result := make(map[string]interface{})

	result["id"] = details.GetId()
	result["org_id"] = details.GetOrganizationId()
	result["env_id"] = details.GetEnvironmentId()
	if val, ok := details.GetAuditOk(); ok {
		result["audit"] = flattenApimAudit(val)
	}
	result["master_organization_id"] = details.GetMasterOrganizationId()
	result["instance_label"] = details.GetInstanceLabel()
	if val, ok := details.GetGroupIdOk(); ok {
		result["asset_group_id"] = *val
	}
	if val, ok := details.GetAssetIdOk(); ok && val != nil {
		result["asset_id"] = *val
	}
	if val, ok := details.GetAssetVersionOk(); ok && val != nil {
		result["asset_version"] = *val
	}
	if val, ok := details.GetProductVersionOk(); ok && val != nil {
		result["product_version"] = *val
	}
	if val, ok := details.GetDescriptionOk(); ok && val != nil {
		result["description"] = *val
	}
	if val, ok := details.GetTagsOk(); ok {
		result["tags"] = val
	}
	if val, ok := details.GetOrderOk(); ok {
		result["order"] = *val
	}
	if val, ok := details.GetProviderIdOk(); ok {
		result["provider_id"] = val
	}
	if val, ok := details.GetDeprecatedOk(); ok {
		result["deprecated"] = *val
	}
	if val, ok := details.GetLastActiveDateOk(); ok && val != nil {
		result["last_active_date"] = *val
	}
	if val, ok := details.GetEndpointUriOk(); ok && val != nil {
		result["endpoint_uri"] = *val
	}
	if val, ok := details.GetIsPublicOk(); ok && val != nil {
		result["is_public"] = *val
	}
	if val, ok := details.GetTechnologyOk(); ok {
		result["technology"] = *val
	}
	if val, ok := details.GetEndpointOk(); ok && val != nil {
		endpoint := flattenApimEndpoint(val)
		maps.Copy(result, endpoint)
	}
	if val, ok := details.GetDeploymentOk(); ok && val != nil {
		deployment := flattenApimDeployment(val)
		maps.Copy(result, deployment)
	}
	if val, ok := details.GetRoutingOk(); ok && val != nil {
		result["routing"] = flattenApimRoutingCollection(val)
	}
	if val, ok := details.GetStatusOk(); ok && val != nil {
		result["status"] = *val
	}
	if val, ok := details.GetAutodiscoveryInstanceNameOk(); ok {
		result["autodiscovery_instance_name"] = *val
	}

	return result
}

// flattens endpoint object of api manager instances
func flattenApimEndpoint(endpoint *apim.Endpoint) map[string]interface{} {
	if endpoint == nil {
		return nil
	}
	result := make(map[string]interface{})
	if val, ok := endpoint.GetAuditOk(); ok && val != nil {
		result["endpoint_audit"] = flattenApimAudit(val)
	}
	if val, ok := endpoint.GetIdOk(); ok {
		result["endpoint_id"] = *val
	}
	if val, ok := endpoint.GetTypeOk(); ok {
		result["endpoint_type"] = *val
	}
	if val, ok := endpoint.GetApiGatewayVersionOk(); ok && val != nil {
		result["endpoint_api_gateway_version"] = *val
	}
	if val, ok := endpoint.GetProxyUriOk(); ok && val != nil {
		result["endpoint_proxy_uri"] = *val
	}
	if val, ok := endpoint.GetProxyRegistrationUriOk(); ok && val != nil {
		result["endpoint_proxy_registration_uri"] = *val
	}
	if val, ok := endpoint.GetLastActiveDateOk(); ok && val != nil {
		result["endpoint_last_active_date"] = *val
	}
	if val, ok := endpoint.GetDeploymentTypeOk(); ok && val != nil {
		result["endpoint_deployment_type"] = *val
	}
	if val, ok := endpoint.GetTlsContextsOk(); ok {
		result["endpoint_tls_inbound_context"] = []map[string]interface{}{flattenApimEndpointTlsContext(val)}
	}
	if val, ok := endpoint.GetApiVersionIdOk(); ok && val != nil {
		result["endpoint_api_version_id"] = *val
	}
	return result
}

// flattens tls context of endpoints in any api manager
func flattenApimEndpointTlsContext(tc *apim.EndpointTlsContexts) map[string]interface{} {
	if tc == nil {
		return nil
	}
	result := make(map[string]interface{})
	if inbound, ok := tc.GetInboundOk(); ok {
		if val, ok := inbound.GetSecretGroupIdOk(); ok {
			result["secret_group_id"] = *val
		}
		if val, ok := inbound.GetTlsContextIdOk(); ok {
			result["tls_context_id"] = *val
		}
		if val, ok := inbound.GetNameOk(); ok {
			result["name"] = *val
		}
		if val, ok := inbound.GetAuthorizedOk(); ok {
			result["authorized"] = *val
		}
	}
	if val, ok := tc.GetAuditOk(); ok {
		result["audit"] = flattenApimAudit(val)
	}
	return result
}

// flattens deployment object of any api manager instance
func flattenApimDeployment(deployment *apim.Deployment) map[string]interface{} {
	result := make(map[string]interface{})
	if val, ok := deployment.GetAuditOk(); ok {
		audit := flattenApimAudit(val)
		if created, ok := audit["created"]; ok {
			result["deployment_audit_created_date"] = created
		}
		if updated, ok := audit["updated"]; ok {
			result["deployment_audit_updated_date"] = updated
		}
	}
	if val, ok := deployment.GetDeploymentIdOk(); ok && val != nil {
		result["deployment_id"] = *val
	}
	if val, ok := deployment.GetApplicationIdOk(); ok && val != nil {
		result["deployment_application_id"] = *val
	}
	if val, ok := deployment.GetApplicationNameOk(); ok && val != nil {
		result["deployment_application_name"] = *val
	}
	if val, ok := deployment.GetGatewayVersionOk(); ok && val != nil {
		result["deployment_gateway_version"] = *val
	}
	if val, ok := deployment.GetEnvironmentNameOk(); ok && val != nil {
		result["deployment_environment_name"] = *val
	}
	if val, ok := deployment.GetEnvironmentIdOk(); ok && val != nil {
		result["deployment_environment_id"] = *val
	}
	if val, ok := deployment.GetTargetIdOk(); ok && val != nil {
		result["deployment_target_id"] = *val
	}
	if val, ok := deployment.GetTargetNameOk(); ok && val != nil {
		result["deployment_target_name"] = *val
	}
	if val, ok := deployment.GetUpdatedDateOk(); ok && val != nil {
		result["deployment_updated_date"] = *val
	}
	if val, ok := deployment.GetTypeOk(); ok && val != nil {
		result["deployment_type"] = *val
	}
	if val, ok := deployment.GetExpectedStatusOk(); ok && val != nil {
		result["deployment_expected_status"] = *val
	}
	if val, ok := deployment.GetApiIdOk(); ok && val != nil {
		result["deployment_api_id"] = *val
	}
	return result
}

// flattens list of routing of any api manager instance
func flattenApimRoutingCollection(routings []apim.Routing) []map[string]interface{} {
	sortApimRouting(routings)
	result := make([]map[string]interface{}, len(routings))
	for i, routing := range routings {
		result[i] = flattenApimRouting(&routing)
	}
	return result
}

// flattens the routing object of any api manager instance
func flattenApimRouting(routing *apim.Routing) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := routing.GetLabelOk(); ok && val != nil {
		item["label"] = *val
	}
	if upstreams, ok := routing.GetUpstreamsOk(); ok && upstreams != nil {
		set := make([]map[string]interface{}, len(upstreams))
		for i, u := range upstreams {
			upstream := make(map[string]interface{})
			if val, ok := u.GetIdOk(); ok {
				upstream["id"] = *val
			}
			if val, ok := u.GetWeightOk(); ok {
				upstream["weight"] = *val
			}
			set[i] = upstream
		}
		item["upstreams"] = set
	}
	if rules, ok := routing.GetRulesOk(); ok && rules != nil {
		r := make(map[string]interface{})
		if val, ok := rules.GetMethodsOk(); ok && val != nil {
			slice := strings.Split(*val, "|")
			sort.Strings(slice)
			r["methods"] = slice
		}
		if val, ok := rules.GetHostOk(); ok && val != nil {
			r["host"] = *val
		}
		if val, ok := rules.GetPathOk(); ok && val != nil {
			r["path"] = *val
		}
		if val, ok := rules.GetHeadersOk(); ok && val != nil {
			r["headers"] = val
		}
		item["rules"] = []interface{}{r}
	}
	return item
}

func flattenApimAudit(audit *apim.Audit) map[string]interface{} {
	result := make(map[string]interface{})
	if audit == nil {
		return result
	}
	if created, ok := audit.GetCreatedOk(); ok && created != nil {
		if val, ok := created.GetDateOk(); ok && val != nil {
			result["created"] = *val
		}
	}
	if updated, ok := audit.GetUpdatedOk(); ok && updated != nil {
		if val, ok := updated.GetDateOk(); ok && updated != nil {
			result["updated"] = *val
		}
	}
	return result
}

func setApimInstanceDetailsAttributesToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getApimInstanceDetailsAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set api manager instance attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getApimInstanceDetailsAttributes() []string {
	attributes := [...]string{
		"org_id", "env_id", "audit", "master_organization_id", "instance_label", "asset_group_id",
		"asset_id", "asset_version", "product_version", "description", "tags", "order", "provider_id",
		"deprecated", "last_active_date", "endpoint_uri", "is_public", "technology", "endpoint_audit",
		"endpoint_id", "endpoint_type", "endpoint_api_gateway_version", "endpoint_proxy_uri",
		"endpoint_proxy_registration_uri", "endpoint_last_active_date", "endpoint_deployment_type",
		"endpoint_tls_inbound_context", "endpoint_api_version_id", "deployment_audit_created_date", "deployment_audit_updated_date", "deployment_id",
		"deployment_application_id", "deployment_application_name", "deployment_gateway_version",
		"deployment_environment_name", "deployment_environment_id", "deployment_target_id",
		"deployment_target_name", "deployment_updated_date", "deployment_type", "deployment_expected_status",
		"deployment_api_id", "routing", "status", "autodiscovery_instance_name",
	}
	return attributes[:]
}

// sorts list of api routing by their label
func sortApimRouting(list []apim.Routing) {
	sort.SliceStable(list, func(i, j int) bool {
		i_label := list[i].GetLabel()
		j_label := list[j].GetLabel()
		return i_label < j_label
	})
}
