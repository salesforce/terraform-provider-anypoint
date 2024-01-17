package anypoint

import (
	"context"
	"fmt"
	"io"
	"maps"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secretgroup_truststore "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_truststore"
)

func dataSourceSecretGroupTruststore() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretGroupTruststoreRead,
		Description: `
		Query a specific truststore for a given secret-group, organization and environment.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Id assigned to this truststore",
			},
			"sg_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The secret-group id where the truststore instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the truststore instance is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the truststore instance is defined.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the truststore",
			},
			"expiration_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The expiration date of the truststore",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The specific type of the truststore",
			},
			"path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "path of this secret, relative to the containing secret group",
			},
			"truststore_file_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "File name of the truststore that is stored in this secret",
			},
			"algorithm": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Algorithm used to create the truststore manager factory which will make use of this truststore. Only present in the case of JKS, JCEKS and PKCS12 types",
			},
			"details": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Details about each of the trusted certificate from the truststore",
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
						"alias": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alias associated with the certificate entry",
						},
					},
				},
			},
		},
	}
}

func dataSourceSecretGroupTruststoreRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	id := d.Get("id").(string)
	authctx := getSgTruststoreAuthCtx(ctx, &pco)
	res, httpr, err := pco.sgtruststoreclient.DefaultApi.GetSecretGroupTruststoreDetails(authctx, orgid, envid, sgid, id).Execute()
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
			Summary:  "Unable to get truststore " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process result
	data := flattenSgTruststore(res)
	if err := setSgTruststoreAttributesToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set truststore " + id + " details attributes",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(*res.GetMeta().Id)

	return diags
}

func flattenSgTruststore(store *secretgroup_truststore.Truststore) map[string]interface{} {
	if isSgTruststorePEM(store.GetType()) {
		return flattenSgTruststorePEM(store)
	} else {
		return flattenSgTruststoreOthers(store)
	}
}

func flattenSgTruststorePEM(store *secretgroup_truststore.Truststore) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := store.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := store.GetExpirationDateOk(); ok {
		item["expiration_date"] = *val
	}
	if val, ok := store.GetTypeOk(); ok {
		item["type"] = *val
	}
	if meta, ok := store.GetMetaOk(); ok {
		maps.Copy(item, flattenSgTruststoreMeta(meta))
	}
	if val, ok := store.GetTruststoreFileNameOk(); ok {
		item["truststore_file_name"] = *val
	}
	if val, ok := store.GetTruststoreFileNameOk(); ok {
		item["truststore_file_name"] = *val
	}
	if val, ok := store.GetDetailsOk(); ok {
		item["details"] = flattenSgTruststoreDetails(val)
	}
	return item
}

func flattenSgTruststoreOthers(store *secretgroup_truststore.Truststore) map[string]interface{} {
	item := make(map[string]interface{})
	maps.Copy(item, flattenSgTruststorePEM(store))
	if val, ok := store.GetAlgorithmOk(); ok {
		item["algorithm"] = *val
	}
	return item
}

func flattenSgTruststoreDetails(details *secretgroup_truststore.TruststoreDetails) []map[string]interface{} {
	certs := make([]map[string]interface{}, 0)
	if entries, ok := details.GetCertificateEntriesOk(); ok {
		for _, entry := range entries {
			item := make(map[string]interface{})
			if val, ok := entry.GetAliasOk(); ok {
				item["alias"] = *val
			}
			if cert, ok := entry.GetCertificateOk(); ok {
				item["certificate"] = []interface{}{flattenSgTruststoreDetailsCertificate(cert)}
			}
		}
	}
	return certs
}

func flattenSgTruststoreDetailsCertificate(cert *secretgroup_truststore.CertificateDetails) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := cert.GetIssuerOk(); ok {
		item["issuer"] = flattenSgTruststoreIssuerSubject(val)
	}
	if val, ok := cert.GetSubjectOk(); ok {
		item["subject"] = flattenSgTruststoreIssuerSubject(val)
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
		item["validity"] = flattenSgTruststoreCertificateValidity(val)
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

func flattenSgTruststoreCertificateValidity(validity *secretgroup_truststore.CertificateValidity) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := validity.GetNotBeforeOk(); ok {
		item["not_before"] = *val
	}
	if val, ok := validity.GetNotAfterOk(); ok {
		item["not_after"] = *val
	}
	return item
}

func flattenSgTruststoreIssuerSubject(is *secretgroup_truststore.IssuerSubject) map[string]interface{} {
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

func isSgTruststorePEM(t string) bool {
	return t == "PEM"
}

func setSgTruststoreAttributesToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getSgTruststoreAttributes()
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

func getSgTruststoreAttributes() []string {
	attributes := [...]string{
		"name", "expiration_date", "truststore_file_name",
		"type", "path", "algorithm", "details",
	}
	return attributes[:]
}
