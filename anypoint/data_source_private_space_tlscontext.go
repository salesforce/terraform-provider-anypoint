package anypoint

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/private_space_tlscontext"
)

var PRIVATE_SPACE_TLSCONTEXT_SCHEMA = map[string]*schema.Schema{
	"org_id": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"private_space_id": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"id": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"name": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"trust_store_file_name": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"trust_store_expiration_date": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"trust_store_type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The type of the trust store.",
	},
	"trust_store_dn_list": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "The list of DNs of the trust store.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"issuer_common_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The issuer common name of the trust store.",
				},
				"issuer_organization_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The issuer organization name of the trust store.",
				},
				"issuer_organization_unit": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The issuer organization unit of the trust store.",
				},
				"issuer_country_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The issuer country name of the trust store.",
				},
				"issuer_state": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The issuer state of the trust store.",
				},
				"issuer_locality_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The issuer locality name of the trust store.",
				},
				"subject_common_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The subject common name of the trust store.",
				},
				"subject_organization_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The subject organization name of the trust store.",
				},
				"subject_organization_unit": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The subject organization unit of the trust store.",
				},
				"subject_country_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The subject country name of the trust store.",
				},
				"subject_state": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The subject state of the trust store.",
				},
				"subject_locality_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The subject locality name of the trust store.",
				},
				"version": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The version of the trust store.",
				},
				"serial_number": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The serial number of the trust store.",
				},
				"signature_algorithm": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The signature algorithm of the trust store.",
				},
				"public_key_algorithm": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The public key algorithm of the trust store.",
				},
				"basic_constraints_is_ca": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "The basic constraints is ca of the trust store.",
				},
				"validity_not_before": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The validity start date of the trust store.",
				},
				"validity_not_after": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The validity end date of the trust store.",
				},
				"key_usage": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "The key usage of the trust store.",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"certificate_type": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The certificate type of the trust store.",
				},
			},
		},
	},
	"key_store_type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The type of the keystore.",
	},
	"key_store_cn": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The common name of the keystore.",
	},
	"key_store_san": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "The subject alternative names of the keystore.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"key_store_file_name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The keystore certificate file name.",
	},
	"key_store_capath_file_name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The keystore capath file name.",
	},
	"key_store_key_file_name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The keystore private key file name.",
	},
	"key_store_expiration_date": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The keystore expiration date.",
	},
	"cipher_aes128_gcm_sha256": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_aes128_sha256": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_aes256_gcm_sha384": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_aes256_sha256": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_dhe_rsa_aes128_sha256": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_dhe_rsa_aes256_gcm_sha384": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_dhe_rsa_aes256_sha256": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_ecdhe_ecdsa_aes128_gcm_sha256": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_ecdhe_ecdsa_aes256_gcm_sha384": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_ecdhe_rsa_aes128_gcm_sha256": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_ecdhe_rsa_aes256_gcm_sha384": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_ecdhe_ecdsa_chacha20_poly1305": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_ecdhe_rsa_chacha20_poly1305": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_dhe_rsa_chacha20_poly1305": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_tls_aes256_gcm_sha384": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_tls_chacha20_poly1305_sha256": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"cipher_tls_aes128_gcm_sha256": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "The ciphers used by the tls context.",
	},
	"type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The type of the tls context.",
	},
}

func dataSourcePrivateSpaceTlsContext() *schema.Resource {
	pstc_schema := cloneSchema(PRIVATE_SPACE_TLSCONTEXT_SCHEMA)
	pstc_schema["org_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The master organization id where the private space is defined.",
	}
	pstc_schema["private_space_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The private space id.",
	}
	pstc_schema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The private space tls context id.",
	}
	return &schema.Resource{
		ReadContext: dataSourcePrivateSpaceTlsContextRead,
		Description: `
		Reads a ` + "`" + `private space tls context` + "`" + ` in your business group.
		`,
		Schema: pstc_schema,
	}
}

func dataSourcePrivateSpaceTlsContextRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	private_space_id := d.Get("private_space_id").(string)
	id := d.Get("id").(string)
	authctx := getPrivateSpaceTlsContextAuthCtx(ctx, &pco)
	//request
	res, httpr, err := pco.privatespacetlscontextclient.DefaultApi.GetTlsContext(authctx, orgid, private_space_id, id).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Get Private Space TLS Context for org " + orgid + " and id " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	private_space_tls_context := flattenPrivateSpaceTlsContextData(res)
	//save in data source schema
	if err := setPrivateSpaceTlsContextAttributesToResourceData(d, private_space_tls_context); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set Private Space TLS Context " + id + " in org " + orgid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(id)
	return diags
}

