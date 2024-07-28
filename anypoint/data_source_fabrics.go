package anypoint

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rtf "github.com/mulesoft-anypoint/anypoint-client-go/rtf"
)

var NodeCapacityDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"cpu": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"cpu_millis": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"memory": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"memory_mi": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"pods": {
			Type:     schema.TypeInt,
			Computed: true,
		},
	},
}

var NodeStatusDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"is_healthy": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"is_ready": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"is_schedulable": {
			Type:     schema.TypeBool,
			Computed: true,
		},
	},
}

var NodeDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"uid": {
			Type:        schema.TypeString,
			Description: "The node id",
			Computed:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Description: "The node name",
			Computed:    true,
		},
		"kubelet_version": {
			Type:        schema.TypeString,
			Description: "The kubelet version of the node",
			Computed:    true,
		},
		"docker_version": {
			Type:        schema.TypeString,
			Description: "The docker version",
			Computed:    true,
		},
		"role": {
			Type:        schema.TypeString,
			Description: "The role of the node in the cluster",
			Computed:    true,
		},
		"status": {
			Type:        schema.TypeList,
			Description: "The status of the node",
			Computed:    true,
			Elem:        NodeStatusDefinition,
		},
		"capacity": {
			Type:        schema.TypeList,
			Description: "The capacity of the node",
			Computed:    true,
			Elem:        NodeCapacityDefinition,
		},
		"allocated_request_capacity": {
			Type:        schema.TypeList,
			Description: "The allocated request capacity of the node",
			Computed:    true,
			Elem:        NodeCapacityDefinition,
		},
		"allocated_limit_capacity": {
			Type:        schema.TypeList,
			Description: "The allocated limit capacity of the node",
			Computed:    true,
			Elem:        NodeCapacityDefinition,
		},
	},
}

var FabricsFeaturesDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"enhanced_security": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether enhanced security feature is active",
		},
		"persistent_store": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether peristent store feature is active",
		},
	},
}

var FabricsIngressDomainsDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"domains": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "The list of domains.",
			Elem:        schema.TypeString,
		},
	},
}

var FabricsUpgradeDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The upgrade status.",
		},
	},
}

