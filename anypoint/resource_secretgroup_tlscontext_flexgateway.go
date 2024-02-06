package anypoint

import (
	"context"
	"io"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	secretgroup_tlscontext "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_tlscontext"
)

var SG_TLS_CONTEXT_FG_TARGET = "FlexGateway"

func resourceSecretGroupTlsContextFG() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretGroupTlsContextFGCreate,
		ReadContext:   resourceSecretGroupTlsContextFGRead,
		UpdateContext: resourceSecretGroupTlsContextFGUpdate,
		DeleteContext: resourceSecretGroupTlsContextFGDelete,
		Description: `
		Create and Manage tls-context of type "FlexGateway" for a secret-group in a given organization and environment.
		This resource doesn't support delete. The delete operation only removes the resource from local terraform state file.
		Only the parent resource (secret-group) can be deleted.
		`,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last time this resource has been updated locally.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Id assigned to this tls-context",
			},
			"sg_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The secret-group id where the tls-context instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organization id where the tls-context's secret group is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The environment id where the tls-context's secret group is defined.",
			},
			"path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The path of the tls-context",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the tls-context",
			},
			"expiration_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The expiration date of the tls-context",
			},
			"target": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The target application for the tls-context",
			},
			"keystore_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Refers to a secret of type keystore. Relative path of the secret to be referenced.",
			},
			"truststore_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Refers to a secret of type truststore. Relative path of the secret to be referenced.",
			},
			"cipher_suites": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of enabled cipher suites for Mule target.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(
						validation.StringInSlice(
							[]string{
								"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_PSK_WITH_CHACHA20_POLY1305_SHA256",
								"TLS_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_RSA_WITH_AES_256_GCM_SHA384",
								"TLS_RSA_WITH_NULL_SHA",
								"TLS_RSA_WITH_AES_128_CBC_SHA",
								"TLS_RSA_WITH_AES_256_CBC_SHA",
								"TLS_PSK_WITH_AES_128_CBC_SHA",
								"TLS_PSK_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_PSK_WITH_AES_256_CBC_SHA",
								"TLS_RSA_WITH_3DES_EDE_CBC_SHA",
							},
							false,
						),
					),
				},
			},
			"min_tls_version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Minimum TLS version supported.",
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice(
						[]string{"TLSv1.1", "TLSv1.2", "TLSv1.3"},
						false,
					),
				),
			},
			"max_tls_version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Maximum TLS version supported.",
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice(
						[]string{"TLSv1.1", "TLSv1.2", "TLSv1.3"},
						false,
					),
				),
			},
			"alpn_protocols": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "supported HTTP versions in the most-to-least preferred order. At least one version must be specified.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(
						validation.StringInSlice(
							[]string{"h2", "http/1.1"},
							false,
						),
					),
				},
			},
			"inbound_settings": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Properties that are applicable only when the TLS context is used to secure inbound traffic.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_client_cert_validation": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Indicates whether the client certificate validation must be enforced.",
						},
					},
				},
			},
			"outbound_settings": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Properties that are applicable only when the TLS context is used to secure outbound traffic.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"skip_server_cert_validation": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "flag that indicates whether the server certificate validation must be skipped.",
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSecretGroupTlsContextFGCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	name := d.Get("name").(string)
	authctx := getSgTlsContextAuthCtx(ctx, &pco)
	//prepare request
	body := newSgTlsContextFGPostBody(d)
	//perform request
	res, httpr, err := pco.sgtlscontextclient.DefaultApi.PostSecretGroupTlsContext(authctx, orgid, envid, sgid).TlsContextPostBody(*body).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create tls-context " + name,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(res.GetId())

	return resourceSecretGroupTlsContextFGRead(ctx, d, m)
}

func resourceSecretGroupTlsContextFGRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	id := d.Get("id").(string)
	authctx := getSgTlsContextAuthCtx(ctx, &pco)
	if isComposedResourceId(id) {
		orgid, envid, sgid, id = decomposeSgTlsContextFGId(d)
	}
	//perform request
	res, httpr, err := pco.sgtlscontextclient.DefaultApi.GetSecretGroupTlsContextDetails(authctx, orgid, envid, sgid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get tls-context " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process result
	if !isSgTlsContextFlexGateway(res) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong target type for tls-context " + id,
			Detail:   "source is not of type FlexGateway",
		})
		return diags
	}
	data := flattenSgTlsContextFlexGateway(res)
	if err := setSgTlsContextFGAttributesToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set tls-context " + id + " details attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(id)
	d.Set("sg_id", sgid)
	d.Set("env_id", envid)
	d.Set("org_id", orgid)
	return diags
}

func resourceSecretGroupTlsContextFGUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if d.HasChanges(getSgTlsContextFGUpdatableAttributes()...) {
		pco := m.(ProviderConfOutput)
		orgid := d.Get("org_id").(string)
		envid := d.Get("env_id").(string)
		sgid := d.Get("sg_id").(string)
		id := d.Get("id").(string)
		authctx := getSgTlsContextAuthCtx(ctx, &pco)
		//prepare body
		body := newSgTlsContextFGPutBody(d)
		// perform request
		_, httpr, err := pco.sgtlscontextclient.DefaultApi.PutSecretGroupTlsContext(authctx, orgid, envid, sgid, id).TlsContextPutBody(*body).Execute()
		if err != nil {
			var details string
			if httpr != nil && httpr.StatusCode >= 400 {
				defer httpr.Body.Close()
				b, _ := io.ReadAll(httpr.Body)
				details = string(b)
			} else {
				details = err.Error()
			}
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update tls-context " + id,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		d.Set("last_updated", time.Now().Format(time.RFC850))
		return resourceSecretGroupTlsContextFGRead(ctx, d, m)
	}
	return diags
}

func resourceSecretGroupTlsContextFGDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// NOTE: The delete action is not supported for this resource.
	// a tls-context cannot be deleted, only secret-group (parent) can be deleted
	// Therefore we are only removing reference here
	d.SetId("")
	return diags
}

func newSgTlsContextFGPostBody(d *schema.ResourceData) *secretgroup_tlscontext.TlsContextPostBody {
	body := secretgroup_tlscontext.TlsContextPostBody{TlsContextFlexGatewayBody: newSgTlsContextFlexGatewayBody(d)}
	return &body
}

func newSgTlsContextFGPutBody(d *schema.ResourceData) *secretgroup_tlscontext.TlsContextPutBody {
	body := secretgroup_tlscontext.TlsContextPutBody{TlsContextFlexGatewayBody: newSgTlsContextFlexGatewayBody(d)}
	return &body
}

func newSgTlsContextFlexGatewayBody(d *schema.ResourceData) *secretgroup_tlscontext.TlsContextFlexGatewayBody {
	body := secretgroup_tlscontext.NewTlsContextFlexGatewayBody()
	body.SetTarget(SG_TLS_CONTEXT_FG_TARGET)
	if val, ok := d.GetOk("name"); ok {
		body.SetName(val.(string))
	}
	if val, ok := d.GetOk("keystore_path"); ok {
		keystore := secretgroup_tlscontext.NewSecretPath()
		keystore.SetPath(val.(string))
		body.SetKeystore(*keystore)
	}
	if val, ok := d.GetOk("truststore_path"); ok {
		truststore := secretgroup_tlscontext.NewSecretPath()
		truststore.SetPath(val.(string))
		body.SetTruststore(*truststore)
	}
	if val, ok := d.GetOk("cipher_suites"); ok {
		set := val.(*schema.Set)
		body.SetCipherSuites(ListInterface2ListStrings(set.List()))
	}
	if val, ok := d.GetOk("min_tls_version"); ok {
		body.SetMinTlsVersion(val.(string))
	}
	if val, ok := d.GetOk("max_tls_version"); ok {
		body.SetMaxTlsVersion(val.(string))
	}
	if val, ok := d.GetOk("alpn_protocols"); ok {
		set := val.(*schema.Set)
		body.SetAlpnProtocols(ListInterface2ListStrings(set.List()))
	}
	if val, ok := d.GetOk("inbound_settings"); ok {
		list := val.([]interface{})
		if len(list) > 0 {
			inbound := list[0].(map[string]interface{})
			if val, ok := inbound["enable_client_cert_validation"]; ok {
				setting := secretgroup_tlscontext.NewTlsContextFlexGatewayBodyInboundSettings()
				setting.SetEnableClientCertValidation(val.(bool))
				body.SetInboundSettings(*setting)
			}
		}
	}
	if val, ok := d.GetOk("outbound_settings"); ok {
		list := val.([]interface{})
		if len(list) > 0 {
			outbound := list[0].(map[string]interface{})
			if val, ok := outbound["skip_server_cert_validation"]; ok {
				setting := secretgroup_tlscontext.NewTlsContextFlexGatewayBodyOutboundSettings()
				setting.SetSkipServerCertValidation(val.(bool))
				body.SetOutboundSettings(*setting)
			}
		}
	}
	return body
}

func getSgTlsContextFGUpdatableAttributes() []string {
	attributes := [...]string{
		"name", "keystore_path", "truststore_path",
		"cipher_suites", "min_tls_version", "max_tls_version",
		"alpn_protocols", "inbound_settings", "outbound_settings",
	}
	return attributes[:]
}

// returns the composed of the secret
func decomposeSgTlsContextFGId(d *schema.ResourceData) (string, string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2], s[3]
}
