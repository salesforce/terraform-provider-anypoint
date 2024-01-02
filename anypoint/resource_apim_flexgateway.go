package anypoint

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mulesoft-anypoint/anypoint-client-go/apim"
	"github.com/mulesoft-anypoint/anypoint-client-go/apim_upstream"
)

const FLEX_GATEWAY_TECHNOLOGY = "flexGateway"

func resourceApimFlexGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApimFlexGatewayCreate,
		ReadContext:   resourceApimFlexGatewayRead,
		UpdateContext: resourceApimFlexGatewayUpdate,
		DeleteContext: resourceApimFlexGatewayDelete,
		Description: `
		Create an API Manager Instance of type Flex Gateway.
		`,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The last time this resource has been updated locally.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Instance's unique id",
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
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organization id where the flex gateway instance is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The environment id where the flex gateway instance is defined.",
			},
			"apim_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The API Manager Flex Gateway instance id.",
			},
			"instance_label": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The Instance's label",
			},
			"asset_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The API specification's business group id",
			},
			"asset_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The API specification's asset id in exchange",
			},
			"asset_version": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The API specification's version number in exchange",
			},
			"product_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's asset major version number ",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the instance",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
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
				Optional:    true,
				Default:     nil,
				Description: "The client identity provider's id to use for this instance",
			},
			"deprecated": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "True if the instance is deprecated",
			},
			"last_active_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date of last activity for this instance",
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
			"endpoint_uri": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The endpoint URI of this instance API",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
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
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Endpoint's Proxy URI",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
			},
			"endpoint_proxy_registration_uri": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Endpoint's Proxy registration URI",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
			},
			"endpoint_last_active_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint's last active date",
			},
			"endpoint_deployment_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "HY",
				Description: "Endpoint's deployment type",
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice(
						[]string{"CH", "HY", "RF", "SM", "CH2"},
						false,
					),
				),
			},
			"endpoint_tls_inbound_context": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"secret_group_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The secret group id",
						},
						"tls_context_id": {
							Type:        schema.TypeString,
							Required:    true,
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
			"deployment_audit": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The instance's deployment auditing data",
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
				Optional:    true,
				Default:     "1.0.0",
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
				Required:    true,
				Description: "The instance's deployment flex gateway target id",
			},
			"deployment_target_name": {
				Type:        schema.TypeString,
				Required:    true,
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
				Optional:    true,
				Default:     "HY",
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice(
						[]string{"CH", "HY", "RF", "SM", "CH2"},
						false,
					),
				),
			},
			"deployment_overwrite": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "The API Manager Instance id",
			},
			"deployment_expected_status": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "deployed",
				Description: "The instance's deployment expected status. \"deployed\" or \"undeployed\"",
			},
			"deployment_api_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The API Manager Instance id",
			},
			"routing": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The instance's routing mapping",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The route label",
						},
						"upstreams": {
							Type:         schema.TypeSet,
							Required:     true,
							Description:  "Upstreams for this particular route",
							RequiredWith: []string{"upstreams"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"index": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "The index of the upstream from the list of declared upstreams. Index starts at 0",
									},
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The id of the upstream",
									},
									"weight": {
										Type:             schema.TypeInt,
										Required:         true,
										Description:      "The weight of this particular upstream. All upstreams for a single route should add up to 100.",
										ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtMost(100)),
									},
								},
							},
						},
						"rules": {
							Type:        schema.TypeSet,
							Optional:    true,
							MaxItems:    1,
							Description: "Define methods, hosts and other settings for this route",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"methods": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "List of request methods that should be matched for this route. Supported values are GET, PUT, DELETE, OPTIONS, POST, PATCH, HEAD, TRACE, CONNECT",
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateDiagFunc: validation.ToDiagFunc(
												validation.StringInSlice(
													[]string{"GET", "PUT", "DELETE", "OPTIONS", "POST", "PATCH", "HEAD", "TRACE", "CONNECT"},
													false,
												),
											),
										},
									},
									"host": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "",
										Description: "A regular expression to match the request host",
									},
									"path": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "",
										Description: "A regular expression to match the request path",
									},
									"headers": {
										Type:        schema.TypeMap,
										Optional:    true,
										Default:     map[string]string{},
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
				Optional:    true,
				Description: "The list of upstreams to be created for this particular api instance",
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
							Required:    true,
							Description: "The upstream's label",
						},
						"uri": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "The upstram URI",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
						},
						"tls_context": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"secret_group_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The secret group id",
									},
									"tls_context_id": {
										Type:        schema.TypeString,
										Required:    true,
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

func resourceApimFlexGatewayCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	authctx := getApimAuthCtx(ctx, &pco)
	body := newApimFlexGatewayPostBody(d)

	// Validate Routing Upstreams before acting
	if err := validateRoutingUpstreams(d); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Flex Gateway of org " + orgid + " and env " + envid,
			Detail:   err.Error(),
		})
		return diags
	}

	res, httpr, err := pco.apimclient.DefaultApi.PostApimInstance(authctx, orgid, envid).ApimInstancePostBody(*body).Execute()
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
			Summary:  "Unable to create Flex Gateway of org " + orgid + " and env " + envid,
			Detail:   details,
		})
		return diags
	}

	id := string(res.GetId())
	d.SetId(ComposeResourceId([]string{orgid, envid, id}))
	d.Set("apim_id", id)

	resourceApimFlexGatewayRead(ctx, d, m)

	_, err = apimFlexGatewayUpstreamsCreate(ctx, d, m)
	if err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Flex Gateway upstreams for instance " + id,
			Detail:   err.Error(),
		})
		return diags
	}

	return resourceApimFlexGatewayRoutingUpdate(ctx, d, m)
}

