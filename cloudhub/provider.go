package cloudhub

import (
	"context"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	auth "github.com/mulesoft-consulting/cloudhub-client-go/authorization"
)

type ProviderConfOutput struct {
	authres *auth.InlineResponse200
	org_id  string
}

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("CLOUDHUB_CLIENT_ID", nil),
			},
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("CLOUDHUB_CLIENT_SECRET", nil),
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   false,
				DefaultFunc: schema.EnvDefaultFunc("CLOUDHUB_ORG_ID", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"cloudhub_vpcs": dataSourceVPCs(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client_id := d.Get("client_id").(string)
	client_secret := d.Get("client_secret").(string)
	org_id := d.Get("org_id").(string)

	if org_id == "" {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Required org id",
			Detail:   "The Organization Id is required.",
		})
		return ProviderConfOutput{
			authres: auth.NewInlineResponse200(),
			org_id:  org_id,
		}, diags
	}

	creds := auth.NewCredentialsWithDefaults()
	if (client_id != "") && (client_secret != "") {
		creds.SetClientId(client_id)
		creds.SetClientSecret(client_secret)
	}
	//authenticate
	cfgauth := auth.NewConfiguration()
	authclient := auth.NewAPIClient(cfgauth)
	authres, httpauthr, err := authclient.DefaultApi.Oauth2TokenPost(ctx).Credentials(*creds).Execute()
	if err != nil {
		b, _ := ioutil.ReadAll(httpauthr.Body)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Authenticate",
			Detail:   string(b),
		})
		return ProviderConfOutput{
			authres: auth.NewInlineResponse200(),
			org_id:  org_id,
		}, diags
	}
	defer httpauthr.Body.Close()
	return ProviderConfOutput{
		authres: &authres,
		org_id:  org_id,
	}, diags
}
