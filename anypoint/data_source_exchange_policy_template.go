package anypoint

import (
	"context"
	"encoding/json"
	"fmt"
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

var EXCHANGE_POLICY_TEMPLATE_INTERFACE_TRANSFORMATION = map[string]*schema.Schema{
	"language": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Snippet language (e.g. ramlSnippet, ramlV1Snippet, oasV2Snippet, oasV3Snippet).",
	},
	"transformation": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Snippet body to merge into the API definition.",
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
	"status": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Lifecycle status of the version (e.g. published, deprecated, development).",
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
			"split_model": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to request the split-asset model. When true, modern policies return `configuration` as a JSON Schema (draft-2019-09) object surfaced via `configuration_schema`; when false, legacy policies return a list of property configurations in `configuration`. Defaults to false for backward compatibility.",
			},
			"api_instance_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional api instance id used to filter applicability of the template.",
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
				Description: "Legacy policy configuration list (mule3/older mule4). Empty for modern policies that return a JSON Schema in `configuration_schema`.",
				Elem: &schema.Resource{
					Schema: EXCHANGE_POLICY_TEMPLATE_CONFIG,
				},
			},
			"configuration_schema": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Modern policy configuration as a JSON Schema (draft-2019-09) string. Populated when the policy returns the new schema shape; empty for legacy policies that use `configuration`.",
			},
			"schema_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identifier of the underlying JSON schema for modern policies.",
			},
			"split_asset_model": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the policy uses the split asset model.",
			},
			"ootb_upgradeable_impl": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the OOTB implementation supports in-place upgrades.",
			},
			"supported_java_versions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Java runtime versions this policy implementation is compatible with.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"interface_scope": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Scopes at which the policy can be applied (e.g. \"api\", \"resource\").",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"interface_transformation": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Modern replacement for the flat raml/oas snippet fields. Each entry pairs a language identifier with a transformation snippet.",
				Elem: &schema.Resource{
					Schema: EXCHANGE_POLICY_TEMPLATE_INTERFACE_TRANSFORMATION,
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

func dataSourceExchangePolicyTemplateRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	groupid := d.Get("group_id").(string)
	version := d.Get("version").(string)
	id := d.Get("id").(string)
	include_all_versions := d.Get("include_all_versions").(bool)
	split_model := d.Get("split_model").(bool)
	api_instance_id := d.Get("api_instance_id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	//perform request
	req := pco.apimpolicyclient.DefaultAPI.GetOrgExchangePolicyTemplateDetails(authctx, orgid, groupid, id, version).
		IncludeAllVersions(include_all_versions).
		SplitModel(split_model)
	if api_instance_id != "" {
		req = req.ApiInstanceId(api_instance_id)
	}
	res, httpr, err := req.Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
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

func flattenExchangePolicyTemplate(template *apim_policy.ExchangePolicyTemplate) map[string]any {
	result := make(map[string]any)
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
	if confPtr, ok := template.GetConfigurationOk(); ok && confPtr != nil {
		switch v := (*confPtr).(type) {
		case []any:
			b, _ := json.Marshal(v)
			var typed []apim_policy.PolicyConfiguration
			if err := json.Unmarshal(b, &typed); err == nil {
				result["configuration"] = flattenExchPolicyTempConfigs(typed)
			}
		case map[string]any:
			if b, err := json.Marshal(v); err == nil {
				result["configuration_schema"] = string(b)
			}
		}
	}
	if val, ok := template.GetSchemaIdOk(); ok {
		result["schema_id"] = *val
	}
	if val, ok := template.GetSplitAssetModelOk(); ok {
		result["split_asset_model"] = *val
	}
	if val, ok := template.GetOotbUpgradeableImplOk(); ok {
		result["ootb_upgradeable_impl"] = *val
	}
	if val, ok := template.GetSupportedJavaVersionsOk(); ok {
		result["supported_java_versions"] = flattenSupportedJavaVersions(val)
	}
	if val, ok := template.GetInterfaceScopeOk(); ok {
		result["interface_scope"] = val
	}
	if val, ok := template.GetInterfaceTransformationOk(); ok {
		result["interface_transformation"] = flattenExchPolicyTempInterfaceTransformation(val)
	}
	if val, ok := template.GetAllVersionsOk(); ok {
		result["all_versions"] = flattenExchPolicyTempAllVersions(val)
	}
	return result
}

func flattenSupportedJavaVersions(versions []*string) []string {
	out := make([]string, 0, len(versions))
	for _, v := range versions {
		if v == nil {
			continue
		}
		out = append(out, *v)
	}
	return out
}

func flattenExchPolicyTempInterfaceTransformation(collection []apim_policy.ExchangePolicyTemplateInterfaceTransformationInner) []any {
	slice := make([]any, len(collection))
	for i, item := range collection {
		data := make(map[string]any)
		if val, ok := item.GetLanguageOk(); ok {
			data["language"] = *val
		}
		if val, ok := item.GetTransformationOk(); ok {
			data["transformation"] = *val
		}
		slice[i] = data
	}
	return slice
}

func flattenExchPolicyTempConfigs(collection []apim_policy.PolicyConfiguration) []any {
	slice := make([]any, len(collection))
	for i, conf := range collection {
		slice[i] = flattenExchPolicyTempConfig(&conf)
	}
	return slice
}

func flattenExchPolicyTempConfig(conf *apim_policy.PolicyConfiguration) map[string]any {
	result := make(map[string]any)
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

func flattenExchPolicyTempConfigOpts(opts []map[string]any) []any {
	slice := make([]any, len(opts))
	for i, opt := range opts {
		data := make(map[string]any)
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

func flattenExchPolicyTempConfigOptsConfig(collection []apim_policy.PolicyConfigurationConfigurationInner) []any {
	slice := make([]any, len(collection))
	for i, c := range collection {
		data := make(map[string]any)
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

func flattenExchPolicyTempAllVersions(collection []apim_policy.ExchangePolicyTemplateAllVersionsInner) []any {
	slice := make([]any, len(collection))
	for i, version := range collection {
		data := make(map[string]any)
		if val, ok := version.GetGroupIdOk(); ok {
			data["group_id"] = *val
		}
		if val, ok := version.GetAssetIdOk(); ok {
			data["asset_id"] = *val
		}
		if val, ok := version.GetVersionOk(); ok {
			data["version"] = *val
		}
		if val, ok := version.GetStatusOk(); ok {
			data["status"] = *val
		}
		slice[i] = data
	}
	return slice
}

func setApimExchPolicyTempToResourceData(d *schema.ResourceData, data map[string]any) error {
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
		"configuration_schema", "schema_id", "split_asset_model",
		"ootb_upgradeable_impl", "supported_java_versions",
		"interface_scope", "interface_transformation",
		"all_versions",
	}
	return attributes[:]
}