// Create upstreams for the apim flex gateway instance if upstreams is set
func apimFlexGatewayUpstreamsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) ([]interface{}, error) {
	if input, ok := d.GetOk("upstreams"); ok {
		pco := m.(ProviderConfOutput)
		orgid := d.Get("org_id").(string)
		envid := d.Get("env_id").(string)
		id := d.Id()
		authctx := getApimUpstreamAuthCtx(ctx, &pco)
		bodies := newApimFlexGatewayUpstreamPostBody(input.([]interface{}))
		upstreams := make([]apim_upstream.UpstreamDetails, len(bodies))
		for i, body := range bodies {
			res, httpr, err := pco.apimupstreamclient.DefaultApi.PostApimInstanceUpstream(authctx, orgid, envid, id).UpstreamPostBody(*body).Execute()
			defer httpr.Body.Close()
			if err != nil {
				var details error
				if httpr != nil {
					b, _ := io.ReadAll(httpr.Body)
					details = fmt.Errorf(string(b))
				} else {
					details = err
				}
				return nil, details
			}
			upstreams[i] = res
		}
		result := flattenApimUpstreamsResult(upstreams)
		if err := d.Set("upstreams", result); err != nil {
			return result, err
		}
		return result, nil
	}
	return []interface{}{}, nil
}

func resourceApimFlexGatewayRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, envid, id := decomposeApimFlexGatewayId(d)
	authctx := getApimAuthCtx(ctx, &pco)

	res, httpr, err := pco.apimclient.DefaultApi.GetApimInstanceDetails(authctx, orgid, envid, id).Execute()
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
			Summary:  "Unable to get API manager's flex gateway instance",
			Detail:   details,
		})
		return diags
	}

	if diagsbis := readApimInstanceUpstreamsOnly(ctx, d, m); diagsbis.HasError() {
		diags = append(diags, diagsbis...)
	}
	details := flattenApimInstanceDetails(&res)
	if err := setApimFlexGatewayAttributesToResourceData(d, details); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set API manager's flex gateway instance details attributes",
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func resourceApimFlexGatewayUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	//todo: finish

	return diags
}

func resourceApimFlexGatewayRoutingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, envid, id := decomposeApimFlexGatewayId(d)
	authctx := getApimAuthCtx(ctx, &pco)

	body := newApimFlexGatewayPatchBody(d)

	if _, ok := d.GetOk("routing"); ok {
		_, httpr, err := pco.apimclient.DefaultApi.PatchApimInstance(authctx, orgid, envid, id).ApimInstancePatchBody(*body).Execute()
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
				Summary:  "unable to update api manager's flex gateway instance with routing parameters ",
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()

		d.Set("last_updated", time.Now().Format(time.RFC850))
		return resourceApimFlexGatewayRead(ctx, d, m)
	}

	return diags
}

func resourceApimFlexGatewayDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, envid, id := decomposeApimFlexGatewayId(d)
	authctx := getApimAuthCtx(ctx, &pco)

	httpr, err := pco.apimclient.DefaultApi.DeleteApimInstance(authctx, orgid, envid, id).Execute()
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
			Summary:  "Unable to Delete API Manager's Flex Gateway Instance",
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

func newApimFlexGatewayPostBody(d *schema.ResourceData) *apim.ApimInstancePostBody {
	body := apim.NewApimInstancePostBody()
	endpoint := newApimFlexGatewayEndpointPostBody(d)
	deployment := newApimFlexGatewayDeploymentPostBody(d)
	spec := newApimFlexGatewaySpecPostBody(d)
	//routing := newApimFlexGatewayRoutingPostBody(d)

	body.SetTechnology(FLEX_GATEWAY_TECHNOLOGY)
	body.SetInstanceLabel(d.Get("instance_label").(string))
	body.SetEndpoint(*endpoint)
	body.SetDeployment(*deployment)
	body.SetSpec(*spec)
	//body.SetRouting(routing)

	return body
}

func newApimFlexGatewayEndpointPostBody(d *schema.ResourceData) *apim.EndpointPostBody {
	body := apim.NewEndpointPostBody()

	if val, ok := d.GetOk("endpoint_uri"); ok {
		body.SetUri(val.(string))
	}
	if val, ok := d.GetOk("endpoint_proxy_uri"); ok {
		body.SetProxyUri(val.(string))
	} else {
		body.SetProxyUriNil()
	}
	if val, ok := d.GetOk("endpoint_proxy_registration_uri"); ok {
		body.SetProxyRegistrationUri(val.(string))
	} else {
		body.SetProxyRegistrationUriNil()
	}
	if val, ok := d.GetOk("endpoint_deployment_type"); ok {
		body.SetDeploymentType(val.(string))
	}
	if _, ok := d.GetOk("endpoint_tls_inbound_context"); ok {
		tlscontext := newApimFlexGatewayEndpointTlsContextPostBody(d)
		body.SetTlsContexts(*tlscontext)
	} else {
		body.SetTlsContextsNil()
	}

	return body
}

func newApimFlexGatewayEndpointTlsContextPostBody(d *schema.ResourceData) *apim.EndpointPostBodyTlsContexts {
	body := apim.NewEndpointPostBodyTlsContexts()

	if val, ok := d.GetOk("endpoint_tls_inbound_context"); ok {
		set := val.(*schema.Set)
		list := set.List()
		if len(list) > 0 {
			inbound := apim.NewEndpointPostBodyTlsContextsInbound()
			item := list[0].(map[string]interface{})
			if sgi, ok := item["secret_group_id"]; ok {
				inbound.SetSecretGroupId(sgi.(string))
			}
			if tci, ok := item["tls_context_id"]; ok {
				inbound.SetTlsContextId(tci.(string))
			}
			body.SetInbound(*inbound)
		} else {
			body.SetInboundNil()
		}
	} else {
		body.SetInboundNil()
	}

	return body
}

func newApimFlexGatewayDeploymentPostBody(d *schema.ResourceData) *apim.DeploymentPostBody {
	body := apim.NewDeploymentPostBody()

	if val, ok := d.GetOk("deployment_target_id"); ok {
		body.SetTargetId(val.(string))
	}
	if val, ok := d.GetOk("deployment_target_name"); ok {
		body.SetTargetName(val.(string))
	}
	if val, ok := d.GetOk("deployment_target_name"); ok {
		body.SetTargetName(val.(string))
	}
	if val, ok := d.GetOk("deployment_type"); ok {
		body.SetType(val.(string))
	}
	body.SetEnvironmentId(d.Get("env_id").(string))
	body.SetOverwrite(d.Get("deployment_overwrite").(bool))
	body.SetGatewayVersion(d.Get("deployment_gateway_version").(string))
	body.SetExpectedStatus(d.Get("deployment_expected_status").(string))

	return body
}

