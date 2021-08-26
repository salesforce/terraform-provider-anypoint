package anypoint

import (
	"context"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/cloudhub-client-go/user_rolegroups"
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
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rolegroup_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role_group_id": {
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
			"external_names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"editable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"context_params": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"user_role_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUserRolegroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
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
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
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

	d.SetId(orgid + "_" + userid + "_" + rolegroupid)

	resourceUserRead(ctx, d, m)

	return diags
}

func resourceUserRolegroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	userid := d.Get("user_id").(string)
	rolegroupid := d.Get("rolegroup_id").(string)

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

	return diags
}

func resourceUserRolegroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	userid := d.Get("user_id").(string)
	rolegroupid := d.Get("rolegroup_id").(string)

	authctx := getUserRolegroupsAuthCtx(ctx, &pco)

	httpr, err := pco.userrgpclient.DefaultApi.OrganizationsOrgIdUsersUserIdRolegroupsRolegroupIdDelete(authctx, orgid, userid, rolegroupid).Execute()
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
			Summary:  "Unable to Delete user " + userid + " rolegroup " + rolegroupid,
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
	return context.WithValue(ctx, user_rolegroups.ContextAccessToken, pco.access_token)
}
