package anypoint

import (
	"context"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/cloudhub-client-go/rolegroup"
)

func dataSourceRoleGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRoleGroupRead,
		DeprecationMessage: `
		This resource is deprecated, please use ` + "`" + `teams` + "`" + `, ` + "`" + `team_members` + "`" + `team_roles` + "`" + ` instead.
		`,
		Description: `
		Reads a specific ` + "`" + `rolegroup` + "`" + ` in your business group.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
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
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
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
		},
	}
}

func dataSourceRoleGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	rolegroupid := d.Get("id").(string)

	authctx := getRoleGroupAuthCtx(ctx, &pco)

	res, httpr, err := pco.rolegroupclient.DefaultApi.OrganizationsOrgIdRolegroupsRolegroupIdGet(authctx, orgid, rolegroupid).Execute()
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
			Summary:  "Unable to get rolegroup",
			Detail:   details,
		})
		return diags
	}

	//process data
	rolegroup := flattenRoleGroupData(&res)
	//save in data source schema
	if err := setRolegroupAttributesToResourceData(d, rolegroup); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read rolegroup",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(rolegroupid)

	return diags
}

/**
 * Returns Assigned Roles attributes (core attributes)
 */
func getRolegroupAttributes() []string {
	attributes := [...]string{
		"role_group_id", "name", "external_names", "desription", "org_id",
		"editable", "created_at", "updated_at",
	}
	return attributes[:]
}

/*
* Transforms a rolegroup to the resourceRoleGroup schema
* @param rolegroup rolegroup.Rolegroup the rolegroup
* @return generic items
 */
func flattenRoleGroupData(rolegroup *rolegroup.Rolegroup) map[string]interface{} {
	item := make(map[string]interface{})

	item["role_group_id"] = rolegroup.GetRoleGroupId()
	item["name"] = rolegroup.GetName()
	item["external_names"] = rolegroup.GetExternalNames()
	item["description"] = rolegroup.GetDescription()
	item["org_id"] = rolegroup.GetOrgId()
	item["editable"] = rolegroup.GetEditable()
	item["created_at"] = rolegroup.GetCreatedAt()
	item["updated_at"] = rolegroup.GetUpdatedAt()

	return item
}
