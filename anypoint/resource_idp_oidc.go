package anypoint

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	idp "github.com/mulesoft-consulting/anypoint-client-go/idp"
)

func resourceOIDC() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOIDCCreate,
		ReadContext:   resourceOIDCRead,
		UpdateContext: resourceOIDCUpdate,
		DeleteContext: resourceOIDCDelete,
		Description: `
		Creates an ` + "`" + `identity provider` + "`" + ` OIDC type instance in your account.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
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
				Required:    true,
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
							Required:    true,
							Description: "The token url of the openid-connect provider",
						},
						"redirect_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The redirect url of the openid-connect provider",
						},
						"userinfo_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The userinfo url of the openid-connect provider",
						},
						"authorize_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The authorization url of the openid-connect provider",
						},
						"client_registration_url": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The registration url, for dynamic client registration, of the openid-connect provider. Mutually exclusive with credentials id/secret.",
						},
						"client_credentials_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The client's credentials id. This should only be provided if manual registration is wanted. Mutually exclusive with registration url",
						},
						"client_credentials_secret": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "The client's credentials secret. This should only be provided if manual registration is wanted. Mutually exclusive with registration url",
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
							Required:    true,
							Description: "The provider token issuer url",
						},
						"group_scope": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The provider group scopes",
						},
						"allow_untrusted_certificates": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "The certification validation trigger",
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

func getIDPAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, idp.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, idp.ContextServerIndex, pco.server_index)
}
