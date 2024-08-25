package anypoint

import (
	"context"
	"io"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	application_manager_v2 "github.com/mulesoft-anypoint/anypoint-client-go/application_manager_v2"
)

var DeplApplicationConfigLoggingRTFDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"artifact_name": {
			Type:        schema.TypeString,
			Description: "The application name.",
			Computed:    true,
		},
		"scope_logging_configurations": {
			Type:        schema.TypeList,
			Description: "Additional log levels and categories to include in logs.",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"scope": {
						Type:        schema.TypeString,
						Description: "The logging package scope",
						Required:    true,
					},
					"log_level": {
						Type:        schema.TypeString,
						Description: "The application log level: INFO / DEBUG / WARNING / ERROR / FATAL",
						Required:    true,
						ValidateDiagFunc: validation.ToDiagFunc(
							validation.StringInSlice([]string{"INFO", "DEBUG", "WARNING", "ERROR", "FATAL"}, false),
						),
					},
				},
			},
		},
	},
}

var DeplApplicationConfigPropsRTFDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"application_name": {
			Type:        schema.TypeString,
			Description: "The application name",
			Computed:    true,
		},
		"properties": {
			Type:        schema.TypeMap,
			Description: "The mule application properties.",
			Optional:    true,
			DefaultFunc: func() (interface{}, error) { return make(map[string]string), nil },
		},
		"secure_properties": {
			Type:        schema.TypeMap,
			Description: "The mule application secured properties.",
			Optional:    true,
			DefaultFunc: func() (interface{}, error) { return make(map[string]string), nil },
		},
	},
}

var DeplApplicationConfigRTFDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"mule_agent_app_props_service": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Description: "The mule app properties",
			Elem:        DeplApplicationConfigPropsRTFDefinition,
			Required:    true,
		},
		"mule_agent_logging_service": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Description: "The mule app logging props",
			Elem:        DeplApplicationConfigLoggingRTFDefinition,
			Optional:    true,
		},
		"mule_agent_scheduling_service": {
			Type:        schema.TypeList,
			Description: "The mule app scheduling",
			Elem:        DeplApplicationConfigSchedulingReadOnlyDefinition,
			Computed:    true,
		},
	},
}

var DeplApplicationRefRTFDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"group_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The groupId of the application.",
		},
		"artifact_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The artifactId of the application.",
		},
		"version": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The version of the application.",
		},
		"packaging": {
			Type:     schema.TypeString,
			Required: true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringInSlice([]string{"jar"}, false),
			),
			Description: "The packaging of the application. Only 'jar' is supported.",
		},
	},
}

var DeplApplicationRTFDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of the application.",
		},
		"desired_state": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "STARTED",
			Description: "The desired state of the application.",
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringInSlice(
					[]string{
						"PARTIALLY_STARTED", "DEPLOYMENT_FAILED", "STARTING", "STARTED", "STOPPING",
						"STOPPED", "UNDEPLOYING", "UNDEPLOYED", "UPDATED", "APPLIED", "APPLYING", "FAILED", "DELETED",
					},
					false,
				),
			),
		},
		"ref": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Required: true,
			Description: `
			The reference to the artifact on Exchange that is to be deployed on Runtime Fabrics.
			Please ensure the application's artifact is deployed on Exchange before using this resource on Runtime Fabrics.
			`,
			Elem: DeplApplicationRefRTFDefinition,
		},
		"configuration": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Required:    true,
			Description: "The configuration of the application.",
			Elem:        DeplApplicationConfigRTFDefinition,
		},
		"vcores": {
			Type:        schema.TypeFloat,
			Description: "The allocated virtual cores.",
			Computed:    true,
		},
		"object_store_v2_enabled": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether object store v2 is enabled. Only for Cloudhub.",
		},
	},
}

var DeplTargetDeplSettHttpRTFDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"inbound_public_url": {
			Type: schema.TypeString,
			Description: `The ingress url(s).
			If you need to use multiple ingress urls, separete them with commas.
			example: http://example.mulesoft.terraform.net/(.+)
			`,
			Optional: true,
			Default:  "",
		},
		"inbound_path_rewrite": {
			Type:        schema.TypeString,
			Description: "The inbound path rewrite. This option is only available for Cloudhub 2.0 with private spaces",
			Computed:    true,
		},
		"inbound_last_mile_security": {
			Type:        schema.TypeBool,
			Description: "Last-mile security means that the connection between ingress and the actual Mule app will be HTTPS.",
			Optional:    true,
			Default:     false,
		},
		"inbound_forward_ssl_session": {
			Type:        schema.TypeBool,
			Description: "Whether to forward the ssl session.",
			Optional:    true,
			Default:     false,
		},
		"inbound_internal_url": {
			Type:        schema.TypeString,
			Description: "The inbound internal url.",
			Computed:    true,
		},
		"inbound_unique_id": {
			Type:        schema.TypeString,
			Description: "The inbound unique id.",
			Computed:    true,
		},
	},
}

var DeplTargetDeplSettRuntimeRTFDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"version": {
			Type: schema.TypeString,
			Description: `
			On deployment operations it can be set to:
				- a full image version with tag (i.e "4.6.0:40e-java17"),
				- a base version with a partial tag not indicating the java version (i.e. "4.6.0:40")
				- or only a base version (i.e. "4.6.0").
			Defaults to the latest image version.
			This field has precedence over the legacy 'target.deploymentSettings.runtimeVersion'.
			Learn more about Mule runtime release notes [here](https://docs.mulesoft.com/release-notes/runtime-fabric/runtime-fabric-runtimes-release-notes)
			`,
			Required: true,
		},
		"release_channel": {
			Type: schema.TypeString,
			Description: `
			On deployment operations it can be set to one of:
				- "LTS"
				- "EDGE"
				- "LEGACY".
			Defaults to "EDGE". This field has precedence over the legacy 'target.deploymentSettings.runtimeReleaseChannel'.
			Learn more on release channels [here](https://docs.mulesoft.com/release-notes/mule-runtime/lts-edge-release-cadence).
			`,
			Optional: true,
			Default:  "EDGE",
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringInSlice([]string{"LTS", "EDGE", "LEGACY"}, false),
			),
		},
		"java": {
			Type: schema.TypeString,
			Description: `
			On deployment operations it can be set to one of:
				- "8"
				- "17"
			Defaults to "8".
			Learn more about Java support [here](https://docs.mulesoft.com/general/java-support).
			`,
			Optional: true,
			Default:  "8",
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringInSlice([]string{"8", "17"}, false),
			),
		},
	},
}

var DeplTargetDeplSettResourcesRTFDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"cpu_limit": {
			Type:        schema.TypeString,
			Description: "The CPU limit",
			Required:    true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringMatch(
					regexp.MustCompile(`^\d+m$`),
					"field value should be a valid cpu representation. ex: 100m (= 0.1 vcores).",
				),
			),
		},
		"cpu_reserved": {
			Type:        schema.TypeString,
			Description: "The CPU reserved.",
			Required:    true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringMatch(
					regexp.MustCompile(`^\d+m$`),
					"field value should be a valid cpu representation. ex: 100m (= 0.1 vcores).",
				),
			),
		},
		"memory_limit": {
			Type:        schema.TypeString,
			Description: "The memory limit",
			Required:    true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringMatch(
					regexp.MustCompile(`^\d+Mi$`),
					"field value should be a valid memory representation. ex: 1000Mi (= 1Gb).",
				),
			),
		},
		"memory_reserved": {
			Type:        schema.TypeString,
			Description: "The memory reserved.",
			Required:    true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringMatch(
					regexp.MustCompile(`^\d+Mi$`),
					"field value should be a valid memory representation. ex: 1000Mi (= 1Gb).",
				),
			),
		},
		"storage_limit": {
			Type:        schema.TypeString,
			Description: "The storage limit",
			Computed:    true,
		},
		"storage_reserved": {
			Type:        schema.TypeString,
			Description: "The storage reserved",
			Computed:    true,
		},
	},
}

var DeplTargetDeplSettAutoscalingRTFDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"enabled": {
			Type:        schema.TypeBool,
			Description: "Enables or disables the Autoscaling feature. The possible values are: true or false.",
			Required:    true,
		},
		"min_replicas": {
			Type:             schema.TypeInt,
			Description:      "Set the minimum amount of replicas for your deployment. The minimum accepted value is 1. The maximum is 3.",
			Optional:         true,
			Default:          1,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 3)),
		},
		"max_replicas": {
			Type:             schema.TypeInt,
			Description:      "Set the maximum amount of replicas your application can scale to. The minimum accepted value is 2. The maximum is 32.",
			Optional:         true,
			Default:          2,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(2, 32)),
		},
	},
}

var DeplTargetDeploymentSettingsRTFDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"clustered": {
			Type:        schema.TypeBool,
			Description: "Whether the application is deployed in clustered mode.",
			Optional:    true,
			Default:     false,
		},
		"enforce_deploying_replicas_across_nodes": {
			Type:        schema.TypeBool,
			Description: "If true, forces the deployment of replicas across the RTF cluster. This option only available for Runtime Fabrics.",
			Optional:    true,
			Default:     false,
		},
		"http": {
			Type:        schema.TypeList,
			Description: "The details about http inbound or outbound configuration",
			Optional:    true,
			MaxItems:    1,
			DefaultFunc: func() (interface{}, error) {
				dict := make(map[string]interface{})
				dict["inbound_last_mile_security"] = false
				dict["inbound_forward_ssl_session"] = false
				return []interface{}{dict}, nil
			},
			Elem: DeplTargetDeplSettHttpRTFDefinition,
		},
		"jvm_args": {
			Type:        schema.TypeString,
			Description: "The java virtual machine arguments",
			Optional:    true,
			Default:     "",
		},
		"runtime": {
			Type:        schema.TypeList,
			Description: "The Mule app runtime version info.",
			Optional:    true,
			MaxItems:    1,
			Elem:        DeplTargetDeplSettRuntimeRTFDefinition,
		},
		"autoscaling": {
			Type: schema.TypeList,
			Description: `
			Use this object to provide CPU Based Horizontal Autoscaling configuration on deployment and redeployment operations. This object is optional.
			If Autoscaling is disabled and the fields "minReplicas" and "maxReplicas" are provided, they must match the value of "target.replicas" field.
			Learn more about Autoscaling [here](https://docs.mulesoft.com/cloudhub-2/ch2-configure-horizontal-autoscaling).
			`,
			Optional: true,
			MaxItems: 1,
			DefaultFunc: func() (interface{}, error) {
				dict := make(map[string]interface{})
				dict["enabled"] = false
				return []interface{}{dict}, nil
			},
			Elem: DeplTargetDeplSettAutoscalingRTFDefinition,
		},
		"update_strategy": {
			Type:        schema.TypeString,
			Description: "The mule app deployment update strategy: rolling or recreate",
			Optional:    true,
			Default:     "rolling",
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringInSlice([]string{"rolling", "recreate"}, false),
			),
		},
		"resources": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Description: "The mule app allocated resources.",
			Elem:        DeplTargetDeplSettResourcesRTFDefinition,
			Required:    true,
		},
		"disable_am_log_forwarding": {
			Type:        schema.TypeBool,
			Description: "Whether log forwarding is disabled.",
			Optional:    true,
			Default:     false,
		},
		"persistent_object_store": {
			Type:        schema.TypeBool,
			Description: "Whether persistent object store is enabled.",
			Optional:    true,
			Default:     false,
		},
		"anypoint_monitoring_scope": {
			Type:        schema.TypeString,
			Description: "The anypoint moniroting scope",
			Computed:    true,
		},
		"sidecars": {
			Type:        schema.TypeList,
			Description: "The mule app sidecars.",
			Elem:        DeplTargetDeplSettSidecarsReadOnlyDefinition,
			Computed:    true,
		},
		"disable_external_log_forwarding": {
			Type:        schema.TypeBool,
			Description: "Whether the log forwarding is disabled.",
			Optional:    true,
			Default:     false,
		},
		"tracing_enabled": {
			Type:        schema.TypeBool,
			Description: "Whether the log tracing is enabled.",
			Computed:    true,
		},
		"generate_default_public_url": {
			Type:        schema.TypeBool,
			Description: "Whether default public url should be generated.",
			Optional:    true,
			Default:     false,
		},
	},
}

var DeplTargetRTFDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"provider": {
			Type:        schema.TypeString,
			Description: "The cloud provider the target belongs to.",
			Optional:    true,
			Default:     "MC",
			ForceNew:    true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringInSlice([]string{"MC"}, false),
			),
		},
		"target_id": {
			Type:        schema.TypeString,
			Description: "The unique identifier of the Runtime Fabrics target.",
			Required:    true,
			ForceNew:    true,
		},
		"deployment_settings": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Description: "The settings of the target for the deployment to perform.",
			Required:    true,
			Elem:        DeplTargetDeploymentSettingsRTFDefinition,
		},
		"replicas": {
			Type:        schema.TypeInt,
			Description: "The number of replicas. Default is 1.",
			Optional:    true,
			Default:     1,
		},
	},
}

func resourceRTFDeployment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRTFDeploymentCreate,
		ReadContext:   resourceRTFDeploymentRead,
		UpdateContext: resourceRTFDeploymentUpdate,
		DeleteContext: resourceRTFDeploymentDelete,
		Description: `
		Creates and manages a ` + "`" + `deployment` + "`" + ` of a mule app on Runtime Fabrics only.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique id of the mule app deployment in the platform.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organization where the mule app is deployed.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The environment where mule app is deployed.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the deployed mule app.",
			},
			"creation_date": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The creation date of the mule app.",
			},
			"last_modified_date": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The last modification date of the mule app.",
			},
			"desired_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The deployment desired version of the mule app.",
			},
			"replicas": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Data of the mule app replicas",
				Elem:        ReplicasReadOnlyDefinition,
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data of the mule app replicas",
			},
			"application": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "The details of the application to deploy",
				Elem:        DeplApplicationRTFDefinition,
			},
			"target": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "The details of the target to perform the deployment on.",
				Elem:        DeplTargetRTFDefinition,
			},
			"last_successful_version": {
				Type:        schema.TypeString,
				Description: "The last successfully deployed version",
				Computed:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceRTFDeploymentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	name := d.Get("name").(string)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	authctx := getAppDeploymentV2AuthCtx(ctx, &pco)
	body := newRTFDeploymentBody(d)
	//Execute post deployment
	res, httpr, err := pco.appmanagerclient.DefaultApi.PostDeployment(authctx, orgid, envid).DeploymentRequestBody(*body).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create " + name + " deployment for runtime fabrics.",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.SetId(res.GetId())
	return resourceRTFDeploymentRead(ctx, d, m)
}

func resourceRTFDeploymentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	id := d.Id()
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	if isComposedResourceId(id) {
		orgid, envid, id = decomposeRTFDeploymentId(d)
	}
	authctx := getAppDeploymentV2AuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.appmanagerclient.DefaultApi.GetDeploymentById(authctx, orgid, envid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read runtime fabrics deployment " + id + ".",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	//process data
	data := flattenAppDeploymentV2(res)
	if err := setAppDeploymentV2AttributesToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set App Deployment details attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	// setting all params required for reading in case of import
	d.SetId(res.GetId())
	d.Set("org_id", orgid)
	d.Set("env_id", envid)

	return diags
}

func resourceRTFDeploymentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if !d.HasChanges(getRTFDeploymentUpdatableAttributes()...) {
		return diags
	}
	pco := m.(ProviderConfOutput)
	id := d.Id()
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	name := d.Get("name").(string)
	authctx := getAppDeploymentV2AuthCtx(ctx, &pco)
	body := newRTFDeploymentBody(d)
	_, httpr, err := pco.appmanagerclient.DefaultApi.PatchDeployment(authctx, orgid, envid, id).DeploymentRequestBody(*body).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to update deployment " + name + " on runtime fabrics.",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	return resourceRTFDeploymentRead(ctx, d, m)
}

func resourceRTFDeploymentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	id := d.Id()
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	name := d.Get("name").(string)
	authctx := getAppDeploymentV2AuthCtx(ctx, &pco)
	httpr, err := pco.appmanagerclient.DefaultApi.DeleteDeployment(authctx, orgid, envid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete deployment " + name + " on cloudhub 2.0 shared-space.",
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

// Prepares Deployment Post Body out of resource data input
func newRTFDeploymentBody(d *schema.ResourceData) *application_manager_v2.DeploymentRequestBody {
	body := application_manager_v2.NewDeploymentRequestBody()
	// -- Parsing Application
	app_list_d := d.Get("application").([]interface{})
	app_d := app_list_d[0].(map[string]interface{})
	application := newRTFDeploymentApplication(app_d)
	// -- Parsing Target
	target_list_d := d.Get("target").([]interface{})
	target_d := target_list_d[0].(map[string]interface{})
	target := newRTFDeploymentTarget(target_d)
	//Set Body Data
	body.SetName(d.Get("name").(string))
	body.SetApplication(*application)
	body.SetTarget(*target)

	return body
}

// Prepares Application object out of map input
func newRTFDeploymentApplication(app_d map[string]interface{}) *application_manager_v2.Application {
	ref_list_d := app_d["ref"].([]interface{})
	ref_d := ref_list_d[0].(map[string]interface{})
	// Ref
	ref := newRTFDeploymentRef(ref_d)
	//Parse Configuration
	configuration_list_d := app_d["configuration"].([]interface{})
	configuration_d := configuration_list_d[0].(map[string]interface{})
	configuration := newRTFDeploymentConfiguration(configuration_d)
	//Object Store V2
	// object_store_v2_enabled_d := app_d["object_store_v2_enabled"].(bool)
	//Application Integration
	// integrations := application_manager_v2.NewApplicationIntegrations()
	// object_store_v2 := application_manager_v2.NewObjectStoreV2()
	// object_store_v2.SetEnabled(object_store_v2_enabled_d)
	// services := application_manager_v2.NewServices()
	// services.SetObjectStoreV2(*object_store_v2)
	// integrations.SetServices(*services)
	//Application
	application := application_manager_v2.NewApplication()
	application.SetDesiredState(app_d["desired_state"].(string))
	application.SetConfiguration(*configuration)
	// application.SetIntegrations(*integrations)
	application.SetRef(*ref)

	return application
}

// Prepares Target object out of map input
func newRTFDeploymentTarget(target_d map[string]interface{}) *application_manager_v2.Target {
	deployment_settings_list_d := target_d["deployment_settings"].([]interface{})
	deployment_settings_d := deployment_settings_list_d[0].(map[string]interface{})
	deployment_settings := newRTFDeploymentDeploymentSettings(deployment_settings_d)
	//Prepare Target data
	target := application_manager_v2.NewTarget()
	target.SetProvider(target_d["provider"].(string))
	target.SetTargetId(target_d["target_id"].(string))
	target.SetDeploymentSettings(*deployment_settings)
	target.SetReplicas(int32(target_d["replicas"].(int)))

	return target
}

// Prepares Ref Object out of map input
func newRTFDeploymentRef(ref_d map[string]interface{}) *application_manager_v2.Ref {
	ref := application_manager_v2.NewRef()
	ref.SetGroupId(ref_d["group_id"].(string))
	ref.SetArtifactId(ref_d["artifact_id"].(string))
	ref.SetVersion(ref_d["version"].(string))
	ref.SetPackaging(ref_d["packaging"].(string))
	return ref
}

// Prepares Application Configuration Object out of map input
func newRTFDeploymentConfiguration(configuration_d map[string]interface{}) *application_manager_v2.AppConfiguration {
	//Mule Agent App Properties Service
	mule_agent_app_props_service_list_d := configuration_d["mule_agent_app_props_service"].([]interface{})
	mule_agent_app_props_service_d := mule_agent_app_props_service_list_d[0].(map[string]interface{})
	mule_agent_app_props_service_properties := mule_agent_app_props_service_d["properties"].(map[string]interface{})
	mule_agent_app_props_service_secure_properties := mule_agent_app_props_service_d["secure_properties"].(map[string]interface{})
	mule_agent_app_props_service := application_manager_v2.NewMuleAgentAppPropService()
	mule_agent_app_props_service.SetProperties(mule_agent_app_props_service_properties)
	mule_agent_app_props_service.SetSecureProperties(mule_agent_app_props_service_secure_properties)
	mule_agent_logging_service_list_d := configuration_d["mule_agent_logging_service"].([]interface{})
	mule_agent_logging_service_d := mule_agent_logging_service_list_d[0].(map[string]interface{})
	//Scope logging configuration
	scope_logging_configurations_list_d := mule_agent_logging_service_d["scope_logging_configurations"].([]interface{})
	scope_logging_configurations := make([]application_manager_v2.ScopeLoggingConfiguration, len(scope_logging_configurations_list_d))
	for i, item := range scope_logging_configurations_list_d {
		data := item.(map[string]interface{})
		conf := application_manager_v2.NewScopeLoggingConfiguration()
		conf.SetScope(data["scope"].(string))
		conf.SetLogLevel(data["log_level"].(string))
		scope_logging_configurations[i] = *conf
	}
	//Mule Agent Logging Service
	mule_agent_logging_service := application_manager_v2.NewMuleAgentLoggingService()
	mule_agent_logging_service.SetScopeLoggingConfigurations(scope_logging_configurations)
	configuration := application_manager_v2.NewAppConfiguration()
	configuration.SetMuleAgentApplicationPropertiesService(*mule_agent_app_props_service)
	configuration.SetMuleAgentLoggingService(*mule_agent_logging_service)

	return configuration
}

// Prepares DeploymentSettings object out of map input
func newRTFDeploymentDeploymentSettings(deployment_settings_d map[string]interface{}) *application_manager_v2.DeploymentSettings {
	//http
	http := newRTFDeploymentHttp(deployment_settings_d)
	//runtime
	runtime := newRTFDeploymentRuntime(deployment_settings_d)
	//autoscaling
	autoscaling := newRTFDeploymentAutoscaling(deployment_settings_d)
	//resources
	resources := newRTFDeploymentResources(deployment_settings_d)
	//Prepare JVM Args data
	jvm := application_manager_v2.NewJvm()
	jvm.SetArgs(deployment_settings_d["jvm_args"].(string))
	deployment_settings := application_manager_v2.NewDeploymentSettings()
	deployment_settings.SetClustered(deployment_settings_d["clustered"].(bool))
	deployment_settings.SetEnforceDeployingReplicasAcrossNodes(deployment_settings_d["enforce_deploying_replicas_across_nodes"].(bool))
	deployment_settings.SetHttp(*http)
	deployment_settings.SetJvm(*jvm)
	deployment_settings.SetUpdateStrategy(deployment_settings_d["update_strategy"].(string))
	deployment_settings.SetDisableAmLogForwarding(deployment_settings_d["disable_am_log_forwarding"].(bool))
	deployment_settings.SetPersistentObjectStore(deployment_settings_d["persistent_object_store"].(bool))
	deployment_settings.SetDisableExternalLogForwarding(deployment_settings_d["disable_external_log_forwarding"].(bool))
	deployment_settings.SetGenerateDefaultPublicUrl(deployment_settings_d["generate_default_public_url"].(bool))
	deployment_settings.SetRuntime(*runtime)
	deployment_settings.SetAutoscaling(*autoscaling)
	deployment_settings.SetResources(*resources)

	return deployment_settings
}

// Prepares Runtime object out of map input
func newRTFDeploymentRuntime(deployment_settings_d map[string]interface{}) *application_manager_v2.Runtime {
	runtime := application_manager_v2.NewRuntime()
	if val, ok := deployment_settings_d["runtime"]; ok {
		runtime_list_d := val.([]interface{})
		if len(runtime_list_d) > 0 {
			runtime_d := runtime_list_d[0].(map[string]interface{})
			runtime.SetVersion(runtime_d["version"].(string))
			runtime.SetReleaseChannel(runtime_d["release_channel"].(string))
			runtime.SetJava(runtime_d["java"].(string))
		}

	}
	return runtime
}

// Prepares Http object out of map input
func newRTFDeploymentHttp(deployment_settings_d map[string]interface{}) *application_manager_v2.Http {
	http_inbound := application_manager_v2.NewHttpInbound()
	http := application_manager_v2.NewHttp()
	if val, ok := deployment_settings_d["http"]; ok {
		http_list_d := val.([]interface{})
		if len(http_list_d) > 0 {
			http_d := http_list_d[0].(map[string]interface{})
			http_inbound.SetLastMileSecurity(http_d["inbound_last_mile_security"].(bool))
			http_inbound.SetForwardSslSession(http_d["inbound_forward_ssl_session"].(bool))
			http.SetInbound(*http_inbound)
		}
	}
	return http
}

func newRTFDeploymentAutoscaling(deployment_settings_d map[string]interface{}) *application_manager_v2.Autoscaling {
	autoscaling := application_manager_v2.NewAutoscaling()
	if val, ok := deployment_settings_d["autoscaling"]; ok {
		autoscaling_list_d := val.([]interface{})
		if len(autoscaling_list_d) > 0 {
			autoscaling_d := autoscaling_list_d[0].(map[string]interface{})
			autoscaling.SetEnabled(autoscaling_d["enabled"].(bool))
			autoscaling.SetMinReplicas(int32(autoscaling_d["min_replicas"].(int)))
			autoscaling.SetMaxReplicas(int32(autoscaling_d["max_replicas"].(int)))
		}
	}
	return autoscaling
}

func newRTFDeploymentResources(deployment_settings_d map[string]interface{}) *application_manager_v2.Resources {
	resources := application_manager_v2.NewResources()
	if val, ok := deployment_settings_d["resources"]; ok {
		resources_list_d := val.([]interface{})
		if len(resources_list_d) > 0 {
			resources_d := resources_list_d[0].(map[string]interface{})
			cpu := application_manager_v2.NewResourcesCpu()
			cpu.SetLimit(resources_d["cpu_limit"].(string))
			cpu.SetReserved(resources_d["cpu_reserved"].(string))
			memory := application_manager_v2.NewResourcesMemory()
			memory.SetLimit(resources_d["memory_limit"].(string))
			memory.SetReserved(resources_d["memory_reserved"].(string))
			resources.SetCpu(*cpu)
			resources.SetMemory(*memory)
		}
	}
	return resources
}

func decomposeRTFDeploymentId(d *schema.ResourceData) (string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2]
}

func getRTFDeploymentUpdatableAttributes() []string {
	attributes := [...]string{"application", "target"}
	return attributes[:]
}
