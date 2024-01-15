package anypoint

import (
	"context"
	"fmt"
	"io"
	"maps"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iancoleman/strcase"
	"github.com/mulesoft-anypoint/anypoint-client-go/apim"
	"github.com/mulesoft-anypoint/anypoint-client-go/apim_upstream"
	flexgateway "github.com/mulesoft-anypoint/anypoint-client-go/flexgateway"
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
		When an API instance of type Flex Gateway is created, it has automatically a default upstream linked to the endpoint_uri and a routing that points to this one.
		This provider will remove all default routings and upstreams.
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
				Description: "The API Manager Flex Gateway instance id.",
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
			"instance_label": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
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
				Required:         true,
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
				ForceNew:    true,
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
				Required:    true,
				Description: "The instance's routing mapping",
				MinItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The route label",
						},
						"upstreams": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "Upstreams for this particular route",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"label": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The exact label of the upstream from the list of declared upstreams.",
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
				Required:    true,
				MinItems:    1,
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
							Description: "The upstream's label. Used to select the upstream inside a routing.",
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
		CustomizeDiff: func(ctx context.Context, rd *schema.ResourceDiff, i interface{}) error {
			return validateRoutingUpstreams(rd)
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceApimFlexGatewayCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// init variables
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	authctx := getApimAuthCtx(ctx, &pco)
	// init request body
	body := newApimFlexGatewayPostBody(d)
	// execute post request
	res, httpr, err := pco.apimclient.DefaultApi.PostApimInstance(authctx, orgid, envid).ApimInstancePostBody(*body).Execute()
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
			Summary:  "Unable to create API flex gateway for org " + orgid + " and env " + envid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//update ids following the creation
	id := res.GetId()
	d.SetId(strconv.Itoa(int(id)))

	//create all upstreams if available
	diags = append(diags, resourceApimFlexGatewayUpstreamsCreate(ctx, d, m)...)
	// updates the routing
	diags = append(diags, resourceApimFlexGatewayRoutingUpdate(ctx, d, m)...)
	// removes default upstream that is systematically created along with the api flex gateway instance
	diags = append(diags, resourceApimFlexGatewayDeleteDefaultUpstream(ctx, d, m)...)
	//perform read
	diags = append(diags, resourceApimFlexGatewayRead(ctx, d, m)...)

	return diags
}

// Create upstreams for the apim flex gateway instance if upstreams is set
func resourceApimFlexGatewayUpstreamsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if input, ok := d.GetOk("upstreams"); ok {
		pco := m.(ProviderConfOutput)
		orgid, envid, id := decomposeApimFlexGatewayId(d)
		authctx := getApimUpstreamAuthCtx(ctx, &pco)
		bodies := newApimFlexGatewayUpstreamPostBody(input.([]interface{}))
		//upstreams := make([]apim_upstream.UpstreamDetails, 0)
		// loop over all new upstreams to create them
		for _, body := range bodies {
			//execute post upstream
			_, httpr, err := pco.apimupstreamclient.DefaultApi.PostApimInstanceUpstream(authctx, orgid, envid, id).UpstreamPostBody(*body).Execute()
			if err != nil {
				var details error
				if httpr != nil {
					b, _ := io.ReadAll(httpr.Body)
					details = fmt.Errorf(string(b))
				} else {
					details = err
				}
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to create Flex Gateway upstream " + body.GetLabel() + " for instance " + id,
					Detail:   details.Error(),
				})
				return diags
			}
			defer httpr.Body.Close()
			//upstreams = append(upstreams, res)
		}
		return readApimInstanceUpstreamsOnly(ctx, d, m)
	}
	return diags
}

// refresh the state of the flex gateway instance
func resourceApimFlexGatewayRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	id := d.Get("id").(string)
	authctx := getApimAuthCtx(ctx, &pco)
	if isComposedResourceId(id) {
		orgid, envid, id = decomposeApimFlexGatewayId(d)
	}

	res, httpr, err := pco.apimclient.DefaultApi.GetApimInstanceDetails(authctx, orgid, envid, id).Execute()
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
	defer httpr.Body.Close()
	// read upstreams
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

	d.SetId(id)
	d.Set("env_id", envid)
	d.Set("org_id", orgid)

	return diags
}

