package anypoint

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTeamRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamRoleRead,
		Description: `
		Reads a single ` + "`" + `role` + "`" + ` assignment on a ` + "`" + `team` + "`" + ` for your ` + "`" + `org` + "`" + `.
		Matches on the tuple (role_id, context_params.org, context_params.envId).
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Composite id of the team role: {org_id}/{team_id}/{role_id}/{context_org}[/{env_id}].",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The master organization id where the team is defined.",
			},
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The id of the team.",
			},
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The role id to look up.",
			},
			"context_params": {
				Type:        schema.TypeMap,
				Required:    true,
				Description: "The role's scope (`org` and optional `envId`). Used to disambiguate role grants whose role_id alone is not unique.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role display name returned by the platform.",
			},
		},
	}
}

func dataSourceTeamRoleRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	roleid := d.Get("role_id").(string)
	ctxOrg, envId, cpDiags := readTeamRoleContextParams(d)
	if cpDiags.HasError() {
		return cpDiags
	}
	authctx := getTeamRolesAuthCtx(ctx, &pco)
	res, httpr, err := pco.teamrolesclient.DefaultAPI.GetTeamRoles(authctx, orgid, teamid).RoleId(roleid).Limit(500).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read team role " + roleid + " for team " + teamid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	match := findTeamRoleAssignment(res.GetData(), roleid, ctxOrg, envId)
	if match == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Team role assignment not found",
			Detail:   "No assignment of role " + roleid + " in scope (org=" + ctxOrg + ", envId=" + envId + ") was found on team " + teamid + ".",
		})
		return diags
	}

	if name, ok := match.GetNameOk(); ok {
		d.Set("name", *name)
	}
	d.SetId(composeTeamRoleId(orgid, teamid, roleid, ctxOrg, envId))
	return diags
}
