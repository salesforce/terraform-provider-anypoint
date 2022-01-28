package anypoint

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/anypoint-client-go/idp"
)

func dataSourceIDP() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIDPRead,
		Description: `
		Reads a specific ` + "`" + `identity provider` + "`" + ` in your business group.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The provider id",
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
				Computed:    true,
				Description: "The name of the provider",
			},
			"type": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The type of the provider, contains description and the name of the type of the provider (saml or oidc)",
			},
			"oidc_provider": {
				Type:        schema.TypeSet,
				Description: "The description of provider specific for OIDC types",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"token_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The token url of the openid-connect provider",
						},
						"redirect_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The redirect url of the openid-connect provider",
						},
						"userinfo_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The userinfo url of the openid-connect provider",
						},
						"authorize_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The authorization url of the openid-connect provider",
						},
						"client_registration_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The registration url, for dynamic client registration, of the openid-connect provider",
						},
						"client_credentials_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The client's credentials id",
						},
						"client_token_endpoint_auth_methods_supported": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The list of authentication methods supported",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"issuer": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The provider token issuer url",
						},
						"group_scope": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The provider group scopes",
						},
						"allow_untrusted_certificates": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "The certification validation trigger",
						},
					},
				},
			},
			"saml": {
				Type:        schema.TypeSet,
				Description: "The description of provider specific for SAML types",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"issuer": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The provider issuer",
						},
						"audience": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The provider audience",
						},
						"public_key": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The list of public keys",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"claims_mapping_email_attribute": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Field name in the SAML AttributeStatements that maps to Email. By default, the email attribute in the SAML assertion is used.",
						},
						"claims_mapping_group_attribute": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Field name in the SAML AttributeStatements that maps to Group.",
						},
						"claims_mapping_lastname_attribute": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Field name in the SAML AttributeStatements that maps to Last Name. By default, the lastname attribute in the SAML assertion is used.",
						},
						"claims_mapping_username_attribute": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Field name in the SAML AttributeStatements that maps to username. By default, the NameID attribute in the SAML assertion is used.",
						},
						"claims_mapping_firstname_attribute": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Field name in the SAML AttributeStatements that maps to First Name. By default, the firstname attribute in the SAML assertion is used.",
						},
						"sp_initiated_sso_enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if the Service Provider initiated SSO enabled",
						},
						"idp_initiated_sso_enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if the Identity Provider initiated SSO enabled",
						},
						"require_encrypted_saml_assertions": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if the encryption of saml assertions requirement is enabled",
						},
					},
				},
			},
			"service_provider_sign_on_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The provider's sign on url",
			},
			"service_provider_sign_out_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The provider's sign out url, only available for SAML",
			},
		},
	}
}

func dataSourceIDPRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	idpid := d.Get("id").(string)
	orgid := d.Get("org_id").(string)
	authctx := getENVAuthCtx(ctx, &pco)

	//request env
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

	d.SetId(idpid)

	return diags
}

/*
* Transforms a idp.Idp object to the dataSourceIDP schema
 */
func flattenIDPData(idpitem *idp.Idp) map[string]interface{} {
	if idpitem != nil {
		item := make(map[string]interface{})

		item["provider_id"] = idpitem.GetProviderId()
		item["name"] = idpitem.GetName()
		t := idpitem.GetType()
		t_tmp := make(map[string]string)
		t_tmp["description"] = t.GetDescription()
		t_tmp["name"] = t.GetName()
		item["type"] = t_tmp

		if _, ok := idpitem.GetOidcProviderOk(); ok {
			item["oidc_provider"] = flattenOIDCData(idpitem)
		} else if _, ok := idpitem.GetSamlOk(); ok {
			item["saml"] = flattenSAMLData(idpitem)
		}

		if sp, ok := idpitem.GetServiceProviderOk(); ok {
			if urls, ok := sp.GetUrlsOk(); ok {
				if signon, ok := urls.GetSignOnOk(); ok {
					item["service_provider_sign_on_url"] = *signon
				} else {
					item["service_provider_sign_on_url"] = ""
				}
				if signout, ok := urls.GetSignOutOk(); ok {
					item["service_provider_sign_out_url"] = *signout
				} else {
					item["service_provider_sign_out_url"] = ""
				}
			} else {
				item["service_provider_sign_on_url"] = ""
				item["service_provider_sign_out_url"] = ""
			}
		} else {
			item["service_provider_sign_on_url"] = ""
			item["service_provider_sign_out_url"] = ""
		}

		return item
	}

	return nil
}

func flattenOIDCData(idpitem *idp.Idp) []interface{} {
	array := make([]interface{}, 0)
	if idpitem != nil {
		item := make(map[string]interface{})
		oidcdata := idpitem.GetOidcProvider()
		if urls, ok := oidcdata.GetUrlsOk(); ok {
			item["token_url"] = urls.GetToken()
			item["redirect_url"] = urls.GetRedirect()
			item["userinfo_url"] = urls.GetUserinfo()
			item["authorize_url"] = urls.GetAuthorize()
		}
		if client, ok := oidcdata.GetClientOk(); ok {
			if urls, ok := client.GetUrlsOk(); ok {
				item["client_registration_url"] = urls.GetRegister()
			}
			if creds, ok := client.GetCredentialsOk(); ok {
				item["client_credentials_id"] = creds.GetId()
			}
			if meth, ok := client.GetTokenEndpointAuthMethodsSupportedOk(); ok {
				item["client_token_endpoint_auth_methods_supported"] = *meth
			}
		}
		item["issuer"] = oidcdata.GetIssuer()
		item["group_scope"] = oidcdata.GetGroupScope()

		if atc, ok := idpitem.GetAllowUntrustedCertificatesOk(); ok {
			item["allow_untrusted_certificates"] = *atc
		}

		array = append(array, item)
	}

	return array
}

func flattenSAMLData(idpitem *idp.Idp) []interface{} {
	array := make([]interface{}, 0)
	if idpitem != nil {
		item := make(map[string]interface{})
		samldata := idpitem.GetSaml()

		item["issuer"] = samldata.GetIssuer()
		item["audience"] = samldata.GetAudience()
		item["public_key"] = samldata.GetPublicKey()
		if claims, ok := samldata.GetClaimsMappingOk(); ok {
			if email, ok := claims.GetEmailAttributeOk(); ok {
				item["claims_mapping_email_attribute"] = *email
			}
			if group, ok := claims.GetGroupAttributeOk(); ok {
				item["claims_mapping_group_attribute"] = *group
			}
			if lastname, ok := claims.GetLastnameAttributeOk(); ok {
				item["claims_mapping_lastname_attribute"] = *lastname
			}
			if username, ok := claims.GetUsernameAttributeOk(); ok {
				item["claims_mapping_username_attribute"] = *username
			}
			if firstname, ok := claims.GetFirstnameAttributeOk(); ok {
				item["claims_mapping_firstname_attribute"] = *firstname
			}
		}
		item["sp_initiated_sso_enabled"] = samldata.GetSpInitiatedSsoEnabled()
		item["idp_initiated_sso_enabled"] = samldata.GetIdpInitiatedSsoEnabled()
		item["require_encrypted_saml_assertions"] = samldata.GetRequireEncryptedSamlAssertions()

		array = append(array, item)
	}

	return array
}

/*
* Copies the given idp instance into the given resource data
 */
func setIDPAttributesToResourceData(d *schema.ResourceData, idpitem map[string]interface{}) error {
	attributes := getIDPAttributes()
	if idpitem != nil {
		for _, attr := range attributes {
			if err := d.Set(attr, idpitem[attr]); err != nil {
				return fmt.Errorf("unable to set IDP attribute %s\n details: %s", attr, err)
			}
		}
	}
	return nil
}

func getIDPAttributes() []string {
	attributes := [...]string{
		"provider_id", "name", "type", "oidc_provider", "saml", "service_provider_sign_on_url", "service_provider_sign_out_url",
	}
	return attributes[:]
}
