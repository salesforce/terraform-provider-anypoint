package anypoint

import (
	"context"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	team_roles "github.com/mulesoft-consulting/anypoint-client-go/team_roles"
)

const BG_VIEWER_ROLE = "833ab9ca-0c72-45ba-9764-1df83240db57"

func resourceTeamRoles() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamRolesCreate,
		ReadContext:   resourceTeamRolesRead,
		DeleteContext: resourceTeamRolesDelete,
		Description: `
		Attributes ` + "`" + `roles` + "`" + ` to your selected ` + "`" + `team` + "`" + ` for your ` + "`" + `org` + "`" + `.

Depending on the ` + "`" + `role` + "`" + `, some roles are environment scoped others are business group scoped :
* For environment scoped roles, the org id and environment id needs to be specified.
* For business group scoped roles, only the org id is needed.
		`,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last time this resource has been updated locally.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique id of this team roles composed by `org_id`_`team_id`_roles",
			},
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The id of the team. team_id is globally unique.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The master organization id where the team is defined.",
			},
			"roles": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return equalTeamRoles(d.GetChange("roles"))
				},
				Description: "The roles (permissions) of the team.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role name",
						},
						"role_id": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The role id",
						},
						"context_params": {
							Type:        schema.TypeMap,
							Required:    true,
							ForceNew:    true,
							Description: "The role's scope. Contains the organisation id to which the role is applied and optionally if the role spans environments, the environment within the organization id.",
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Description: "The total number of roles within the team",
				Computed:    true,
			},
		},
	}
}

func resourceTeamRolesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	authctx := getTeamRolesAuthCtx(ctx, &pco)
	body := newTeamRolesPostBody(d)

	//request user creation
	httpr, err := pco.teamrolesclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdRolesPost(authctx, orgid, teamid).RequestBody(body).Execute()
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
			Summary:  "Unable to create team roles ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(orgid + "_" + teamid + "_roles")

	resourceTeamRolesRead(ctx, d, m)

	return diags
}

func resourceTeamRolesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	id := d.Id()
	split := strings.Split(id, "_")
	orgid := split[0]
	teamid := split[1]
	authctx := getTeamRolesAuthCtx(ctx, &pco)
	//request roles
	res, httpr, err := pco.teamrolesclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdRolesGet(authctx, orgid, teamid).Limit(500).Execute()
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
			Summary:  "Unable to get team " + teamid + " roles",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	//process data
	roles := flattenTeamRolesData(res.Data)
	//save in data source schema
	if err := d.Set("roles", roles); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set roles for team " + teamid,
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("total", res.GetTotal()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number of team " + teamid + " roles",
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func resourceTeamRolesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	id := d.Id()
	split := strings.Split(id, "_")
	orgid := split[0]
	teamid := split[1]
	authctx := getTeamRolesAuthCtx(ctx, &pco)

	body := newTeamRolesDeleteBody(d)

	httpr, err := pco.teamrolesclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdRolesDelete(authctx, orgid, teamid).RequestBody(body).Execute()
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
			Summary:  "Unable to delete team " + teamid + " roles",
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

func newTeamRolesPostBody(d *schema.ResourceData) []map[string]interface{} {
	roles := d.Get("roles").([]interface{})

	if roles == nil || len(roles) <= 0 {
		return make([]map[string]interface{}, 0)
	}

	body := make([]map[string]interface{}, len(roles))

	for i, role := range roles {
		content := role.(map[string]interface{})
		item := make(map[string]interface{})
		item["role_id"] = content["role_id"]
		item["context_params"] = content["context_params"]
		body[i] = item
	}

	return body
}

func newTeamRolesDeleteBody(d *schema.ResourceData) []map[string]interface{} {
	roles := d.Get("roles").([]interface{})

	if roles == nil || len(roles) <= 0 {
		return make([]map[string]interface{}, 0)
	}

	body := make([]map[string]interface{}, 0) // It is forbidden to remove the Business Group Viewer role

	for _, role := range roles {
		content := role.(map[string]interface{})
		if content["role_id"] == BG_VIEWER_ROLE { // It is forbidden to remove the Business Group Viewer role
			continue
		}
		item := make(map[string]interface{})
		item["role_id"] = content["role_id"]
		item["context_params"] = content["context_params"]
		body = append(body, item)
	}

	return body
}

// Compares old and new values of roles
// returns true if they are the same, false otherwise
func equalTeamRoles(old, new interface{}) bool {
	old_list := old.([]interface{})
	new_list := new.([]interface{})
	old_list = FilterMapList(old_list, rolesSkipFilter)
	new_list = FilterMapList(new_list, rolesSkipFilter)
	sortMapListRoles(old_list)
	sortMapListRoles(new_list)
	if len(old_list) != len(new_list) {
		return false
	}
	for i := range old_list {
		if !equalTeamRole(old_list[i], new_list[i]) {
			return false
		}
	}
	return true
}

// compares 2 singles roles
func equalTeamRole(old, new interface{}) bool {
	old_role := old.(map[string]interface{})
	new_role := new.(map[string]interface{})

	ridkey := "role_id"
	cparamskey := "context_params"

	if old_role[ridkey].(string) != new_role[ridkey].(string) {
		return false
	}
	if !equalTeamRoleContextParams(old_role[cparamskey], new_role[cparamskey]) {
		return false
	}
	return true
}

// compares 2 role contexts
func equalTeamRoleContextParams(old, new interface{}) bool {
	old_cparams := old.(map[string]interface{})
	new_cparams := new.(map[string]interface{})
	for k := range old_cparams {
		if old_cparams[k].(string) != new_cparams[k].(string) {
			return false
		}
	}
	return true
}

// filter for roles to skip when attempting to calculate the diffin
// the role ids in this function are automatically added when a team is created. Therefore should be skipped
func rolesSkipFilter(item map[string]interface{}) bool {
	skip := []string{BG_VIEWER_ROLE}
	ridkey := "role_id"
	return !StringInSlice(skip, item[ridkey].(string), false)
}

// sorts a list of roles by role_id, org, envId
func sortMapListRoles(roles []interface{}) {
	sort.SliceStable(roles, func(i, j int) bool {
		i_elem := roles[i].(map[string]interface{})
		j_elem := roles[j].(map[string]interface{})

		sortAttrA := "role_id"
		sortAttrB := "context_params"
		if i_elem[sortAttrA].(string) != j_elem[sortAttrA].(string) {
			return i_elem[sortAttrA].(string) < j_elem[sortAttrA].(string)
		}

		sortAttrC := "org"
		sortAttrD := "envId"
		i_context := i_elem[sortAttrB].(map[string]interface{})
		j_context := j_elem[sortAttrB].(map[string]interface{})
		if  i_context[sortAttrC] != nil && j_context[sortAttrC] != nil && i_context[sortAttrC].(string) != j_context[sortAttrC].(string) {
			return i_context[sortAttrC].(string) < j_context[sortAttrC].(string)
		}
		if i_context[sortAttrD] != nil && j_context[sortAttrD] != nil && i_context[sortAttrD].(string) != j_context[sortAttrD].(string) {
			return i_context[sortAttrD].(string) < j_context[sortAttrD].(string)
		}

		return true
	})
}

/*
 * Returns authentication context (includes authorization header)
 */
func getTeamRolesAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, team_roles.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, team_roles.ContextServerIndex, pco.server_index)
}
