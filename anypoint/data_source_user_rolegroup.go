package anypoint

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/anypoint-client-go/user_rolegroups"
)

func dataSourceUserRolegroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRolegroupRead,
		DeprecationMessage: `
		This resource is deprecated, please use ` + "`" + `teams` + "`" + `, ` + "`" + `team_members` + "`" + `team_roles` + "`" + ` instead.
		`,
		Description: `
		Reads the ` + "`" + `user` + "`" + ` related ` + "`" + `rolegroup` + "`" + ` in the business group.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rolegroup_id": {
				Type:     schema.TypeString,
				Required: true,
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

func dataSourceUserRolegroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

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
