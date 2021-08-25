package anypoint

import (
	"context"
	"io/ioutil"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	team_members "github.com/mulesoft-consulting/cloudhub-client-go/team_members"
)

func resourceTeamMembers() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamMembersCreate,
		ReadContext:   resourceTeamMembersRead,
		DeleteContext: resourceTeamMembersDelete,
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
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"membership_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "member",
				ForceNew: true,
			},
			"identity_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
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
	}
}

func resourceTeamMembersCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	teamid := d.Get("team_id").(string)
	userid := d.Get("user_id").(string)
	authctx := getTeamMembersAuthCtx(ctx, &pco)
	body := newTeamMembersPutBody(d)

	//request user creation
	httpr, err := pco.teammembersclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdMembersUserIdPut(authctx, orgid, teamid, userid).TeamMemberPutBody(*body).Execute()
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
			Summary:  "Unable to add team member ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(orgid + "_" + teamid + "_" + userid + "_members")
	d.Set("last_updated", time.Now().Format(time.RFC850))

	resourceTeamMembersRead(ctx, d, m)

	return diags
}

func resourceTeamMembersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	id := d.Id()
	split := strings.Split(id, "_")
	orgid := split[0]
	teamid := split[1]
	authctx := getTeamMembersAuthCtx(ctx, &pco)
	//request members
	res, httpr, err := pco.teammembersclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdMembersGet(authctx, orgid, teamid).Execute()

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
			Summary:  "Unable to get team " + teamid + " members",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	//save name
	if err := d.Set("name", *res.GetData()[0].Name); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set member name for team " + teamid,
			Detail:   err.Error(),
		})
		return diags
	}
	//save created_at
	if err := d.Set("created_at", *res.GetData()[0].CreatedAt); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set member created_at for team " + teamid,
			Detail:   err.Error(),
		})
		return diags
	}
	//save identity_type
	if err := d.Set("identity_type", *res.GetData()[0].IdentityType); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set member identity_type for team " + teamid,
			Detail:   err.Error(),
		})
		return diags
	}
	//save updated_at
	if err := d.Set("updated_at", *res.GetData()[0].UpdatedAt); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set member updated_at for team " + teamid,
			Detail:   err.Error(),
		})
		return diags
	}
	//save is_assigned_via_external_groups
	if err := d.Set("is_assigned_via_external_groups", *res.GetData()[0].IsAssignedViaExternalGroups); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set member is_assigned_via_external_groups for team " + teamid,
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func resourceTeamMembersDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	id := d.Id()
	split := strings.Split(id, "_")
	orgid := split[0]
	teamid := split[1]
	userid := split[2]
	authctx := getTeamMembersAuthCtx(ctx, &pco)

	httpr, err := pco.teammembersclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdMembersUserIdDelete(authctx, orgid, teamid, userid).Execute()
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
			Summary:  "Unable to delete team " + teamid + " members",
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

func newTeamMembersPutBody(d *schema.ResourceData) *team_members.TeamMemberPutBody {
	body := team_members.NewTeamMemberPutBodyWithDefaults()
	body.SetMembershipType(d.Get("membership_type").(string))

	return body
}

/*
 * Returns authentication context (includes authorization header)
 */
func getTeamMembersAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	return context.WithValue(ctx, team_members.ContextAccessToken, pco.access_token)
}