func newApimFlexGatewaySpecPostBody(d *schema.ResourceData) *apim.Spec {
	body := apim.NewSpec()

	if val, ok := d.GetOk("asset_group_id"); ok {
		body.SetGroupId(val.(string))
	}
	if val, ok := d.GetOk("asset_id"); ok {
		body.SetAssetId(val.(string))
	}
	if val, ok := d.GetOk("asset_version"); ok {
		body.SetVersion(val.(string))
	}

	return body
}

func newApimFlexGatewayRoutingPostBody(d *schema.ResourceData) []map[string]interface{} {
	upstreamsdata := d.Get("upstreams")
	upstreams := upstreamsdata.([]interface{})
	if routings, ok := d.GetOk("routing"); ok {
		routingslist := routings.([]interface{})
		body := make([]map[string]interface{}, len(routingslist))
		for i, routingsitem := range routingslist {
			routinginput := routingsitem.(map[string]interface{})
			routingoutput := make(map[string]interface{})
			if val, ok := routinginput["label"]; ok {
				routingoutput["label"] = val
			}
			if upstreaminput, ok := routinginput["upstreams"]; ok {
				upstreaminputset := upstreaminput.(*schema.Set)
				upstreaminputlist := upstreaminputset.List()
				upstreamsoutput := make([]map[string]interface{}, len(upstreaminputlist))
				for j, upstreamitem := range upstreaminputlist {
					record := make(map[string]interface{})
					upstream := upstreamitem.(map[string]interface{})
					if v, ok := upstream["weight"]; ok {
						record["weight"] = v
					}
					if v, ok := upstream["index"]; ok {
						u := upstreams[v.(int32)]
						record["id"] = u
					}
					upstreamsoutput[j] = record
				}
				routingoutput["upstreams"] = upstreamsoutput
			}
			if val, ok := routinginput["rules"]; ok {
				routingoutput["rules"] = val
			}
			body[i] = routingoutput
		}
		return body
	} else {
		return nil
	}
}

func newApimFlexGatewayUpstreamPostBody(upstreams []interface{}) []*apim_upstream.UpstreamPostBody {
	if len(upstreams) == 0 {
		return nil
	}
	bodies := make([]*apim_upstream.UpstreamPostBody, len(upstreams))
	for i, item := range upstreams {
		upstream := item.(map[string]interface{})
		bodies[i] = apim_upstream.NewUpstreamPostBody()
		if val, ok := upstream["label"]; ok {
			bodies[i].SetLabel(val.(string))
		}
		if val, ok := upstream["uri"]; ok {
			bodies[i].SetUri(val.(string))
		}
		if val, ok := upstream["tls_context"]; ok {
			set := val.(*schema.Set)
			tlscontextinput := set.List()[0].(map[string]interface{})
			tlcbody := apim_upstream.NewUpstreamPostBodyTlsContext()
			if v, ok := tlscontextinput["secret_group_id"]; ok {
				tlcbody.SetSecretGroupId(v.(string))
			}
			if v, ok := tlscontextinput["tls_context_id"]; ok {
				tlcbody.SetTlsContextId(v.(string))
			}
			bodies[i].SetTlsContext(*tlcbody)
		} else {
			bodies[i].SetTlsContextNil()
		}
	}
	return bodies
}

func newApimFlexGatewayPatchBody(d *schema.ResourceData) *apim.ApimInstancePatchBody {
	body := apim.NewApimInstancePatchBody()
	endpoint := newApimFlexGatewayEndpointPostBody(d)
	deployment := newApimFlexGatewayDeploymentPostBody(d)
	spec := newApimFlexGatewaySpecPostBody(d)
	routing := newApimFlexGatewayRoutingPostBody(d)

	body.SetTechnology(FLEX_GATEWAY_TECHNOLOGY)
	body.SetInstanceLabel(d.Get("instance_label").(string))
	body.SetEndpoint(*endpoint)
	body.SetDeployment(*deployment)
	body.SetSpec(*spec)
	body.SetRouting(routing)

	return body
}

func decomposeApimFlexGatewayId(d *schema.ResourceData) (string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2]
}

