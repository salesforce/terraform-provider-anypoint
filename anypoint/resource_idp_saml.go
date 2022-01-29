package anypoint

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/anypoint-client-go/idp"
)

func resourceSAML() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSAMLCreate,
		ReadContext:   resourceSAMLRead,
		UpdateContext: resourceSAMLUpdate,
		DeleteContext: resourceSAMLDelete,
		Description: `
		Creates an ` + "`" + `identity provider` + "`" + ` SAML type configuration in your account.
		`,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The business group id",
			},
			"provider_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The provider id",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the provider",
			},
			"type": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The type of the provider, contains description and the name of the type of the provider (saml or oidc)",
			},
			"saml": {
				Type:        schema.TypeSet,
				Description: "The description of provider specific for SAML types",
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"issuer": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The provider issuer",
						},
						"audience": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The provider audience",
						},
						"public_key": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "The list of public keys",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"claims_mapping_email_attribute": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Field name in the SAML AttributeStatements that maps to Email. By default, the email attribute in the SAML assertion is used.",
						},
						"claims_mapping_group_attribute": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Field name in the SAML AttributeStatements that maps to Group.",
						},
						"claims_mapping_lastname_attribute": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Field name in the SAML AttributeStatements that maps to Last Name. By default, the lastname attribute in the SAML assertion is used.",
						},
						"claims_mapping_username_attribute": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Field name in the SAML AttributeStatements that maps to username. By default, the NameID attribute in the SAML assertion is used.",
						},
						"claims_mapping_firstname_attribute": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Field name in the SAML AttributeStatements that maps to First Name. By default, the firstname attribute in the SAML assertion is used.",
						},
						"sp_initiated_sso_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "True if the Service Provider initiated SSO enabled",
						},
						"idp_initiated_sso_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "True if the Identity Provider initiated SSO enabled",
						},
						"require_encrypted_saml_assertions": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "True if the encryption of saml assertions requirement is enabled",
						},
					},
				},
			},
			"sp_sign_on_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The provider's sign on url",
			},
			"sp_sign_out_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The provider's sign out url, only available for SAML",
			},
		},
	}
}

func resourceSAMLCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)

	authctx := getIDPAuthCtx(ctx, &pco)
	body, errDiags := newSAMLPostBody(d)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}

	res, httpr, err := pco.idpclient.DefaultApi.OrganizationsOrgIdIdentityProvidersPost(authctx, orgid).IdpPostBody(*body).Execute()
	defer httpr.Body.Close()
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
			Summary:  "Unable to create OIDC provider for org " + orgid,
			Detail:   details,
		})
		return diags
	}

	d.SetId(res.GetProviderId())

	return resourceSAMLRead(ctx, d, m)
}

func resourceSAMLRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	idpid := d.Id()
	orgid := d.Get("org_id").(string)
	authctx := getIDPAuthCtx(ctx, &pco)

	//request idp
	res, httpr, err := pco.idpclient.DefaultApi.OrganizationsOrgIdIdentityProvidersIdpIdGet(authctx, orgid, idpid).Execute()
	defer httpr.Body.Close()
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
			Summary:  "Unable to Get IDP " + idpid + " in org " + orgid,
			Detail:   details,
		})
		return diags
	}
	//process data
	idpinstance := flattenIDPData(&res)
	//save in data source schema
	if err := setIDPAttributesToResourceData(d, idpinstance); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set IDP " + idpid + " in org " + orgid,
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func resourceSAMLUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	idpid := d.Id()
	orgid := d.Get("org_id").(string)

	if d.HasChanges(getIDPAttributes()...) {
		authctx := getIDPAuthCtx(ctx, &pco)
		body, errDiags := newSAMLPatchBody(d)
		if errDiags.HasError() {
			diags = append(diags, errDiags...)
			return diags
		}
		_, httpr, err := pco.idpclient.DefaultApi.OrganizationsOrgIdIdentityProvidersIdpIdPatch(authctx, orgid, idpid).IdpPatchBody(*body).Execute()
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
				Summary:  "Unable to Update IDP " + idpid + " in org " + orgid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceSAMLRead(ctx, d, m)
}

func resourceSAMLDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	idpid := d.Id()
	orgid := d.Get("org_id").(string)
	authctx := getIDPAuthCtx(ctx, &pco)

	httpr, err := pco.idpclient.DefaultApi.OrganizationsOrgIdIdentityProvidersIdpIdDelete(authctx, orgid, idpid).Execute()
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
			Summary:  "Unable to Delete OIDC provider " + idpid,
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

