package anypoint

import (
	"context"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/user_rolegroups"
)

func resourceUserRolegroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserRolegroupCreate,
		ReadContext:   resourceUserRolegroupRead,
		DeleteContext: resourceUserRolegroupDelete,
		DeprecationMessage: `
		This resource is deprecated, please use ` + "`" + `teams` + "`" + `, ` + "`" + `team_members` + "`" + `team_roles` + "`" + ` instead.
		`,
		Description: `
		Assignes a ` + "`" + `user` + "`" + ` to a ` + "`" + `rolegroup` + "`" + `
		`,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The last time this resource has been updated locally.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique id of this user-rolegroup resource composed by {org_id}/{user_id}/{rolegroup_id}",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The master organization id where the role-group is defined.",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The user id.",
			},
			"rolegroup_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The role-group id.",
			},
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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceUserRolegroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	userid := d.Get("user_id").(string)
	rolegroupid := d.Get("rolegroup_id").(string)
	authctx := getUserRolegroupsAuthCtx(ctx, &pco)
	//request user creation
	httpr, err := pco.userrgpclient.DefaultApi.OrganizationsOrgIdUsersUserIdRolegroupsRolegroupIdPost(authctx, orgid, userid, rolegroupid).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to assign user " + userid + " rolegroup " + rolegroupid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.SetId(ComposeResourceId([]string{orgid, userid, rolegroupid}))

	return resourceUserRead(ctx, d, m)
}

func resourceUserRolegroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	orgid := d.Get("org_id").(string)
	userid := d.Get("user_id").(string)
	rolegroupid := d.Get("rolegroup_id").(string)
	id := d.Id()
	if isComposedResourceId(id) {
		orgid, userid, rolegroupid = decomposeUserRolegroupId(d)
	} else if isComposedResourceId(id, "_") { // retro-compatibility with versions < 1.6.x
		orgid, userid, rolegroupid = decomposeUserRolegroupId(d, "_")
	}
	rg, errDiags := searchUserRolegroup(ctx, d, m)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}
	//process data
	rolegroup := flattenUserRolegroupData(rg)
	//save in data source schema
	if err := setUserRolegroupAttributesToResourceData(d, rolegroup); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set user " + userid + " rolegroup " + rolegroupid,
			Detail:   err.Error(),
		})
		return diags
	}
	// set main identifiers parameters
	d.SetId(ComposeResourceId([]string{orgid, userid, rolegroupid}))
	d.Set("rolegroup_id", rolegroupid)
	d.Set("user_id", userid)
	d.Set("org_id", orgid)

	return diags
}

func resourceUserRolegroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	userid := d.Get("user_id").(string)
	rolegroupid := d.Get("rolegroup_id").(string)
	authctx := getUserRolegroupsAuthCtx(ctx, &pco)
	//perform request
	httpr, err := pco.userrgpclient.DefaultApi.OrganizationsOrgIdUsersUserIdRolegroupsRolegroupIdDelete(authctx, orgid, userid, rolegroupid).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete user " + userid + " rolegroup " + rolegroupid,
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

/*
Returns authentication context (includes authorization header)
*/
func getUserRolegroupsAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, user_rolegroups.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, user_rolegroups.ContextServerIndex, pco.server_index)
}
