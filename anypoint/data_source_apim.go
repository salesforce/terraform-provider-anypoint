package anypoint

import (
	"context"
	"io"
	"maps"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	apim "github.com/mulesoft-anypoint/anypoint-client-go/apim"
)

func dataSourceApim() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApimRead,
		Description: `
		Query all or part of available api manager instances for a given organization and environment.
		`,
		Schema: map[string]*schema.Schema{
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
			"params": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The search parameters. Should only provide one occurrence of the block.",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A string that will be checked for a partial or similar matches of the name, description, label and tags",
						},
						"group_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A string that will be checked for an exact match of the group_id",
						},
						"asset_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A string that will be checked for an exact match of the asset_id",
						},
						"asset_version": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A string that will be checked for an exact match of the asset_version",
						},
						"instance_label": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A string that will be checked for an exact match of the instanceLabel",
						},
						"product_version": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A string that will be checked for an exact match of the product_version",
						},
						"autodiscovery_instance_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A string that will be checked for an exact match of the autodiscovery_instance_name",
						},
						"filters": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "list of filters, which can be \"active\" and/or \"pinned\"",
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateDiagFunc: validation.ToDiagFunc(
									validation.StringInSlice(
										[]string{"active", "pinned"},
										false,
									),
								),
							},
						},
						"offset": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Skip over a number of elements by specifying an offset value for the query.",
						},
						"limit": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     20,
							Description: "Limit the number of elements in the response.",
						},
						"sort": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The field to sort on",
						},
						"ascending": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to sort ascending or descending",
						},
					},
				},
			},
			"assets": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List assets result of the query",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The asset's unique id",
						},
						"audit": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The asset's auditing data",
						},
						"master_organization_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The root business group id",
						},
						"organization_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The organization id where asset is published",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The asset name",
						},
						"exchange_asset_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The exchange asset name",
						},
						"group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The asset business group id",
						},
						"asset_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The asset id",
						},
						"apis": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The list of api manager instances related to this particular asset",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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
									"organization_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The organization id where asset is published",
									},
									"environment_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The environment id where asset is published",
									},
									"id": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The API manager instance unique id",
									},
									"instance_label": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The API manager Instance label",
									},
									"group_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The API manager instance's asset business group id",
									},
									"asset_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The API manager instance's asset id",
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
									"status": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The API Instance status",
									},
									"deployment_audit": {
										Type:        schema.TypeMap,
										Computed:    true,
										Description: "The api manager instance's deployment auditing data",
									},
									"deployment_application_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The api manager instance's deployment application id",
									},
									"deployment_target_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The api manager instance's deployment flex gateway target id",
									},
									"deployment_expected_status": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The api manager instance's deployment expected status",
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
									"pinned": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "True if the api manager instance is pinned on the anypoint manager's UI",
									},
									"active_contracts_count": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The number of active contracts on this api manager instance",
									},
									"autodiscovery_instance_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The instance's discovery name",
									},
								},
							},
						},
						"total_apis": {
							Type:        schema.TypeInt,
							Description: "The total number of apis related to this particular asset",
							Computed:    true,
						},
						"autodiscovery_api_name": {
							Type:        schema.TypeInt,
							Description: "The asset's autodiscovery api name. Basically a combenation of group id and artifact id",
							Computed:    true,
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Description: "The total number of available results",
				Computed:    true,
			},
		},
	}
}

func dataSourceApimRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//init vars
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("params").(*schema.Set)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	authctx := getApimAuthCtx(ctx, &pco)
	//prepare request
	req := pco.apimclient.DefaultApi.GetEnvApimInstances(authctx, orgid, envid)
	req, errDiags := parseApimSearchOpts(req, searchOpts)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}
	//execut request
	res, httpr, err := req.Execute()
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
			Summary:  "Unable to get api manager instances for org " + orgid + " and env " + envid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	assets := flattenApimAssetsResult(res.GetAssets())
	if err := d.Set("assets", assets); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set assets of apim instances for org " + orgid + " and env " + envid,
			Detail:   err.Error(),
		})
		return diags
	}
	if err := d.Set("total", res.GetTotal()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number of apim instances for org " + orgid + " and env " + envid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

/*
Parses the api manager search options in order to check if the required search parameters are set correctly.
Appends the parameters to the given request
*/
func parseApimSearchOpts(req apim.DefaultApiApiGetEnvApimInstancesRequest, params *schema.Set) (apim.DefaultApiApiGetEnvApimInstancesRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if params.Len() == 0 {
		return req, diags
	}
	opts := params.List()[0]
	for k, v := range opts.(map[string]interface{}) {
		if k == "query" {
			req = req.Query(v.(string))
			continue
		}
		if k == "group_id" {
			req = req.GroupId(v.(string))
			continue
		}
		if k == "asset_id" {
			req = req.AssetId(v.(string))
			continue
		}
		if k == "asset_version" {
			req = req.AssetVersion(v.(string))
			continue
		}
		if k == "instance_label" {
			req = req.InstanceLabel(v.(string))
			continue
		}
		if k == "product_version" {
			req = req.ProductVersion(v.(string))
			continue
		}
		if k == "autodiscovery_instance_name" {
			req = req.AutodiscoveryInstanceName(v.(string))
			continue
		}
		if k == "filters" {
			set := v.(*schema.Set)
			list := set.List()
			filters := make([]string, len(list))
			for i, f := range list {
				filters[i] = f.(string)
			}
			req = req.Filters(filters)
			continue
		}
		if k == "offset" {
			req = req.Offset(int32(v.(int)))
			continue
		}
		if k == "limit" {
			req = req.Limit(int32(v.(int)))
			continue
		}
		if k == "sort" {
			req = req.Sort(v.(string))
			continue
		}
		if k == "ascending" {
			req = req.Ascending(v.(bool))
			continue
		}
	}
	return req, diags
}

func flattenApimAssetsResult(assets []apim.ApimInstanceCollectionAssets) []interface{} {
	if len(assets) > 0 {
		res := make([]interface{}, len(assets))
		for i, asset := range assets {
			res[i] = flattenApimAssetResult(&asset)
		}
		return res
	}
	return make([]interface{}, 0)
}

func flattenApimAssetResult(asset *apim.ApimInstanceCollectionAssets) map[string]interface{} {
	item := make(map[string]interface{})
	if asset == nil {
		return item
	}
	if val, ok := asset.GetIdOk(); ok {
		item["id"] = *val
	}
	if val, ok := asset.GetAuditOk(); ok {
		item["audit"] = flattenApimAudit(val)
	}
	if val, ok := asset.GetMasterOrganizationIdOk(); ok {
		item["master_organization_id"] = *val
	}
	if val, ok := asset.GetOrganizationIdOk(); ok {
		item["organization_id"] = *val
	}
	if val, ok := asset.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := asset.GetExchangeAssetNameOk(); ok {
		item["exchange_asset_name"] = *val
	}
	if val, ok := asset.GetGroupIdOk(); ok {
		item["group_id"] = *val
	}
	if val, ok := asset.GetAssetIdOk(); ok {
		item["asset_id"] = *val
	}
	if val, ok := asset.GetApisOk(); ok {
		item["apis"] = flattenApimAssetApisResult(val)
	}
	return item
}

func flattenApimAssetApisResult(apis *[]apim.ApimInstanceCollectionApis) []map[string]interface{} {
	if apis == nil || len(*apis) == 0 {
		return []map[string]interface{}{}
	}
	result := make([]map[string]interface{}, len(*apis))
	for i, api := range *apis {
		item := make(map[string]interface{})
		if val, ok := api.GetAuditOk(); ok {
			item["audit"] = flattenApimAudit(val)
		}
		if val, ok := api.GetMasterOrganizationIdOk(); ok {
			item["master_organization_id"] = *val
		}
		if val, ok := api.GetOrganizationIdOk(); ok {
			item["organization_id"] = *val
		}
		if val, ok := api.GetEnvironmentIdOk(); ok {
			item["environment_id"] = *val
		}
		if val, ok := api.GetIdOk(); ok {
			item["id"] = *val
		}
		if val, ok := api.GetInstanceLabelOk(); ok {
			item["instance_label"] = *val
		}
		if val, ok := api.GetGroupIdOk(); ok {
			item["group_id"] = *val
		}
		if val, ok := api.GetAssetIdOk(); ok {
			item["asset_id"] = *val
		}
		if val, ok := api.GetAssetVersionOk(); ok {
			item["asset_version"] = *val
		}
		if val, ok := api.GetProductVersionOk(); ok {
			item["product_version"] = *val
		}
		if val, ok := api.GetDescriptionOk(); ok && val != nil {
			item["description"] = *val
		}
		if val, ok := api.GetTagsOk(); ok {
			item["tags"] = *val
		}
		if val, ok := api.GetOrderOk(); ok {
			item["order"] = *val
		}
		if val, ok := api.GetProviderIdOk(); ok && val != nil {
			item["provider_id"] = *val
		}
		if val, ok := api.GetDeprecatedOk(); ok {
			item["deprecated"] = *val
		}
		if val, ok := api.GetLastActiveDateOk(); ok && val != nil {
			item["last_active_date"] = *val
		}
		if val, ok := api.GetEndpointUriOk(); ok && val != nil {
			item["endpoint_uri"] = *val
		}
		if val, ok := api.GetIsPublicOk(); ok {
			item["is_public"] = *val
		}
		if val, ok := api.GetTechnologyOk(); ok {
			item["technology"] = *val
		}
		if val, ok := api.GetStatusOk(); ok {
			item["status"] = *val
		}
		if val, ok := api.GetDeploymentOk(); ok {
			deployment := flattenApimDeployment(val)
			maps.Copy(item, deployment)
		}
		if val, ok := api.GetRoutingOk(); ok && val != nil {
			item["routing"] = flattenApimRoutingCollection(*val)
		}
		result[i] = item
	}
	return result
}

/*
 * Returns authentication context (includes authorization header)
 */
func getApimAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, apim.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, apim.ContextServerIndex, pco.server_index)
}