// updates the whole apim flex gateway in case of changes
func resourceApimFlexGatewayUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("id").(string)
	tocreate, toupdate, todelete := calcUpstreamsDiff(d)
	if len(toupdate) > 0 {
		authctx := getApimUpstreamAuthCtx(ctx, &pco)
		bodies := newApimFlexGatewayUpstreamPatchBody(toupdate)
		for i, body := range bodies {
			item := toupdate[i].(map[string]interface{})
			id := item["id"].(string)
			_, httpr, err := pco.apimupstreamclient.DefaultApi.PatchApimInstanceUpstream(authctx, orgid, envid, apimid, id).UpstreamPatchBody(*body).Execute()
			if err != nil {
				var details error
				if httpr != nil {
					b, _ := io.ReadAll(httpr.Body)
					details = fmt.Errorf(string(b))
				} else {
					details = err
				}
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to update Flex Gateway upstream " + id + " for instance " + apimid,
					Detail:   details.Error(),
				})
			}
			defer httpr.Body.Close()
		}
	}
	if len(tocreate) > 0 && !diags.HasError() {
		authctx := getApimUpstreamAuthCtx(ctx, &pco)
		bodies := newApimFlexGatewayUpstreamPostBody(tocreate)
		for _, body := range bodies {
			_, httpr, err := pco.apimupstreamclient.DefaultApi.PostApimInstanceUpstream(authctx, orgid, envid, apimid).UpstreamPostBody(*body).Execute()
			if err != nil {
				var details error
				if httpr != nil {
					b, _ := io.ReadAll(httpr.Body)
					details = fmt.Errorf(string(b))
				} else {
					details = err
				}
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to create Flex Gateway upstream " + body.GetLabel() + " for instance " + apimid,
					Detail:   details.Error(),
				})
			}
			defer httpr.Body.Close()
		}
	}
	if d.HasChanges(getApimFlexGatewayUpdatableAttributes()...) {
		body := newApimFlexGatewayPatchBody(d)
		authctx := getApimAuthCtx(ctx, &pco)
		_, httpr, err := pco.apimclient.DefaultApi.PatchApimInstance(authctx, orgid, envid, apimid).Body(body).Execute()
		if err != nil {
			var details error
			if httpr != nil {
				b, _ := io.ReadAll(httpr.Body)
				details = fmt.Errorf(string(b))
			} else {
				details = err
			}
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update Flex Gateway instance " + apimid,
				Detail:   details.Error(),
			})
		}
		defer httpr.Body.Close()
	}
	if len(todelete) > 0 && !diags.HasError() {
		authctx := getApimUpstreamAuthCtx(ctx, &pco)
		for _, item_todelete := range todelete {
			item := item_todelete.(map[string]interface{})
			id := item["id"].(string)
			httpr, err := pco.apimupstreamclient.DefaultApi.DeleteApimInstanceUpstream(authctx, orgid, envid, apimid, id).Execute()
			if err != nil {
				var details error
				if httpr != nil {
					b, _ := io.ReadAll(httpr.Body)
					details = fmt.Errorf(string(b))
				} else {
					details = err
				}
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to remove Flex Gateway upstream " + id + " for instance " + apimid,
					Detail:   details.Error(),
				})
			}
			defer httpr.Body.Close()
		}
	}
	diags = append(diags, resourceApimFlexGatewayRead(ctx, d, m)...)
	return diags
}

