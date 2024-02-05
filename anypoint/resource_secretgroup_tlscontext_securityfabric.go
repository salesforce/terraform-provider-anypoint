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

var SG_TLS_CONTEXT_SF_TARGET = "SecurityFabric"

func resourceSecretGroupTlsContextSF() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretGroupTlsContextSFCreate,
		ReadContext:   resourceSecretGroupTlsContextSFRead,
		UpdateContext: resourceSecretGroupTlsContextSFUpdate,
		DeleteContext: resourceSecretGroupTlsContextSFDelete,
		Description: `
		Create and manage tls-context of type security-fabric for a secret-group in a given organization and environment.
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
			"enable_mutual_authentication": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "This flag is to enable client authentication.",
			},
			"acceptable_cipher_suites": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Description: `
				List of accepted cipher suites by Security Fabric target, at least one should be set to true. If you are are not using the defaults and select individual ciphers, please select ciphers that match the configured keystore to ensure that TLS can setup a connection.
        For a keystore with an RSA key (the most common type), select ciphers which contain the string RSA (there are some exceptions). If using ECC ciphers, select ciphers which contain the string "ECDSA".
        TLS standards and documentation can be consulted for more background information.
				`,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aes128_gcm_sha256": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"aes128_sha256": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"aes256_gcm_sha384": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"aes256_sha256": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"dhe_rsa_aes128_gcm_sha256": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"dhe_rsa_aes128_sha256": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"dhe_rsa_aes256_gcm_sha384": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"dhe_rsa_aes256_sha256": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_ecdsa_aes128_gcm_sha256": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_ecdsa_aes128_sha1": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_ecdsa_aes256_gcm_sha384": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_ecdsa_aes256_sha1": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_rsa_aes128_gcm_sha256": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_rsa_aes128_sha1": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_rsa_aes256_gcm_sha384": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_rsa_aes256_sha1": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_ecdsa_chacha20_poly1305": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_rsa_chacha20_poly1305": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"dhe_rsa_chacha20_poly1305": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"tls_aes256_gcm_sha384": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"tls_chacha20_poly1305_sha256": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"tls_aes128_gcm_sha256": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
					},
				},
			},
			"mutual_authentication": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Configuration for client authentication.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_policies": {
							Type:     schema.TypeList,
							Optional: true,
							Description: `
							List of Object identifier (OID).
							OIDs are intended to be globally unique.
							They are formed by taking a unique numeric string (e.g. 1.3.5.7.9.24.68) and adding additional digits in a unique fashion (e.g. 1.3.5.7.9.24.68.1, 1.3.5.7.9.24.68.2, 1.3.5.7.9.24.68.1.1, etc.) An institution will acquire an arc (eg 1.3.5.7.9.24.68) and then extend the arc (called subarcs) as indicated above to create additional OID’s and arcs.
							There is no limit to the length of an OID, and virtually no computational burden to having a long OID.
							`,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"cert_checking_strength": {
							Type:     schema.TypeString,
							Required: true,
							Description: `
							Allows application to control if strict or lax certificate checking will be performed during chain-of-trust processing. 
							Supported values are "Strict" and "Lax"
							`,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"Strict", "Lax"}, false),
							),
						},
						"verification_depth": {
							Type:             schema.TypeInt,
							Optional:         true,
							Default:          1,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 9)),
							Description:      "Maximum allowed chain length for the certificates. Allowed values between 0-9",
						},
						"perform_domain_checking": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether or not to perform domain checking",
						},
						"certificate_policy_checking": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							Description: `
							Controls certificate policy processing as defined in RFC 3280, 5280. A certificate can contain zero or more policies.
							A policy is represented as an object identifier (OID). In an end entity certificate, this policy information indicate the policy under which the certificate has been issued and the purposes for which the certificate may be used.
							In a CA certificate, this policy information limits the set of policies for certification paths that include this certificate. Applications with specific policy requirements are expected to have a list of those policies that they will accept and to compare the policy OIDs in the certificate to that list.
							If this extension is critical, the path validation software MUST be able to interpret this extension (including the optional qualifier), or MUST reject the certificate
							`,
						},
						"require_initial_explicit_policy": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Indicates if the path must be valid for at least one of the certificate policies in the user-initial-policy-set.",
						},
						"revocation_checking": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Indicates if certificate revocation checking should be enabled or not.",
						},
						"revocation_checking_method": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Protocol used for certificate revocation checking. Must be set if revocationChecking is set to 'true'.",
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"CRL"}, false),
							),
						},
						"crl_distributor_config_path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Reference to a secret of type crlDistributorConfig. Must be set if revocationCheckingMethod is set to 'CRL'.",
						},
						"require_crl_for_all_ca": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Indicates if a valid CRL file must be in effect for every immediate and root Certificate Authority (CA) in the chain-of-trust",
						},
						"send_truststore": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Should the truststore i.e. trusted certificate authorities be sent to far-end during mutual authentication",
						},
						"certificate_pinning": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Referes to pinned certificates",
						},
						"authentication_overrides": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Override failing authentication when mutual authentication is being performed",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"certificate_bad_format": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Allow processing of certificates with bad format",
									},
									"certificate_bad_signature": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Allow processing of certificates with bad signature",
									},
									"certificate_not_yet_valid": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Allow processing of certificates that are not yet valid",
									},
									"certificate_has_expired": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Allow processing of certificates that are expired",
									},
									"allow_self_signed": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Allow self signed certificates",
									},
									"certificate_unresolved": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Allow unresolved certificates",
									},
									"certificate_untrusted": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Allow untrusted certificates",
									},
									"invalid_ca": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Allow invalid certificate authority certificates",
									},
									"invalid_purpose": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Allow certificates with invalid purpose",
									},
									"other": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Override any miscellaneous error condition encountered",
									},
								},
							},
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

func resourceSecretGroupTlsContextSFCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	name := d.Get("name").(string)
	authctx := getSgTlsContextAuthCtx(ctx, &pco)
	//prepare request
	body := newSgTlsContextSFPostBody(d)
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

	return resourceSecretGroupTlsContextSFRead(ctx, d, m)
}

func resourceSecretGroupTlsContextSFRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	id := d.Get("id").(string)
	if isComposedResourceId(id) {
		orgid, envid, sgid, id = decomposeSgTlsContextSFId(d)
	}
	authctx := getSgTlsContextAuthCtx(ctx, &pco)
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
			Summary:  "Unable to read tls-context " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process result
	if !isSgTlsContextSecurityFabric(res) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong target type for tls-context " + id,
			Detail:   "source is not of type SecuritFabric",
		})
		return diags
	}
	data := flattenSgTlsContextSF(res)
	if err := setSgTlsContextSFAttributesToResourceData(d, data); err != nil {
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

func resourceSecretGroupTlsContextSFUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if d.HasChanges(getSgTlsContextSFUpdatableAttributes()...) {
		pco := m.(ProviderConfOutput)
		orgid := d.Get("org_id").(string)
		envid := d.Get("env_id").(string)
		sgid := d.Get("sg_id").(string)
		id := d.Get("id").(string)
		authctx := getSgTlsContextAuthCtx(ctx, &pco)
		//prepare body
		body := newSgTlsContextSFPutBody(d)
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
		return resourceSecretGroupTlsContextMuleRead(ctx, d, m)
	}

	return diags
}

func resourceSecretGroupTlsContextSFDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// NOTE: The delete action is not supported for this resource.
	// a tls-context cannot be deleted, only secret-group (parent) can be deleted
	// Therefore we are only removing reference here
	d.SetId("")
	return diags
}

func newSgTlsContextSFPostBody(d *schema.ResourceData) *secretgroup_tlscontext.TlsContextPostBody {
	body := secretgroup_tlscontext.TlsContextPostBody{TlsContextSfBody: newSgTlsContextSFBody(d)}
	return &body
}

func newSgTlsContextSFPutBody(d *schema.ResourceData) *secretgroup_tlscontext.TlsContextPutBody {
	body := secretgroup_tlscontext.TlsContextPutBody{TlsContextSfBody: newSgTlsContextSFBody(d)}
	return &body
}

func newSgTlsContextSFBody(d *schema.ResourceData) *secretgroup_tlscontext.TlsContextSfBody {
	body := secretgroup_tlscontext.NewTlsContextSfBody()
	body.SetTarget(SG_TLS_CONTEXT_SF_TARGET)
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
	if val, ok := d.GetOk("enable_mutual_authentication"); ok {
		body.SetEnableMutualAuthentication(val.(bool))
	}
	if val, ok := d.GetOk("acceptable_cipher_suites"); ok {
		list := val.([]interface{})
		if len(list) > 0 {
			acs := newSgTlsContextSFAcceptableCipherSuites(list[0].(map[string]interface{}))
			body.SetAcceptableCipherSuites(*acs)
		}
	}
	if val, ok := d.GetOk("mutual_authentication"); ok {
		list := val.([]interface{})
		if len(list) > 0 {
			ma := newSgTlsContextSFMutualAuthentication(list[0].(map[string]interface{}))
			body.SetMutualAuthentication(*ma)
		}
	}
	return body
}

func newSgTlsContextSFAcceptableCipherSuites(data map[string]interface{}) *secretgroup_tlscontext.AcceptableCipherSuites {
	body := secretgroup_tlscontext.NewAcceptableCipherSuites()
	if val, ok := data["aes128_gcm_sha256"]; ok {
		body.SetAes128GcmSha256(val.(bool))
	}
	if val, ok := data["aes128_sha256"]; ok {
		body.SetAes128Sha256(val.(bool))
	}
	if val, ok := data["aes256_gcm_sha384"]; ok {
		body.SetAes256GcmSha384(val.(bool))
	}
	if val, ok := data["aes256_sha256"]; ok {
		body.SetAes256Sha256(val.(bool))
	}
	if val, ok := data["dhe_rsa_aes128_gcm_sha256"]; ok {
		body.SetDheRsaAes128GcmSha256(val.(bool))
	}
	if val, ok := data["dhe_rsa_aes128_sha256"]; ok {
		body.SetDheRsaAes128Sha256(val.(bool))
	}
	if val, ok := data["dhe_rsa_aes256_gcm_sha384"]; ok {
		body.SetDheRsaAes256GcmSha384(val.(bool))
	}
	if val, ok := data["dhe_rsa_aes256_sha256"]; ok {
		body.SetDheRsaAes256Sha256(val.(bool))
	}
	if val, ok := data["ecdhe_ecdsa_aes128_gcm_sha256"]; ok {
		body.SetEcdheEcdsaAes128GcmSha256(val.(bool))
	}
	if val, ok := data["ecdhe_ecdsa_aes128_sha1"]; ok {
		body.SetEcdheEcdsaAes128Sha1(val.(bool))
	}
	if val, ok := data["ecdhe_ecdsa_aes256_gcm_sha384"]; ok {
		body.SetEcdheEcdsaAes256GcmSha384(val.(bool))
	}
	if val, ok := data["ecdhe_ecdsa_aes256_sha1"]; ok {
		body.SetEcdheEcdsaAes256Sha1(val.(bool))
	}
	if val, ok := data["ecdhe_rsa_aes128_gcm_sha256"]; ok {
		body.SetEcdheRsaAes128GcmSha256(val.(bool))
	}
	if val, ok := data["ecdhe_rsa_aes128_sha1"]; ok {
		body.SetEcdheRsaAes128Sha1(val.(bool))
	}
	if val, ok := data["ecdhe_rsa_aes256_gcm_sha384"]; ok {
		body.SetEcdheRsaAes256GcmSha384(val.(bool))
	}
	if val, ok := data["ecdhe_rsa_aes256_sha1"]; ok {
		body.SetEcdheRsaAes256Sha1(val.(bool))
	}
	if val, ok := data["ecdhe_ecdsa_chacha20_poly1305"]; ok {
		body.SetEcdheEcdsaChacha20Poly1305(val.(bool))
	}
	if val, ok := data["ecdhe_rsa_chacha20_poly1305"]; ok {
		body.SetEcdheRsaChacha20Poly1305(val.(bool))
	}
	if val, ok := data["dhe_rsa_chacha20_poly1305"]; ok {
		body.SetDheRsaChacha20Poly1305(val.(bool))
	}
	if val, ok := data["tls_aes256_gcm_sha384"]; ok {
		body.SetTlsAes256GcmSha384(val.(bool))
	}
	if val, ok := data["tls_chacha20_poly1305_sha256"]; ok {
		body.SetTlsChacha20Poly1305Sha256(val.(bool))
	}
	if val, ok := data["tls_aes128_gcm_sha256"]; ok {
		body.SetTlsAes128GcmSha256(val.(bool))
	}
	return body
}

func newSgTlsContextSFMutualAuthentication(data map[string]interface{}) *secretgroup_tlscontext.MutualAuthentication {
	body := secretgroup_tlscontext.NewMutualAuthentication()
	if val, ok := data["certificate_policies"]; ok {
		body.SetCertificatePolicies(ListInterface2ListStrings(val.([]interface{})))
	}
	if val, ok := data["cert_checking_strength"]; ok {
		body.SetCertCheckingStrength(val.(string))
	}
	if val, ok := data["verification_depth"]; ok {
		body.SetVerificationDepth(int32(val.(int)))
	}
	if val, ok := data["perform_domain_checking"]; ok {
		body.SetPerformDomainChecking(val.(bool))
	}
	if val, ok := data["certificate_policy_checking"]; ok {
		body.SetCertificatePolicyChecking(val.(bool))
	}
	if val, ok := data["require_initial_explicit_policy"]; ok {
		body.SetRequireInitialExplicitPolicy(val.(bool))
	}
	if val, ok := data["revocation_checking"]; ok {
		body.SetRevocationChecking(val.(bool))
	}
	if val, ok := data["revocation_checking_method"]; ok {
		body.SetRevocationCheckingMethod(val.(string))
	}
	if val, ok := data["crl_distributor_config_path"]; ok && val != nil && val.(string) != "" {
		distr := secretgroup_tlscontext.NewSecretPath()
		distr.SetPath(val.(string))
		body.SetCrlDistributorConfig(*distr)
	}
	if val, ok := data["require_crl_for_all_ca"]; ok {
		body.SetRequireCrlForAllCa(val.(bool))
	}
	if val, ok := data["send_truststore"]; ok {
		body.SetSendTruststore(val.(bool))
	}
	if val, ok := data["authentication_overrides"]; ok {
		list := val.([]interface{})
		if len(list) > 0 {
			ao := newSgTlsContextSFAuthOverrides(list[0].(map[string]interface{}))
			body.SetAuthenticationOverrides(*ao)
		}
	}
	return body
}

func newSgTlsContextSFAuthOverrides(data map[string]interface{}) *secretgroup_tlscontext.AuthenticationOverrides {
	body := secretgroup_tlscontext.NewAuthenticationOverrides()
	if val, ok := data["certificate_bad_format"]; ok {
		body.SetCertificateBadFormat(val.(bool))
	}
	if val, ok := data["certificate_bad_signature"]; ok {
		body.SetCertificateBadSignature(val.(bool))
	}
	if val, ok := data["certificate_not_yet_valid"]; ok {
		body.SetCertificateNotYetValid(val.(bool))
	}
	if val, ok := data["certificate_has_expired"]; ok {
		body.SetCertificateHasExpired(val.(bool))
	}
	if val, ok := data["allow_self_signed"]; ok {
		body.SetAllowSelfSigned(val.(bool))
	}
	if val, ok := data["certificate_unresolved"]; ok {
		body.SetCertificateUnresolved(val.(bool))
	}
	if val, ok := data["certificate_untrusted"]; ok {
		body.SetCertificateUntrusted(val.(bool))
	}
	if val, ok := data["invalid_ca"]; ok {
		body.SetInvalidCa(val.(bool))
	}
	if val, ok := data["invalid_purpose"]; ok {
		body.SetInvalidPurpose(val.(bool))
	}
	if val, ok := data["other"]; ok {
		body.SetOther(val.(bool))
	}
	return body
}

func getSgTlsContextSFUpdatableAttributes() []string {
	attributes := [...]string{
		"name", "keystore_path", "truststore_path", "acceptable_tls_versions",
		"enable_mutual_authentication", "acceptable_cipher_suites", "mutual_authentication",
	}
	return attributes[:]
}

// returns the composed of the secret
func decomposeSgTlsContextSFId(d *schema.ResourceData) (string, string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2], s[3]
}
