package anypoint

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	connected_app "github.com/mulesoft-anypoint/anypoint-client-go/connected_app"
)

func resourceConnectedApp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectedAppCreate,
		ReadContext:   resourceConnectedAppRead,
		UpdateContext: resourceConnectedAppUpdate,
		DeleteContext: resourceConnectedAppDelete,
		Description: `
		Creates and manage a ` + "`" + `connected app` + "`" + `.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique id of this connected app generated by the anypoint platform.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organization id where the connected app's owner is defined.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the connected app.",
			},
			"secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The secret of the connected app.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "" {
						return true
					} else {
						return old == new
					}
				},
			},
			"user_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The id of the user who owns the connected app",
			},
			"grant_types": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(
						validation.StringInSlice(
							[]string{
								"implicit", "authorization_code", "refresh_token",
								"client_credentials", "password", "urn:ietf:params:oauth:grant-type:jwt-bearer",
							},
							false,
						),
					),
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return equalStrList(d.GetChange("grant_types"))
				},
				Description: `
				List of grant types. For "on its own behalf" connected apps the only allowed value is "client_credentials".
				The allowed values for "on behalf of user" connected apps are: "authorization_code", "refresh_token",
				"password", and "urn:ietf:params:oauth:grant-type:jwt-bearer".
				`,
			},
			"redirect_uris": {
				Description: "Configure which URIs users may be directed to after authorization",
				Type:        schema.TypeList,
				Optional:    true,
				DefaultFunc: func() (interface{}, error) {
					return make([]string, 0), nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return equalStrList(d.GetChange("redirect_uris"))
				},
			},
			"scope": {
				Description: "The scopes this connected app has authorization to work on",
				Type:        schema.TypeList,
				Optional:    true,
				DefaultFunc: func() (interface{}, error) {
					return make([]interface{}, 0), nil
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return equalsConnectedAppScopes(d.GetChange("scope"))
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"scope": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Scope",
						},
						"org_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The id of the business group the scope is valid. Only required for particular scopes",
						},
						"env_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The id of the environment the scope is valid. Only required for particular scopes",
						},
					},
				},
			},
			"public_keys": {
				Type:     schema.TypeList,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					return make([]string, 0), nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: `
				Application public key (PEM format). Used to validate JWT authorization grants.
				Required when grant type jwt-bearer is selected.
				`,
			},
			"client_uri": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `
				Users can visit this URL to learn more about your app. Required for "on behalf of user"
				connected apps
				`,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "True if the connected app is enabled",
			},
			"audience": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Who can use this application",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"internal", "everyone"}, true)),
			},
			"policy_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tos_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cert_expiry": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceConnectedAppCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	authctx := getConnectedAppAuthCtx(ctx, &pco)
	body := newConnectedAppPostBody(d)
	//request connected app creation
	res, httpr, err := pco.connectedappclient.DefaultApi.CreateConnectedApp(authctx, orgid).ConnectedAppCore(*body).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create connected-app",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.SetId(res.GetClientId())
	// Is it a "on its own behalf" connected apps?
	if grant_types, ok := body.GetGrantTypesOk(); ok && StringInSlice(grant_types, "client_credentials", true) {
		// Are there scopes to be saved?
		if scopes := d.Get("scope"); scopes != nil && len(scopes.([]interface{})) > 0 {
			// Save the connected app scopes
			if error := replaceConnectedAppScopes(authctx, d, m); error != nil {
				diags := append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to Create Connected App Scopes",
					Detail:   error.Error(),
				})
				return diags
			}
		}
	}

	return resourceConnectedAppRead(ctx, d, m)
}

func resourceConnectedAppRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	connappid := d.Id()
	if isComposedResourceId(connappid) {
		orgid, connappid = decomposeConnectedAppId(d)
	}
	authctx := getConnectedAppAuthCtx(ctx, &pco)
	//execute request
	var res *connected_app.ConnectedAppRespExt
	var httpr *http.Response
	var err error
	// perform request depending on if org_id exists or not
	if len(orgid) > 0 {
		res, httpr, err = pco.connectedappclient.DefaultApi.GetConnectedApp(authctx, orgid, connappid).Execute()
	} else {
		// NOTE: Ensuring backwards compatibility
		res, httpr, err = pco.connectedappclient.DefaultApi.GetConnectedAppByIdOnly(authctx, connappid).Execute()
	}
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read connected-app " + connappid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	connappinstance := flattenConnectedAppData(res)
	orgid = res.GetOrgId() // NOTE: Ensuring backwards compatibility
	// Is it a "on behalf of user" connected apps?
	if granttypes := connappinstance["grant_types"]; granttypes != nil && StringInSlice(granttypes.([]string), "client_credentials", true) {
		// Yes, then load the scopes using connapps/{connapp_id}/scopes
		if scopes, err := readScopesByConnectedAppId(authctx, orgid, connappid, m); err != nil {
			diags := append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read connected app " + connappid + " scopes",
				Detail:   err.Error(),
			})
			return diags
		} else {
			connappinstance["scope"] = scopes
		}
	}
	//save in data source schema
	if err := setConnectedAppAttributesToResourceData(d, connappinstance); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set connected-app " + connappid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(connappid)
	d.Set("org_id", orgid)
	return diags
}

func resourceConnectedAppUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	connappid := d.Id()
	authctx := getConnectedAppAuthCtx(ctx, &pco)
	if d.HasChanges(getConnectedAppAttributes()...) {
		body := newConnectedAppPatchBody(d)
		//perform request
		_, httpr, err := pco.connectedappclient.DefaultApi.UpdateConnectedApp(authctx, orgid, connappid).ConnectedAppPatchExt(*body).Execute()
		if err != nil {
			var details string
			if httpr != nil && httpr.StatusCode >= 400 {
				defer httpr.Body.Close()
				b, _ := io.ReadAll(httpr.Body)
				details = string(b)
			} else {
				details = err.Error()
			}
			diags := append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update connected-app " + connappid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		// Is it a "on its own behalf" connected apps?
		if grant_types, ok := body.GetGrantTypesOk(); ok && StringInSlice(grant_types, "client_credentials", true) {
			// Are there scopes to be saved?
			if scopes := d.Get("scope"); scopes != nil && len(scopes.([]interface{})) > 0 {
				// Save the connected app scopes
				if error := replaceConnectedAppScopes(authctx, d, m); error != nil {
					diags := append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Unable to create connected-app " + connappid + " scopes",
						Detail:   error.Error(),
					})
					return diags
				}
			}
		}
		return resourceConnectedAppRead(ctx, d, m)
	}
	return diags
}

func resourceConnectedAppDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	connappid := d.Id()
	authctx := getConnectedAppAuthCtx(ctx, &pco)
	// perform request
	httpr, err := pco.connectedappclient.DefaultApi.DeleteConnectedApp(authctx, orgid, connappid).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete connected-app " + connappid,
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

func replaceConnectedAppScopes(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	connappid := d.Id()
	authctx := getConnectedAppAuthCtx(ctx, &pco)
	body := newConnectedAppScopesPutBody(d)
	//request scopes replacement
	httpr, err := pco.connectedappclient.DefaultApi.UpdateConnectedAppScopes(authctx, orgid, connappid).ConnectedAppScopesPutBody(*body).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		return errors.New(details)
	}
	defer httpr.Body.Close()
	return nil
}

func newConnectedAppScopesPutBody(d *schema.ResourceData) *connected_app.ConnectedAppScopesPutBody {
	body := connected_app.NewConnectedAppScopesPutBodyWithDefaults()
	// Is there any scope set in the resource?
	if scopes, ok := d.GetOk("scope"); ok {
		scopes_list := scopes.([]interface{})
		scopes_body := make([]connected_app.ScopeCore, len(scopes_list))
		for i, scope := range scopes_list {
			scope_map := scope.(map[string]interface{})
			scope_core := connected_app.NewScopeCoreWithDefaults()
			if val, ok := scope_map["scope"]; ok {
				scope_core.SetScope(val.(string))
			}
			context_params := connected_app.NewContextParamsWithDefaults()
			if org_id, ok := scope_map["org_id"]; ok && org_id != "" {
				context_params.SetOrg(org_id.(string))
			}
			if env_id, ok := scope_map["env_id"]; ok && env_id != "" {
				context_params.SetEnvId(env_id.(string))
			}
			scope_core.SetContextParams(*context_params)
			scopes_body[i] = *scope_core
		}
		body.SetScopes(scopes_body)
	}

	return body
}

/*
 * Creates a new connected app Core Struct from the resource data schema
 */
func newConnectedAppPostBody(d *schema.ResourceData) *connected_app.ConnectedAppCore {
	body := connected_app.NewConnectedAppCoreWithDefaults()

	// connected_app.ConnectedAppCore attributes
	body.SetClientName(d.Get("name").(string))
	body.SetGrantTypes(ListInterface2ListStrings(d.Get("grant_types").([]interface{})))
	body.SetAudience(d.Get("audience").(string))

	// Required by Anypoint endpoint, but value can be an empty list
	if publickeys, ok := d.GetOk("public_keys"); ok {
		body.SetPublicKeys(ListInterface2ListStrings(publickeys.([]interface{})))
	} else {
		body.SetPublicKeys(make([]string, 0))
	}

	// Required by Anypoint endpoint, but value can be an empty list
	if redirecturis, ok := d.GetOk("redirect_uris"); ok {
		body.SetRedirectUris(ListInterface2ListStrings(redirecturis.([]interface{})))
	} else {
		body.SetRedirectUris(make([]string, 0))
	}
	// Is it a "on behalf of user" connected apps?
	if grant_types, ok := body.GetGrantTypesOk(); ok && !StringInSlice(grant_types, "client_credentials", true) {
		// Is there any scope set in the resource?
		if scopes := d.Get("scope"); scopes != nil {
			scopes_list := scopes.([]interface{})
			scopes_body := make([]string, len(scopes_list))
			for i, scope := range scopes_list {
				scope_map := scope.(map[string]interface{})
				if val := scope_map["scope"]; val != nil {
					scopes_body[i] = val.(string)
				}
			}
			body.SetScopes(scopes_body)
		}
	} else {
		body.SetScopes(make([]string, 0))
	}
	if clienturi, ok := d.GetOk("client_uri"); ok {
		body.SetClientUri(clienturi.(string))
	}
	return body
}

/*
 * Creates a new connected app patch Struct from the resource data schema
 */
func newConnectedAppPatchBody(d *schema.ResourceData) *connected_app.ConnectedAppPatchExt {
	body := connected_app.NewConnectedAppPatchExtWithDefaults()

	// connected_app.ConnectedAppCore attributes
	body.SetClientName(d.Get("name").(string))
	body.SetGrantTypes(ListInterface2ListStrings(d.Get("grant_types").([]interface{})))
	body.SetAudience(d.Get("audience").(string))

	// Required by Anypoint endpoint, but value can be an empty list
	if publickeys, ok := d.GetOk("public_keys"); ok {
		body.SetPublicKeys(ListInterface2ListStrings(publickeys.([]interface{})))
	} else {
		body.SetPublicKeys(make([]string, 0))
	}

	// Required by Anypoint endpoint, but value can be an empty list
	if redirecturis, ok := d.GetOk("redirect_uris"); ok {
		body.SetRedirectUris(ListInterface2ListStrings(redirecturis.([]interface{})))
	} else {
		body.SetRedirectUris(make([]string, 0))
	}
	// Is it a "on behalf of user" connected apps?
	if val, ok := body.GetGrantTypesOk(); ok && !StringInSlice(val, "client_credentials", true) {
		// Is there any scope set in the resource?
		if scopes := d.Get("scope"); scopes != nil {
			scopes_list := scopes.([]interface{})
			scopes_body := make([]string, len(scopes_list))
			for i, scope := range scopes_list {
				scope_map := scope.(map[string]interface{})
				if val, ok := scope_map["scope"]; ok {
					scopes_body[i] = val.(string)
				}
			}
			body.SetScopes(scopes_body)
		}
	} else {
		body.SetScopes(make([]string, 0))
	}
	if clienturi, ok := d.GetOk("client_uri"); ok {
		body.SetClientUri(clienturi.(string))
	}
	// connected_app.ConnectedAppPatchExt extra attributes
	if secret, ok := d.GetOk("secret"); ok {
		body.SetClientSecret(secret.(string))
	}
	if enabled, ok := d.GetOk("enabled"); ok {
		body.SetEnabled(enabled.(bool))
	}
	return body
}

// Compares 2 scopes lists
// returns true if they are the same, false otherwise
func equalsConnectedAppScopes(old, new interface{}) bool {
	old_list := old.([]interface{})
	new_list := new.([]interface{})

	old_list = removeProfileScope(old_list)
	new_list = removeProfileScope(new_list)

	if len(new_list) != len(old_list) {
		return false
	}

	if len(new_list) == 0 {
		return true
	}

	sortScopes(old_list)
	sortScopes(new_list)

	for i, val := range old_list {
		o := val.(map[string]interface{})
		n := new_list[i].(map[string]interface{})

		old_scope := o["scope"].(string)
		new_scope := n["scope"].(string)

		if old_scope != new_scope {
			return false
		}

		old_org_id := o["org_id"]
		new_org_id := n["org_id"]

		if old_org_id != new_org_id {
			return false
		}

		old_env_id := o["env_id"]
		new_env_id := n["env_id"]

		if old_env_id != new_env_id {
			return false
		}
	}

	return true
}

// "Profile" is a defaul scope added to "act on its own behalf" connected apps.
// It should not be considered when dealing with this kind of connecte app
func removeProfileScope(scopes []interface{}) []interface{} {
	if len(scopes) == 0 {
		return scopes
	}

	for i, scope := range scopes {
		scope_map := scope.(map[string]interface{})

		if strings.EqualFold(scope_map["scope"].(string), "profile") {
			scopes[i] = scopes[len(scopes)-1]
			return scopes[:len(scopes)-1]
		}
	}

	return scopes
}

func sortScopes(list []interface{}) {
	sort.SliceStable(list, func(i, j int) bool {
		i_elem := list[i].(map[string]interface{})
		j_elem := list[j].(map[string]interface{})
		i_scope := i_elem["scope"].(string)
		j_scope := j_elem["scope"].(string)
		if i_scope != j_scope {
			return i_scope < j_scope
		}
		i_org_id := i_elem["org_id"].(string)
		j_org_id := j_elem["org_id"].(string)
		if i_org_id != j_org_id {
			return i_org_id < j_org_id
		}
		i_env_id := i_elem["env_id"].(string)
		j_env_id := j_elem["env_id"].(string)
		if i_env_id != j_env_id {
			return i_env_id < j_env_id
		}
		return true
	})
}

/*
 * Returns authentication context (includes authorization header)
 */
func getConnectedAppAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, connected_app.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, connected_app.ContextServerIndex, pco.server_index)
}

func decomposeConnectedAppId(d *schema.ResourceData) (string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1]
}
