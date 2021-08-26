package anypoint

import (
	"context"
	"io/ioutil"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	team_roles "github.com/mulesoft-consulting/cloudhub-client-go/team_roles"
)

func resourceTeamRoles() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamRolesCreate,
		ReadContext:   resourceTeamRolesRead,
		DeleteContext: resourceTeamRolesDelete,
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
			"roles": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"context_params": {
							Type:     schema.TypeMap,
							Required: true,
							ForceNew: true,
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

func resourceTeamRolesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	authctx := getTeamRolesAuthCtx(ctx, &pco)
	body := newTeamRolesPostBody(d)

	//request user creation
	httpr, err := pco.teamrolesclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdRolesPost(authctx, orgid, teamid).RequestBody(body).Execute()
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
			Summary:  "Unable to create team roles ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(orgid + "_" + teamid + "_roles")

	resourceTeamRolesRead(ctx, d, m)

	return diags
}

func resourceTeamRolesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	id := d.Id()
	split := strings.Split(id, "_")
	orgid := split[0]
	teamid := split[1]
	authctx := getTeamRolesAuthCtx(ctx, &pco)
	//request roles
	res, httpr, err := pco.teamrolesclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdRolesGet(authctx, orgid, teamid).Limit(500).Execute()
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
			Summary:  "Unable to get team " + teamid + " roles",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	//process data
	roles := flattenTeamRolesData(res.Data)
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

	return diags
}

func resourceTeamRolesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	id := d.Id()
	split := strings.Split(id, "_")
	orgid := split[0]
	teamid := split[1]
	authctx := getTeamRolesAuthCtx(ctx, &pco)

	body := newTeamRolesDeleteBody(d)

	httpr, err := pco.teamrolesclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdRolesDelete(authctx, orgid, teamid).RequestBody(body).Execute()
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

func newTeamRolesPostBody(d *schema.ResourceData) []map[string]interface{} {
	roles := d.Get("roles").([]interface{})

	if roles == nil || len(roles) <= 0 {
		return make([]map[string]interface{}, 0)
	}

	body := make([]map[string]interface{}, len(roles))

	for i, role := range roles {
		content := role.(map[string]interface{})
		item := make(map[string]interface{})
		item["role_id"] = content["role_id"]
		item["context_params"] = content["context_params"]
		body[i] = item
	}

	return body
}

func newTeamRolesDeleteBody(d *schema.ResourceData) []map[string]interface{} {
	roles := d.Get("roles").([]interface{})

	if roles == nil || len(roles) <= 0 {
		return make([]map[string]interface{}, 0)
	}

	body := make([]map[string]interface{}, 0) // It is forbidden to remove the Business Group Viewer role

	for _, role := range roles {
		content := role.(map[string]interface{})
		if content["role_id"] == "833ab9ca-0c72-45ba-9764-1df83240db57" { // It is forbidden to remove the Business Group Viewer role
			continue
		}
		item := make(map[string]interface{})
		item["role_id"] = content["role_id"]
		item["context_params"] = content["context_params"]
		body = append(body, item)
	}

	return body
}

/*
 * Returns authentication context (includes authorization header)
 */
func getTeamRolesAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	return context.WithValue(ctx, team_roles.ContextAccessToken, pco.access_token)
}