// Updates the routing only for the apim flex gateway
func resourceApimFlexGatewayRoutingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	id := d.Get("id").(string)
	authctx := getApimAuthCtx(ctx, &pco)

	if _, ok := d.GetOk("routing"); ok {
		body := newApimFlexGatewayRoutingPostBody(d) // creating body
		// patch
		_, httpr, err := pco.apimclient.DefaultApi.PatchApimInstance(authctx, orgid, envid, id).Body(body).Execute()
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
				Summary:  "unable to update api manager's flex gateway instance " + id + " with routing parameters ",
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

// deletes the apim flex gateway
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
		diags = append(diags, diag.Diagnostic{
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

// removes the default upstream that is created upon the creation of a flex gateway instance. it has an empty label.
// the list of upstreams should be updated before calling this function
func resourceApimFlexGatewayDeleteDefaultUpstream(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if upstreams, ok := d.GetOk("upstreams"); ok {
		filtered := FilterMapList(upstreams.([]interface{}), func(m map[string]interface{}) bool { return len(m["label"].(string)) == 0 })
		if len(filtered) > 0 {
			item := filtered[0].(map[string]interface{})
			id := item["id"].(string)
			pco := m.(ProviderConfOutput)
			orgid, envid, apimid := decomposeApimFlexGatewayId(d)
			authctx := getApimUpstreamAuthCtx(ctx, &pco)
			httpr, err := pco.apimupstreamclient.DefaultApi.DeleteApimInstanceUpstream(authctx, orgid, envid, apimid, id).Execute()
			if err != nil {
				var details error
				if httpr != nil {
					b, _ := io.ReadAll(httpr.Body)
					details = fmt.Errorf(string(b))
				} else {
					details = err
				}
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to delete API flex gateway's upstream " + id + " for instance " + apimid,
					Detail:   details.Error(),
				})
				return diags
			}
			defer httpr.Body.Close()
			return readApimInstanceUpstreamsOnly(ctx, d, m)
		}
	}
	return diags
}

func newApimFlexGatewayPostBody(d *schema.ResourceData) *apim.ApimInstancePostBody {
	body := apim.NewApimInstancePostBody()
	endpoint := newApimFlexGatewayEndpointPostBody(d)
	deployment := newApimFlexGatewayDeploymentPostBody(d)
	spec := newApimFlexGatewaySpecPostBody(d)
	//routing := newApimFlexGatewayRoutingPostBody(d)

	if val, ok := d.GetOk("instance_label"); ok {
		body.SetInstanceLabel(val.(string))
	} else {
		body.SetInstanceLabelNil()
	}
	body.SetTechnology(FLEX_GATEWAY_TECHNOLOGY)
	body.SetEndpoint(*endpoint)
	body.SetDeployment(*deployment)
	body.SetSpec(*spec)
	//body.SetRouting(routing)

	return body
}

func newApimFlexGatewayEndpointPostBody(d *schema.ResourceData) *apim.EndpointPostBody {
	body := apim.NewEndpointPostBody()

	body.SetIsCloudHubNil()
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

func newApimFlexGatewayRoutingPostBody(d *schema.ResourceData) map[string]interface{} {
	upstreams_data := d.Get("upstreams").([]interface{})
	body := make(map[string]interface{})
	if routings, ok := d.GetOk("routing"); ok {
		routings_list := routings.([]interface{})
		routing_elements := make([]map[string]interface{}, len(routings_list))
		for i, routings_item := range routings_list {
			routing_input := routings_item.(map[string]interface{})
			routing_output := make(map[string]interface{})
			if val, ok := routing_input["label"]; ok {
				routing_output["label"] = val
			}
			if upstream_input, ok := routing_input["upstreams"]; ok {
				upstream_input_set := upstream_input.(*schema.Set)
				upstream_input_list := upstream_input_set.List()
				upstreams_output := make([]map[string]interface{}, len(upstream_input_list))
				for j, upstream_item := range upstream_input_list {
					record := make(map[string]interface{})
					upstream := upstream_item.(map[string]interface{})
					if v, ok := upstream["weight"]; ok {
						record["weight"] = v
					}
					if v, ok := upstream["label"]; ok {
						filtered := FilterMapList(upstreams_data, func(m map[string]interface{}) bool { return m["label"].(string) == v.(string) })
						if len(filtered) > 0 {
							u := filtered[0].(map[string]interface{})
							record["id"] = u["id"]
						}
					}
					upstreams_output[j] = record
				}
				routing_output["upstreams"] = upstreams_output
			}
			if val, ok := routing_input["rules"]; ok && val != nil {
				rules := newApimFlexGatewayRoutingRulesPostBody(val.(*schema.Set))
				if len(rules) > 0 {
					routing_output["rules"] = rules
				}
			}
			routing_elements[i] = routing_output
		}
		body["routing"] = routing_elements
	}
	return body
}

func newApimFlexGatewayRoutingRulesPostBody(rules *schema.Set) map[string]interface{} {
	body := make(map[string]interface{})
	if rules.Len() > 0 {
		rules_list := rules.List()
		rules_map := rules_list[0].(map[string]interface{})
		if val, ok := rules_map["methods"]; ok && val != nil {
			set := val.(*schema.Set)
			body["methods"] = JoinStringInterfaceSlice(set.List(), "|")
		}
		if val, ok := rules_map["host"]; ok && val != nil {
			body["host"] = val.(string)
		}
		if val, ok := rules_map["path"]; ok && val != nil {
			body["path"] = val.(string)
		}
		if val, ok := rules_map["headers"]; ok && val != nil {
			body["headers"] = val.(map[string]interface{})
		}
	}
	return body
}

func newApimFlexGatewayUpstreamPostBody(upstreams []interface{}) []*apim_upstream.UpstreamPostBody {
	length := len(upstreams)
	if length == 0 {
		return []*apim_upstream.UpstreamPostBody{}
	}
	bodies := make([]*apim_upstream.UpstreamPostBody, length)
	for i, item := range upstreams {
		upstream := item.(map[string]interface{})
		b := apim_upstream.NewUpstreamPostBody()
		if val, ok := upstream["label"]; ok {
			b.SetLabel(val.(string))
		}
		if val, ok := upstream["uri"]; ok {
			b.SetUri(val.(string))
		}
		if val, ok := upstream["tls_context"]; ok && val != nil {
			set := val.(*schema.Set)
			if set.Len() > 0 {
				tls_context_input := set.List()[0].(map[string]interface{})
				tlc_body := apim_upstream.NewUpstreamPostBodyTlsContext()
				if v, ok := tls_context_input["secret_group_id"]; ok {
					tlc_body.SetSecretGroupId(v.(string))
				}
				if v, ok := tls_context_input["tls_context_id"]; ok {
					tlc_body.SetTlsContextId(v.(string))
				}
				b.SetTlsContext(*tlc_body)
			} else {
				b.SetTlsContextNil()
			}
		} else {
			b.SetTlsContextNil()
		}
		bodies[i] = b
	}
	return bodies
}

// creates patch body depending on the changes occured on the updatable attributes
func newApimFlexGatewayPatchBody(d *schema.ResourceData) map[string]interface{} {
	body := make(map[string]interface{})
	attributes := FilterStrList(getApimFlexGatewayUpdatableAttributes(), func(s string) bool {
		return !strings.HasPrefix(s, "endpoint") && !strings.HasPrefix(s, "deployment") && s != "routing" && s != "upstreams"
	})
	for _, attr := range attributes {
		if d.HasChange(attr) {
			body[strcase.ToCamel(attr)] = d.Get(attr)
		}
	}
	maps.Copy(body, newPatchBodyMap4FlattenedAttr("endpoint", d))
	maps.Copy(body, newPatchBodyMap4FlattenedAttr("deployment", d))
	maps.Copy(body, newApimFlexGatewayRoutingPatchBody(d))
	return body
}

// returns a routing patch body depending on if there's changes
func newApimFlexGatewayRoutingPatchBody(d *schema.ResourceData) map[string]interface{} {
	if d.HasChange("routing") {
		return newApimFlexGatewayRoutingPostBody(d)
	}
	return map[string]interface{}{}
}

// returns map constructed for the given prefix of attributes (ex endpoint or deployment).
// the function will get all attributes prefixed by the given prefix
// the function will return a map with the prefix as its root attribute and all attributes transformed to caml case without the prefix as the object elements.
// if no changes occured on the prefixed attributes, an empty map is returned
func newPatchBodyMap4FlattenedAttr(prefix string, d *schema.ResourceData) map[string]interface{} {
	separator := "_"
	endpoint_params := make(map[string]interface{})
	attributes := FilterStrList(getApimFlexGatewayUpdatableAttributes(), func(s string) bool {
		return strings.HasPrefix(s, prefix)
	})
	for _, attr := range attributes {
		if d.HasChange(attr) {
			attr_without_prefix, _ := strings.CutPrefix(attr, prefix+separator)
			endpoint_params[strcase.ToCamel(attr_without_prefix)] = d.Get(attr)
		}
	}
	body := make(map[string]interface{})
	if len(endpoint_params) > 0 {
		body[prefix] = endpoint_params
	}
	return body
}

func newApimFlexGatewayUpstreamPatchBody(upstreams []interface{}) []*apim_upstream.UpstreamPatchBody {
	if len(upstreams) == 0 {
		return []*apim_upstream.UpstreamPatchBody{}
	}
	bodies := make([]*apim_upstream.UpstreamPatchBody, len(upstreams))
	for i, item := range upstreams {
		upstream := item.(map[string]interface{})
		bodies[i] = apim_upstream.NewUpstreamPatchBody()
		if val, ok := upstream["label"]; ok {
			bodies[i].SetLabel(val.(string))
		}
		if val, ok := upstream["uri"]; ok {
			bodies[i].SetUri(val.(string))
		}
		if val, ok := upstream["tls_context"]; ok && val != nil {
			set := val.(*schema.Set)
			if set.Len() > 0 {
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
		} else {
			bodies[i].SetTlsContextNil()
		}
	}
	return bodies
}

func decomposeApimFlexGatewayId(d *schema.ResourceData) (string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2]
}

func setApimFlexGatewayAttributesToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getApimInstanceDetailsAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if attr == "routing" {
					result, err := mergeRoutingDetails2ResourceData(d, val.([]map[string]interface{}))
					if err != nil {
						return err
					}
					if err := d.Set(attr, result); err != nil {
						return fmt.Errorf("unable to set flex gateway instance attribute %s\n details: %s", attr, err)
					}
					continue
				}
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set api manager instance attribute %s\n details: %s", attr, err)
				}
			}
		}
	}
	return nil
}