/* Prepares the body required to post an OIDC provider*/
func newSAMLPostBody(d *schema.ResourceData) (*idp.IdpPostBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	saml_input := d.Get("saml")
	sp_sign_on_url := d.Get("sp_sign_on_url").(string)
	sp_sign_out_url := d.Get("sp_sign_out_url").(string)

	body := idp.NewIdpPostBody()

	saml_type := idp.NewIdpPostBodyType()
	saml_type.SetName("saml")
	saml_type.SetDescription("SAML 2.0")

	saml := idp.NewSaml1()

	if saml_input != nil {
		list := saml_input.([]interface{})
		if len(list) > 0 {
			item := list[0]
			data := item.(map[string]interface{})
			if issuer, ok := data["issuer"]; ok {
				saml.SetIssuer(issuer.(string))
			}
			if audience, ok := data["audience"]; ok {
				saml.SetAudience(audience.(string))
			}
			if public_key, ok := data["public_key"]; ok {
				saml.SetPublicKey(public_key.([]string))
			}
			//parsing claims
			claims := idp.NewClaimsMapping2()
			if claims_mapping_email_attribute, ok := data["claims_mapping_email_attribute"]; ok {
				claims.SetEmailAttribute(claims_mapping_email_attribute.(string))
			}
			if claims_mapping_group_attribute, ok := data["claims_mapping_group_attribute"]; ok {
				claims.SetGroupAttribute(claims_mapping_group_attribute.(string))
			}
			if claims_mapping_lastname_attribute, ok := data["claims_mapping_lastname_attribute"]; ok {
				claims.SetLastnameAttribute(claims_mapping_lastname_attribute.(string))
			}
			if claims_mapping_username_attribute, ok := data["claims_mapping_username_attribute"]; ok {
				claims.SetUsernameAttribute(claims_mapping_username_attribute.(string))
			}
			if claims_mapping_firstname_attribute, ok := data["claims_mapping_firstname_attribute"]; ok {
				claims.SetFirstnameAttribute(claims_mapping_firstname_attribute.(string))
			}
			saml.SetClaimsMapping(*claims)

			if sp_initiated_sso_enabled, ok := data["sp_initiated_sso_enabled"]; ok {
				saml.SetSpInitiatedSsoEnabled(sp_initiated_sso_enabled.(bool))
			}
			if idp_initiated_sso_enabled, ok := data["idp_initiated_sso_enabled"]; ok {
				saml.SetIdpInitiatedSsoEnabled(idp_initiated_sso_enabled.(bool))
			}
			if require_encrypted_saml_assertions, ok := data["require_encrypted_saml_assertions"]; ok {
				saml.SetRequireEncryptedSamlAssertions(require_encrypted_saml_assertions.(bool))
			}
		}
	}

	sp := idp.NewServiceProvider1()
	sp_urls := idp.NewUrls4()
	sp_urls.SetSignOn(sp_sign_on_url)
	sp_urls.SetSignOut(sp_sign_out_url)
	sp.SetUrls(*sp_urls)
	body.SetServiceProvider(*sp)
	body.SetSaml(*saml)
	body.SetName(name)
	body.SetType(*saml_type)

	return body, diags
}

/* Prepares the body required to patch an OIDC provider*/
func newSAMLPatchBody(d *schema.ResourceData) (*idp.IdpPatchBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	oidc_provider_input := d.Get("oidc_provider")

	body := idp.NewIdpPatchBody()

	oidc_type := idp.NewIdpPostBodyType()
	oidc_type.SetName("saml")
	oidc_type.SetDescription("SAML 2.0")

	oidc_provider := idp.NewOidcProvider1()

	body.SetName(name)
	if oidc_provider_input != nil {
		list := oidc_provider_input.([]interface{})
		for _, item := range list {
			if item != nil {
				data := item.(map[string]interface{})
				// reads client registration or credentials depending on which one is added
				client := idp.NewClient1()
				client_urls := idp.NewUrls1()
				if client_registration_url, ok := data["client_registration_url"]; ok {
					client_urls.SetRegister(client_registration_url.(string))
					client.SetUrls(*client_urls)
				} else {
					credentials := idp.NewCredentials1()
					if client_credentials_id, ok := data["client_credentials_id"]; ok {
						credentials.SetId(client_credentials_id.(string))
					}
					if client_credentials_secret, ok := data["client_credentials_secret"]; ok {
						credentials.SetSecret(client_credentials_secret.(string))
					}
					client.SetCredentials(*credentials)
				}
				oidc_provider.SetClient(*client)

				//Parsing URLs
				urls := idp.NewUrls3()
				if token_url, ok := data["token_url"]; ok {
					urls.SetToken(token_url.(string))
				}
				if userinfo_url, ok := data["userinfo_url"]; ok {
					urls.SetUserinfo(userinfo_url.(string))
				}
				if authorize_url, ok := data["authorize_url"]; ok {
					urls.SetAuthorize(authorize_url.(string))
				}
				oidc_provider.SetUrls(*urls)

				if issuer, ok := data["issuer"]; ok {
					oidc_provider.SetIssuer(issuer.(string))
				}
				if group_scope, ok := data["group_scope"]; ok {
					oidc_provider.SetGroupScope(group_scope.(string))
				}
				if allow_untrusted_certificates, ok := data["allow_untrusted_certificates"]; ok {
					body.SetAllowUntrustedCertificates(allow_untrusted_certificates.(bool))
				}
			}
		}
	}
	body.SetOidcProvider(*oidc_provider)

	return body, diags
}
