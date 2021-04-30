package anypoint

import (
	"context"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	auth "github.com/mulesoft-consulting/cloudhub-client-go/authorization"
	org "github.com/mulesoft-consulting/cloudhub-client-go/org"
	vpc "github.com/mulesoft-consulting/cloudhub-client-go/vpc"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ANYPOINT_CLIENT_ID", nil),
			},
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ANYPOINT_CLIENT_SECRET", nil),
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ANYPOINT_USERNAME", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ANYPOINT_PASSWORD", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"anypoint_vpc": resourceVPC(),
			"anypoint_bg":  resourceBG(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"anypoint_vpcs": dataSourceVPCs(),
			"anypoint_vpc":  dataSourceVPC(),
			"anypoint_bg":   dataSourceBG(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client_id := d.Get("client_id").(string)
	client_secret := d.Get("client_secret").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	if (username != "") && (password != "") {
		authres, d := userPwdAuth(ctx, username, password)
		if d != nil {
			return newProviderConfOutput(""), d
		}
		return newProviderConfOutput(authres.GetAccessToken()), diags
	}

	if (client_id != "") && (client_secret != "") {
		authres, d := connectedAppAuth(ctx, client_id, client_secret)
		if d != nil {
			return newProviderConfOutput(""), d
		}
		return newProviderConfOutput(authres.GetAccessToken()), diags
	}

	return newProviderConfOutput(""), diags

}

/*
 * Authenticates a user using username and password
 */
func userPwdAuth(ctx context.Context, username string, password string) (*auth.InlineResponse2001, diag.Diagnostics) {
	var diags diag.Diagnostics
	creds := auth.NewUserPwdCredentialsWithDefaults()
	creds.SetUsername(username)
	creds.SetPassword(password)
	//authenticate
	cfgauth := auth.NewConfiguration()
	authclient := auth.NewAPIClient(cfgauth)
	authres, httpr, err := authclient.DefaultApi.LoginPost(ctx).UserPwdCredentials(*creds).Execute()
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
			Summary:  "Unable to Authenticate Using User Password",
			Detail:   details,
		})
		return auth.NewInlineResponse2001(), diags
	}
	return &authres, diags
}

/*
 * Authenticates a connected app
 */
func connectedAppAuth(ctx context.Context, client_id string, client_secret string) (*auth.InlineResponse200, diag.Diagnostics) {
	var diags diag.Diagnostics
	creds := auth.NewCredentialsWithDefaults()
	creds.SetClientId(client_id)
	creds.SetClientSecret(client_secret)
	//authenticate
	cfgauth := auth.NewConfiguration()
	authclient := auth.NewAPIClient(cfgauth)
	authres, httpr, err := authclient.DefaultApi.ApiV2Oauth2TokenPost(ctx).Credentials(*creds).Execute()
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
			Summary:  "Unable to Authenticate Using Connected App",
			Detail:   details,
		})
		return auth.NewInlineResponse200(), diags
	}
	return &authres, diags
}

type ProviderConfOutput struct {
	access_token string
	vpcclient    *vpc.APIClient
	orgclient    *org.APIClient
}

func newProviderConfOutput(access_token string) ProviderConfOutput {
	//prepare request to get vpcs
	//ctx := context.Background()
	//authctx := context.WithValue(ctx, vpc.ContextAccessToken, access_token)
	vpccfg := vpc.NewConfiguration()
	orgcfg := org.NewConfiguration()
	vpcclient := vpc.NewAPIClient(vpccfg)
	orgclient := org.NewAPIClient(orgcfg)

	return ProviderConfOutput{
		access_token: access_token,
		vpcclient:    vpcclient,
		orgclient:    orgclient,
	}
}