// merge the given routing data coming from remote read operation with existing resource data for the case of routing.
func mergeRoutingDetails2ResourceData(d *schema.ResourceData, data []map[string]interface{}) ([]map[string]interface{}, error) {
	// the upstreams reference
	var upstreams_ref []interface{}
	if val, ok := d.GetOk("upstreams"); ok {
		upstreams_ref = val.([]interface{})
	}
	//Merging routing-details (coming from remote) with resourceData in a new list
	result := make([]map[string]interface{}, len(data))
	for i, rdetails := range data {
		rdetails_upstreams := rdetails["upstreams"].([]map[string]interface{})
		label := rdetails["label"]
		rupstream_result := make([]map[string]interface{}, 0)
		for _, rdetails_upstream := range rdetails_upstreams {
			id := rdetails_upstream["id"]
			clone := maps.Clone(rdetails_upstream)
			filtered := FilterMapList(upstreams_ref, func(m map[string]interface{}) bool { return m["id"].(string) == id.(string) })
			if len(filtered) > 0 {
				u := filtered[0].(map[string]interface{})
				clone["label"] = u["label"].(string)
				rupstream_result = append(rupstream_result, clone)
			}
		}
		result[i] = map[string]interface{}{
			"label":     label,
			"upstreams": rupstream_result,
		}
		if rules, ok := rdetails["rules"]; ok {
			result[i]["rules"] = rules
		}
	}
	return result, nil
}

