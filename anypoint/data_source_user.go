package anypoint

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/cloudhub-client-go/user"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"first_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"phone_number": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"idprovider_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"deleted": {
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
			"last_login": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mfa_verifiers_configured": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mfa_verification_excluded": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_federated": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_preferences": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"member_of_organizations": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
			"contributor_of_organizations": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
			"organization": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"properties": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	orgid := d.Get("org_id").(string)
	userid := d.Get("user_id").(string)
	authctx := getUserAuthCtx(ctx, &pco)

	//request roles
	res, httpr, err := pco.userclient.DefaultApi.OrganizationsOrgIdUsersUserIdGet(authctx, orgid, userid).Execute()
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
			Summary:  "Unable to get user " + userid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	user := flattenUserData(&res)
	//save in data source schema
	if err := setUserAttributesToResourceData(d, user); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set user " + userid,
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

/*
 Transforms a set of users to the dataSourceUsers schema
*/
func flattenUserData(usr *user.User) map[string]interface{} {
	res := make(map[string]interface{})
	if usr == nil {
		return res
	}

	res["id"] = usr.GetId()
	res["created_at"] = usr.GetCreatedAt()
	res["updated_at"] = usr.GetUpdatedAt()
	res["organization_id"] = usr.GetOrganizationId()
	res["phone_number"] = usr.GetPhoneNumber()
	res["enabled"] = usr.GetEnabled()
	res["deleted"] = usr.GetDeleted()
	res["idprovider_id"] = usr.GetIdproviderId()
	res["last_login"] = usr.GetLastLogin()
	res["mfa_verifiers_configured"] = usr.GetMfaVerifiersConfigured()
	res["mfa_verification_excluded"] = usr.GetMfaVerificationExcluded()
	res["is_federated"] = usr.GetIsFederated()
	res["username"] = usr.GetUsername()
	res["type"] = usr.GetType()
	res["organization_preferences"] = usr.GetOrganizationPreferences()
	usrOrgData := usr.GetOrganization()
	res["organization"] = flattenUserOrganizationData(&usrOrgData)
	jsonProps, _ := json.Marshal(usr.GetProperties())
	res["properties"] = string(jsonProps)
	res["member_of_organizations"] = flattenUserOrgsData(usr.GetMemberOfOrganizations())
	res["contributor_of_organizations"] = flattenUserOrgsData(usr.GetContributorOfOrganizations())

	return res
}

/*
 * Transforms a user organization to a generic map
 */
func flattenUserOrgsData(userOrgs []user.Org) []map[string]interface{} {
	if userOrgs == nil || len(userOrgs) <= 0 {
		return make([]map[string]interface{}, 0)
	}
	res := make([]map[string]interface{}, len(userOrgs))

	for i, usrOrgData := range userOrgs {
		item := make(map[string]interface{})
		item["parent_name"] = usrOrgData.GetParentName()
		item["parent_id"] = usrOrgData.GetParentId()
		item["domain"] = usrOrgData.GetDomain()
		item["name"] = usrOrgData.GetName()
		item["id"] = usrOrgData.GetId()
		item["created_at"] = usrOrgData.GetCreatedAt()
		item["updated_at"] = usrOrgData.GetUpdatedAt()
		item["owner_id"] = usrOrgData.GetOwnerId()
		item["client_id"] = usrOrgData.GetClientId()
		item["idprovider_id"] = usrOrgData.GetIdproviderId()
		item["is_federated"] = strconv.FormatBool(usrOrgData.GetIsFederated())
		jsonParentOrgs, _ := json.Marshal(usrOrgData.GetParentOrganizationIds())
		item["parent_organization_ids"] = string(jsonParentOrgs)
		jsonSubOrgIds, _ := json.Marshal(usrOrgData.GetSubOrganizationIds())
		item["sub_organization_ids"] = string(jsonSubOrgIds)
		jsonTenantOrgIds, _ := json.Marshal(usrOrgData.GetTenantOrganizationIds())
		item["tenant_organization_ids"] = string(jsonTenantOrgIds)
		item["mfa_required"] = usrOrgData.GetMfaRequired()
		item["is_automatic_admin_promotion_exempt"] = strconv.FormatBool(usrOrgData.GetIsAutomaticAdminPromotionExempt())
		item["is_master"] = strconv.FormatBool(usrOrgData.GetIsMaster())
		jsonSub, _ := json.Marshal(usrOrgData.GetSubscription())
		item["subscription"] = string(jsonSub)

		res[i] = item
	}

	return res
}

/*
 * Copies the given user instance into the given resource data
 */
func setUserAttributesToResourceData(d *schema.ResourceData, usr map[string]interface{}) error {
	attributes := getUserAttributes()
	if usr != nil {
		for _, attr := range attributes {
			if err := d.Set(attr, usr[attr]); err != nil {
				return fmt.Errorf("unable to set user attribute %s\n details: %s", attr, err)
			}
		}
	}
	return nil
}

func getUserAttributes() []string {
	attributes := [...]string{
		"created_at", "updated_at", "organization_id", "first_name", "last_name", "email", "phone_number",
		"username", "idprovider_id", "enabled", "deleted", "last_login", "mfa_verifiers_configured", "mfa_verification_excluded",
		"is_federated", "type", "organization_preferences", "organization", "properties", "member_of_organizations", "contributor_of_organizations",
	}
	return attributes[:]
}
