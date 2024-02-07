package anypoint

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	team_members "github.com/mulesoft-anypoint/anypoint-client-go/team_members"
)

func resourceTeamMember() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamMemberCreate,
		ReadContext:   resourceTeamMemberRead,
		DeleteContext: resourceTeamMemberDelete,
		Description: `
		Assignes a ` + "`" + `user` + "`" + ` to a ` + "`" + `team` + "`" + ` for your ` + "`" + `org` + "`" + `.
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
				Description: "The unique id of this team membership composed by {org_id}/{team_id}/{user_id}",
			},
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The id of the team. team_id is globally unique.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The master organization id where the team is defined.",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The owner id",
			},
			"membership_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "member",
				ForceNew:         true,
				Description:      "Whether the member is a regular member or a maintainer. Only users may be team maintainers. Enum values: member, maintainer",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"member", "maintainer"}, true)),
			},
			"identity_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The member's identity type.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the team",
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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceTeamMemberCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	userid := d.Get("user_id").(string)
	authctx := getTeamMembersAuthCtx(ctx, &pco)
	body := newTeamMemberPutBody(d)
	//request user creation
	httpr, err := pco.teammembersclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdMembersUserIdPut(authctx, orgid, teamid, userid).TeamMemberPutBody(*body).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to add team member ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(ComposeResourceId([]string{orgid, teamid, userid}))
	d.Set("last_updated", time.Now().Format(time.RFC850))

	return resourceTeamMemberRead(ctx, d, m)
}

func resourceTeamMemberRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	userid := d.Get("user_id").(string)
	id := d.Id()
	if isComposedResourceId(id) {
		orgid, teamid, userid = decomposeTeamMemberId(d)
	} else if isComposedResourceId(id, "_") { // retro-compatibility with versions < 1.6.x
		orgid, teamid, userid = decomposeTeamMemberId(d, "_")
	}
	authctx := getTeamMembersAuthCtx(ctx, &pco)
	//request members
	res, httpr, err := pco.teammembersclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdMembersGet(authctx, orgid, teamid).Execute()
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
			Summary:  "Unable to get team " + teamid + " members",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//parse result
	item := search4MemberByIdInSlice(res.GetData(), userid)
	if item == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find team member " + userid + " for team " + teamid,
			Detail:   err.Error(),
		})
		return diags
	}
	teammember := flattenTeamMemberData(item)
	if err := setTeamMemberAttributesToResourceData(d, teammember); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set member " + id + " attributes for team " + teamid,
			Detail:   err.Error(),
		})
		return diags
	}

	d.Set("org_id", orgid)
	d.Set("team_id", teamid)
	d.Set("user_id", userid)
	d.SetId(ComposeResourceId([]string{orgid, teamid, userid}))

	return diags
}

func resourceTeamMemberDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	userid := d.Get("user_id").(string)
	authctx := getTeamMembersAuthCtx(ctx, &pco)
	//perform request
	httpr, err := pco.teammembersclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdMembersUserIdDelete(authctx, orgid, teamid, userid).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete team " + teamid + " member" + userid,
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

func newTeamMemberPutBody(d *schema.ResourceData) *team_members.TeamMemberPutBody {
	body := team_members.NewTeamMemberPutBodyWithDefaults()
	body.SetMembershipType(d.Get("membership_type").(string))
	return body
}

/*
 * Copies the given user instance into the given Source data
 */
func setTeamMemberAttributesToResourceData(d *schema.ResourceData, teammember map[string]interface{}) error {
	attributes := getTeamMemberAttributes()
	if teammember != nil {
		for _, attr := range attributes {
			if err := d.Set(attr, teammember[attr]); err != nil {
				return fmt.Errorf("unable to set team member attribute %s\n details: %s", attr, err)
			}
		}
	}
	return nil
}

func getTeamMemberAttributes() []string {
	attributes := [...]string{
		"identity_type", "name", "is_assigned_via_external_groups", "created_at", "updated_at",
	}
	return attributes[:]
}

// returns a member from the given team_member with the given user_id.
// if not found, returns nil
func search4MemberByIdInSlice(members []team_members.TeamMember, user_id string) *team_members.TeamMember {
	for _, member := range members {
		if member.GetId() == user_id {
			return &member
		}
	}
	return nil
}

/*
 * Returns authentication context (includes authorization header)
 */
func getTeamMembersAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, team_members.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, team_members.ContextServerIndex, pco.server_index)
}

func decomposeTeamMemberId(d *schema.ResourceData, separator ...string) (string, string, string) {
	s := DecomposeResourceId(d.Id(), separator...)
	return s[0], s[1], s[2]
}