// validates if routing have a declared upstream
func validateRoutingUpstreams(d *schema.ResourceDiff) error {
	var upstreams []interface{}
	var routings []interface{}
	weight_limit := 100
	if val, ok := d.GetOk("upstreams"); ok {
		upstreams = val.([]interface{})
	}
	if val, ok := d.GetOk("routing"); ok {
		routings = val.([]interface{})
	} else {
		return nil
	}
	for _, routing_item := range routings {
		ritem := routing_item.(map[string]interface{})
		label := ritem["label"]
		if rupstreams, ok := ritem["upstreams"]; ok {
			set := rupstreams.(*schema.Set)
			rupstream_list := set.List()
			weight_total := 0
			for _, rupstream_item := range rupstream_list {
				upitem := rupstream_item.(map[string]interface{})
				if val, ok := upitem["label"]; ok {
					filtered := FilterMapList(upstreams, func(m map[string]interface{}) bool { return val.(string) == m["label"].(string) })
					if len(filtered) == 0 {
						return fmt.Errorf("could not find upstream with label %s for routing %s in your list of upstreams", val.(string), label.(string))
					}
				} else {
					return fmt.Errorf("routings label is mandatory ")
				}
				weight_total = weight_total + upitem["weight"].(int)
			}
			if weight_total > weight_limit {
				return fmt.Errorf("routing upstreams weight should add-up to %d. The routing \"%s\" upstreams weight sum exceeds %d", weight_limit, label.(string), weight_limit)
			}
			if weight_total < weight_limit {
				return fmt.Errorf("routing upstreams weight should add-up to %d. The routing \"%s\" upstreams weight sum is below %d", weight_limit, label.(string), weight_limit)
			}
		}
	}
	return nil
}