func setApimFlexGatewayAttributesToResourceData(d *schema.ResourceData, details map[string]interface{}) error {
	attributes := getApimInstanceDetailsAttributes()
	if details != nil {
		for _, attr := range attributes {
			if attr == "routing" {
				routing := details[attr]
				result, err := mergeRoutingDetails2ResourceData(d, routing.([]interface{}))
				if err != nil {
					return err
				}
				if err := d.Set(attr, result); err != nil {
					return fmt.Errorf("unable to set flex gateway instance attribute %s\n details: %s", attr, err)
				}
				continue
			}
			if err := d.Set(attr, details[attr]); err != nil {
				return fmt.Errorf("unable to set api manager instance attribute %s\n details: %s", attr, err)
			}
		}
	}
	return nil
}

// Merge routingDetails coming from remote read operation with existing resource data for the case of routing.
func mergeRoutingDetails2ResourceData(d *schema.ResourceData, routingdetails []interface{}) ([]map[string]interface{}, error) {
	dict, err := geApimUpstreamsDictionary(d)
	if err != nil {
		return nil, err
	}

	//Merging routing-details (coming from remote) with resourceData in a new list
	result := make([]map[string]interface{}, len(routingdetails))
	for i, rdetails := range routingdetails {
		rdetailsmap := rdetails.(map[string]interface{})
		rdetailsupstreams := rdetailsmap["upstreams"]
		label := rdetailsmap["label"]
		rdetailsupstreamslist := rdetailsupstreams.([]interface{})
		rupstreamresult := make([]interface{}, len(rdetailsupstreamslist))
		for _, rdetailsupstream := range rdetailsupstreamslist {
			rdetailsupstreammap := rdetailsupstream.(map[string]interface{})
			id := rdetailsupstreammap["id"]
			index := dict[id.(string)]
			rdetailsupstreammap["index"] = index
			rupstreamresult[i] = rdetailsupstreammap
		}
		result[i] = map[string]interface{}{
			"label":     label,
			"upstreams": rupstreamresult,
		}
		if rules, ok := rdetailsmap["rules"]; ok {
			result[i]["rules"] = rules
		}
	}
	return result, nil
}

// returns a dictionary of upstreams ids and their indexes in the list
func geApimUpstreamsDictionary(d *schema.ResourceData) (map[string]int32, error) {
	var upstreams []interface{}
	var routings []interface{}
	dict := make(map[string]int32)
	if val, ok := d.GetOk("upstreams"); ok {
		upstreams = val.([]interface{})
	}
	if val, ok := d.GetOk("routing"); ok {
		routings = val.([]interface{})
	} else {
		return nil, nil
	}
	// creating dictionary of upstreams and routing-upstreams from the resourceData
	for _, routingitem := range routings {
		ritem := routingitem.(map[string]interface{})
		label := ritem["label"]
		if rupstreams, ok := ritem["upstreams"]; ok {
			set := rupstreams.(*schema.Set)
			rupstreamlist := set.List()
			for _, rupstreamitem := range rupstreamlist {
				upitem := rupstreamitem.(map[string]interface{})
				if val, ok := upitem["index"]; ok {
					index := val.(int32)
					found := upstreams[index]
					if found == nil {
						return nil, fmt.Errorf("the routing with label %s has an invalid index. could not find upstream with index %d ", label.(string), index)
					} else {
						upstream := found.(map[string]interface{})
						id := upstream["id"]
						dict[id.(string)] = index
					}
				}
			}
		}
	}
	return dict, nil
}

func validateRoutingUpstreams(d *schema.ResourceData) error {
	var upstreams []interface{}
	var routings []interface{}
	if val, ok := d.GetOk("upstreams"); ok {
		upstreams = val.([]interface{})
	}
	if val, ok := d.GetOk("routing"); ok {
		routings = val.([]interface{})
	} else {
		return nil
	}
	for _, routingitem := range routings {
		ritem := routingitem.(map[string]interface{})
		label := ritem["label"]
		if rupstreams, ok := ritem["upstreams"]; ok {
			set := rupstreams.(*schema.Set)
			rupstreamlist := set.List()
			for _, rupstreamitem := range rupstreamlist {
				upitem := rupstreamitem.(map[string]interface{})
				if val, ok := upitem["index"]; ok {
					if upstreams[val.(int32)] == nil {
						return fmt.Errorf("the routing with label %s has an invalid index. could not find upstream with index %d ", label.(string), val.(int32))
					}
				}
			}
		}
	}
	return nil
}
