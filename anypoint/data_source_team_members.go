package anypoint

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/team_members"
)

func dataSourceTeamMembers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamMembersRead,
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
						"membership_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Include the group access mappings that grant the provided membership type By default, all group access mappings are returned",
						},
						"identity_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A search string to use for case-insensitive partial matches on external group name",
						},
						"member_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Include the members of the team that have ids in this list",
						},
						"search": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Maximum records to retrieve per request.",
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
							Description: "The field to sort on",
						},
						"ascending": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to sort ascending or descending",
						},
					},
				},
			},
			"teammembers": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of resulting team-members.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identity_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The member's identity type.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique id of this team membership composed by `org_id`_`team_id`_`user_id`_members",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the team",
						},
						"membership_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Whether the member is a regular member or a maintainer. Only users may be team maintainers. Enum values: member, maintainer",
						},
						"is_assigned_via_external_groups": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the member was assigned to the team via a external group mapping",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The member team assignment creation date",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The member team assignment update date",
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

func dataSourceTeamMembersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("params").(*schema.Set)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	authctx := getTeamMembersAuthCtx(ctx, &pco)
	//prepare request
	req := pco.teammembersclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdMembersGet(authctx, orgid, teamid)
	req, errDiags := parseTeamMembersSearchOpts(req, searchOpts)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}
	//perform request
	res, httpr, err := req.Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get team " + teamid + " member ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	teammembers := flattenTeamMembersData(res.Data)
	//save in data source schema
	if err := d.Set("teammembers", teammembers); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set members for team " + teamid,
			Detail:   err.Error(),
		})
		return diags
	}
	if err := d.Set("total", res.GetTotal); err != nil {
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
Parses the team members search options in order to check if the required search parameters are set correctly.
Appends the parameters to the given request
*/
func parseTeamMembersSearchOpts(req team_members.DefaultApiApiOrganizationsOrgIdTeamsTeamIdMembersGetRequest, params *schema.Set) (team_members.DefaultApiApiOrganizationsOrgIdTeamsTeamIdMembersGetRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if params.Len() == 0 {
		return req, diags
	}

	opts := params.List()[0]

	for k, v := range opts.(map[string]interface{}) {
		if k == "membership_type" {
			req = req.MembershipType(v.(string))
			continue
		}
		if k == "identity_type" {
			req = req.IdentityType(v.(string))
			continue
		}
		if k == "member_ids" {
			req = req.MemberIds(v.([]string))
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

func flattenTeamMembersData(teammembers *[]team_members.TeamMember) []interface{} {
	if teammembers != nil && len(*teammembers) > 0 {
		res := make([]interface{}, len(*teammembers))
		for i, teammember := range *teammembers {
			res[i] = flattenTeamMemberData(&teammember)
		}
		return res
	}
	return make([]interface{}, 0)
}

func flattenTeamMemberData(teammember *team_members.TeamMember) map[string]interface{} {
	item := make(map[string]interface{})
	if teammember == nil {
		//log.Printf("tm nil")
		return item
	}
	if val, ok := teammember.GetIdentityTypeOk(); ok {
		item["identity_type"] = *val
	}
	if val, ok := teammember.GetIdOk(); ok {
		item["id"] = *val
	}
	if val, ok := teammember.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := teammember.GetMembershipTypeOk(); ok {
		item["membership_type"] = *val
	}
	if val, ok := teammember.GetIsAssignedViaExternalGroupsOk(); ok {
		item["is_assigned_via_external_groups"] = *val
	}
	if val, ok := teammember.GetCreatedAtOk(); ok {
		item["created_at"] = *val
	}
	if val, ok := teammember.GetUpdatedAtOk(); ok {
		item["updated_at"] = *val
	}
	return item
}
