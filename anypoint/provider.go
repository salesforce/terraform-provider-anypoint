package anypoint

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	ame "github.com/mulesoft-anypoint/anypoint-client-go/ame"
	ame_binding "github.com/mulesoft-anypoint/anypoint-client-go/ame_binding"
	amq "github.com/mulesoft-anypoint/anypoint-client-go/amq"
	auth "github.com/mulesoft-anypoint/anypoint-client-go/authorization"
	connected_app "github.com/mulesoft-anypoint/anypoint-client-go/connected_app"
	dlb "github.com/mulesoft-anypoint/anypoint-client-go/dlb"
	env "github.com/mulesoft-anypoint/anypoint-client-go/env"
	idp "github.com/mulesoft-anypoint/anypoint-client-go/idp"
	org "github.com/mulesoft-anypoint/anypoint-client-go/org"
	role "github.com/mulesoft-anypoint/anypoint-client-go/role"
	rolegroup "github.com/mulesoft-anypoint/anypoint-client-go/rolegroup"
	team "github.com/mulesoft-anypoint/anypoint-client-go/team"
	team_group_mappings "github.com/mulesoft-anypoint/anypoint-client-go/team_group_mappings"
	team_members "github.com/mulesoft-anypoint/anypoint-client-go/team_members"
	team_roles "github.com/mulesoft-anypoint/anypoint-client-go/team_roles"
	user "github.com/mulesoft-anypoint/anypoint-client-go/user"
	user_rolegroups "github.com/mulesoft-anypoint/anypoint-client-go/user_rolegroups"
	vpc "github.com/mulesoft-anypoint/anypoint-client-go/vpc"
	vpn "github.com/mulesoft-anypoint/anypoint-client-go/vpn"
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
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ANYPOINT_USERNAME", nil),
				Description: "the user's username",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ANYPOINT_PASSWORD", nil),
				Description: "the user's password",
			},
			"cplane": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ANYPOINT_CPLANE", "us"),
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if v != "us" && v != "eu" && v != "gov" {
						errs = append(errs, fmt.Errorf("%q must be 'eu‘ or 'us', got: %s", key, v))
					}
					return
				},
				Description: "the anypoint control plane",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"anypoint_vpc":                 resourceVPC(),
			"anypoint_vpn":                 resourceVPN(),
			"anypoint_bg":                  resourceBG(),
			"anypoint_rolegroup_roles":     resourceRoleGroupRoles(),
			"anypoint_rolegroup":           resourceRoleGroup(),
			"anypoint_env":                 resourceENV(),
			"anypoint_user":                resourceUser(),
			"anypoint_user_rolegroup":      resourceUserRolegroup(),
			"anypoint_team":                resourceTeam(),
			"anypoint_team_roles":          resourceTeamRoles(),
			"anypoint_team_member":         resourceTeamMember(),
			"anypoint_team_group_mappings": resourceTeamGroupMappings(),
			"anypoint_dlb":                 resourceDLB(),
			"anypoint_idp_oidc":            resourceOIDC(),
			"anypoint_idp_saml":            resourceSAML(),
			"anypoint_connected_app":       resourceConnectedApp(),
			"anypoint_amq":                 resourceAMQ(),
			"anypoint_ame":                 resourceAME(),
			"anypoint_ame_binding":         resourceAMEBinding(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"anypoint_vpcs":                dataSourceVPCs(),
			"anypoint_vpc":                 dataSourceVPC(),
			"anypoint_vpn":                 dataSourceVPN(),
			"anypoint_bg":                  dataSourceBG(),
			"anypoint_roles":               dataSourceRoles(),
			"anypoint_rolegroup":           dataSourceRoleGroup(),
			"anypoint_rolegroups":          dataSourceRoleGroups(),
			"anypoint_users":               dataSourceUsers(),
			"anypoint_user":                dataSourceUser(),
			"anypoint_env":                 dataSourceENV(),
			"anypoint_user_rolegroup":      dataSourceUserRolegroup(),
			"anypoint_user_rolegroups":     dataSourceUserRolegroups(),
			"anypoint_team":                dataSourceTeam(),
			"anypoint_teams":               dataSourceTeams(),
			"anypoint_team_roles":          dataSourceTeamRoles(),
			"anypoint_team_members":        dataSourceTeamMembers(),
			"anypoint_team_group_mappings": dataSourceTeamGroupMappings(),
			"anypoint_dlb":                 dataSourceDLB(),
			"anypoint_dlbs":                dataSourceDLBs(),
			"anypoint_idp":                 dataSourceIDP(),
			"anypoint_idps":                dataSourceIDPs(),
			"anypoint_connected_app":       dataSourceConnectedApp(),
			"anypoint_amq":                 dataSourceAMQ(),
			"anypoint_ame":                 dataSourceAME(),
		},
		ConfigureContextFunc: providerConfigure,
		TerraformVersion:     "v1.0.1",
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client_id := d.Get("client_id").(string)
	client_secret := d.Get("client_secret").(string)
	access_token := d.Get("access_token").(string)
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
	defer httpr.Body.Close()
	return &authres, diags
}

