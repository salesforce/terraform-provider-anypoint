package anypoint

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	role "github.com/mulesoft-anypoint/anypoint-client-go/role"
)

func resourceRoleGroupRoles() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleGroupRolesCreate,
		ReadContext:   resourceRoleGroupRolesRead,
		DeleteContext: resourceRoleGroupRolesDelete,
		DeprecationMessage: `
		This resource is deprecated, please use ` + "`" + `teams` + "`" + `, ` + "`" + `team_members` + "`" + `team_roles` + "`" + ` instead.
		`,
		Description: `
		Assignes ` + "`" + `roles` + "`" + ` to a ` + "`" + `rolegroup` + "`" + ` for your ` + "`" + `org` + "`" + `.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique id of this rolegroup-roles resource composed by {org_id}/{role_group_id}",
			},
			"role_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The role-group id",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The business group id",
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of roles attributed to the role group",
			},
			"roles": {
				Type:        schema.TypeList,
				ForceNew:    true,
				Required:    true,
				Description: "List of roles in the role group",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"context_params": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_group_assignment_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"org_id": {
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
					},
				},
			},
		},
	}
}

func resourceRoleGroupRolesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	org_id := d.Get("org_id").(string)
	rolegroup_id := d.Get("role_group_id").(string)
	authctx := getRoleAuthCtx(ctx, &pco)
	//prepare request body
	body, errDiags := newRolegroupRolesPostBody(org_id, rolegroup_id, d)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}
	//perform request
	_, httpr, err := pco.roleclient.DefaultApi.OrganizationsOrgIdRolegroupsRolegroupIdRolesPost(authctx, org_id, rolegroup_id).RequestBody(body).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to assign roles to rolegroup " + rolegroup_id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.SetId(ComposeResourceId([]string{org_id, rolegroup_id}))
	return resourceRoleGroupRolesRead(ctx, d, m)
}

func resourceRoleGroupRolesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	org_id := d.Get("org_id").(string)
	rolegroup_id := d.Get("role_group_id").(string)
	id := d.Id()
	if isComposedResourceId(id) {
		org_id, rolegroup_id = decomposeRolegroupRolesId(d)
	} else if isComposedResourceId(id, "_") {
		org_id, rolegroup_id = decomposeRolegroupRolesId(d, "_")
	}
	authctx := getRoleAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.roleclient.DefaultApi.OrganizationsOrgIdRolegroupsRolegroupIdRolesGet(authctx, org_id, rolegroup_id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get rolegroup " + rolegroup_id + " assigned roles",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	data := res.GetData()
	assigned_roles, errDiags := flattenRoleGroupRolesData(data)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}
	//save in data source schema
	if err := setAssignedRolesAttributesToResourceData(d, assigned_roles); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to assign roles to rolegroup " + rolegroup_id,
			Detail:   err.Error(),
		})
		return diags
	}
	if err := d.Set("total", res.GetTotal()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number rolegroup " + rolegroup_id + " roles",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(ComposeResourceId([]string{org_id, rolegroup_id}))
	d.Set("org_id", org_id)
	d.Set("role_group_id", rolegroup_id)
	return diags
}

func resourceRoleGroupRolesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	org_id := d.Get("org_id").(string)
	rolegroup_id := d.Get("role_group_id").(string)

	authctx := getRoleAuthCtx(ctx, &pco)

	body, errDiags := newRolegroupRolesDeleteBody(d)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}

	_, httpr, err := pco.roleclient.DefaultApi.OrganizationsOrgIdRolegroupsRolegroupIdRolesDelete(authctx, org_id, rolegroup_id).RequestBody(body).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Delete rolegroup roles",
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

/**
 * Generates body object for creating rolegroup roles
 */
func newRolegroupRolesPostBody(org_id string, rolegroup_id string, d *schema.ResourceData) ([]map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	roles := d.Get("roles").([]interface{})

	if len(roles) == 0 {
		return nil, diags
	}
	res := make([]map[string]interface{}, len(roles))
	for i, role := range roles {
		item := make(map[string]interface{})
		item["role_id"] = role.(map[string]interface{})["role_id"].(string)
		item["context_params"] = map[string]string{
			"org": org_id,
		}
		res[i] = item
	}
	return res, diags
}

/**
 * Generates body object for deleting rolegroup roles
 */
func newRolegroupRolesDeleteBody(d *schema.ResourceData) ([]map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	roles := d.Get("roles").([]interface{})

	if len(roles) == 0 {
		return nil, diags
	}
	res := make([]map[string]interface{}, len(roles))
	for i, role := range roles {
		item := make(map[string]interface{})
		item["role_id"] = role.(map[string]interface{})["role_id"]
		item["context_params"] = map[string]string{
			"org": role.(map[string]interface{})["context_params"].(map[string]interface{})["org"].(string),
		}
		item["role_group_assignment_id"] = role.(map[string]interface{})["role_group_assignment_id"]
		item["role_group_id"] = role.(map[string]interface{})["role_group_id"]
		res[i] = item
	}
	return res, diags
}

func flattenRoleGroupRolesData(assigned_roles []role.AssignedRole) ([]map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(assigned_roles) > 0 {
		res := make([]map[string]interface{}, len(assigned_roles))

		for i, role := range assigned_roles {
			item := make(map[string]interface{})
			item["context_params"] = map[string]string{
				"org": *role.GetContextParams().Org,
			}
			item["created_at"] = role.GetCreatedAt()
			item["role_group_assignment_id"] = role.GetRoleGroupAssignmentId()
			item["role_group_id"] = role.GetRoleGroupId()
			item["role_id"] = role.GetRoleId()
			item["org_id"] = role.GetOrgId()
			item["name"] = role.GetName()
			item["description"] = role.GetDescription()
			item["internal"] = role.GetInternal()
			res[i] = item
		}
		return res, diags
	}

	return make([]map[string]interface{}, 0), diags
}

/*
* Copies the given rolegroup assigned roles into the given resource data
* @param d *schema.ResourceData the resource data schema
* @param assigned_roles map[string]interface{} the rolegroup assigned roles
 */
func setAssignedRolesAttributesToResourceData(d *schema.ResourceData, assigned_roles []map[string]interface{}) error {
	attributes := getAssignedRolesAttributes()

	if len(assigned_roles) == 0 {
		return nil
	}
	roles := make([]map[string]interface{}, len(assigned_roles))
	for i, assigned_role := range assigned_roles {
		role := make(map[string]interface{})
		for _, attr := range attributes {
			role[attr] = assigned_role[attr]
		}
		roles[i] = role
	}
	if err := d.Set("roles", roles); err != nil {
		return fmt.Errorf("unable to set assigned roles attribute \n details: %s", err)
	}
	return nil
}

/**
 * Returns Assigned Roles attributes (core attributes)
 */
func getAssignedRolesAttributes() []string {
	attributes := [...]string{
		"context_params", "created_at", "role_group_assignment_id", "role_group_id", "role_id",
		"org_id", "name", "description", "internal",
	}
	return attributes[:]
}

/*
 * Returns authentication context (includes authorization header)
 */
func getRoleAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, role.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, role.ContextServerIndex, pco.server_index)
}

func decomposeRolegroupRolesId(d *schema.ResourceData, separator ...string) (string, string) {
	s := DecomposeResourceId(d.Id(), separator...)
	return s[0], s[1]
}
