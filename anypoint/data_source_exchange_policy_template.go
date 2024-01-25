package anypoint

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/apim_policy"
)

var EXCHANGE_POLICY_TEMPLATE_CONFIG = map[string]*schema.Schema{
	"property_name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The property name.",
	},
	"name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The property name.",
	},
	"description": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The property description.",
	},
	"type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The property type.",
	},
	"options": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "The property options.",
		Elem: &schema.Schema{
			Type: schema.TypeMap,
		},
	},
	"optional": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the property is optional.",
	},
	"default_value": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The property default value.",
	},
	"sensitive": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the property is sensitive.",
	},
	"allow_multiple": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the property allows multiple values.",
	},
	"configuration": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "The property configuration.",
		Elem: &schema.Schema{
			Type: schema.TypeMap,
		},
	},
}

var EXCHANGE_POLICY_TEMPLATE_ALL_VERSIONS = map[string]*schema.Schema{
	"group_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The group id.",
	},
	"asset_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The asset id.",
	},
	"version": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The version.",
	},
}

func dataSourceExchangePolicyTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceExchangePolicyTemplateRead,
		Description: `
		Query a specific exchange policy template.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The exchange policy template id.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id.",
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The policy template group id in exchange.",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The policy template version.",
			},
			"include_all_versions": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to include all versions of the asset.",
			},
			"audit": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The exchange policy template auditing data.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy template name.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy template description.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy template type.",
			},
			"is_ootb": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the policy template is out of the box.",
			},
			"stage": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy template stage.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy template status.",
			},
			"yaml_md5": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The yaml file checksum for data integrity.",
			},
			"jar_md5": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The jar file checksum for data integrity.",
			},
			"min_mule_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The minimum mule version to use the policy.",
			},
			"supported_policies_versions": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The supported policies versions.",
			},
			"category": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The category of the policy template.",
			},
			"violation_category": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The violation category of the policy template.",
			},
			"resource_level_supported": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the policy template supports resource level.",
			},
			"encryption_supported": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the policy template supports encryption.",
			},
			"standalone": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the policy template is standalone.",
			},
			"required_characteristics": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of required characteristics.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"identity_management_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of identity management.",
			},
			"provided_characteristics": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of provided characteristics.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"raml_snippet": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy template snippet in RAML.",
			},
			"raml_v1_snippet": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy template snippet in RAML v1.",
			},
			"oas_v2_snippet": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy template snippet in OAS v2.",
			},
			"oas_v3_snippet": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy template snippet in OAS v3.",
			},
			"applicable": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the template snippet is applicable.",
			},
			"configuration": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The policy template list of property configurations.",
				Elem: &schema.Resource{
					Schema: EXCHANGE_POLICY_TEMPLATE_CONFIG,
				},
			},
			"all_versions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The policy template list of versions.",
				Elem: &schema.Resource{
					Schema: EXCHANGE_POLICY_TEMPLATE_ALL_VERSIONS,
				},
			},
		},
	}
}

func dataSourceExchangePolicyTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	groupid := d.Get("group_id").(string)
	version := d.Get("version").(string)
	id := d.Get("id").(string)
	include_all_versions := d.Get("include_all_versions").(bool)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.apimpolicyclient.DefaultApi.GetOrgExchangePolicyTemplateDetails(authctx, orgid, groupid, id, version).IncludeAllVersions(include_all_versions).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get policy template " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	data := flattenExchangePolicyTemplate(res)
	if err := setApimExchPolicyTempToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set policy template " + id + " details attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(id)
	return diags
}

func flattenExchangePolicyTemplate(template *apim_policy.ExchangePolicyTemplate) map[string]interface{} {
	result := make(map[string]interface{})
	if val, ok := template.GetIdOk(); ok {
		result["id"] = strconv.Itoa(int(*val))
	}
	if val, ok := template.GetAuditOk(); ok {
		result["audit"] = flattenApimInstancePolicyAudit(val)
	}
	if val, ok := template.GetGroupIdOk(); ok {
		result["group_id"] = *val
	}
	if val, ok := template.GetAssetIdOk(); ok {
		result["asset_id"] = *val
	}
	if val, ok := template.GetVersionOk(); ok {
		result["version"] = *val
	}
	if val, ok := template.GetNameOk(); ok {
		result["name"] = *val
	}
	if val, ok := template.GetDescriptionOk(); ok {
		result["description"] = *val
	}
	if val, ok := template.GetTypeOk(); ok {
		result["type"] = *val
	}
	if val, ok := template.GetIsOOTBOk(); ok {
		result["is_ootb"] = *val
	}
	if val, ok := template.GetStageOk(); ok {
		result["stage"] = *val
	}
	if val, ok := template.GetStatusOk(); ok {
		result["status"] = *val
	}
	if val, ok := template.GetYamlMd5Ok(); ok {
		result["yaml_md5"] = *val
	}
	if val, ok := template.GetJarMd5Ok(); ok {
		result["jar_md5"] = *val
	}
	if val, ok := template.GetOrgIdOk(); ok {
		result["org_id"] = *val
	}
	if val, ok := template.GetMinMuleVersionOk(); ok {
		result["min_mule_version"] = *val
	}
	if val, ok := template.GetSupportedPoliciesVersionsOk(); ok {
		result["supported_policies_versions"] = *val
	}
	if val, ok := template.GetCategoryOk(); ok {
		result["category"] = *val
	}
	if val, ok := template.GetViolationCategoryOk(); ok {
		result["violation_category"] = *val
	}
	if val, ok := template.GetResourceLevelSupportedOk(); ok {
		result["resource_level_supported"] = *val
	}
	if val, ok := template.GetEncryptionSupportedOk(); ok {
		result["encryption_supported"] = *val
	}
	if val, ok := template.GetStandaloneOk(); ok {
		result["standalone"] = *val
	}
	if val, ok := template.GetRequiredCharacteristicsOk(); ok {
		result["required_characteristics"] = val
	}
	if idp, ok := template.GetIdentityManagementOk(); ok {
		result["identity_management_type"] = idp.GetType()
	}
	if val, ok := template.GetProvidedCharacteristicsOk(); ok {
		result["provided_characteristics"] = val
	}
	if val, ok := template.GetRamlSnippetOk(); ok {
		result["raml_snippet"] = *val
	}
	if val, ok := template.GetRamlV1SnippetOk(); ok {
		result["raml_v1_snippet"] = *val
	}
	if val, ok := template.GetOasV2SnippetOk(); ok {
		result["oas_v2_snippet"] = *val
	}
	if val, ok := template.GetOasV3SnippetOk(); ok {
		result["oas_v3_snippet"] = *val
	}
	if val, ok := template.GetApplicableOk(); ok {
		result["applicable"] = *val
	}
	if conf, ok := template.GetConfigurationOk(); ok {
		result["configuration"] = flattenExchPolicyTempConfigs(conf)
	}
	if val, ok := template.GetAllVersionsOk(); ok {
		result["all_versions"] = flattenExchPolicyTempAllVersions(val)
	}
	return result
}

func flattenExchPolicyTempConfigs(collection []apim_policy.PolicyConfiguration) []interface{} {
	slice := make([]interface{}, len(collection))
	for i, conf := range collection {
		slice[i] = flattenExchPolicyTempConfig(&conf)
	}
	return slice
}

func flattenExchPolicyTempConfig(conf *apim_policy.PolicyConfiguration) map[string]interface{} {
	result := make(map[string]interface{})
	if val, ok := conf.GetPropertyNameOk(); ok {
		result["property_name"] = *val
	}
	if val, ok := conf.GetNameOk(); ok {
		result["name"] = *val
	}
	if val, ok := conf.GetDescriptionOk(); ok {
		result["description"] = *val
	}
	if val, ok := conf.GetTypeOk(); ok {
		result["type"] = *val
	}
	if opts, ok := conf.GetOptionsOk(); ok {
		result["options"] = flattenExchPolicyTempConfigOpts(opts)
	}
	if val, ok := conf.GetOptionalOk(); ok {
		result["optional"] = *val
	}
	if val, ok := conf.GetSensitiveOk(); ok {
		result["sensitive"] = *val
	}
	if val, ok := conf.GetAllowMultipleOk(); ok {
		result["allow_multiple"] = *val
	}
	if val, ok := conf.GetConfigurationOk(); ok {
		result["configuration"] = flattenExchPolicyTempConfigOptsConfig(val)
	}
	return result
}

func flattenExchPolicyTempConfigOpts(opts []map[string]interface{}) []interface{} {
	slice := make([]interface{}, len(opts))
	for i, opt := range opts {
		data := make(map[string]interface{})
		if val, ok := opt["name"]; ok {
			data["name"] = ConvPrimtiveInterface2String(val)
		}
		if val, ok := opt["value"]; ok {
			data["value"] = ConvPrimtiveInterface2String(val)
		}
		slice[i] = data
	}
	return slice
}

func flattenExchPolicyTempConfigOptsConfig(collection []apim_policy.PolicyConfigurationConfigurationInner) []interface{} {
	slice := make([]interface{}, len(collection))
	for i, c := range collection {
		data := make(map[string]interface{})
		if val, ok := c.GetPropertyNameOk(); ok {
			data["property_name"] = *val
		}
		if val, ok := c.GetTypeOk(); ok {
			data["type"] = *val
		}
		slice[i] = data
	}
	return slice
}

func flattenExchPolicyTempAllVersions(collection []apim_policy.ExchangePolicyTemplateAllVersionsInner) []interface{} {
	slice := make([]interface{}, len(collection))
	for i, version := range collection {
		data := make(map[string]interface{})
		if val, ok := version.GetGroupIdOk(); ok {
			data["group_id"] = *val
		}
		if val, ok := version.GetAssetIdOk(); ok {
			data["asset_id"] = *val
		}
		if val, ok := version.GetVersionOk(); ok {
			data["version"] = *val
		}
		slice[i] = data
	}
	return slice
}

func setApimExchPolicyTempToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getExchPolicyTempDetailsAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set policy template attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getExchPolicyTempDetailsAttributes() []string {
	attributes := [...]string{
		"audit", "name", "description", "type", "is_ootb",
		"stage", "status", "yaml_md5", "jar_md5", "min_mule_version",
		"supported_policies_versions", "category", "violation_category",
		"resource_level_supported", "encryption_supported", "standalone",
		"required_characteristics", "identity_management_type",
		"provided_characteristics", "raml_snippet", "raml_v1_snippet",
		"oas_v2_snippet", "oas_v3_snippet", "applicable", "configuration",
		"all_versions",
	}
	return attributes[:]
}