func getApimFlexGatewayUpdatableAttributes() []string {
	attributes := [...]string{
		"instance_label", "description", "tags", "provider_id",
		"deprecated", "endpoint_uri", "endpoint_proxy_uri",
		"endpoint_proxy_registration_uri", "endpoint_deployment_type",
		"endpoint_tls_inbound_context", "deployment_gateway_version",
		"deployment_target_id", "deployment_expected_status",
		"deployment_type", "deployment_overwrite",
		"routing", "upstreams",
	}
	return attributes[:]
}

// calculates the difference between old and new upstreams declaration if any
// returns list of upstreams to be created, another list of upstreams to be updated and a list of upstreams to be deleted.
func calcUpstreamsDiff(d *schema.ResourceData) ([]interface{}, []interface{}, []interface{}) {
	toupdate := make([]interface{}, 0)
	todelete := make([]interface{}, 0)
	tocreate := make([]interface{}, 0)
	if d.HasChange("upstreams") {
		old_upstreams, new_upstreams := d.GetChange("upstreams")
		old_upstreams_list := old_upstreams.([]interface{})
		new_upstreams_list := new_upstreams.([]interface{})
		for _, new_upstream := range new_upstreams_list {
			new := new_upstream.(map[string]interface{})
			if id, ok := new["id"]; ok {
				filtered := FilterMapList(old_upstreams_list, func(m map[string]interface{}) bool { return id.(string) == m["id"].(string) })
				if len(filtered) > 0 {
					old_match := filtered[0].(map[string]interface{})
					if !isUpstreamsEqual(new, old_match) {
						toupdate = append(toupdate, new)
					}
				}
			} else {
				tocreate = append(tocreate, new)
			}
		}
		for _, old_upstream := range old_upstreams_list {
			old := old_upstream.(map[string]interface{})
			filtered := FilterMapList(new_upstreams_list, func(m map[string]interface{}) bool {
				if id, ok := m["id"]; ok {
					return old["id"] == id
				}
				return false
			})
			if len(filtered) == 0 {
				todelete = append(todelete, old)
			}
		}
	}
	return tocreate, toupdate, todelete
}

// returns true if the given upstreams are equal
func isUpstreamsEqual(a, b map[string]interface{}) bool {
	attributes := [...]string{
		"id", "label", "uri", "tls_context",
	}
	for _, attr := range attributes {
		if attr == "tls_context" {
			if !isUpstreamTlsContextEqual(a[attr].(*schema.Set), b[attr].(*schema.Set)) {
				return false
			}
		} else {
			if a[attr].(string) != b[attr].(string) {
				return false
			}
		}
	}
	return true
}

// returns true if given upstream tls-contexts are equal
func isUpstreamTlsContextEqual(a *schema.Set, b *schema.Set) bool {
	if a.Len() != b.Len() {
		return false
	}
	if a.Len() == 0 && b.Len() == 0 {
		return true
	}
	attributes := [...]string{
		"secret_group_id", "tls_context_id",
	}
	a_list := a.List()
	b_list := b.List()
	a_item := a_list[0].(map[string]interface{})
	b_item := b_list[0].(map[string]interface{})
	for _, attr := range attributes {
		if a_item[attr].(string) != b_item[attr].(string) {
			return false
		}
	}

	return true
}

/*
 * Returns authentication context (includes authorization header)
 */
func getFlexGatewayAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, flexgateway.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, flexgateway.ContextServerIndex, pco.server_index)
}
