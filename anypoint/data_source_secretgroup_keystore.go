package anypoint

import (
	"context"
	"fmt"
	"io"
	"maps"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_keystore"
)

var SG_CERTIFICATE_DETAILS_SCHEMA = map[string]*schema.Schema{
	"issuer": {
		Type:        schema.TypeMap,
		Computed:    true,
		Description: "Details about the entity that issued the certificate.",
	},
	"subject": {
		Type:        schema.TypeMap,
		Computed:    true,
		Description: "Details about the entity to which the certificate is issued.",
	},
	"subject_alternative_name": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Collection of subject alternative names from the SubjectAltName x509 extension",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"version": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "version",
	},
	"serial_number": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Serial number assigned by the CA to this certificate, in hex format",
	},
	"signature_algorithm": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Name of the signature algorithm",
	},
	"public_key_algorithm": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The standard algorithm name for the public key of this certificate",
	},
	"is_certificate_authority": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "If set to true, indicates that this is a CA certificate.",
	},
	"validity": {
		Type:        schema.TypeMap,
		Computed:    true,
		Description: "Details about validity period of this certificate",
	},
	"key_usage": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "A list of values defining the purpose of the public key i.e. the key usage extensions from this certificate",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"extended_key_usage": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "A list of values providing details about the extended key usage extensions from this certificate.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"certificate_type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The type of this certificate",
	},
}

