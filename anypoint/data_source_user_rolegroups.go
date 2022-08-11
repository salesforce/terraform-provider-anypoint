package anypoint

import (
	"context"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/anypoint-client-go/user_rolegroups"
)

func dataSourceUserRolegroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRolegroupsRead,
		DeprecationMessage: `
		This resource is deprecated, please use ` + "`" + `teams` + "`" + `, ` + "`" + `team_members` + "`" + `team_roles` + "`" + ` instead.
		`,
		Description: `
		Reads all ` + "`" + `user` + "`" + ` related ` + "`" + `rolegroups` + "`" + ` in the business group.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The master organization id where the role-group is defined.",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The user id.",
			},
			"params": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The search parameters. Should only provide one occurrence of the block.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"offset": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "The number of records to omit from the response.",
						},
						"limit": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     200,
							Description: "Maximum records to retrieve per request.",
						},
					},
				},
			},
			"rolegroups": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of resulted rolegroups.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role-group id.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the role-group.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the role-group",
						},
						"external_names": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of external names of the role-group",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The master organization id where the role-group is defined.",
						},
						"editable": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the role-group is editable",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The time when the user was assigned to the role-group.",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The time when the user assignment to the role-group was updated.",
						},
						"context_params": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The role-group scope.",
						},
						"user_role_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique if of the user assignment to the role-group",
						},
					},
				},
			},
			"len": {
				Type:        schema.TypeInt,
				Description: "The number of loaded results (pagination purposes).",
				Computed:    true,
			},
			"total": {
				Type:        schema.TypeInt,
				Description: "The total number of available results.",
				Computed:    true,
			},
		},
	}
}

func dataSourceUserRolegroupsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("params").(*schema.Set)
	orgid := d.Get("org_id").(string)
	userid := d.Get("user_id").(string)
	authctx := getUserRolegroupsAuthCtx(ctx, &pco)

	req := pco.userrgpclient.DefaultApi.OrganizationsOrgIdUsersUserIdRolegroupsGet(authctx, orgid, userid)
	req, errDiags := parseUserRolegroupsSearchOpts(req, searchOpts)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}

	//request roles
	res, httpr, err := req.Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get user rolegroups",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	data := res.GetData()
	rolegroups := flattenUserRolegroupsData(&data)
	//save in data source schema
	if err := d.Set("rolegroups", rolegroups); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set user rolegroups",
			Detail:   err.Error(),
		})
		return diags
	}
	if err := d.Set("len", len(rolegroups)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set length of user rolegroups",
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("total", res.GetTotal()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number of user rolegroups",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

/*
 Parses the users search options in order to check if the required search parameters are set correctly.
 Appends the parameters to the given request
*/
func parseUserRolegroupsSearchOpts(req user_rolegroups.DefaultApiApiOrganizationsOrgIdUsersUserIdRolegroupsGetRequest, params *schema.Set) (user_rolegroups.DefaultApiApiOrganizationsOrgIdUsersUserIdRolegroupsGetRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if params.Len() == 0 {
		return req, diags
	}

	opts := params.List()[0]

	for k, v := range opts.(map[string]interface{}) {
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

/*
 Transforms a set of users to the dataSourceUsers schema
*/
func flattenUserRolegroupsData(rolegroups *[]user_rolegroups.Rolegroup) []interface{} {
	if rolegroups == nil || len(*rolegroups) <= 0 {
		return make([]interface{}, 0)
	}
	res := make([]interface{}, len(*rolegroups))
	for i, rg := range *rolegroups {
		res[i] = flattenUserRolegroupData(&rg)
	}
	return res
}

func flattenUserRolegroupData(rg *user_rolegroups.Rolegroup) map[string]interface{} {
	if rg == nil {
		return nil
	}
	res := make(map[string]interface{})
	if val, ok := rg.GetRoleGroupIdOk(); ok {
		res["role_group_id"] = *val
	}
	if val, ok := rg.GetNameOk(); ok {
		res["name"] = *val
	}
	if val, ok := rg.GetDescriptionOk(); ok {
		res["description"] = *val
	}
	if val, ok := rg.GetExternalNamesOk(); ok {
		res["external_names"] = *val
	}
	if val, ok := rg.GetOrgIdOk(); ok {
		res["org_id"] = *val
	}
	if val, ok := rg.GetEditableOk(); ok {
		res["editable"] = *val
	}
	if val, ok := rg.GetCreatedAtOk(); ok {
		res["created_at"] = *val
	}
	if val, ok := rg.GetUpdatedAtOk(); ok {
		res["updated_at"] = *val
	}
	if val, ok := rg.GetContextParamsOk(); ok {
		res["context_params"] = *val
	}
	if val, ok := rg.GetUserRoleGroupIdOk(); ok {
		res["user_role_group_id"] = *val
	}

	return res
}
