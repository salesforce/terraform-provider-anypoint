package anypoint

import (
	"context"
	"fmt"
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
  Searches for the rolegroup in the list of results that has the same id as the one given by the user (rolegroup_id)
*/
func searchUserRolegroup(ctx context.Context, d *schema.ResourceData, m interface{}) (*user_rolegroups.Rolegroup, diag.Diagnostics) {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	userid := d.Get("user_id").(string)
	orgid := d.Get("org_id").(string)
	rolegroupid := d.Get("rolegroup_id").(string)
	authctx := getUserRolegroupsAuthCtx(ctx, &pco)

	limit := 50
	offset := 0
	count := 0
	end := false

	for !end {
		req := pco.userrgpclient.DefaultApi.OrganizationsOrgIdUsersUserIdRolegroupsGet(authctx, orgid, userid)
		req = req.Limit(int32(limit))
		req = req.Offset(int32(offset))
		res, httpr, err := req.Execute()
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
				Summary:  "Unable to get user " + userid + " rolegroup " + rolegroupid,
				Detail:   details,
			})
			return nil, diags
		}
		data := res.GetData()
		for _, rg := range data {
			if rg.GetRoleGroupId() == rolegroupid {
				end = true
				return &rg, diags
			}
		}
		l := len(data)
		count += l
		if count >= int(res.GetTotal()) || l == 0 {
			end = true
		} else {
			offset += limit
		}
	}
	return nil, diags
}

/*
 Copies the given user rolegroup instance into the given Source data
*/
func setUserRolegroupAttributesToResourceData(d *schema.ResourceData, rg map[string]interface{}) error {
	attributes := getUserRolegroupAttributes()
	if rg != nil {
		for _, attr := range attributes {
			if err := d.Set(attr, rg[attr]); err != nil {
				return fmt.Errorf("unable to set user rolegroup attribute %s\n details: %s", attr, err)
			}
		}
	}
	return nil
}

func getUserRolegroupAttributes() []string {
	attributes := [...]string{
		"role_group_id", "name", "description", "external_names", "editable", "created_at",
		"updated_at", "context_params", "user_role_group_id",
	}
	return attributes[:]
}

/*
  Returns authentication context (includes authorization header)
*/
func getUserRolegroupsAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	return context.WithValue(ctx, user_rolegroups.ContextAccessToken, pco.access_token)
}
