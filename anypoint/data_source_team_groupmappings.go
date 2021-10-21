package anypoint

import (
	"context"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/cloudhub-client-go/team_group_mappings"
)

func dataSourceTeamGroupMappings() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamGroupMappingsRead,
		Description: `
		Reads all ` + "`" + `groupmappings` + "`" + ` in the team.
		`,
		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"params": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
					},
				},
			},
			"teamgroupmappings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external_group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"provider_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"membership_type": {
							Type:     schema.TypeString,
							Computed: true,
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

func dataSourceTeamGroupMappingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("params").(*schema.Set)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)

	authctx := getTeamMembersAuthCtx(ctx, &pco)
	req := pco.teamgroupmappingsclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdGroupmappingsGet(authctx, orgid, teamid)
	req, errDiags := parseTeamGroupMappingsSearchOpts(req, searchOpts)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}

	//request members
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
			Summary:  "Unable to get team " + teamid + " groupmappings ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	teamgroupmappings := flattenTeamGroupMappingsData(res.Data)
	//save in data source schema
	if err := d.Set("teamgroupmappings", teamgroupmappings); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set groupmappings for team " + teamid,
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("total", res.GetTotal); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number of team " + teamid + " gropumappings",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

/*
 Parses the team members search options in order to check if the required search parameters are set correctly.
 Appends the parameters to the given request
*/
func parseTeamGroupMappingsSearchOpts(req team_group_mappings.DefaultApiApiOrganizationsOrgIdTeamsTeamIdGroupmappingsGetRequest, params *schema.Set) (team_group_mappings.DefaultApiApiOrganizationsOrgIdTeamsTeamIdGroupmappingsGetRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if params.Len() == 0 {
		return req, diags
	}

	opts := params.List()[0]

	for k, v := range opts.(map[string]interface{}) {
		if k == "offset" {
			req = req.Offset(int32(v.(int)))
			continue
		}
		if k == "limit" {
			req = req.Limit(int32(v.(int)))
			continue
		}
	}

	return req, diags
}

func flattenTeamGroupMappingsData(teamgroupmappings *[]team_group_mappings.TeamGroupMapping) []interface{} {
	if teamgroupmappings != nil && len(*teamgroupmappings) > 0 {
		res := make([]interface{}, len(*teamgroupmappings))
		for i, teamgroupmapping := range *teamgroupmappings {
			res[i] = flattenTeamGroupMappingData(&teamgroupmapping)
		}
		return res
	}
	return make([]interface{}, 0)
}

func flattenTeamGroupMappingData(teamgroupmapping *team_group_mappings.TeamGroupMapping) map[string]interface{} {
	item := make(map[string]interface{})
	if teamgroupmapping == nil {
		//log.Printf("tm nil")
		return item
	}
	if val, ok := teamgroupmapping.GetMembershipTypeOk(); ok {
		item["membership_type"] = *val
	}
	if val, ok := teamgroupmapping.GetExternalGroupNameOk(); ok {
		item["external_group_name"] = *val
	}
	if val, ok := teamgroupmapping.GetProviderIdOk(); ok {
		item["provider_id"] = *val
	}
	return item
}
