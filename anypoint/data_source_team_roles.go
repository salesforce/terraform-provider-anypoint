package anypoint

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mulesoft-anypoint/anypoint-client-go/team_roles"
)

func dataSourceTeamRoles() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamRolesRead,
		Description: `
		Reads all ` + "`" + `team` + "`" + ` roles in the business group.
		`,
		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The id of the team. team_id is globally unique.",
			},
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
						"role_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "return only role assignments containing one of the supplied role_ids",
						},
						"search": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A search string to use for case-insensitive partial matches on role name",
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
					},
				},
			},
			"roles": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The resulted list of roles.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role name.",
						},
						"role_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role id.",
						},
						"context_params": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The role's scope. Contains the organisation id to which the role is applied and optionally if the role spans environments, the environment within the organization id.",
						},
					},
				},
			},
			"len": {
				Type:        schema.TypeInt,
				Description: "The number of loaded results (pagination purpose).",
				Computed:    true,
			},
			"total": {
				Type:        schema.TypeInt,
				Description: "The total number of available results.",
				Computed:    true,
			},
		},
	}
}

func dataSourceTeamRolesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("params").(*schema.Set)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	authctx := getTeamRolesAuthCtx(ctx, &pco)
	//prepare request
	req := pco.teamrolesclient.DefaultAPI.GetTeamRoles(authctx, orgid, teamid)
	req, errDiags := parseTeamRolesSearchOpts(req, searchOpts)
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
			Summary:  "Unable to get team " + teamid + " roles ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	roles := flattenTeamRolesData(res.GetData())
	//save in data source schema
	if err := d.Set("roles", roles); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set roles for team " + teamid,
			Detail:   err.Error(),
		})
		return diags
	}
	if err := d.Set("len", len(roles)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set length of team " + teamid + " roles",
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

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

/*
Parses the team roles search options in order to check if the required search parameters are set correctly.
Appends the parameters to the given request
*/
func parseTeamRolesSearchOpts(req team_roles.DefaultAPIGetTeamRolesRequest, params *schema.Set) (team_roles.DefaultAPIGetTeamRolesRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if params.Len() == 0 {
		return req, diags
	}

	opts := params.List()[0]

	for k, v := range opts.(map[string]any) {
		if k == "role_id" && toString(v) != "" {
			req = req.RoleId(toString(v))
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
	}

	return req, diags
}

func flattenTeamRolesData(roles []team_roles.TeamRole) []any {
	if len(roles) == 0 {
		return make([]any, 0)
	}
	res := make([]any, len(roles))
	for i, role := range roles {
		res[i] = flattenTeamRoleData(&role)
	}
	return res
}

func flattenTeamRoleData(role *team_roles.TeamRole) map[string]any {
	item := make(map[string]any)
	if role == nil {
		return item
	}
	if val, ok := role.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := role.GetRoleIdOk(); ok {
		item["role_id"] = *val
	}
	if val, ok := role.GetContextParamsOk(); ok {
		if env, ok := val.GetEnvIdOk(); ok {
			item["context_params"] = map[string]any{
				"org":   val.GetOrg(),
				"envId": *env,
			}
		} else {
			item["context_params"] = map[string]any{
				"org": val.GetOrg(),
			}
		}
	}
	return item
}
