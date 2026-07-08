package anypoint

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	team_roles "github.com/mulesoft-anypoint/anypoint-client-go/team_roles"
)

func resourceTeamRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamRoleCreate,
		ReadContext:   resourceTeamRoleRead,
		DeleteContext: resourceTeamRoleDelete,
		Description: `
		Grants a single ` + "`" + `role` + "`" + ` to a ` + "`" + `team` + "`" + ` for your ` + "`" + `org` + "`" + `.

This resource is additive: it manages exactly one role assignment identified by
` + "`" + `role_id` + "`" + ` and ` + "`" + `context_params` + "`" + `. Use it when you want to grant individual roles
from multiple Terraform modules without one of them clobbering the others.

Depending on the ` + "`" + `role` + "`" + `, some roles are environment scoped while others are
business group scoped:
* For environment scoped roles, ` + "`" + `context_params.org` + "`" + ` and ` + "`" + `context_params.envId` + "`" + ` are required.
* For business group scoped roles, only ` + "`" + `context_params.org` + "`" + ` is required.

**Do not mix ` + "`" + `anypoint_team_role` + "`" + ` and ` + "`" + `anypoint_team_roles` + "`" + ` on the same team.**
` + "`" + `anypoint_team_roles` + "`" + ` declares the exhaustive set of roles for a team and will
fight with any ` + "`" + `anypoint_team_role` + "`" + ` grants outside its list.

The Business Group Viewer role is auto-granted by Anypoint when a team is created
and cannot be removed — managing it explicitly with this resource is rejected.
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
				Description: "The unique id of this team role composed by {org_id}/{team_id}/{role_id}/{context_org}[/{env_id}]",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The master organization id where the team is defined.",
			},
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The id of the team. team_id is globally unique.",
			},
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The role id to grant to the team.",
				ValidateFunc: func(v any, k string) ([]string, []error) {
					if v.(string) == BG_VIEWER_ROLE_ID {
						return nil, []error{fmt.Errorf("%q cannot manage the Business Group Viewer role (auto-granted by Anypoint, removal rejected). Remove this resource declaration", k)}
					}
					return nil, nil
				},
			},
			"context_params": {
				Type:        schema.TypeMap,
				Required:    true,
				ForceNew:    true,
				Description: "The role's scope. Contains the organisation id (`org`) to which the role is applied and, for environment-scoped roles, the environment id (`envId`).",
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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceTeamRoleCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
	body := []team_roles.TeamRolePostBody{newTeamRolePostBody(roleid, ctxOrg, envId)}
	httpr, err := pco.teamrolesclient.DefaultAPI.AssignTeamRoles(authctx, orgid, teamid).TeamRolePostBody(body).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create team role " + roleid + " for team " + teamid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(composeTeamRoleId(orgid, teamid, roleid, ctxOrg, envId))
	d.Set("last_updated", time.Now().Format(time.RFC850))

	return resourceTeamRoleRead(ctx, d, m)
}

func resourceTeamRoleRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, teamid, roleid, ctxOrg, envId, idDiags := resolveTeamRoleId(d)
	if idDiags.HasError() {
		return idDiags
	}
	authctx := getTeamRolesAuthCtx(ctx, &pco)
	res, httpr, err := pco.teamrolesclient.DefaultAPI.GetTeamRoles(authctx, orgid, teamid).RoleId(roleid).Limit(500).Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == 404 {
			d.SetId("")
			return nil
		}
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
		d.SetId("")
		return nil
	}

	d.Set("org_id", orgid)
	d.Set("team_id", teamid)
	d.Set("role_id", roleid)
	if name, ok := match.GetNameOk(); ok {
		d.Set("name", *name)
	}
	cp := map[string]any{"org": ctxOrg}
	if envId != "" {
		cp["envId"] = envId
	}
	if err := d.Set("context_params", cp); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set context_params for team role " + roleid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(composeTeamRoleId(orgid, teamid, roleid, ctxOrg, envId))
	return diags
}

func resourceTeamRoleDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, teamid, roleid, ctxOrg, envId, idDiags := resolveTeamRoleId(d)
	if idDiags.HasError() {
		return idDiags
	}
	authctx := getTeamRolesAuthCtx(ctx, &pco)
	body := []team_roles.TeamRoleDeleteBody{newTeamRoleDeleteBody(roleid, ctxOrg, envId)}
	httpr, err := pco.teamrolesclient.DefaultAPI.DeleteTeamRoles(authctx, orgid, teamid).TeamRoleDeleteBody(body).Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete team role " + roleid + " for team " + teamid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.SetId("")
	return diags
}

func newTeamRolePostBody(roleid, ctxOrg, envId string) team_roles.TeamRolePostBody {
	item := team_roles.TeamRolePostBody{}
	item.SetRoleId(roleid)
	cp := team_roles.ContextParams{Org: &ctxOrg}
	if envId != "" {
		cp.EnvId = &envId
	}
	item.SetContextParams(cp)
	return item
}

func newTeamRoleDeleteBody(roleid, ctxOrg, envId string) team_roles.TeamRoleDeleteBody {
	item := team_roles.TeamRoleDeleteBody{}
	item.SetRoleId(roleid)
	cp := team_roles.ContextParams{Org: &ctxOrg}
	if envId != "" {
		cp.EnvId = &envId
	}
	item.SetContextParams(cp)
	return item
}

func findTeamRoleAssignment(roles []team_roles.TeamRole, roleid, ctxOrg, envId string) *team_roles.TeamRole {
	for i := range roles {
		r := roles[i]
		if r.GetRoleId() != roleid {
			continue
		}
		cp := r.GetContextParams()
		if cp.GetOrg() != ctxOrg {
			continue
		}
		if cp.GetEnvId() != envId {
			continue
		}
		return &r
	}
	return nil
}

// readTeamRoleContextParams pulls org + envId out of the schema map. Org is
// required; envId is empty for business-group-scoped roles.
func readTeamRoleContextParams(d *schema.ResourceData) (string, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	raw, ok := d.Get("context_params").(map[string]any)
	if !ok || raw == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid context_params",
			Detail:   "context_params must be a map containing at least an 'org' key.",
		})
		return "", "", diags
	}
	org, _ := raw["org"].(string)
	if org == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Missing context_params.org",
			Detail:   "context_params.org is required and must be the business group id the role applies to.",
		})
		return "", "", diags
	}
	envId, _ := raw["envId"].(string)
	return org, envId, diags
}

func composeTeamRoleId(orgid, teamid, roleid, ctxOrg, envId string) string {
	parts := []string{orgid, teamid, roleid, ctxOrg}
	if envId != "" {
		parts = append(parts, envId)
	}
	return ComposeResourceId(parts)
}

// resolveTeamRoleId rebuilds (org, team, role, ctxOrg, envId) either from the
// composite Id (import) or from the typed schema (steady state). The composite
// is authoritative when present so imports work without separate field setting.
func resolveTeamRoleId(d *schema.ResourceData) (string, string, string, string, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	id := d.Id()
	if isComposedResourceId(id) {
		parts := DecomposeResourceId(id)
		if len(parts) != 4 && len(parts) != 5 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid Team Role ID format",
				Detail:   fmt.Sprintf("Expected ORG_ID/TEAM_ID/ROLE_ID/CONTEXT_ORG[/ENV_ID], got %s", id),
			})
			return "", "", "", "", "", diags
		}
		envId := ""
		if len(parts) == 5 {
			envId = parts[4]
		}
		return parts[0], parts[1], parts[2], parts[3], envId, diags
	}
	ctxOrg, envId, cpDiags := readTeamRoleContextParams(d)
	if cpDiags.HasError() {
		return "", "", "", "", "", cpDiags
	}
	return d.Get("org_id").(string), d.Get("team_id").(string), d.Get("role_id").(string), ctxOrg, envId, diags
}
