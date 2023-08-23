package anypoint

import (
	"context"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/role"
)

func dataSourceRoles() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRolesRead,
		Description: `
		Reads all ` + "`" + `roles` + "`" + ` availabble.
		`,
		Schema: map[string]*schema.Schema{
			"params": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The name of a role",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The description of a role",
						},
						"include_internal": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Include internal roles",
						},
						"search": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "A search string to use for partial matches of role names",
						},
						"offset": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Pagination parameter to start returning results from this position of matches",
						},
						"limit": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     200,
							Description: "Pagination parameter for how many results to return",
						},
						"ascending": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Sort order for filtering",
						},
					},
				},
			},
			"roles": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique id of the role in the platform.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of this role.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the role.",
						},
						"internal": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this role is intended for internal use only.",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The business group id",
						},
						"namespaces": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "The list of namespaces related to this role.",
						},
						"shareable": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this role is shareable.",
						},
					},
				},
			},
			"len": {
				Type:        schema.TypeInt,
				Description: "The number of loaded results (pagination purpose).",
				Computed:    true,
			},
			"total": {
				Type:        schema.TypeInt,
				Description: "The total number of available results",
				Computed:    true,
			},
		},
	}
}

func dataSourceRolesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("params").(*schema.Set)

	authctx := getRoleAuthCtx(ctx, &pco)
	req := pco.roleclient.DefaultApi.RolesGet(authctx)
	req, errDiags := parseRoleSearchOpts(req, searchOpts)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}

	//request roles
	res, httpr, err := req.Execute()
	defer httpr.Body.Close()
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
			Summary:  "Unable to Get Roles",
			Detail:   details,
		})
		return diags
	}
	//process data
	data := res.GetData()
	roles := flattenRolesData(&data)
	//save in data source schema
	if err := d.Set("roles", roles); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set roles",
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("len", len(roles)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set length of roles",
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("total", res.GetTotal()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number roles",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

/**
 * Parses the roles search options in order to check if the required search parameters are set correctly
 * Appends the parameters to the given request
 */
func parseRoleSearchOpts(req role.DefaultApiApiRolesGetRequest, params *schema.Set) (role.DefaultApiApiRolesGetRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if params.Len() == 0 {
		return req, diags
	}
	opts := params.List()[0]

	for k, v := range opts.(map[string]interface{}) {
		if k == "name" && len(v.(string)) > 0 {
			req = req.Name(v.(string))
			continue
		}
		if k == "description" && len(v.(string)) > 0 {
			req = req.Description(v.(string))
			continue
		}
		if k == "include_internal" {
			req = req.IncludeInternal(v.(bool))
			continue
		}
		if k == "search" && len(v.(string)) > 0 {
			req = req.Search(v.(string))
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
		if k == "ascending" {
			req = req.Ascending(v.(bool))
			continue
		}
	}

	return req, diags
}

/*
* Transforms a set of roles to the dataSourceRoles schema
* @param roles *[]role.Role the list of roles
* @return list of generic items
 */
func flattenRolesData(roles *[]role.Role) []interface{} {
	if roles != nil && len(*roles) > 0 {
		res := make([]interface{}, len(*roles))

		for i, role := range *roles {
			item := make(map[string]interface{})

			item["role_id"] = role.GetRoleId()
			item["name"] = role.GetName()
			item["description"] = role.GetDescription()
			item["internal"] = role.GetInternal()
			item["org_id"] = role.GetOrgId()
			item["namespaces"] = role.GetNamespaces()
			item["shareable"] = role.GetShareable()

			res[i] = item
		}
		return res
	}

	return make([]interface{}, 0)
}
