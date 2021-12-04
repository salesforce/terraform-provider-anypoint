package anypoint

import (
	"context"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mulesoft-consulting/anypoint-client-go/team"
)

func dataSourceTeams() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamsRead,
		Description: `
		Reads all ` + "`" + `teams` + "`" + ` available in the business group.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"params": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ancestor_team_id": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"parent_team_id": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"team_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"team_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"search": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"offset": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"limit": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  200,
						},
						"sort": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ascending": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"teams": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"org_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"team_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"team_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"team_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ancestor_team_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
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
				},
			},
			"len": {
				Type:        schema.TypeInt,
				Description: "The number of loaded results",
				Computed:    true,
			},
			"total": {
				Type:        schema.TypeInt,
				Description: "The total number of available results",
				Computed:    true,
			},
		},
	}
}

func dataSourceTeamsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("params").(*schema.Set)
	orgid := d.Get("org_id").(string)
	authctx := getTeamAuthCtx(ctx, &pco)

	req := pco.teamclient.DefaultApi.OrganizationsOrgIdTeamsGet(authctx, orgid)
	req, errDiags := parseTeamSearchOpts(req, searchOpts)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}

	//request roles
	res, httpr, err := req.Execute()
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
			Summary:  "Unable to get teams",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	data := res.GetData()
	teams := flattenTeamsData(&data)
	//save in data source schema
	if err := d.Set("teams", teams); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set teams",
			Detail:   err.Error(),
		})
		return diags
	}
	if err := d.Set("len", len(teams)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set length of teams",
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("total", res.GetTotal()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number of teams",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func parseTeamSearchOpts(req team.DefaultApiApiOrganizationsOrgIdTeamsGetRequest, params *schema.Set) (team.DefaultApiApiOrganizationsOrgIdTeamsGetRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if params.Len() == 0 {
		return req, diags
	}

	opts := params.List()[0]

	for k, v := range opts.(map[string]interface{}) {
		if k == "ancestor_team_id" {
			req = req.AncestorTeamId(v.([]string))
			continue
		}
		if k == "parent_team_id" {
			req = req.ParentTeamId(v.([]string))
			continue
		}
		if k == "team_id" {
			req = req.TeamId(v.(string))
			continue
		}
		if k == "team_type" {
			req = req.TeamType(v.(string))
			continue
		}
		if k == "search" {
			req = req.Search(v.(string))
			continue
		}
		if k == "offset" {
			req = req.Offset(int32(v.(int)))
			continue
		}
		if k == "limit" {
			req = req.Limit(int32(v.(int)))
			continue
		}
		if k == "sort" {
			req = req.Sort(v.(string))
			continue
		}
		if k == "ascending" {
			req = req.Ascending(v.(bool))
			continue
		}
	}

	return req, diags
}

func flattenTeamsData(teams *[]team.Team) []interface{} {
	if teams != nil && len(*teams) > 0 {
		res := make([]interface{}, len(*teams))
		for i, team := range *teams {
			res[i] = flattenTeamData(&team)
		}
		return res
	}

	return make([]interface{}, 0)
}
