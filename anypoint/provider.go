package anypoint

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	auth "github.com/mulesoft-anypoint/anypoint-client-go/authorization"
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
				Description: "the connected app's id",
			},
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ANYPOINT_CLIENT_SECRET", nil),
				Description: "the connected app's secret",
			},
			"access_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ANYPOINT_ACCESS_TOKEN", nil),
				Description: "the connected app's access token",
			},
			"username": {
				Type:        schema.TypeString,
				Deprecated:  "Remove this attribute's configuration as it no longer is used and the attribute will be removed in the next major version of the provider.",
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ANYPOINT_USERNAME", nil),
				Description: "the user's username",
			},
			"password": {
				Type:        schema.TypeString,
				Deprecated:  "Remove this attribute's configuration as it no longer is used and the attribute will be removed in the next major version of the provider.",
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ANYPOINT_PASSWORD", nil),
				Description: "the user's password",
			},
			"cplane": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ANYPOINT_CPLANE", "us"),
				ValidateFunc: func(val any, key string) (warns []string, errs []error) {
					v := val.(string)
					if v != "us" && v != "eu" && v != "gov" {
						errs = append(errs, fmt.Errorf("%q must be 'eu‘ or 'us', got: %s", key, v))
					}
					return
				},
				Description: "the anypoint control plane",
			},
		},
		ResourcesMap:         RESOURCES_MAP,
		DataSourcesMap:       DATASOURCES_MAP,
		ConfigureContextFunc: providerConfigure,
		TerraformVersion:     "v1.0.1",
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client_id := d.Get("client_id").(string)
	client_secret := d.Get("client_secret").(string)
	access_token := d.Get("access_token").(string)
	//Deprecated
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	cplane := d.Get("cplane").(string)

	server_index := cplane2serverindex(cplane)
	auth_ctx := context.WithValue(ctx, auth.ContextServerIndex, server_index)

	if access_token != "" {
		return newProviderConfOutput(access_token, server_index), diags
	}

	if (username != "") && (password != "") {
		authres, d := userPwdAuth(auth_ctx, username, password)
		if d != nil {
			return newProviderConfOutput("", server_index), d
		}
		return newProviderConfOutput(authres.GetAccessToken(), server_index), diags
	}

	if (client_id != "") && (client_secret != "") {
		authres, d := connectedAppAuth(auth_ctx, client_id, client_secret)
		if d != nil {
			return newProviderConfOutput("", server_index), d
		}
		return newProviderConfOutput(authres.GetAccessToken(), server_index), diags
	}

	return newProviderConfOutput("", server_index), diags

}

/*
Authenticates a user using username and password
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
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := io.ReadAll(httpr.Body)
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
	defer httpr.Body.Close()
	return &authres, diags
}

/*
Authenticates a connected app
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
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := io.ReadAll(httpr.Body)
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
	defer httpr.Body.Close()
	return &authres, diags
}

/*
returns the server index depending on the control plane name
if the control plane is not recognized, returns -1
*/
func cplane2serverindex(cplane string) int {
	switch cplane {
	case "eu":
		return 1
	case "us":
		return 0
	case "gov":
		return 2
	}
	return -1
}
