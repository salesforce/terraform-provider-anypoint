package anypoint

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	team_group_mappings "github.com/mulesoft-anypoint/anypoint-client-go/team_group_mappings"
)

func resourceTeamGroupMappings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamGroupMappingsCreate,
		ReadContext:   resourceTeamGroupMappingsRead,
		DeleteContext: resourceTeamGroupMappingsDelete,
		UpdateContext: resourceTeamGroupMappingsUpdate,
		Description: `
		Maps identity providers' groups to a team.
		You can map users in a federated organization’s group to a team or role. Your Anypoint Platform organization must use an external identity provider, such as PingFederate.
		After you have mapped them, users in an organization can log in to Anypoint Platform using the same organizational credentials and access permissions that an organization maintains using SAML, OpenID Connect (OIDC), or LDAP.
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
				Description: "The unique id of this group mappings composed by {org_id}/{team_id}/groupmappings",
			},
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The id of the team. team_id is globally unique",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The master organization id where the team is defined.",
			},
			"groupmappings": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "The list of external identity provider groups that should be mapped to the given team.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return equalTeamGroupMappings(d.GetChange("groupmappings"))
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external_group_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The group name in the external identity provider that should be mapped to this team.",
						},
						"provider_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The id of the identity provider in anypoint platform.",
						},
						"membership_type": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Whether the mapped member is a regular member or a maintainer. Only users may be team maintainers. Enum values: member, maintainer",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"member", "maintainer"}, true)),
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Description: "The total number of group-mappings",
				Computed:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceTeamGroupMappingsCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	authctx := getTeamGroupMappingsAuthCtx(ctx, &pco)
	body := newTeamGroupMappingsPutBody(d)

	//request put
	httpr, err := pco.teamgroupmappingsclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdGroupmappingsPut(authctx, orgid, teamid).RequestBody(body).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create team group mappings for team" + teamid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.SetId(ComposeResourceId([]string{orgid, teamid}))
	return resourceTeamGroupMappingsRead(ctx, d, m)
}

func resourceTeamGroupMappingsUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	authctx := getTeamGroupMappingsAuthCtx(ctx, &pco)
	body := newTeamGroupMappingsPutBody(d)
	//perform request
	httpr, err := pco.teamgroupmappingsclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdGroupmappingsPut(authctx, orgid, teamid).RequestBody(body).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to update team group mappings for team " + teamid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.Set("last_updated", time.Now().Format(time.RFC850))
	return resourceTeamGroupMappingsRead(ctx, d, m)
}

func resourceTeamGroupMappingsDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	authctx := getTeamGroupMappingsAuthCtx(ctx, &pco)
	body := newTeamGroupMappingsPutBody(d)
	//perform request
	httpr, err := pco.teamgroupmappingsclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdGroupmappingsPut(authctx, orgid, teamid).RequestBody(body).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete team group mappings for team " + teamid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.SetId("")
	return diags
}

func resourceTeamGroupMappingsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	id := d.Id()
	if isComposedResourceId(id) {
		orgid, teamid, diags = decomposeTeamGroupMappingId(d)
	} else if isComposedResourceId(id, "_") { // retro-compatibility with versions < 1.6.x
		orgid, teamid, diags = decomposeTeamGroupMappingId(d, "_")
	}
	if diags.HasError() {
		return diags
	}
	authctx := getTeamGroupMappingsAuthCtx(ctx, &pco)
	//request get
	res, httpr, err := pco.teamgroupmappingsclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdGroupmappingsGet(authctx, orgid, teamid).Limit(500).Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get team " + teamid + " groupmappings",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	teamgroupmappings := flattenTeamGroupMappingsData(res.Data)
	//save in data source schema
	if err := d.Set("groupmappings", teamgroupmappings); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set groupmappings for team " + teamid,
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("total", res.GetTotal()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number of team " + teamid + " groupmappings",
			Detail:   err.Error(),
		})
		return diags
	}

	d.Set("org_id", orgid)
	d.Set("team_id", teamid)
	d.SetId(ComposeResourceId([]string{orgid, teamid}))

	return diags
}

func newTeamGroupMappingsPutBody(d *schema.ResourceData) []map[string]any {
	teamgroupmappings := d.Get("groupmappings").([]any)

	if len(teamgroupmappings) <= 0 {
		return make([]map[string]any, 0)
	}

	body := make([]map[string]any, len(teamgroupmappings))

	for i, teamgroupmapping := range teamgroupmappings {
		content := teamgroupmapping.(map[string]any)
		item := make(map[string]any)
		item["membership_type"] = content["membership_type"]
		item["external_group_name"] = content["external_group_name"]
		if val, ok := content["provider_id"]; ok {
			item["provider_id"] = val
		}
		body[i] = item
	}

	return body
}

// Compares old and new values of group mappings
// returns true if they are the same, false otherwise
func equalTeamGroupMappings(old, new any) bool {
	old_list := old.([]any)
	new_list := new.([]any)
	sortAttrs := []string{"membership_type", "external_group_name", "provider_id"}
	SortMapListAl(old_list, sortAttrs)
	SortMapListAl(new_list, sortAttrs)
	if len(old_list) != len(new_list) {
		return false
	}
	for i := range old_list {
		if !equalTeamGroupMapping(old_list[i], new_list[i]) {
			return false
		}
	}
	return true
}

// compares 2 single group mappings
func equalTeamGroupMapping(old, new any) bool {
	old_role := old.(map[string]any)
	new_role := new.(map[string]any)

	keys := []string{"membership_type", "external_group_name", "provider_id"}

	for _, k := range keys {
		if old_role[k].(string) != new_role[k].(string) {
			return false
		}
	}

	return true
}

/*
 * Returns authentication context (includes authorization header)
 */
func getTeamGroupMappingsAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, team_group_mappings.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, team_group_mappings.ContextServerIndex, pco.server_index)
}

func decomposeTeamGroupMappingId(d *schema.ResourceData, separator ...string) (string, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	s := DecomposeResourceId(d.Id(), separator...)
	if len(s) != 2 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid Team Group Mapping ID format",
			Detail:   fmt.Sprintf("Expected ORG_ID/TEAM_ID, got %s", d.Id()),
		})
		return "", "", diags
	}
	return s[0], s[1], diags
}
