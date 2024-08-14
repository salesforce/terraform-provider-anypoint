package anypoint

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mulesoft-anypoint/anypoint-client-go/application_manager_v2"
)

func dataSourceAppDeploymentsV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAppDeploymentsV2Read,
		Description: `
		Reads ` + "`" + `Deployments` + "`" + ` from the runtime manager for a given organization and environment.
		This only works for Cloudhub V2 and Runtime Fabrics Apps.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization where to query deployments.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where to get deployments from",
			},
			"params": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The search parameters. Should only provide one occurrence of the block.",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The id of the target the deployments are deployed to.",
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
							Default:     25,
							Description: "Limit the number of elements in the response.",
						},
					},
				},
			},
			"deployments": {
				Type:        schema.TypeList,
				Description: "The result of the query with the list of all deployments.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "The id of the mule app deployment",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
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
						"target_provider": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cloud provider the target belongs to.",
						},
						"target_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The target id",
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
							Description: `Data of the mule app replicas
							- PARTIALLY_STARTED
							- DEPLOYMENT_FAILED
							- STARTING
							- STARTED
							- STOPPING
							- STOPPED
							- UNDEPLOYING
							- UNDEPLOYED
							- UPDATED
							- APPLIED
							- APPLYING
							- FAILED
							- DELETED
							`,
						},
						"application_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the application.",
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice(
									[]string{"RUNNING", "NOT_RUNNING"},
									false,
								),
							),
						},
						"current_runtime_version": {
							Type:        schema.TypeString,
							Description: "The mule app's runtime version",
							Computed:    true,
						},
						"last_successful_runtime_version": {
							Type:        schema.TypeString,
							Description: "The last successful runtime version",
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

func dataSourceAppDeploymentsV2Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("params").(*schema.Set)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	authctx := getAppDeploymentV2AuthCtx(ctx, &pco)
	//prepare request
	req := pco.appmanagerclient.DefaultApi.GetAllDeployments(authctx, orgid, envid)
	req, errDiags := parseAppDeploymentSearchOpts(req, searchOpts)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}
	//execut request
	res, httpr, err := req.Execute()
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
			Summary:  "Unable to get deployments for org " + orgid + " and env " + envid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	assets := flattenDeploymentItemsResult(res.GetItems())
	if err := d.Set("assets", assets); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set deployment items for org " + orgid + " and env " + envid,
			Detail:   err.Error(),
		})
		return diags
	}
	if err := d.Set("total", res.GetTotal()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number of deployment items for org " + orgid + " and env " + envid,
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
func parseAppDeploymentSearchOpts(req application_manager_v2.DefaultApiGetAllDeploymentsRequest, params *schema.Set) (application_manager_v2.DefaultApiGetAllDeploymentsRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if params.Len() == 0 {
		return req, diags
	}
	opts := params.List()[0]
	for k, v := range opts.(map[string]interface{}) {
		if k == "target_id" {
			req = req.TargetId(v.(string))
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
	}
	return req, diags
}

func flattenDeploymentItemsResult(items []application_manager_v2.DeploymentItem) []interface{} {
	if len(items) > 0 {
		res := make([]interface{}, len(items))
		for i, item := range items {
			res[i] = flattenDeploymentItemResult(&item)
		}
		return res
	}
	return make([]interface{}, 0)
}

func flattenDeploymentItemResult(data *application_manager_v2.DeploymentItem) map[string]interface{} {
	item := make(map[string]interface{})
	if data == nil {
		return item
	}
	if val, ok := data.GetIdOk(); ok {
		item["id"] = *val
	}
	if val, ok := data.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := data.GetCreationDateOk(); ok {
		item["creation_date"] = *val
	}
	if val, ok := data.GetLastModifiedDateOk(); ok {
		item["last_modified_date"] = *val
	}
	if target, ok := data.GetTargetOk(); ok {
		if val, ok := target.GetProviderOk(); ok {
			item["target_provider"] = *val
		}
		if val, ok := target.GetTargetIdOk(); ok {
			item["target_id"] = *val
		}
	}
	if val, ok := data.GetStatusOk(); ok {
		item["status"] = *val
	}
	if app, ok := data.GetApplicationOk(); ok {
		if val, ok := app.GetStatusOk(); ok {
			item["application_status"] = *val
		}
	}
	if val, ok := data.GetCurrentRuntimeVersionOk(); ok {
		item["current_runtime_version"] = *val
	}
	if val, ok := data.GetLastSuccessfulRuntimeVersionOk(); ok {
		item["last_successful_runtime_version"] = *val
	}

	return item
}
