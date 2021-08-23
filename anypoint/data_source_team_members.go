package anypoint

import (
	"context"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/cloudhub-client-go/team_members"
)

func dataSourceTeamMembers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamMembersRead,
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
						"membership_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"identity_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"member_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
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
			"teammebers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identity_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"membership_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"is_assigned_via_external_groups": {
							Type:     schema.TypeBool,
							Computed: true,
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
	req := pco.teammembersclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdMembersGet(authctx, orgid, teamid)
	req, errDiags := parseTeamMembersSearchOpts(req, searchOpts)
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
			Summary:  "Unable to get team " + teamid + " member ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	teammebers := flattenTeamMembersData(res.Data)
	//save in data source schema
	if err := d.Set("teammebers", teammebers); err != nil {
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

func flattenTeamMembersData(teammebers *[]team_members.TeamMember) []interface{} {
	log.Printf("err1")
	if teammebers != nil && len(*teammebers) > 0 {
		res := make([]interface{}, len(*teammebers))
		for i, teammeber := range *teammebers {
			log.Printf("loop")
			res[i] = flattenTeamMemberData(&teammeber)
		}
		return res
	}
	log.Printf("tm empty")
	return make([]interface{}, 0)
}

func flattenTeamMemberData(teammeber *team_members.TeamMember) map[string]interface{} {
	item := make(map[string]interface{})
	if teammeber == nil {
		log.Printf("tm nil")
		return item
	}
	log.Printf("tm not null")
	if val, ok := teammeber.GetIdentityTypeOk(); ok {
		item["identity_type"] = *val
	}
	if val, ok := teammeber.GetIdOk(); ok {
		item["id"] = *val
	}
	if val, ok := teammeber.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := teammeber.GetMembershipTypeOk(); ok {
		item["membership_type"] = *val
	}
	if val, ok := teammeber.GetIsAssignedViaExternalGroupsOk(); ok {
		item["is_assigned_via_external_groups"] = *val
	}
	if val, ok := teammeber.GetCreatedAtOk(); ok {
		item["created_at"] = *val
	}
	if val, ok := teammeber.GetUpdatedAtOk(); ok {
		item["updated_at"] = *val
	}
	return item
}
