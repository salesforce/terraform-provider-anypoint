package anypoint

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/cloudhub-client-go/user"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"first_name": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"last_name": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"email": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"phone_number": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": {
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

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)

	authctx := getUserAuthCtx(ctx, &pco)

	body := newUserPostBody(d)

	//request user creation
	res, httpr, err := pco.userclient.DefaultApi.OrganizationsOrgIdUsersPost(authctx, orgid).UserPostBody(*body).Execute()
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
			Summary:  "Unable to Create User",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(res.GetId())

	resourceUserRead(ctx, d, m)

	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	userid := d.Id()
	orgid := d.Get("org_id").(string)

	authctx := getUserAuthCtx(ctx, &pco)

	res, httpr, err := pco.userclient.DefaultApi.OrganizationsOrgIdUsersUserIdGet(authctx, orgid, userid).Execute()
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
			Summary:  "Unable to Get User " + userid,
			Detail:   details,
		})
		return diags
	}

	//process data
	user := flattenUserData(&res)
	//save in data source schema
	if err := setUserAttributesToResourceData(d, user); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set User",
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	userid := d.Id()
	orgid := d.Get("org_id").(string)

	authctx := getUserAuthCtx(ctx, &pco)

	if d.HasChanges(getUserWatchAttributes()...) {
		body := newUserPutBody(d)
		//request user creation
		_, httpr, err := pco.userclient.DefaultApi.OrganizationsOrgIdUsersUserIdPut(authctx, orgid, userid).UserPutBody(*body).Execute()
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
				Summary:  "Unable to Update User",
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	userid := d.Id()
	orgid := d.Get("orgid").(string)

	authctx := getUserAuthCtx(ctx, &pco)

	httpr, err := pco.userclient.DefaultApi.OrganizationsOrgIdUsersUserIdDelete(authctx, orgid, userid).Execute()
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
			Summary:  "Unable to Delete User",
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

func newUserPostBody(d *schema.ResourceData) *user.UserPostBody {
	body := new(user.UserPostBody)

	if username := d.Get("username"); username != nil {
		body.SetUsername(username.(string))
	}
	if firstname := d.Get("first_name"); firstname != nil {
		body.SetFirstName(firstname.(string))
	}
	if lastname := d.Get("last_name"); lastname != nil {
		body.SetLastName(lastname.(string))
	}
	if email := d.Get("email"); email != nil {
		body.SetEmail(email.(string))
	}
	if phone_number := d.Get("phone_number"); phone_number != nil {
		body.SetPhoneNumber(d.Get("phone_number").(string))
	}
	if password := d.Get("password"); password != nil {
		body.SetPassword(password.(string))
	}

	return body
}

func newUserPutBody(d *schema.ResourceData) *user.UserPutBody {
	body := new(user.UserPutBody)

	if username := d.Get("username"); username != nil {
		body.SetUsername(username.(string))
	}
	if firstname := d.Get("first_name"); firstname != nil {
		body.SetFirstName(firstname.(string))
	}
	if lastname := d.Get("last_name"); lastname != nil {
		body.SetLastName(lastname.(string))
	}
	if email := d.Get("email"); email != nil {
		body.SetEmail(email.(string))
	}
	if phone_number := d.Get("phone_number"); phone_number != nil {
		body.SetPhoneNumber(d.Get("phone_number").(string))
	}
	if password := d.Get("password"); password != nil {
		body.SetPassword(password.(string))
	}

	return body
}

func getUserWatchAttributes() []string {
	attributes := [...]string{
		"first_name", "last_name", "properties", "email", "phone_number",
	}
	return attributes[:]
}

/*
 * Returns authentication context (includes authorization header)
 */
func getUserAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	return context.WithValue(ctx, user.ContextAccessToken, pco.access_token)
}
