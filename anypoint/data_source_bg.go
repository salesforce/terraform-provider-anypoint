package anypoint

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	org "github.com/mulesoft-consulting/anypoint-client-go/org"
)

func dataSourceBG() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBGRead,
		Description: `
		Reads a specific ` + "`" + `business group` + "`" + `.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "This organization's generated unique id.",
			},
			"owner_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user id of the owner of this organization.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of this organization.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time when this organization was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time when this organization was updated.",
			},
			"client_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization client id.",
			},
			"idprovider_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The identity provider if of this organization",
			},
			"is_federated": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this organization is federated.",
			},
			"parent_organization_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The immediate parent organization id of this organization.",
			},
			"sub_organization_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Array of descendant organizations.",
			},
			"tenant_organization_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Array of tenant organizations",
			},
			"mfa_required": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Whether MFA is enforced in this organization",
			},
			"is_automatic_admin_promotion_exempt": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the admin promotion exemption is enabled on this organization",
			},
			"domain": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization's domain",
			},
			"is_master": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this organization is the master org.",
			},
			"subscription_category": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The anypoint platform subscription category",
			},
			"subscription_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The anypoint platform subscription type.",
			},
			"subscription_expiration": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The anypoint platform subscription expiration date.",
			},
			"properties": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organiztion's general properties.",
			},
			"environments": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The organization's list of environments",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The environment unique id.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The environment name",
						},
						"organization_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The environment's organization id.",
						},
						"is_production": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this environment is a production environment.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the environment (e.g sandbox or production)",
						},
						"client_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The environment's client id",
						},
					},
				},
			},
			"entitlements_createenvironments": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this organization can have additional environments.",
			},
			"entitlements_globaldeployment": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this organization can have global deployments.",
			},
			"entitlements_createsuborgs": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this organization can create sub organizations (descendants).",
			},
			"entitlements_hybridenabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this organization has hybrid enabled.",
			},
			"entitlements_hybridinsight": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this organization has hybrid insight.",
			},
			"entitlements_hybridautodiscoverproperties": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this organization has hybrid auto-discovery properties enabled",
			},
			"entitlements_vcoresproduction_assigned": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The number of production vcores assigned to this organization.",
			},
			"entitlements_vcoresproduction_reassigned": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The number of production vcores reassigned to this organization.",
			},
			"entitlements_vcoressandbox_assigned": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The number of sandbox vcores assigned to this organization.",
			},
			"entitlements_vcoressandbox_reassigned": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The number of sandbox vcores reassigned to this organization.",
			},
			"entitlements_vcoresdesign_assigned": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The number of design vcores assigned to this organization.",
			},
			"entitlements_vcoresdesign_reassigned": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The number of design vcores reassigned to this organization.",
			},
			"entitlements_staticips_assigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of static IPs assigned to this organization.",
			},
			"entitlements_staticips_reassigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of static IPs reassigned to this organization.",
			},
			"entitlements_vpcs_assigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of VPCs assigned to this organization.",
			},
			"entitlements_vpcs_reassigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of VPCs reassigned to this organization.",
			},
			"entitlements_vpns_assigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of VPNs assigned to this organization.",
			},
			"entitlements_vpns_reassigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of VPNs reassigned to this organization.",
			},
			"entitlements_workerloggingoverride_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the loggin override on workers is enabled for this organization.",
			},
			"entitlements_mqmessages_base": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of basic MQ messages assigned to this organization.",
			},
			"entitlements_mqmessages_addon": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of MQ messages addons assigned to this organization.",
			},
			"entitlements_mqrequests_base": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of MQ requests base assigned to this organization.",
			},
			"entitlements_mqrequests_addon": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of MQ requests addon assigned to this organization.",
			},
			"entitlements_objectstorerequestunits_base": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of object store requests unists base for this organization.",
			},
			"entitlements_objectstorerequestunits_addon": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of object store requests units addon for this organization.",
			},
			"entitlements_objectstorekeys_base": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of object store keys base for this organization.",
			},
			"entitlements_objectstorekeys_addon": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of object store keys addon for this organization.",
			},
			"entitlements_mqadvancedfeatures_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the Anypoint MQ advanced features are enabled for this organization.",
			},
			"entitlements_gateways_assigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of gateways assigned to this organization.",
			},
			"entitlements_designcenter_api": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether te design center api is enabled for this organization.",
			},
			"entitlements_designcenter_mozart": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the design center mozart is enabled for this organization.",
			},
			"entitlements_partnersproduction_assigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of partners production vcores assigned to this organization.",
			},
			"entitlements_partnerssandbox_assigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of partners sandbox vcores assigned to this organization.",
			},
			"entitlements_tradingpartnersproduction_assigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of traded partners production vcores assigned to this organization.",
			},
			"entitlements_tradingpartnerssandbox_assigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of traded partners sandbox vcores assigned to this organization.",
			},
			"entitlements_loadbalancer_assigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of dedicated load balancers (DLB) assigned to this organization.",
			},
			"entitlements_loadbalancer_reassigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of dedicated load balancers (DLB) reassigned to this organization.",
			},
			"entitlements_externalidentity": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether an external identity provider (IDP) was assigned to this organization.",
			},
			"entitlements_autoscaling": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether autoscaling is enabled for this organization",
			},
			"entitlements_armalerts": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether arm alerts are enabled for this organization.",
			},
			"entitlements_apis_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "whether APIs are enabled for this organization.",
			},
			"entitlements_apimonitoring_schedules": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of api monitoring schedules for this organization.",
			},
			"entitlements_apicommunitymanager_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether api community manager is enabled for this organization.",
			},
			"entitlements_monitoringcenter_productsku": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of monitoring center products sku for this organization.",
			},
			"entitlements_apiquery_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether api queries are enabled for this organization.",
			},
			"entitlements_apiquery_productsku": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of api query product sku for this organization.",
			},
			"entitlements_apiqueryc360_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether api query C360 is enabled for this organization.",
			},
			"entitlements_anggovernance_level": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"entitlements_crowd_hideapimanagerdesigner": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"entitlements_crowd_hideformerapiplatform": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"entitlements_crowd_environments": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"entitlements_cam_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether cam is enabled for this organization.",
			},
			"entitlements_exchange2_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether exchange v2 is enabled for this organization.",
			},
			"entitlements_crowdselfservicemigration_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether crow self service migration is enabled for this organization.",
			},
			"entitlements_kpidashboard_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether KPI dashboard is enabled for this organization.",
			},
			"entitlements_pcf": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether PCF is included for this organization.",
			},
			"entitlements_appviz": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the app vizualize if enabled for this organization.",
			},
			"entitlements_runtimefabric": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether Runtime Fabrics (RTF) is enabled for this organization.",
			},
			"entitlements_anypointsecuritytokenization_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "whether Anypoint securirty tokenization is enabled for this organization.",
			},
			"entitlements_anypointsecurityedgepolicies_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether Anypoint security edge policies is enabled for this organization.",
			},
			"entitlements_runtimefabriccloud_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether Runtime Fabrics (RTF) is enabled for this organization.",
			},
			"entitlements_servicemesh_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether Service Mesh is enabled for this organization.",
			},
			"entitlements_messaging_assigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of messaging assigned to this organization.",
			},
			"entitlements_workerclouds_assigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of worker clouds assigned to this organization",
			},
			"entitlements_workerclouds_reassigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of worker clouds reassigned to this organization",
			},
			"owner_created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the organization owner creation date",
			},
			"owner_updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization owner update date.",
			},
			"owner_organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization owner's organization id.",
			},
			"owner_firstname": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The organization owner's firstname",
			},
			"owner_lastname": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The organization owner's lastname.",
			},
			"owner_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The organization owner's email.",
			},
			"owner_phonenumber": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The organization owner's phone number.",
			},
			"owner_username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization owner username.",
			},
			"owner_idprovider_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization owner identity provider id.",
			},
			"owner_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the organization owner account is enabled.",
			},
			"owner_deleted": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the organization owner account is deleted.",
			},
			"owner_lastlogin": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last time the organization owner logged in.",
			},
			"owner_mfaverification_excluded": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the organization owner MFA verification is excluded.",
			},
			"owner_mfaverifiers_configured": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization owner MFA verification configuration",
			},
			"owner_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization owner account type.",
			},
			"session_timeout": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The organization's session timeout",
			},
		},
	}
}

/*
 * Reads a Business Group. Required the bg_id as input
 */
func dataSourceBGRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("id").(string)
	authctx := getBGAuthCtx(ctx, &pco)

	res, httpr, err := pco.orgclient.DefaultApi.OrganizationsOrgIdGet(authctx, orgid).Execute()
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
			Summary:  "Unable to Get Business Group",
			Detail:   details,
		})
		return diags
	}

	bg := flattenBGData(&res)

	if err := setBGCoreAttributesToResourceData(d, bg); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set Business Group",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(orgid)

	return diags
}

func setBGCoreAttributesToResourceData(d *schema.ResourceData, bg map[string]interface{}) error {
	attributes := getBGCoreAttributes()
	if bg != nil {
		for _, attr := range attributes {
			if err := d.Set(attr, bg[attr]); err != nil {
				return fmt.Errorf("unable to set Business Group attribute %s\n details: %s", attr, err)
			}
		}
	}
	return nil
}

/*
 * Flattens the Business Group object
 */
func flattenBGData(bg *org.MasterBGDetail) map[string]interface{} {
	if bg != nil {
		item := make(map[string]interface{})

		item["id"] = bg.GetId()
		item["name"] = bg.GetName()
		item["created_at"] = bg.GetCreatedAt()
		item["updated_at"] = bg.GetUpdatedAt()
		item["owner_id"] = bg.GetOwnerId()
		item["client_id"] = bg.GetClientId()
		item["idprovider_id"] = bg.GetIdproviderId()
		item["is_federated"] = bg.GetIsFederated()
		item["parent_organization_ids"] = bg.GetParentOrganizationIds()
		item["sub_organization_ids"] = bg.GetSubOrganizationIds()
		item["tenant_organization_ids"] = bg.GetTenantOrganizationIds()
		item["mfa_required"] = bg.GetMfaRequired()
		item["is_automatic_admin_promotion_exempt"] = bg.GetIsAutomaticAdminPromotionExempt()
		item["domain"] = bg.GetDomain()
		item["is_master"] = bg.GetIsMaster()

		item["properties"] = fmt.Sprint(bg.GetProperties())

		subscription := bg.GetSubscription()
		item["subscription_category"] = subscription.GetCategory()
		item["subscription_type"] = subscription.GetType()
		item["subscription_expiration"] = subscription.GetExpiration()

		environments := make([]interface{}, len(bg.GetEnvironments()))
		for i, currentEnv := range bg.GetEnvironments() {
			env := make(map[string]interface{})
			env["id"] = currentEnv.GetId()
			env["name"] = currentEnv.GetName()
			env["organization_id"] = currentEnv.GetOrganizationId()
			env["is_production"] = currentEnv.GetIsProduction()
			env["type"] = currentEnv.GetType()
			env["client_id"] = currentEnv.GetClientId()
			environments[i] = env
		}
		item["environments"] = environments

		entitlements := bg.GetEntitlements()
		item["entitlements_createenvironments"] = entitlements.GetCreateEnvironments()
		item["entitlements_globaldeployment"] = entitlements.GetGlobalDeployment()
		item["entitlements_createsuborgs"] = entitlements.GetCreateSubOrgs()
		hybrid := entitlements.GetHybrid()
		item["entitlements_hybridenabled"] = hybrid.GetEnabled()
		item["entitlements_hybridinsight"] = entitlements.GetHybridInsight()
		item["entitlements_hybridautodiscoverproperties"] = entitlements.GetHybridAutoDiscoverProperties()
		vCoresProduction := entitlements.GetVCoresProduction()
		item["entitlements_vcoresproduction_assigned"] = vCoresProduction.GetAssigned()
		item["entitlements_vcoresproduction_reassigned"] = vCoresProduction.GetReassigned()
		vCoresSandbox := entitlements.GetVCoresSandbox()
		item["entitlements_vcoressandbox_assigned"] = vCoresSandbox.GetAssigned()
		item["entitlements_vcoressandbox_reassigned"] = vCoresSandbox.GetReassigned()
		vCoreDesign := entitlements.GetVCoresDesign()
		item["entitlements_vcoresdesign_assigned"] = vCoreDesign.GetAssigned()
		item["entitlements_vcoresdesign_reassigned"] = vCoreDesign.GetReassigned()
		staticIps := entitlements.GetStaticIps()
		item["entitlements_staticips_assigned"] = staticIps.GetAssigned()
		item["entitlements_staticips_reassigned"] = staticIps.GetReassigned()
		vpcs := entitlements.GetVpcs()
		item["entitlements_vpcs_assigned"] = vpcs.GetAssigned()
		item["entitlements_vpcs_reassigned"] = vpcs.GetReassigned()
		vpns := entitlements.GetVpns()
		item["entitlements_vpns_assigned"] = vpns.GetAssigned()
		item["entitlements_vpns_reassigned"] = vpns.GetReassigned()
		workerLoggingOverride := entitlements.GetWorkerLoggingOverride()
		item["entitlements_workerloggingoverride_enabled"] = workerLoggingOverride.GetEnabled()
		mqMessages := entitlements.GetMqMessages()
		item["entitlements_mqmessages_base"] = mqMessages.GetBase()
		item["entitlements_mqmessages_addon"] = mqMessages.GetAddOn()
		mqRequests := entitlements.GetMqRequests()
		item["entitlements_mqrequests_base"] = mqRequests.GetBase()
		item["entitlements_mqrequests_addon"] = mqRequests.GetAddOn()
		objectStoreRequestUnits := entitlements.GetObjectStoreRequestUnits()
		item["entitlements_objectstorerequestunits_base"] = objectStoreRequestUnits.GetBase()
		item["entitlements_objectstorerequestunits_addon"] = objectStoreRequestUnits.GetAddOn()
		objectStoreKeys := entitlements.GetObjectStoreKeys()
		item["entitlements_objectstorekeys_base"] = objectStoreKeys.GetBase()
		item["entitlements_objectstorekeys_addon"] = objectStoreKeys.GetAddOn()
		mqAdvancedFeatures := entitlements.GetMqAdvancedFeatures()
		item["entitlements_mqadvancedfeatures_enabled"] = mqAdvancedFeatures.GetEnabled()
		gateways := entitlements.GetGateways()
		item["entitlements_gateways_assigned"] = gateways.GetAssigned()
		designCenter := entitlements.GetDesignCenter()
		item["entitlements_designcenter_api"] = designCenter.GetApi()
		item["entitlements_designcenter_mozart"] = designCenter.GetMozart()
		partnersProduction := entitlements.GetPartnersProduction()
		item["entitlements_partnersproduction_assigned"] = partnersProduction.GetAssigned()
		partnersSandbox := entitlements.GetPartnersSandbox()
		item["entitlements_partnerssandbox_assigned"] = partnersSandbox.GetAssigned()
		tradingPartnersProduction := entitlements.GetTradingPartnersProduction()
		item["entitlements_tradingpartnersproduction_assigned"] = tradingPartnersProduction.GetAssigned()
		tradingPartnersSandbox := entitlements.GetTradingPartnersSandbox()
		item["entitlements_tradingpartnerssandbox_assigned"] = tradingPartnersSandbox.GetAssigned()
		loadBalancer := entitlements.GetLoadBalancer()
		item["entitlements_loadbalancer_assigned"] = loadBalancer.GetAssigned()
		item["entitlements_loadbalancer_reassigned"] = loadBalancer.GetReassigned()
		item["entitlements_externalidentity"] = entitlements.GetExternalIdentity()
		item["entitlements_autoscaling"] = entitlements.GetAutoscaling()
		item["entitlements_armalerts"] = entitlements.GetArmAlerts()
		apis := entitlements.GetApis()
		item["entitlements_apis_enabled"] = apis.GetEnabled()
		apiMonitoring := entitlements.GetApiMonitoring()
		item["entitlements_apimonitoring_schedules"] = apiMonitoring.GetSchedules()
		apiCommunityManager := entitlements.GetApiCommunityManager()
		item["entitlements_apicommunitymanager_enabled"] = apiCommunityManager.GetEnabled()
		monitoringCenter := entitlements.GetMonitoringCenter()
		item["entitlements_monitoringcenter_productsku"] = monitoringCenter.GetProductSKU()
		apiQuery := entitlements.GetApiQuery()
		item["entitlements_apiquery_enabled"] = apiQuery.GetEnabled()
		item["entitlements_apiquery_productsku"] = apiQuery.GetProductSKU()
		apiQueryC360 := entitlements.GetApiQueryC360()
		item["entitlements_apiqueryc360_enabled"] = apiQueryC360.GetEnabled()
		angGovernance := entitlements.GetAngGovernance()
		item["entitlements_anggovernance_level"] = angGovernance.GetLevel()
		crowd := entitlements.GetCrowd()
		item["entitlements_crowd_hideapimanagerdesigner"] = crowd.GetHideApiManagerDesigner()
		item["entitlements_crowd_hideformerapiplatform"] = crowd.GetHideFormerApiPlatform()
		item["entitlements_crowd_environments"] = crowd.GetEnvironments()
		cam := entitlements.GetCam()
		item["entitlements_cam_enabled"] = cam.GetEnabled()
		exchange2 := entitlements.GetExchange2()
		item["entitlements_exchange2_enabled"] = exchange2.GetEnabled()
		crowdSelfServiceMigration := entitlements.GetCrowdSelfServiceMigration()
		item["entitlements_crowdselfservicemigration_enabled"] = crowdSelfServiceMigration.GetEnabled()
		kpiDashboard := entitlements.GetKpiDashboard()
		item["entitlements_kpidashboard_enabled"] = kpiDashboard.GetEnabled()
		item["entitlements_pcf"] = entitlements.GetPcf()
		item["entitlements_appviz"] = entitlements.GetAppViz()
		item["entitlements_runtimefabric"] = entitlements.GetRuntimeFabric()
		anypointSecurityTokenization := entitlements.GetAnypointSecurityTokenization()
		item["entitlements_anypointsecuritytokenization_enabled"] = anypointSecurityTokenization.GetEnabled()
		anypointSecurityEdgePolicies := entitlements.GetAnypointSecurityEdgePolicies()
		item["entitlements_anypointsecurityedgepolicies_enabled"] = anypointSecurityEdgePolicies.GetEnabled()
		runtimeFabricCloud := entitlements.GetRuntimeFabricCloud()
		item["entitlements_runtimefabriccloud_enabled"] = runtimeFabricCloud.GetEnabled()
		serviceMesh := entitlements.GetServiceMesh()
		item["entitlements_servicemesh_enabled"] = serviceMesh.GetEnabled()
		messaging := entitlements.GetMessaging()
		item["entitlements_messaging_assigned"] = messaging.GetAssigned()
		workerClouds := entitlements.GetWorkerClouds()
		item["entitlements_workerclouds_assigned"] = workerClouds.GetAssigned()
		item["entitlements_workerclouds_reassigned"] = workerClouds.GetReassigned()

		owner := bg.GetOwner()
		item["owner_id"] = owner.GetId()
		item["owner_created_at"] = owner.GetCreatedAt()
		item["owner_updated_at"] = owner.GetUpdatedAt()
		item["owner_organization_id"] = owner.GetOrganizationId()
		item["owner_firstname"] = owner.GetFirstName()
		item["owner_lastname"] = owner.GetLastName()
		item["owner_email"] = owner.GetEmail()
		item["owner_phonenumber"] = owner.GetPhoneNumber()
		item["owner_username"] = owner.GetUsername()
		item["owner_idprovider_id"] = owner.GetIdproviderId()
		item["owner_enabled"] = owner.GetEnabled()
		item["owner_deleted"] = owner.GetDeleted()
		item["owner_lastlogin"] = owner.GetLastLogin()
		item["owner_mfaverification_excluded"] = owner.GetMfaVerificationExcluded()
		item["owner_mfaverifiers_configured"] = owner.GetMfaVerifiersConfigured()
		item["owner_type"] = owner.GetType()

		item["session_timeout"] = bg.GetSessionTimeout()

		return item

	}
	return nil
}

func getBGCoreAttributes() []string {
	attributes := [...]string{
		"name", "created_at", "updated_at", "owner_id", "client_id", "idprovider_id",
		"is_federated", "parent_organization_ids", "sub_organization_ids", "tenant_organization_ids",
		"mfa_required", "is_automatic_admin_promotion_exempt", "domain", "is_master", "subscription_category",
		"subscription_type", "subscription_expiration", "properties", "environments",
		"entitlements_createenvironments", "entitlements_globaldeployment", "entitlements_createsuborgs",
		"entitlements_hybridenabled", "entitlements_hybridinsight", "entitlements_hybridautodiscoverproperties",
		"entitlements_vcoresproduction_assigned", "entitlements_vcoresproduction_reassigned",
		"entitlements_vcoressandbox_assigned", "entitlements_vcoressandbox_reassigned",
		"entitlements_vcoresdesign_assigned", "entitlements_vcoresdesign_reassigned",
		"entitlements_staticips_assigned", "entitlements_staticips_reassigned", "entitlements_vpcs_assigned",
		"entitlements_vpcs_reassigned", "entitlements_vpns_assigned", "entitlements_vpns_reassigned",
		"entitlements_workerloggingoverride_enabled", "entitlements_mqmessages_base", "entitlements_mqmessages_addon",
		"entitlements_mqrequests_base", "entitlements_mqrequests_addon", "entitlements_objectstorerequestunits_base",
		"entitlements_objectstorerequestunits_addon", "entitlements_objectstorekeys_base", "entitlements_objectstorekeys_addon",
		"entitlements_mqadvancedfeatures_enabled", "entitlements_gateways_assigned", "entitlements_designcenter_api",
		"entitlements_designcenter_mozart", "entitlements_partnersproduction_assigned", "entitlements_partnerssandbox_assigned",
		"entitlements_tradingpartnersproduction_assigned", "entitlements_tradingpartnerssandbox_assigned", "entitlements_loadbalancer_assigned",
		"entitlements_loadbalancer_reassigned", "entitlements_externalidentity", "entitlements_autoscaling",
		"entitlements_armalerts", "entitlements_apis_enabled", "entitlements_apimonitoring_schedules",
		"entitlements_apicommunitymanager_enabled", "entitlements_monitoringcenter_productsku", "entitlements_apiquery_enabled",
		"entitlements_apiquery_productsku", "entitlements_apiqueryc360_enabled", "entitlements_anggovernance_level",
		"entitlements_crowd_hideapimanagerdesigner", "entitlements_crowd_hideformerapiplatform",
		"entitlements_crowd_environments", "entitlements_cam_enabled", "entitlements_exchange2_enabled",
		"entitlements_crowdselfservicemigration_enabled", "entitlements_kpidashboard_enabled", "entitlements_pcf",
		"entitlements_appviz", "entitlements_runtimefabric", "entitlements_anypointsecuritytokenization_enabled",
		"entitlements_anypointsecurityedgepolicies_enabled", "entitlements_runtimefabriccloud_enabled",
		"entitlements_servicemesh_enabled", "entitlements_messaging_assigned", "entitlements_workerclouds_assigned",
		"entitlements_workerclouds_reassigned", "owner_created_at", "owner_updated_at", "owner_organization_id",
		"owner_firstname", "owner_lastname", "owner_email", "owner_phonenumber", "owner_username", "owner_idprovider_id",
		"owner_enabled", "owner_deleted", "owner_lastlogin", "owner_mfaverification_excluded", "owner_mfaverifiers_configured",
		"owner_type", "session_timeout",
	}
	return attributes[:]
}

func getBGUpdatableAttributes() []string {
	attributes := [...]string{
		"name", "owner_id", "entitlements_createenvironments", "entitlements_createsuborgs",
		"entitlements_globaldeployment", "entitlements_vcoresproduction_assigned", "entitlements_vcoressandbox_assigned",
		"entitlements_vcoresdesign_assigned", "entitlements_vpcs_assigned", "entitlements_loadbalancer_assigned", "entitlements_vpns_assigned",
	}
	return attributes[:]
}
