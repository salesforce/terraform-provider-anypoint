package anypoint

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/cloudhub-client-go/user"
)

func dataSourceUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUsersRead,
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
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "all",
						},
					},
				},
			},
			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
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
						"organization_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"idprovider_id": {
							Type:     schema.TypeString,
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
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"primary_organization": {
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

func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("params").(*schema.Set)
	orgid := d.Get("org_id").(string)
	authctx := getUserAuthCtx(ctx, &pco)

	req := pco.userclient.DefaultApi.OrganizationsOrgIdUsersGet(authctx, orgid)
	req, errDiags := parseUsersSearchOpts(req, searchOpts)
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
			Summary:  "Unable to get users",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	data := res.GetData()
	users := flattenUsersData(&data)
	//save in data source schema
	if err := d.Set("users", users); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set users",
			Detail:   err.Error(),
		})
		return diags
	}
	if err := d.Set("len", len(users)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set length of users",
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("total", res.GetTotal()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number of users",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

/*
 Parses the users search options in order to check if the required search parameters are set correctly.
 Appends the parameters to the given request
*/
func parseUsersSearchOpts(req user.DefaultApiApiOrganizationsOrgIdUsersGetRequest, params *schema.Set) (user.DefaultApiApiOrganizationsOrgIdUsersGetRequest, diag.Diagnostics) {
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
		if k == "type" {
			req = req.Type_(v.(string))
			continue
		}
	}

	return req, diags
}

/*
 Transforms a set of users to the dataSourceUsers schema
*/
func flattenUsersData(users *[]user.User) []interface{} {
	if users == nil || len(*users) <= 0 {
		return make([]interface{}, 0)
	}

	res := make([]interface{}, len(*users))
	for i, usr := range *users {
		item := make(map[string]interface{})

		item["id"] = usr.GetId()
		item["created_at"] = usr.GetCreatedAt()
		item["updated_at"] = usr.GetUpdatedAt()
		item["organization_id"] = usr.GetOrganizationId()
		item["enabled"] = usr.GetEnabled()
		item["idprovider_id"] = usr.GetIdproviderId()
		item["last_login"] = usr.GetLastLogin()
		item["mfa_verifiers_configured"] = usr.GetMfaVerifiersConfigured()
		item["mfa_verification_excluded"] = usr.GetMfaVerificationExcluded()
		item["is_federated"] = usr.GetIsFederated()
		item["username"] = usr.GetUsername()
		item["type"] = usr.GetType()
		primaryOrg := usr.GetPrimaryOrganization()
		item["primary_organization"] = map[string]string{
			"id":   primaryOrg.GetId(),
			"name": primaryOrg.GetName(),
		}
		usrOrgData := usr.GetOrganization()
		item["organization"] = flattenUserOrganizationData(&usrOrgData)

		res[i] = item
	}
	return res
}

/*
 * Transforms a user organization to a generic map
 */
func flattenUserOrganizationData(usrOrgData *user.Organization) map[string]interface{} {
	if usrOrgData == nil {
		return nil
	}
	res := make(map[string]interface{})

	res["name"] = usrOrgData.GetName()
	res["id"] = usrOrgData.GetId()
	res["created_at"] = usrOrgData.GetCreatedAt()
	res["updated_at"] = usrOrgData.GetUpdatedAt()
	res["owner_id"] = usrOrgData.GetOwnerId()
	res["client_id"] = usrOrgData.GetClientId()
	res["idprovider_id"] = usrOrgData.GetIdproviderId()
	res["is_federated"] = strconv.FormatBool(usrOrgData.GetIsFederated())
	jsonParentOrgs, _ := json.Marshal(usrOrgData.GetParentOrganizationIds())
	res["parent_organization_ids"] = string(jsonParentOrgs)
	jsonSubOrgIds, _ := json.Marshal(usrOrgData.GetSubOrganizationIds())
	res["sub_organization_ids"] = string(jsonSubOrgIds)
	jsonTenantOrgIds, _ := json.Marshal(usrOrgData.GetTenantOrganizationIds())
	res["tenant_organization_ids"] = string(jsonTenantOrgIds)
	res["mfa_required"] = usrOrgData.GetMfaRequired()
	res["is_automatic_admin_promotion_exempt"] = strconv.FormatBool(usrOrgData.GetIsAutomaticAdminPromotionExempt())
	res["domain"] = usrOrgData.GetDomain()
	res["is_master"] = strconv.FormatBool(usrOrgData.GetIsMaster())
	jsonSub, _ := json.Marshal(usrOrgData.GetSubscription())
	res["subscription"] = string(jsonSub)
	jsonProps, _ := json.Marshal(usrOrgData.GetProperties())
	res["properties"] = string(jsonProps)
	jsonEntitlments, _ := json.Marshal(usrOrgData.GetEntitlements())
	res["entitlements"] = string(jsonEntitlments)

	return res
}