/*
returns the server index depending on the control plane name
if the control plane is not recognized, returns -1
*/
func cplane2serverindex(cplane string) int {
	if cplane == "eu" {
		return 1
	} else if cplane == "us" {
		return 0
	} else if cplane == "gov" {
		return 2
	}
	return -1
}

type ProviderConfOutput struct {
	access_token            string
	server_index            int
	vpcclient               *vpc.APIClient
	vpnclient               *vpn.APIClient
	orgclient               *org.APIClient
	roleclient              *role.APIClient
	rolegroupclient         *rolegroup.APIClient
	userclient              *user.APIClient
	envclient               *env.APIClient
	userrgpclient           *user_rolegroups.APIClient
	teamclient              *team.APIClient
	teammembersclient       *team_members.APIClient
	teamrolesclient         *team_roles.APIClient
	teamgroupmappingsclient *team_group_mappings.APIClient
	dlbclient               *dlb.APIClient
	idpclient               *idp.APIClient
	connectedappclient      *connected_app.APIClient
	amqclient               *amq.APIClient
	ameclient               *ame.APIClient
	amebindingclient        *ame_binding.APIClient
}

func newProviderConfOutput(access_token string, server_index int) ProviderConfOutput {
	//preparing clients
	vpccfg := vpc.NewConfiguration()
	vpncfg := vpn.NewConfiguration()
	orgcfg := org.NewConfiguration()
	rolecfg := role.NewConfiguration()
	rolegroupcfg := rolegroup.NewConfiguration()
	usercfg := user.NewConfiguration()
	envcfg := env.NewConfiguration()
	userrolegroupscfg := user_rolegroups.NewConfiguration()
	teamcfg := team.NewConfiguration()
	teammemberscfg := team_members.NewConfiguration()
	teamrolescfg := team_roles.NewConfiguration()
	teamgroupmappingscfg := team_group_mappings.NewConfiguration()
	dlbcfg := dlb.NewConfiguration()
	idpcfg := idp.NewConfiguration()
	connectedappcfg := connected_app.NewConfiguration()
	amqcfg := amq.NewConfiguration()
	amecfg := ame.NewConfiguration()
	amebindingcfg := ame_binding.NewConfiguration()

	vpcclient := vpc.NewAPIClient(vpccfg)
	vpnclient := vpn.NewAPIClient(vpncfg)
	orgclient := org.NewAPIClient(orgcfg)
	roleclient := role.NewAPIClient(rolecfg)
	rolegroupclient := rolegroup.NewAPIClient(rolegroupcfg)
	userclient := user.NewAPIClient(usercfg)
	envclient := env.NewAPIClient(envcfg)
	userrgpclient := user_rolegroups.NewAPIClient(userrolegroupscfg)
	teamclient := team.NewAPIClient(teamcfg)
	teammembersclient := team_members.NewAPIClient(teammemberscfg)
	teamrolesclient := team_roles.NewAPIClient(teamrolescfg)
	teamgroupmappingsclient := team_group_mappings.NewAPIClient(teamgroupmappingscfg)
	dlbclient := dlb.NewAPIClient(dlbcfg)
	idpclient := idp.NewAPIClient(idpcfg)
	connectedappclient := connected_app.NewAPIClient(connectedappcfg)
	amqclient := amq.NewAPIClient(amqcfg)
	ameclient := ame.NewAPIClient(amecfg)
	amebindingclient := ame_binding.NewAPIClient(amebindingcfg)

	return ProviderConfOutput{
		access_token:            access_token,
		server_index:            server_index,
		vpcclient:               vpcclient,
		vpnclient:               vpnclient,
		orgclient:               orgclient,
		roleclient:              roleclient,
		rolegroupclient:         rolegroupclient,
		userclient:              userclient,
		envclient:               envclient,
		userrgpclient:           userrgpclient,
		teamclient:              teamclient,
		teammembersclient:       teammembersclient,
		teamrolesclient:         teamrolesclient,
		teamgroupmappingsclient: teamgroupmappingsclient,
		dlbclient:               dlbclient,
		idpclient:               idpclient,
		connectedappclient:      connectedappclient,
		amqclient:               amqclient,
		ameclient:               ameclient,
		amebindingclient:        amebindingclient,
	}
}
