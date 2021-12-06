package anypoint

import (
	"context"
	"io/ioutil"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	team_group_mappings "github.com/mulesoft-consulting/anypoint-client-go/team_group_mappings"
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
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"groupmappings": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external_group_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"provider_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"membership_type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Description: "The total number of available results",
				Computed:    true,
			},
		},
	}
}

func resourceTeamGroupMappingsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create team group mappings ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(orgid + "_" + teamid + "_groupmappings")

	resourceTeamGroupMappingsRead(ctx, d, m)

	return diags
}

func resourceTeamGroupMappingsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	authctx := getTeamGroupMappingsAuthCtx(ctx, &pco)
	body := newTeamGroupMappingsPutBody(d)
	id := d.Id()
	split := strings.Split(id, "_")
	orgid := split[0]
	teamid := split[1]

	//request put
	httpr, err := pco.teamgroupmappingsclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdGroupmappingsPut(authctx, orgid, teamid).RequestBody(body).Execute()
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
			Summary:  "Unable to create team group mappings ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	resourceTeamGroupMappingsRead(ctx, d, m)

	return diags
}

func resourceTeamGroupMappingsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create team group mappings ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId("")

	return diags
}

func resourceTeamGroupMappingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	id := d.Id()
	split := strings.Split(id, "_")
	orgid := split[0]
	teamid := split[1]
	authctx := getTeamGroupMappingsAuthCtx(ctx, &pco)
	//request get
	res, httpr, err := pco.teamgroupmappingsclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdGroupmappingsGet(authctx, orgid, teamid).Limit(500).Execute()
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

	return diags
}

func newTeamGroupMappingsPutBody(d *schema.ResourceData) []map[string]interface{} {
	teamgroupmappings := d.Get("groupmappings").([]interface{})

	if teamgroupmappings == nil || len(teamgroupmappings) <= 0 {
		return make([]map[string]interface{}, 0)
	}

	body := make([]map[string]interface{}, len(teamgroupmappings))

	for i, teamgroupmapping := range teamgroupmappings {
		content := teamgroupmapping.(map[string]interface{})
		item := make(map[string]interface{})
		item["membership_type"] = content["membership_type"]
		item["external_group_name"] = content["external_group_name"]
		body[i] = item
	}

	return body
}

/*
 * Returns authentication context (includes authorization header)
 */
func getTeamGroupMappingsAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, team_group_mappings.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, team_group_mappings.ContextServerIndex, pco.server_index)
}
