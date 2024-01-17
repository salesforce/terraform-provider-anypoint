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

var SG_TLS_CONTEXT_MULE_TARGET = "Mule"

func resourceSecretGroupTlsContextMule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretGroupTlsContextMuleCreate,
		ReadContext:   resourceSecretGroupTlsContextMuleRead,
		UpdateContext: resourceSecretGroupTlsContextMuleUpdate,
		DeleteContext: resourceSecretGroupTlsContextMuleDelete,
		Description: `
		Create and manage tls-context of type "Mule" for a secret-group in a given organization and environment.
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
				Description: "The secret-group id where the tls-context instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the tls-context's secret group is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
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
			"acceptable_tls_versions": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "TLS versions supported.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tls_v1_dot1": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "TLS version 1.1",
						},
						"tls_v1_dot2": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "TLS version 1.2",
						},
						"tls_v1_dot3": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "TLS version 1.3",
						},
					},
				},
			},
			"cipher_suites": {
				Type:        schema.TypeSet,
				Optional:    true,
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
			"insecure": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Setting this flag to true indicates that certificate validation should not be enforced, i.e. the truststore, even though set, is ignored at runtime. Only available for \"Mule\" target",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSecretGroupTlsContextMuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	name := d.Get("name").(string)
	authctx := getSgTlsContextAuthCtx(ctx, &pco)
	//prepare request
	body := newSgTlsContextMulePostBody(d)
	//perform request
	res, httpr, err := pco.sgtlscontextclient.DefaultApi.PostSecretGroupTlsContext(authctx, orgid, envid, sgid).TlsContextPostBody(*body).Execute()
	if err != nil {
		var details string
		if httpr != nil {
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
	return resourceSecretGroupTlsContextMuleRead(ctx, d, m)
}

func resourceSecretGroupTlsContextMuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	id := d.Get("id").(string)
	authctx := getSgTlsContextAuthCtx(ctx, &pco)
	if isComposedResourceId(id) {
		orgid, envid, sgid, id = decomposeSgTlsContextMuleId(d)
	}
	//perform request
	res, httpr, err := pco.sgtlscontextclient.DefaultApi.GetSecretGroupTlsContextDetails(authctx, orgid, envid, sgid, id).Execute()
	if err != nil {
		// var details string
		// if httpr != nil {
		// 	b, _ := io.ReadAll(httpr.Body)
		// 	details = string(b)
		// } else {
		// 	details = err.Error()
		// }
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read tls-context " + id,
			Detail:   err.Error(),
		})
		return diags
	}
	defer httpr.Body.Close()
	//process result
	if !isSgTlsContextMule(res) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong target type for tls-context " + id,
			Detail:   "source is not of type Mule",
		})
		return diags
	}
	data := flattenSgTlsContextMule(res)
	if err := setSgTlsContextMuleAttributesToResourceData(d, data); err != nil {
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

func resourceSecretGroupTlsContextMuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if d.HasChanges(getSgTlsContextMuleUpdatableAttributes()...) {
		pco := m.(ProviderConfOutput)
		orgid := d.Get("org_id").(string)
		envid := d.Get("env_id").(string)
		sgid := d.Get("sg_id").(string)
		id := d.Get("id").(string)
		authctx := getSgTlsContextAuthCtx(ctx, &pco)
		//prepare body
		body := newSgTlsContextMulePutBody(d)
		// perform request
		_, httpr, err := pco.sgtlscontextclient.DefaultApi.PutSecretGroupTlsContext(authctx, orgid, envid, sgid, id).TlsContextPutBody(*body).Execute()
		if err != nil {
			var details string
			if httpr != nil {
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
		return resourceSecretGroupTlsContextMuleRead(ctx, d, m)
	}
	return diags
}

func resourceSecretGroupTlsContextMuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// NOTE: The delete action is not supported for this resource.
	// a tls-context cannot be deleted, only secret-group (parent) can be deleted
	// Therefore we are only removing reference here
	d.SetId("")
	return diags
}

func newSgTlsContextMulePostBody(d *schema.ResourceData) *secretgroup_tlscontext.TlsContextPostBody {
	body := secretgroup_tlscontext.TlsContextPostBody{TlsContextMuleBody: newSgTlsContextMuleBody(d)}
	return &body
}

func newSgTlsContextMulePutBody(d *schema.ResourceData) *secretgroup_tlscontext.TlsContextPutBody {
	body := secretgroup_tlscontext.TlsContextPutBody{TlsContextMuleBody: newSgTlsContextMuleBody(d)}
	return &body
}

func newSgTlsContextMuleBody(d *schema.ResourceData) *secretgroup_tlscontext.TlsContextMuleBody {
	body := secretgroup_tlscontext.NewTlsContextMuleBody()
	body.SetTarget(SG_TLS_CONTEXT_MULE_TARGET)
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
	if val, ok := d.GetOk("acceptable_tls_versions"); ok {
		list := val.([]interface{})
		if len(list) > 0 {
			item := list[0].(map[string]interface{})
			versions := secretgroup_tlscontext.NewAcceptableTlsVersions()
			if val, ok := item["tls_v1_dot1"]; ok {
				versions.SetTlsV1Dot1(val.(bool))
			}
			if val, ok := item["tls_v1_dot2"]; ok {
				versions.SetTlsV1Dot2(val.(bool))
			}
			if val, ok := item["tls_v1_dot3"]; ok {
				versions.SetTlsV1Dot3(val.(bool))
			}
			body.SetAcceptableTlsVersions(*versions)
		}
	}
	if val, ok := d.GetOk("cipher_suites"); ok {
		body.SetCipherSuites(ListInterface2ListStrings(val.([]interface{})))
	}
	if val, ok := d.GetOk("insecure"); ok {
		body.SetInsecure(val.(bool))
	}
	return body
}

func getSgTlsContextMuleUpdatableAttributes() []string {
	attributes := [...]string{
		"name", "keystore_path", "truststore_path",
		"cipher_suites", "acceptable_tls_versions", "insecure",
	}
	return attributes[:]
}

// returns the composed of the secret
func decomposeSgTlsContextMuleId(d *schema.ResourceData) (string, string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2], s[3]
}
