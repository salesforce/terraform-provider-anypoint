package anypoint

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mulesoft-anypoint/anypoint-client-go/team"
)

func dataSourceTeams() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamsRead,
		Description: `
		Reads all ` + "`" + `teams` + "`" + ` available in the business group.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The master organization id where the team is defined.",
			},
			"params": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The search parameters. Should only provide one occurrence of the block.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ancestor_team_id": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of team_ids that must appear in the team's ancestor_team_ids",
						},
						"parent_team_id": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of team_ids of the immediate parent of the team to return.",
						},
						"team_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "id of the team to return.",
						},
						"team_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "return only teams that are of the given type. Enum values are: internal, shared, private, external.",
						},
						"search": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A search string to use for case-insensitive partial matches on team name",
						},
						"offset": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "The number of records to omit from the response.",
						},
						"limit": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     200,
							Description: "Maximum records to retrieve per request.",
						},
						"sort": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The field to sort on. Default is team_name.",
						},
						"ascending": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether to sort ascending or descending.",
						},
					},
				},
			},
			"teams": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of resulting teams.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The master organization id where the team is defined.",
						},
						"team_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The id of the team. team_id is globally unique",
						},
						"team_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the team. Name is unique among teams within the organization.",
						},
						"team_type": {
							Type:     schema.TypeString,
							Computed: true,
							Description: `
							The type of the team. Internal teams are visible to all members of the organziation. 
							All internal teams of an organization are under the root internal team. 
							Private teams are internal teams but are only visible by maintainers/members of the team. 
							Shared teams are internal teams that can be mapped to external teams in other organizations where a trust relationship has been formed.
							Enum values are: internal, private and shared.
							`,
						},
						"ancestor_team_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Array of ancestor teams ids starting from either the internal or external root team down to this team's parent.",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The time the team was created.",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The time the team was last modified.",
						},
					},
				},
			},
			"len": {
				Type:        schema.TypeInt,
				Description: "The number of loaded results (for pagination purposes).",
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

func dataSourceTeamsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("params").(*schema.Set)
	orgid := d.Get("org_id").(string)
	authctx := getTeamAuthCtx(ctx, &pco)
	//prepare request
	req := pco.teamclient.DefaultAPI.GetTeams(authctx, orgid)
	req, errDiags := parseTeamSearchOpts(req, searchOpts)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}
	//request roles
	res, httpr, err := req.Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
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

func parseTeamSearchOpts(req team.DefaultAPIGetTeamsRequest, params *schema.Set) (team.DefaultAPIGetTeamsRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if params.Len() == 0 {
		return req, diags
	}

	opts := params.List()[0]

	for k, v := range opts.(map[string]any) {
		if k == "ancestor_team_id" {
			req = req.AncestorTeamId(ListInterface2ListStrings(v.([]any)))
			continue
		}
		if k == "parent_team_id" {
			req = req.ParentTeamId(ListInterface2ListStrings(v.([]any)))
			continue
		}
		if k == "team_id" && toString(v) != "" {
			req = req.TeamId(toString(v))
			continue
		}
		if k == "team_type" && toString(v) != "" {
			req = req.TeamType(toString(v))
			continue
		}
		if k == "search" && toString(v) != "" {
			req = req.Search(toString(v))
			continue
		}
		if k == "offset" {
			req = req.Offset(toInt32(v))
			continue
		}
		if k == "limit" {
			req = req.Limit(toInt32(v))
			continue
		}
		if k == "sort" && toString(v) != "" {
			req = req.Sort(toString(v))
			continue
		}
		if k == "ascending" {
			req = req.Ascending(toBool(v))
			continue
		}
	}

	return req, diags
}

func flattenTeamsData(teams *[]team.Team) []any {
	if teams != nil && len(*teams) > 0 {
		res := make([]any, len(*teams))
		for i, team := range *teams {
			res[i] = flattenTeamData(&team)
		}
		return res
	}

	return make([]any, 0)
}
