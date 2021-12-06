package anypoint

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mulesoft-consulting/anypoint-client-go/team"
)

func dataSourceTeam() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamRead,
		Description: `
		Reads a specific ` + "`" + `team` + "`" + ` in the business group.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
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
	}
}

func dataSourceTeamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("id").(string)
	authctx := getTeamAuthCtx(ctx, &pco)

	//request roles
	res, httpr, err := pco.teamclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdGet(authctx, orgid, teamid).Execute()
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
			Summary:  "Unable to get team " + teamid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	team := flattenTeamData(&res)
	//save in data source schema
	if err := setTeamAttributesToResourceData(d, team); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set team " + teamid,
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenTeamData(team *team.Team) map[string]interface{} {
	item := make(map[string]interface{})
	if team == nil {
		return item
	}
	if val, ok := team.GetOrgIdOk(); ok {
		item["org_id"] = *val
	}
	if val, ok := team.GetTeamIdOk(); ok {
		item["team_id"] = *val
	}
	if val, ok := team.GetTeamNameOk(); ok {
		item["team_name"] = *val
	}
	if val, ok := team.GetTeamTypeOk(); ok {
		item["team_type"] = *val
	}
	if val, ok := team.GetAncestorTeamIdsOk(); ok {
		item["ancestor_team_ids"] = *val
	}
	if val, ok := team.GetCreatedAtOk(); ok {
		item["created_at"] = *val
	}
	if val, ok := team.GetUpdatedAtOk(); ok {
		item["updated_at"] = *val
	}
	return item
}

/*
 Copies the given team instance into the given Source data
*/
func setTeamAttributesToResourceData(d *schema.ResourceData, team map[string]interface{}) error {
	attributes := getTeamAttributes()
	if team != nil {
		for _, attr := range attributes {
			if err := d.Set(attr, team[attr]); err != nil {
				return fmt.Errorf("unable to set team attribute %s\n details: %s", attr, err)
			}
		}
	}
	return nil
}

func getTeamAttributes() []string {
	attributes := [...]string{
		"org_id", "team_id", "team_name", "team_type",
		"ancestor_team_ids", "created_at", "updated_at",
	}
	return attributes[:]
}