func dataSourceSecretGroupKeystore() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretGroupKeystoreRead,
		Description: `
		Query a specific keystore for a secret-group in a given organization and environment.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Id assigned to this keystore",
			},
			"sg_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The secret-group id where the keystore instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the keystore's secret group is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the keystore's secret group is defined.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the keystore",
			},
			"expiration_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The expiration date of the keystore",
			},
			"certificate_file_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The file name of the certificate file that is stored in this keystore",
			},
			"key_file_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The file name of the encrypted private key that is stored in this keystore",
			},
			"capath_file_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The file name of the CA file that is stored in this keystore",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The specific type of the keystore",
			},
			"path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The path of the keystore",
			},
			"keystore_file_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "File name of the keystore that is stored in this secret",
			},
			"algorithm": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Algorithm used to create the keystore manager factory which will make use of this keystore",
			},
			"alias": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The alias name of the entry that contains the certificate",
			},
			"details": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Details about the public certificate and capath from the keystore",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Details of certificate.",
							Elem: &schema.Resource{
								Schema: SG_CERTIFICATE_DETAILS_SCHEMA,
							},
						},
						"capath": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Details of Certificate Authority",
							Elem: &schema.Resource{
								Schema: SG_CERTIFICATE_DETAILS_SCHEMA,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceSecretGroupKeystoreRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	id := d.Get("id").(string)
	authctx := getSgKeystoreAuthCtx(ctx, &pco)
	res, httpr, err := pco.sgkeystoreclient.DefaultApi.GetSecretGroupKeystoreDetails(authctx, orgid, envid, sgid, id).Execute()
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
			Summary:  "Unable to get keystore " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process result
	data := flattenSgKeystore(res)
	if err := setSgKeystoreAttributesToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set keystore " + id + " details attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(*res.GetMeta().Id)

	return diags
}

func flattenSgKeystore(keystore *secretgroup_keystore.Keystore) map[string]interface{} {
	if isSgKeystorePEM(keystore.GetType()) {
		return flattenSgKeystorePEM(keystore)
	} else {
		return flattenSgKeystoreOthers(keystore)
	}
}

func flattenSgKeystorePEM(keystore *secretgroup_keystore.Keystore) map[string]interface{} {
	item := make(map[string]interface{})

	if val, ok := keystore.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := keystore.GetExpirationDateOk(); ok {
		item["expiration_date"] = *val
	}
	if val, ok := keystore.GetMetaOk(); ok {
		maps.Copy(item, flattenSgKeystoreMeta(val))
	}
	if val, ok := keystore.GetTypeOk(); ok {
		item["type"] = *val
	}
	if val, ok := keystore.GetDetailsOk(); ok {
		item["details"] = []interface{}{flattenSgKeystoreDetails(val)}
	}
	if val, ok := keystore.GetCertificateFileNameOk(); ok {
		item["certificate_file_name"] = *val
	}
	if val, ok := keystore.GetKeyFileNameOk(); ok {
		item["key_file_name"] = *val
	}
	if val, ok := keystore.GetCapathFileNameOk(); ok {
		item["capath_file_name"] = *val
	}

	return item
}

func flattenSgKeystoreOthers(keystore *secretgroup_keystore.Keystore) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := keystore.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := keystore.GetExpirationDateOk(); ok {
		item["expiration_date"] = *val
	}
	if val, ok := keystore.GetMetaOk(); ok {
		maps.Copy(item, flattenSgKeystoreMeta(val))
	}
	if val, ok := keystore.GetTypeOk(); ok {
		item["type"] = *val
	}
	if val, ok := keystore.GetDetailsOk(); ok {
		item["details"] = []interface{}{flattenSgKeystoreDetails(val)}
	}
	if val, ok := keystore.GetKeystoreFileNameOk(); ok {
		item["keystore_file_name"] = *val
	}
	if val, ok := keystore.GetAlgorithmOk(); ok {
		item["algorithm"] = *val
	}
	if val, ok := keystore.GetAliasOk(); ok {
		item["alias"] = *val
	}
	return item
}

func flattenSgKeystoreDetails(details *secretgroup_keystore.KeystoreDetails) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := details.GetCertificateOk(); ok {
		item["certificate"] = []interface{}{flattenSgKeystoreDetailsCertificate(val)}
	}
	if val, ok := details.GetCapathOk(); ok {
		item["capath"] = flattenSgKeystoreDetailsCaPath(val)
	}
	return item
}

func flattenSgKeystoreDetailsCaPath(capath []secretgroup_keystore.CertificateDetails) []map[string]interface{} {
	length := len(capath)
	result := make([]map[string]interface{}, length)
	for i, item := range capath {
		result[i] = flattenSgKeystoreDetailsCertificate(&item)
	}
	return result
}

func flattenSgKeystoreDetailsCertificate(cert *secretgroup_keystore.CertificateDetails) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := cert.GetIssuerOk(); ok {
		item["issuer"] = flattenSgKeystoreIssuerSubject(val)
	}
	if val, ok := cert.GetSubjectOk(); ok {
		item["subject"] = flattenSgKeystoreIssuerSubject(val)
	}
	if val, ok := cert.GetSubjectAlternativeNameOk(); ok {
		item["subject_alternative_name"] = val
	}
	if val, ok := cert.GetVersionOk(); ok {
		item["version"] = *val
	}
	if val, ok := cert.GetSerialNumberOk(); ok {
		item["serial_number"] = *val
	}
	if val, ok := cert.GetSignatureAlgorithmOk(); ok {
		item["signature_algorithm"] = *val
	}
	if val, ok := cert.GetPublicKeyAlgorithmOk(); ok {
		item["public_key_algorithm"] = *val
	}
	if bc, ok := cert.GetBasicConstraintsOk(); ok {
		if val, ok := bc.GetCertificateAuthorityOk(); ok {
			item["is_certificate_authority"] = *val
		}
	}
	if val, ok := cert.GetValidityOk(); ok {
		item["validity"] = flattenSgKeystoreCertificateValidity(val)
	}
	if val, ok := cert.GetKeyUsageOk(); ok {
		item["key_usage"] = val
	}
	if val, ok := cert.GetExtendedKeyUsageOk(); ok {
		item["extended_key_usage"] = val
	}
	if val, ok := cert.GetCertificateTypeOk(); ok {
		item["certificate_type"] = *val
	}
	return item
}

func flattenSgKeystoreCertificateValidity(validity *secretgroup_keystore.CertificateValidity) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := validity.GetNotBeforeOk(); ok {
		item["not_before"] = *val
	}
	if val, ok := validity.GetNotAfterOk(); ok {
		item["not_after"] = *val
	}
	return item
}

func flattenSgKeystoreIssuerSubject(is *secretgroup_keystore.IssuerSubject) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := is.GetCommonNameOk(); ok {
		item["common_name"] = *val
	}
	if val, ok := is.GetOrganizationNameOk(); ok {
		item["organization_name"] = *val
	}
	if val, ok := is.GetOrganizationUnitOk(); ok {
		item["organization_unit"] = *val
	}
	if val, ok := is.GetLocalityNameOk(); ok {
		item["locality_name"] = *val
	}
	if val, ok := is.GetCountryNameOk(); ok {
		item["country_name"] = *val
	}
	if val, ok := is.GetStateOk(); ok {
		item["state"] = *val
	}
	return item
}

func flattenSgKeystoreMeta(meta *secretgroup_keystore.Meta) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := meta.GetIdOk(); ok {
		item["id"] = *val
	}
	if val, ok := meta.GetPathOk(); ok {
		item["path"] = *val
	}
	return item
}

func setSgKeystoreAttributesToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getSgKeystoreAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set keystore attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getSgKeystoreAttributes() []string {
	attributes := [...]string{
		"name", "expiration_date", "certificate_file_name",
		"key_file_name", "capath_file_name", "type", "path",
		"keystore_file_name", "algorithm", "alias", "details",
	}
	return attributes[:]
}

func isSgKeystorePEM(t string) bool {
	return t == "PEM"
}
