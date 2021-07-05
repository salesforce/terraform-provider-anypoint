package anypoint

import (
	"context"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/cloudhub-client-go/role"
)

func dataSourceRoles() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRolesRead,
		Schema: map[string]*schema.Schema{
			"opts": {
				Type:     schema.TypeMap,
				Required: false,
			},
			"roles": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"internal": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"org_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"namespaces": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"shareable": {
							Type:     schema.TypeBool,
							Computed: true,
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

func dataSourceRolesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("opts").(map[string]interface{})

	authctx := getRoleAuthCtx(ctx, &pco)
	req := pco.roleclient.DefaultApi.RolesGet(authctx)
	errDiags := parseRoleSearchOpts(&req, searchOpts)
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
			Summary:  "Unable to set Roles",
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("total", res.GetTotal()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set Total Roles",
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
func parseRoleSearchOpts(req *role.DefaultApiApiRolesGetRequest, params map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	if params == nil {
		return diags
	}

	for k, v := range params {
		if k == "name" && IsString(v) {
			req.Name(v.(string))
			continue
		}
		if k == "description" && IsString(v) {
			req.Description(v.(string))
			continue
		}
		if k == "include_internal" && IsBool(v) {
			req.IncludeInternal(v.(bool))
			continue
		}
		if k == "search" && IsString(v) {
			req.Search(v.(string))
			continue
		}
		if k == "offset" && IsInt32(v) {
			req.Offset(v.(int32))
			continue
		}
		if k == "limit" && IsInt32(v) {
			req.Limit(v.(int32))
			continue
		}
		if k == "ascending" && IsBool(v) {
			req.Ascending(v.(bool))
			continue
		}
	}

	return diags
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
