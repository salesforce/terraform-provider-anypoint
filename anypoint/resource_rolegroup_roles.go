package anypoint

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	role "github.com/mulesoft-consulting/cloudhub-client-go/role"
)

func resourceRoleGroupRoles() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleGroupRolesCreate,
		ReadContext:   resourceRoleGroupRolesRead,
		// UpdateContext: resourceRoleGroupRolesUpdate,
		DeleteContext: resourceRoleGroupRolesDelete,
		Schema: map[string]*schema.Schema{
			"role_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"total": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"roles": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"context_params": {
							Type: schema.TypeSet,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"org": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
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
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	org_id := d.Get("org_id").(string)
	rolegroup_id := d.Get("role_group_id").(string)

	authctx := getRoleAuthCtx(ctx, &pco)

	body, errDiags := newRolegroupRolesPostBody(org_id, rolegroup_id, d)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}

	//request vpc creation
	_, httpr, err := pco.roleclient.DefaultApi.OrganizationsOrgIdRolegroupsRolegroupIdRolesPost(authctx, org_id, rolegroup_id).RequestBody(body).Execute()
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
			Summary:  "Unable to assign roles to rolegroup",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(org_id + "_" + rolegroup_id)

	resourceRoleGroupRolesRead(ctx, d, m)

	return diags
}

func resourceRoleGroupRolesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	org_id := d.Get("org_id").(string)
	rolegroup_id := d.Get("role_group_id").(string)

	authctx := getRoleAuthCtx(ctx, &pco)

	res, httpr, err := pco.roleclient.DefaultApi.OrganizationsOrgIdRolegroupsRolegroupIdRolesGet(authctx, org_id, rolegroup_id).Execute()
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
			Summary:  "Unable to get rolegroup assigned roles",
			Detail:   details,
		})
		return diags
	}

	//process data
	data := res.GetData()
	assigned_roles, errDiags := flattenRoleGroupRolesData(&data)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}
	//save in data source schema
	if err := setAssignedRolesAttributesToResourceData(d, assigned_roles); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to assign roles to rolegroup",
			Detail:   err.Error(),
		})
		return diags
	}

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
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
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
	roles := d.Get("roles").([]map[string]interface{})

	if roles == nil || len(roles) == 0 {
		return nil, diags
	}
	res := make([]map[string]interface{}, len(roles))
	for i, role := range roles {
		item := make(map[string]interface{})
		item["role_id"] = role["role_id"].(string)
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
	roles := d.Get("roles").([]map[string]interface{})

	if roles == nil || len(roles) == 0 {
		return nil, diags
	}
	res := make([]map[string]interface{}, len(roles))
	for i, role := range roles {
		item := make(map[string]interface{})
		item["role_id"] = role["role_id"]
		item["context_params"] = map[string]string{
			"org": role["context_params"].(map[string]interface{})["org"].(string),
		}
		item["role_group_assignment_id"] = role["role_group_assignment_id"]
		item["role_group_id"] = role["role_group_id"]
		res[i] = item
	}
	return res, diags
}

func flattenRoleGroupRolesData(assigned_roles *[]role.AssignedRole) ([]map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if assigned_roles != nil && len(*assigned_roles) > 0 {
		res := make([]map[string]interface{}, len(*assigned_roles))

		for i, role := range *assigned_roles {
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
	if assigned_roles == nil {
		return nil
	}
	roles := make([]map[string]interface{}, len(assigned_roles))
	for i, role := range roles {
		for _, attr := range attributes {
			role[attr] = assigned_roles[i][attr]
		}
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
	return context.WithValue(ctx, role.ContextAccessToken, pco.access_token)
}