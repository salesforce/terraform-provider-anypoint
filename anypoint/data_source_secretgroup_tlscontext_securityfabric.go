package anypoint

import (
	"context"
	"fmt"
	"io"
	"maps"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secretgroup_tlscontext "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_tlscontext"
)

func dataSourceSecretGroupTlsContextSF() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretGroupTlsContextSFRead,
		Description: `
		Query a specific tls-context of type security-fabric for a secret-group in a given organization and environment.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
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
				Computed:    true,
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
				Computed:    true,
				Description: "Refers to a secret of type keystore. Relative path of the secret to be referenced.",
			},
			"truststore_path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Refers to a secret of type truststore. Relative path of the secret to be referenced.",
			},
			"acceptable_tls_versions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "TLS versions supported.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tls_v1_dot1": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "TLS version 1.1",
						},
						"tls_v1_dot2": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "TLS version 1.2",
						},
						"tls_v1_dot3": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "TLS version 1.3",
						},
					},
				},
			},
			"enable_mutual_authentication": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "This flag is to enable client authentication.",
			},
			"acceptable_cipher_suites": {
				Type:     schema.TypeList,
				Computed: true,
				Description: `
				List of accepted cipher suites by Security Fabric target, at least one should be set to true. If you are are not using the defaults and select individual ciphers, please select ciphers that match the configured keystore to ensure that TLS can setup a connection.
        For a keystore with an RSA key (the most common type), select ciphers which contain the string RSA (there are some exceptions). If using ECC ciphers, select ciphers which contain the string "ECDSA".
        TLS standards and documentation can be consulted for more background information.
				`,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aes128_gcm_sha256": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"aes128_sha256": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"aes256_gcm_sha384": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"aes256_sha256": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"dhe_rsa_aes128_gcm_sha256": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"dhe_rsa_aes128_sha256": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"dhe_rsa_aes256_gcm_sha384": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"dhe_rsa_aes256_sha256": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_ecdsa_aes128_gcm_sha256": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_ecdsa_aes128_sha1": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_ecdsa_aes256_gcm_sha384": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_ecdsa_aes256_sha1": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_rsa_aes128_gcm_sha256": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_rsa_aes128_sha1": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_rsa_aes256_gcm_sha384": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_rsa_aes256_sha1": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_ecdsa_chacha20_poly1305": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"ecdhe_rsa_chacha20_poly1305": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"dhe_rsa_chacha20_poly1305": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"tls_aes256_gcm_sha384": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"tls_chacha20_poly1305_sha256": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
						"tls_aes128_gcm_sha256": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allowed to be enabled only if tlsV1Dot2 is enabled",
						},
					},
				},
			},
			"mutual_authentication": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Configuration for client authentication.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_policies": {
							Type:     schema.TypeList,
							Computed: true,
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
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Allows application to control if strict or lax certificate checking will be performed during chain-of-trust processing",
						},
						"verification_depth": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "maximum allowed chain length for the certificates",
						},
						"perform_domain_checking": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether or not to perform domain checking",
						},
						"certificate_policy_checking": {
							Type:     schema.TypeBool,
							Computed: true,
							Description: `
							Controls certificate policy processing as defined in RFC 3280, 5280. A certificate can contain zero or more policies.
							A policy is represented as an object identifier (OID). In an end entity certificate, this policy information indicate the policy under which the certificate has been issued and the purposes for which the certificate may be used.
							In a CA certificate, this policy information limits the set of policies for certification paths that include this certificate. Applications with specific policy requirements are expected to have a list of those policies that they will accept and to compare the policy OIDs in the certificate to that list.
							If this extension is critical, the path validation software MUST be able to interpret this extension (including the optional qualifier), or MUST reject the certificate
							`,
						},
						"require_initial_explicit_policy": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates if the path must be valid for at least one of the certificate policies in the user-initial-policy-set.",
						},
						"revocation_checking": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates if certificate revocation checking should be enabled or not.",
						},
						"revocation_checking_method": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Protocol used for certificate revocation checking.",
						},
						"crl_distributor_config_path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Reference to a secret of type crlDistributorConfig.",
						},
						"require_crl_for_all_ca": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates if a valid CRL file must be in effect for every immediate and root Certificate Authority (CA) in the chain-of-trust",
						},
						"send_truststore": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Should the truststore i.e. trusted certificate authorities be sent to far-end during mutual authentication",
						},
						"certificate_pinning": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Referes to pinned certificates",
						},
						"authentication_overrides": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Override failing authentication when mutual authentication is being performed",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"certificate_bad_format": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Allow processing of certificates with bad format",
									},
									"certificate_bad_signature": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Allow processing of certificates with bad signature",
									},
									"certificate_not_yet_valid": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Allow processing of certificates that are not yet valid",
									},
									"certificate_has_expired": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Allow processing of certificates that are expired",
									},
									"allow_self_signed": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Allow self signed certificates",
									},
									"certificate_unresolved": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Allow unresolved certificates",
									},
									"certificate_untrusted": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Allow untrusted certificates",
									},
									"invalid_ca": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Allow invalid certificate authority certificates",
									},
									"invalid_purpose": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Allow certificates with invalid purpose",
									},
									"other": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Override any miscellaneous error condition encountered",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceSecretGroupTlsContextSFRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	id := d.Get("id").(string)
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
			Summary:  "Unable to get tls-context " + id,
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
	d.SetId(*res.GetMeta().Id)
	return diags
}

func flattenSgTlsContextSF(sf *secretgroup_tlscontext.TlsContextDetails) map[string]interface{} {
	item := make(map[string]interface{})
	if meta, ok := sf.GetMetaOk(); ok {
		maps.Copy(item, flattenSgTlsContextMeta(meta))
	}
	if val, ok := sf.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := sf.GetExpirationDateOk(); ok {
		item["expiration_date"] = *val
	}
	if val, ok := sf.GetTargetOk(); ok {
		item["target"] = *val
	}
	if val, ok := sf.GetKeystoreOk(); ok {
		item["keystore_path"] = val.GetPath()
	}
	if val, ok := sf.GetTruststoreOk(); ok {
		item["truststore_path"] = val.GetPath()
	}
	if val, ok := sf.GetAcceptableTlsVersionsOk(); ok {
		item["acceptable_tls_versions"] = []interface{}{flattenSgTlsContextSFAcceptableTlsVersions(val)}
	}
	if val, ok := sf.GetEnableMutualAuthenticationOk(); ok {
		item["enable_mutual_authentication"] = *val
	}
	if val, ok := sf.GetAcceptableCipherSuitesOk(); ok {
		item["acceptable_cipher_suites"] = []interface{}{flattenSgTlsContextSFAcceptableCipherSuites(val)}
	}
	if val, ok := sf.GetMutualAuthenticationOk(); ok {
		item["mutual_authentication"] = []interface{}{flattenSgTlsContextSFMutualAuthentication(val)}
	}
	return item
}

func flattenSgTlsContextSFAcceptableTlsVersions(atv *secretgroup_tlscontext.AcceptableTlsVersions) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := atv.GetTlsV1Dot1Ok(); ok {
		item["tls_v1_dot1"] = *val
	}
	if val, ok := atv.GetTlsV1Dot2Ok(); ok {
		item["tls_v1_dot2"] = *val
	}
	if val, ok := atv.GetTlsV1Dot3Ok(); ok {
		item["tls_v1_dot3"] = *val
	}
	return item
}

func flattenSgTlsContextSFAcceptableCipherSuites(acs *secretgroup_tlscontext.AcceptableCipherSuites) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := acs.GetAes128GcmSha256Ok(); ok {
		item["aes128_gcm_sha256"] = *val
	}
	if val, ok := acs.GetAes128Sha256Ok(); ok {
		item["aes128_sha256"] = *val
	}
	if val, ok := acs.GetAes256GcmSha384Ok(); ok {
		item["aes256_gcm_sha384"] = *val
	}
	if val, ok := acs.GetAes256Sha256Ok(); ok {
		item["aes256_sha256"] = *val
	}
	if val, ok := acs.GetDheRsaAes128GcmSha256Ok(); ok {
		item["dhe_rsa_aes128_gcm_sha256"] = *val
	}
	if val, ok := acs.GetDheRsaAes128Sha256Ok(); ok {
		item["dhe_rsa_aes128_sha256"] = *val
	}
	if val, ok := acs.GetDheRsaAes256GcmSha384Ok(); ok {
		item["dhe_rsa_aes256_gcm_sha384"] = *val
	}
	if val, ok := acs.GetDheRsaAes256Sha256Ok(); ok {
		item["dhe_rsa_aes256_sha256"] = *val
	}
	if val, ok := acs.GetEcdheEcdsaAes128GcmSha256Ok(); ok {
		item["ecdhe_ecdsa_aes128_gcm_sha256"] = *val
	}
	if val, ok := acs.GetEcdheEcdsaAes128Sha1Ok(); ok {
		item["ecdhe_ecdsa_aes128_sha1"] = *val
	}
	if val, ok := acs.GetEcdheEcdsaAes256GcmSha384Ok(); ok {
		item["ecdhe_ecdsa_aes256_gcm_sha384"] = *val
	}
	if val, ok := acs.GetEcdheEcdsaAes256Sha1Ok(); ok {
		item["ecdhe_ecdsa_aes256_sha1"] = *val
	}
	if val, ok := acs.GetEcdheRsaAes128GcmSha256Ok(); ok {
		item["ecdhe_rsa_aes128_gcm_sha256"] = *val
	}
	if val, ok := acs.GetEcdheRsaAes128Sha1Ok(); ok {
		item["ecdhe_rsa_aes128_sha1"] = *val
	}
	if val, ok := acs.GetEcdheRsaAes256GcmSha384Ok(); ok {
		item["ecdhe_rsa_aes256_gcm_sha384"] = *val
	}
	if val, ok := acs.GetEcdheRsaAes256Sha1Ok(); ok {
		item["ecdhe_rsa_aes256_sha1"] = *val
	}
	if val, ok := acs.GetEcdheEcdsaChacha20Poly1305Ok(); ok {
		item["ecdhe_ecdsa_chacha20_poly1305"] = *val
	}
	if val, ok := acs.GetEcdheRsaChacha20Poly1305Ok(); ok {
		item["ecdhe_rsa_chacha20_poly1305"] = *val
	}
	if val, ok := acs.GetDheRsaChacha20Poly1305Ok(); ok {
		item["dhe_rsa_chacha20_poly1305"] = *val
	}
	if val, ok := acs.GetTlsAes256GcmSha384Ok(); ok {
		item["tls_aes256_gcm_sha384"] = *val
	}
	if val, ok := acs.GetTlsChacha20Poly1305Sha256Ok(); ok {
		item["tls_chacha20_poly1305_sha256"] = *val
	}
	if val, ok := acs.GetTlsAes128GcmSha256Ok(); ok {
		item["tls_aes128_gcm_sha256"] = *val
	}
	return item
}

func flattenSgTlsContextSFMutualAuthentication(ma *secretgroup_tlscontext.MutualAuthentication) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := ma.GetCertificatePoliciesOk(); ok {
		item["certificate_policies"] = val
	}
	if val, ok := ma.GetCertCheckingStrengthOk(); ok {
		item["cert_checking_strength"] = *val
	}
	if val, ok := ma.GetVerificationDepthOk(); ok {
		item["verification_depth"] = *val
	}
	if val, ok := ma.GetPerformDomainCheckingOk(); ok {
		item["perform_domain_checking"] = *val
	}
	if val, ok := ma.GetCertificatePolicyCheckingOk(); ok {
		item["certificate_policy_checking"] = *val
	}
	if val, ok := ma.GetRequireInitialExplicitPolicyOk(); ok {
		item["require_initial_explicit_policy"] = *val
	}
	if val, ok := ma.GetRevocationCheckingOk(); ok {
		item["revocation_checking"] = *val
	}
	if val, ok := ma.GetRevocationCheckingMethodOk(); ok {
		item["revocation_checking_method"] = *val
	}
	if val, ok := ma.GetCrlDistributorConfigOk(); ok {
		item["crl_distributor_config_path"] = val.GetPath()
	}
	if val, ok := ma.GetRequireCrlForAllCaOk(); ok {
		item["require_crl_for_all_ca"] = *val
	}
	if val, ok := ma.GetSendTruststoreOk(); ok {
		item["send_truststore"] = *val
	}
	if val, ok := ma.GetCertificatePinningOk(); ok {
		item["certificate_pinning"] = flattenSgTlsContextSFCertPinning(val)
	}
	if val, ok := ma.GetAuthenticationOverridesOk(); ok {
		item["authentication_overrides"] = []interface{}{flattenSgTlsContextSFAuthOverrides(val)}
	}

	return item
}

func flattenSgTlsContextSFCertPinning(certp *secretgroup_tlscontext.CertificatePinning) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := certp.GetCertificatePinsetOk(); ok {
		item["certificate_pinset"] = val.GetPath()
	}
	if val, ok := certp.GetPerformCertificatePinningOk(); ok {
		item["perform_certificate_pinning"] = *val
	}
	return item
}

func flattenSgTlsContextSFAuthOverrides(authoverrides *secretgroup_tlscontext.AuthenticationOverrides) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := authoverrides.GetCertificateBadFormatOk(); ok {
		item["certificate_bad_format"] = *val
	}
	if val, ok := authoverrides.GetCertificateBadSignatureOk(); ok {
		item["certificate_bad_signature"] = *val
	}
	if val, ok := authoverrides.GetCertificateNotYetValidOk(); ok {
		item["certificate_not_yet_valid"] = *val
	}
	if val, ok := authoverrides.GetCertificateHasExpiredOk(); ok {
		item["certificate_has_expired"] = *val
	}
	if val, ok := authoverrides.GetAllowSelfSignedOk(); ok {
		item["allow_self_signed"] = *val
	}
	if val, ok := authoverrides.GetCertificateUnresolvedOk(); ok {
		item["certificate_unresolved"] = *val
	}
	if val, ok := authoverrides.GetCertificateUntrustedOk(); ok {
		item["certificate_untrusted"] = *val
	}
	if val, ok := authoverrides.GetInvalidCaOk(); ok {
		item["invalid_ca"] = *val
	}
	if val, ok := authoverrides.GetInvalidPurposeOk(); ok {
		item["invalid_purpose"] = *val
	}
	if val, ok := authoverrides.GetOtherOk(); ok {
		item["other"] = *val
	}
	return item
}

func setSgTlsContextSFAttributesToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getSgTlsContextSFAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set tls-context attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getSgTlsContextSFAttributes() []string {
	attributes := [...]string{
		"name", "expiration_date", "target", "path", "keystore_path",
		"truststore_path", "acceptable_tls_versions", "enable_mutual_authentication",
		"acceptable_cipher_suites", "mutual_authentication", "cipher_suites",
		"insecure", "min_tls_version", "max_tls_version", "alpn_protocols", "inbound_settings",
		"outboundSettings",
	}
	return attributes[:]
}

// returns true if target is of type SecurityFabric
func isSgTlsContextSecurityFabric(tls *secretgroup_tlscontext.TlsContextDetails) bool {
	return tls.GetTarget() == "SecurityFabric"
}
