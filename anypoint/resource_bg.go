package anypoint

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	org "github.com/mulesoft-consulting/anypoint-client-go/org"
)

func resourceBG() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBGCreate,
		ReadContext:   resourceBGRead,
		UpdateContext: resourceBGUpdate,
		DeleteContext: resourceBGDelete,
		Description: `
		Creates a business group (org).
		`,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"idprovider_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_federated": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"parent_organization_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"parent_organization_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"sub_organization_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tenant_organization_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"mfa_required": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_automatic_admin_promotion_exempt": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_master": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"subscription_category": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subscription_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subscription_expiration": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"properties": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"environments": {
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
						"organization_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"is_production": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"entitlements_createenvironments": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"entitlements_globaldeployment": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"entitlements_createsuborgs": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"entitlements_hybridenabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"entitlements_hybridinsight": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"entitlements_hybridautodiscoverproperties": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"entitlements_vcoresproduction_assigned": {
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  0,
			},
			"entitlements_vcoresproduction_reassigned": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"entitlements_vcoressandbox_assigned": {
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  0,
			},
			"entitlements_vcoressandbox_reassigned": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"entitlements_vcoresdesign_assigned": {
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  0,
			},
			"entitlements_vcoresdesign_reassigned": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"entitlements_staticips_assigned": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"entitlements_staticips_reassigned": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"entitlements_vpcs_assigned": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"entitlements_vpcs_reassigned": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"entitlements_vpns_assigned": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"entitlements_vpns_reassigned": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"entitlements_workerloggingoverride_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_mqmessages_base": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_mqmessages_addon": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_mqrequests_base": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_mqrequests_addon": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_objectstorerequestunits_base": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_objectstorerequestunits_addon": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_objectstorekeys_base": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_objectstorekeys_addon": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_mqadvancedfeatures_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_gateways_assigned": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_designcenter_api": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_designcenter_mozart": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_partnersproduction_assigned": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_partnerssandbox_assigned": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_tradingpartnersproduction_assigned": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_tradingpartnerssandbox_assigned": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_loadbalancer_assigned": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"entitlements_loadbalancer_reassigned": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_externalidentity": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_autoscaling": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_armalerts": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_apis_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_apimonitoring_schedules": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_apicommunitymanager_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_monitoringcenter_productsku": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_apiquery_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_apiquery_productsku": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_apiqueryc360_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_anggovernance_level": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_crowd_hideapimanagerdesigner": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_crowd_hideformerapiplatform": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_crowd_environments": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_cam_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_exchange2_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_crowdselfservicemigration_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_kpidashboard_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_pcf": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_appviz": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_runtimefabric": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_anypointsecuritytokenization_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_anypointsecurityedgepolicies_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_runtimefabriccloud_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_servicemesh_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"entitlements_messaging_assigned": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_workerclouds_assigned": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"entitlements_workerclouds_reassigned": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"owner_created_at": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_updated_at": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_organization_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_firstname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_lastname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_phonenumber": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_idprovider_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"owner_deleted": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"owner_lastlogin": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_mfaverification_excluded": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"owner_mfaverifiers_configured": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"session_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  60,
			},
		},
	}
}

func resourceBGCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	authctx := getBGAuthCtx(ctx, &pco)
	body := newBGPostBody(d)

	res, httpr, err := pco.orgclient.DefaultApi.OrganizationsPost(authctx).BGPostReqBody(*body).Execute()
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
			Summary:  "Unable to Create Business Group",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(res.GetId())
	resourceBGRead(ctx, d, m)

	return diags
}

func resourceBGRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Id()

	authctx := getBGAuthCtx(ctx, &pco)

	res, httpr, err := pco.orgclient.DefaultApi.OrganizationsOrgIdGet(authctx, orgid).Execute()
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
			Summary:  "Unable to Get Business Group",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	orginstance := flattenBGData(&res)

	if err := setBGCoreAttributesToResourceData(d, orginstance); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set Business Group",
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func resourceBGUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Id()

	authctx := getBGAuthCtx(ctx, &pco)

	if d.HasChanges(getBGUpdatableAttributes()...) {
		body := newBGPutBody(d)
		_, httpr, err := pco.orgclient.DefaultApi.OrganizationsOrgIdPut(authctx, orgid).BGPutReqBody(*body).Execute()
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
				Summary:  "Unable to Update Business Group",
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceBGRead(ctx, d, m)
}

func resourceBGDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Id()

	authctx := getBGAuthCtx(ctx, &pco)

	_, httpr, err := pco.orgclient.DefaultApi.OrganizationsOrgIdDelete(authctx, orgid).Execute()
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
			Summary:  "Unable to Delete Business Group",
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
 * Creates body for B.G POST request
 */
func newBGPostBody(d *schema.ResourceData) *org.BGPostReqBody {
	body := org.NewBGPostReqBodyWithDefaults()

	body.SetName(d.Get("name").(string))
	body.SetOwnerId(d.Get("owner_id").(string))
	body.SetParentOrganizationId((d.Get("parent_organization_id").(string)))
	body.SetEntitlements(*newEntitlementsFromD(d))

	return body
}

/*
 * Creates body for B.G PUT request
 */
func newBGPutBody(d *schema.ResourceData) *org.BGPutReqBody {
	body := org.NewBGPutReqBodyWithDefaults()
	body.SetName(d.Get("name").(string))
	body.SetOwnerId(d.Get("owner_id").(string))
	body.SetEntitlements(*newEntitlementsFromD(d))
	body.SetSessionTimeout(int32(d.Get("session_timeout").(int)))

	return body
}

/*
 * Creates Entitlements from Resource Data Schema
 */
func newEntitlementsFromD(d *schema.ResourceData) *org.EntitlementsCore {
	loadbalancer := org.NewLoadBalancerWithDefaults()
	loadbalancer.SetAssigned(int32(d.Get("entitlements_loadbalancer_assigned").(int)))
	staticips := org.NewStaticIpsWithDefaults()
	staticips.SetAssigned(int32(d.Get("entitlements_staticips_assigned").(int)))
	vcoresandbox := org.NewVCoresSandboxWithDefaults()
	vcoresandbox.SetAssigned(float32(d.Get("entitlements_vcoressandbox_assigned").(float64)))
	vcoredesign := org.NewVCoresDesignWithDefaults()
	vcoredesign.SetAssigned(float32(d.Get("entitlements_vcoresdesign_assigned").(float64)))
	vpns := org.NewVpnsWithDefaults()
	vpns.SetAssigned(int32(d.Get("entitlements_vpns_assigned").(int)))
	vpcs := org.NewVpcsWithDefaults()
	vpcs.SetAssigned(int32(d.Get("entitlements_vpcs_assigned").(int)))
	vcoreprod := org.NewVCoresProductionWithDefaults()
	vcoreprod.SetAssigned(float32(d.Get("entitlements_vcoresproduction_assigned").(float64)))
	entitlements := org.NewEntitlementsCore(
		d.Get("entitlements_globaldeployment").(bool),
		d.Get("entitlements_createenvironments").(bool),
		d.Get("entitlements_createsuborgs").(bool),
		*loadbalancer,
		*staticips,
		*vcoredesign,
		*vcoreprod,
		*vcoresandbox,
		*vpcs,
		*vpns,
	)

	return entitlements
}

/*
 * Returns authentication context (includes authorization header)
 */
func getBGAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, org.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, org.ContextServerIndex, pco.server_index)
}
