package anypoint

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	team_roles "github.com/mulesoft-anypoint/anypoint-client-go/team_roles"
)

const BG_VIEWER_ROLE_ID = "833ab9ca-0c72-45ba-9764-1df83240db57"

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
				Description: "The unique id of this team roles composed by {org_id}/{team_id}",
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
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceTeamRolesCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	authctx := getTeamRolesAuthCtx(ctx, &pco)
	body := newTeamRolesPostBody(d)
	//request user creation
	httpr, err := pco.teamrolesclient.DefaultAPI.AssignTeamRoles(authctx, orgid, teamid).TeamRolePostBody(body).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create team " + teamid + " roles ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.SetId(ComposeResourceId([]string{orgid, teamid}))

	return resourceTeamRolesRead(ctx, d, m)
}

func resourceTeamRolesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	id := d.Id()
	if isComposedResourceId(id) {
		orgid, teamid, diags = decomposeTeamRolesId(d)
	} else if isComposedResourceId(id, "_") { // retro-compatibility with versions < 1.6.x
		orgid, teamid, diags = decomposeTeamRolesId(d, "_")
	}
	if diags.HasError() {
		return diags
	}
	authctx := getTeamRolesAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.teamrolesclient.DefaultAPI.GetTeamRoles(authctx, orgid, teamid).Limit(500).Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get team " + teamid + " roles",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process result
	roles := flattenTeamRolesData(res.Data)
	roles = filterOwnedTeamRoles(roles, d.Get("roles").([]any))
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

	d.Set("org_id", orgid)
	d.Set("team_id", teamid)
	d.SetId(ComposeResourceId([]string{orgid, teamid}))

	return diags
}

func resourceTeamRolesDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	authctx := getTeamRolesAuthCtx(ctx, &pco)
	//prepare request body
	body := newTeamRolesDeleteBody(d)
	//perform requeset
	httpr, err := pco.teamrolesclient.DefaultAPI.DeleteTeamRoles(authctx, orgid, teamid).TeamRoleDeleteBody(body).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
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

func newTeamRolesPostBody(d *schema.ResourceData) []team_roles.TeamRolePostBody {
	roles := d.Get("roles").([]any)

	if len(roles) <= 0 {
		return make([]team_roles.TeamRolePostBody, 0)
	}

	body := make([]team_roles.TeamRolePostBody, len(roles))

	for i, role := range roles {
		content := role.(map[string]any)
		item := team_roles.TeamRolePostBody{}
		item.SetRoleId(content["role_id"].(string))
		contextParam := team_roles.ContextParams{}
		if inputContextParams, ok := content["context_params"].(map[string]any); ok && inputContextParams != nil {
			org := inputContextParams["org"].(string)
			contextParam = team_roles.ContextParams{
				Org: &org,
			}
			if envId, ok := inputContextParams["envId"].(string); ok && envId != "" {
				contextParam.EnvId = &envId
			}
		}
		item.SetContextParams(contextParam)
		body[i] = item
	}

	return body
}

func newTeamRolesDeleteBody(d *schema.ResourceData) []team_roles.TeamRoleDeleteBody {
	roles := d.Get("roles").([]any)
	body := make([]team_roles.TeamRoleDeleteBody, 0)

	if len(roles) <= 0 {
		return body
	}

	// It is forbidden to remove the Business Group Viewer role
	for _, role := range roles {
		content := role.(map[string]any)
		if content["role_id"] == BG_VIEWER_ROLE_ID { // It is forbidden to remove the Business Group Viewer role
			continue
		}
		item := team_roles.TeamRoleDeleteBody{}
		item.SetRoleId(content["role_id"].(string))
		contextParam := team_roles.ContextParams{}
		if inputContextParams, ok := content["context_params"].(map[string]any); ok && inputContextParams != nil {
			org := inputContextParams["org"].(string)
			contextParam = team_roles.ContextParams{
				Org: &org,
			}
			if envId, ok := inputContextParams["envId"].(string); ok && envId != "" {
				contextParam.EnvId = &envId
			}
		}
		item.SetContextParams(contextParam)
		body = append(body, item)
	}

	return body
}

// filterOwnedTeamRoles restricts the API response to roles owned by this
// resource. Anypoint auto-grants roles (notably the Business Group Viewer)
// when a team is created or modified; leaving them in state surfaces as
// permanent drift against the user's configuration and forces a recreate.
//
// When state already lists roles (steady state), keep only API entries whose
// (role_id, org, envId) identity matches the owned set. On import the owned
// list is empty — fall back to dropping only the well-known auto-grants so
// the imported state mirrors what the user can realistically manage.
func filterOwnedTeamRoles(apiRoles, ownedRoles []any) []any {
	if len(ownedRoles) == 0 {
		out := make([]any, 0, len(apiRoles))
		for _, r := range apiRoles {
			m, ok := r.(map[string]any)
			if !ok {
				continue
			}
			if rid, _ := m["role_id"].(string); rid == BG_VIEWER_ROLE_ID {
				continue
			}
			out = append(out, r)
		}
		return out
	}
	owned := make(map[string]bool, len(ownedRoles))
	for _, r := range ownedRoles {
		owned[teamRoleIdentity(r)] = true
	}
	out := make([]any, 0, len(apiRoles))
	for _, r := range apiRoles {
		if owned[teamRoleIdentity(r)] {
			out = append(out, r)
		}
	}
	return out
}

// teamRoleIdentity builds a stable identity key for a role assignment from
// role_id and the context scope (org, envId). Two assignments are considered
// the same iff this key matches.
func teamRoleIdentity(role any) string {
	m, ok := role.(map[string]any)
	if !ok {
		return ""
	}
	rid, _ := m["role_id"].(string)
	var org, env string
	if cp, ok := m["context_params"].(map[string]any); ok {
		if v, ok := cp["org"].(string); ok {
			org = v
		}
		if v, ok := cp["envId"].(string); ok {
			env = v
		}
	}
	return rid + "|" + org + "|" + env
}

/*
 * Returns authentication context (includes authorization header)
 */
func getTeamRolesAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, team_roles.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, team_roles.ContextServerIndex, pco.server_index)
}

func decomposeTeamRolesId(d *schema.ResourceData, separator ...string) (string, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	s := DecomposeResourceId(d.Id(), separator...)
	if len(s) != 2 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid Team Roles ID format",
			Detail:   fmt.Sprintf("Expected ORG_ID/TEAM_ID, got %s", d.Id()),
		})
		return "", "", diags
	}
	return s[0], s[1], diags
}