func flattenPrivateSpaceTlsContextData(data *private_space_tlscontext.TlsContext) map[string]any {
	result := make(map[string]any)
	if data == nil {
		return result
	}
	result["id"] = data.GetId()
	result["name"] = data.GetName()

	if truststore, ok := data.GetTrustStoreOk(); ok {
		result["trust_store_file_name"] = truststore.GetFileName()
		result["trust_store_expiration_date"] = truststore.GetExpirationDate()
		result["trust_store_type"] = truststore.GetType()
		if dnList, ok := truststore.GetDnListOk(); ok {
			result["trust_store_dn_list"] = flattenPrivateSpaceTlsContextDnList(dnList)
		} else {
			result["trust_store_dn_list"] = []any{}
		}
	} else {
		result["trust_store_dn_list"] = []any{}
		result["trust_store_file_name"] = ""
		result["trust_store_expiration_date"] = ""
		result["trust_store_type"] = ""
	}
	if keystore, ok := data.GetKeyStoreOk(); ok {
		result["key_store_type"] = keystore.GetType()
		result["key_store_cn"] = keystore.GetCn()
		result["key_store_san"] = keystore.GetSan()
		result["key_store_file_name"] = keystore.GetFileName()
		result["key_store_capath_file_name"] = keystore.GetCapathFileName()
		result["key_store_key_file_name"] = keystore.GetKeyFileName()
		result["key_store_expiration_date"] = keystore.GetExpirationDate()
	}
	if ciphers, ok := data.GetCiphersOk(); ok {
		result["cipher_aes128_gcm_sha256"] = ciphers.GetAes128GcmSha256()
		result["cipher_aes128_sha256"] = ciphers.GetAes128Sha256()
		result["cipher_aes256_gcm_sha384"] = ciphers.GetAes256GcmSha384()
		result["cipher_aes256_sha256"] = ciphers.GetAes256Sha256()
		result["cipher_dhe_rsa_aes128_sha256"] = ciphers.GetDheRsaAes128Sha256()
		result["cipher_dhe_rsa_aes256_gcm_sha384"] = ciphers.GetDheRsaAes256GcmSha384()
		result["cipher_dhe_rsa_aes256_sha256"] = ciphers.GetDheRsaAes256Sha256()
		result["cipher_ecdhe_ecdsa_aes128_gcm_sha256"] = ciphers.GetEcdheEcdsaAes128GcmSha256()
		result["cipher_ecdhe_ecdsa_aes256_gcm_sha384"] = ciphers.GetEcdheEcdsaAes256GcmSha384()
		result["cipher_ecdhe_rsa_aes128_gcm_sha256"] = ciphers.GetEcdheRsaAes128GcmSha256()
		result["cipher_ecdhe_rsa_aes256_gcm_sha384"] = ciphers.GetEcdheRsaAes256GcmSha384()
		result["cipher_ecdhe_ecdsa_chacha20_poly1305"] = ciphers.GetEcdheEcdsaChacha20Poly1305()
		result["cipher_ecdhe_rsa_chacha20_poly1305"] = ciphers.GetEcdheRsaChacha20Poly1305()
		result["cipher_dhe_rsa_chacha20_poly1305"] = ciphers.GetDheRsaChacha20Poly1305()
		result["cipher_tls_aes256_gcm_sha384"] = ciphers.GetTlsAes256GcmSha384()
		result["cipher_tls_chacha20_poly1305_sha256"] = ciphers.GetTlsChacha20Poly1305Sha256()
		result["cipher_tls_aes128_gcm_sha256"] = ciphers.GetTlsAes128GcmSha256()
	}
	result["type"] = data.GetType()
	return result
}

func flattenPrivateSpaceTlsContextDnList(dnList []private_space_tlscontext.TrustStoreDnListInner) []map[string]any {
	result := make([]map[string]any, len(dnList))
	for i, dn := range dnList {
		m := make(map[string]any)
		if issuer, ok := dn.GetIssuerOk(); ok {
			m["issuer_common_name"] = issuer.GetCommonName()
			m["issuer_organization_unit"] = issuer.GetOrganizationUnit()
			m["issuer_organization_name"] = issuer.GetOrganizationName()
			m["issuer_locality_name"] = issuer.GetLocalityName()
			m["issuer_state"] = issuer.GetState()
			m["issuer_country_name"] = issuer.GetCountryName()
		}
		if subject, ok := dn.GetSubjectOk(); ok {
			m["subject_common_name"] = subject.GetCommonName()
			m["subject_organization_unit"] = subject.GetOrganizationUnit()
			m["subject_organization_name"] = subject.GetOrganizationName()
			m["subject_locality_name"] = subject.GetLocalityName()
			m["subject_state"] = subject.GetState()
			m["subject_country_name"] = subject.GetCountryName()
		}
		m["version"] = dn.GetVersion()
		m["serial_number"] = dn.GetSerialNumber()
		m["signature_algorithm"] = dn.GetSignatureAlgorithm()
		m["public_key_algorithm"] = dn.GetPublicKeyAlgorithm()
		if basicConstraints, ok := dn.GetBasicConstraintsOk(); ok {
			m["basic_constraints_is_ca"] = basicConstraints.GetCertificateAuthority()
		}
		if validity, ok := dn.GetValidityOk(); ok {
			m["validity_not_before"] = validity.GetNotBefore()
			m["validity_not_after"] = validity.GetNotAfter()
		}
		m["key_usage"] = dn.GetKeyUsage()
		m["certificate_type"] = dn.GetCertificateType()
		result[i] = m
	}
	return result
}

func setPrivateSpaceTlsContextAttributesToResourceData(d *schema.ResourceData, data map[string]any) error {
	attributes := getPrivateSpaceTlsContextAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set private space tls context attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getPrivateSpaceTlsContextAttributes() []string {
	return getSchemaKeys(PRIVATE_SPACE_TLSCONTEXT_SCHEMA)
}