func dataSourceFabrics() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFabricsRead,
		Description: `
		Reads a specific ` + "`" + `Runtime Fabrics'` + "`" + ` instance.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique id of the fabrics instance in the platform.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the fabrics is hosted.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of this fabrics instance.",
			},
			"region": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The region where fabrics instance is hosted.",
			},
			"vendor": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The vendor name of the kubernetes instance hosting fabrics.",
			},
			"vendor_metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The vendor metadata",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of fabrics.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the farbics instance.",
			},
			"desired_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The desired version of fabrics.",
			},
			"available_upgrade_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The available upgrade version of fabrics.",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The creation date of the fabrics instance.",
			},
			"upgrade": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The status of the fabrics. Only available when instance is created and not activated yet.",
				Elem:        FabricsUpgradeDefinition,
			},
			"nodes": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        NodeDefinition,
				Description: "The list of fabrics nodes.",
			},
			"activation_data": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The activation data to use during installation of fabrics on the kubernetes cluster. Only available when instance is created and not activated yet.",
			},
			"seconds_since_heartbeat": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of seconds since last heartbeat.",
			},
			"kubernetes_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The kubernetes version of the cluster.",
			},
			"namespace": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The namespace where runtime fabrics is installed.",
			},
			"license_expiry_date": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The expiry date of the license (timestamp).",
			},
			"is_managed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this cluster is managed.",
			},
			"is_helm_managed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this cluster is managed by helmet.",
			},
			"app_scoped_log_forwarding": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether app scoped log forwarding is active.",
			},
			"cluster_configuration_level": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The configuration level of the cluster (production or development).",
			},
			"features": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The features of this cluster.",
				Elem:        FabricsFeaturesDefinition,
			},
			"ingress": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The ingress configurations of this cluster.",
				Elem:        FabricsIngressDomainsDefinition,
			},
		},
	}
}

func dataSourceFabricsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	id := d.Get("id").(string)
	orgid := d.Get("org_id").(string)
	authctx := getFabricsAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.rtfclient.DefaultApi.GetFabrics(authctx, orgid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get fabrics " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	data := flattenFabricsData(res)
	if err := setFabricsResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set fabrics " + id + " attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(id)

	return diags
}

func flattenFabricsData(fabrics *rtf.Fabrics) map[string]interface{} {
	mappedItem := make(map[string]interface{})

	mappedItem["id"] = fabrics.GetId()
	mappedItem["name"] = fabrics.GetName()
	mappedItem["region"] = fabrics.GetRegion()
	mappedItem["vendor"] = fabrics.GetVendor()
	if val, ok := fabrics.GetVendorMetadataOk(); ok {
		mappedItem["vendor_metadata"] = val
	}
	if val, ok := fabrics.GetOrganizationIdOk(); ok {
		mappedItem["org_id"] = *val
	}
	if val, ok := fabrics.GetVersionOk(); ok {
		mappedItem["version"] = *val
	}
	if val, ok := fabrics.GetStatusOk(); ok {
		mappedItem["status"] = *val
	}
	if val, ok := fabrics.GetDesiredVersionOk(); ok {
		mappedItem["desired_version"] = *val
	}
	if val, ok := fabrics.GetAvailableUpgradeVersionOk(); ok {
		mappedItem["available_upgrade_version"] = *val
	}
	if val, ok := fabrics.GetCreatedAtOk(); ok {
		mappedItem["created_at"] = *val
	}
	if val, ok := fabrics.GetUpgradeOk(); ok {
		mappedItem["upgrade"] = flattenFabricsUpgradeData(val)
	} else {
		mappedItem["upgrade"] = []interface{}{}
	}
	if val, ok := fabrics.GetNodesOk(); ok {
		mappedItem["nodes"] = flattenFabricsNodesData(val)
	}
	if val, ok := fabrics.GetActivationDataOk(); ok {
		mappedItem["activation_data"] = *val
	}
	if val, ok := fabrics.GetSecondsSinceHeartbeatOk(); ok {
		mappedItem["seconds_since_heartbeat"] = *val
	}
	if val, ok := fabrics.GetKubernetesVersionOk(); ok {
		mappedItem["kubernetes_version"] = *val
	}
	if val, ok := fabrics.GetNamespaceOk(); ok {
		mappedItem["namespace"] = *val
	}
	if val, ok := fabrics.GetLicenseExpiryDateOk(); ok {
		mappedItem["license_expiry_date"] = *val
	}
	if val, ok := fabrics.GetIsManagedOk(); ok {
		mappedItem["is_managed"] = *val
	}
	if val, ok := fabrics.GetIsHelmManagedOk(); ok {
		mappedItem["is_helm_managed"] = *val
	}
	if val, ok := fabrics.GetAppScopedLogForwardingOk(); ok {
		mappedItem["app_scoped_log_forwarding"] = *val
	}
	if val, ok := fabrics.GetClusterConfigurationLevelOk(); ok {
		mappedItem["cluster_configuration_level"] = *val
	}
	if val, ok := fabrics.GetFeaturesOk(); ok {
		mappedItem["features"] = flattenFabricsFeaturesData(val)
	}
	if val, ok := fabrics.GetIngressOk(); ok {
		mappedItem["ingress"] = flattenFabricsIngressData(val)
	}

	return mappedItem
}

func flattenFabricsUpgradeData(upgrade *rtf.FabricsUpgrade) []interface{} {
	if upgrade == nil {
		return []interface{}{}
	}
	data := make(map[string]interface{})
	data["status"] = upgrade.GetStatus()

	return []interface{}{data}
}

func flattenFabricsFeaturesData(features *rtf.Features) []interface{} {
	if features == nil {
		return []interface{}{}
	}
	data := make(map[string]interface{})
	data["enhanced_security"] = features.GetEnhancedSecurity()
	data["persistent_store"] = features.GetPersistentStore()

	return []interface{}{data}
}

func flattenFabricsIngressData(ingress *rtf.Ingress) []interface{} {
	if ingress == nil {
		return []interface{}{}
	}
	data := make(map[string]interface{})
	data["domains"] = ingress.GetDomains()

	return []interface{}{data}
}

func flattenFabricsNodesData(nodes []rtf.FabricsNode) []interface{} {
	if len(nodes) == 0 {
		return make([]interface{}, 0)
	}

	res := make([]interface{}, len(nodes))
	for i, node := range nodes {
		item := make(map[string]interface{})

		if val, ok := node.GetUidOk(); ok {
			item["uid"] = *val
		}
		if val, ok := node.GetNameOk(); ok {
			item["name"] = *val
		}
		if val, ok := node.GetKubeletVersionOk(); ok {
			item["kubelet_version"] = *val
		}
		if val, ok := node.GetDockerVersionOk(); ok {
			item["docker_version"] = *val
		}
		if val, ok := node.GetRoleOk(); ok {
			item["role"] = *val
		}
		if val, ok := node.GetStatusOk(); ok {
			item["status"] = flattenFabricsNodeStatusData(val)
		}
		if val, ok := node.GetCapacityOk(); ok {
			item["capacity"] = flattenFabricsNodeCapacityData(val)
		}
		if val, ok := node.GetAllocatedRequestCapacityOk(); ok {
			item["allocated_request_capacity"] = flattenFabricsNodeAllocReqCapacityData(val)
		}
		if val, ok := node.GetAllocatedLimitCapacityOk(); ok {
			item["allocated_limit_capacity"] = flattenFabricsNodeAllocLimitCapacityData(val)
		}

		res[i] = item
	}

	return res
}

func flattenFabricsNodeStatusData(status *rtf.Status) []interface{} {
	if status == nil {
		return []interface{}{}
	}
	data := make(map[string]interface{})
	data["is_healthy"] = status.GetIsHealthy()
	data["is_ready"] = status.GetIsReady()
	data["is_schedulable"] = status.GetIsSchedulable()

	return []interface{}{data}
}

func flattenFabricsNodeAllocReqCapacityData(capacity *rtf.AllocatedRequestCapacity) []interface{} {
	if capacity == nil {
		return []interface{}{}
	}
	data := make(map[string]interface{})
	data["cpu"] = capacity.GetCpu()
	data["cpu_millis"] = capacity.GetCpuMillis()
	data["memory"] = capacity.GetMemory()
	data["memory_mi"] = capacity.GetMemoryMi()
	data["pods"] = capacity.GetPods()

	return []interface{}{data}
}

func flattenFabricsNodeAllocLimitCapacityData(capacity *rtf.AllocatedLimitCapacity) []interface{} {
	if capacity == nil {
		return []interface{}{}
	}
	data := make(map[string]interface{})
	data["cpu"] = capacity.GetCpu()
	data["cpu_millis"] = capacity.GetCpuMillis()
	data["memory"] = capacity.GetMemory()
	data["memory_mi"] = capacity.GetMemoryMi()
	data["pods"] = capacity.GetPods()

	return []interface{}{data}
}

func flattenFabricsNodeCapacityData(capacity *rtf.Capacity) []interface{} {
	if capacity == nil {
		return []interface{}{}
	}
	data := make(map[string]interface{})
	data["cpu"] = capacity.GetCpu()
	data["cpu_millis"] = capacity.GetCpuMillis()
	data["memory"] = capacity.GetMemory()
	data["memory_mi"] = capacity.GetMemoryMi()
	data["pods"] = capacity.GetPods()

	return []interface{}{data}
}

func setFabricsResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getFabricsAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set fabrics attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getFabricsAttributes() []string {
	attributes := [...]string{
		"id", "org_id", "name", "region", "vendor", "vendor_metadata", "version",
		"status", "desired_version", "available_upgrade_version", "created_at",
		"nodes", "seconds_since_heartbeat", "kubernetes_version", "namespace",
		"license_expiry_date", "is_managed", "is_helm_managed", "app_scoped_log_forwarding",
		"cluster_configuration_level", "features", "ingress",
	}
	return attributes[:]
}
