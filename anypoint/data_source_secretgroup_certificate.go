package anypoint

import (
	"context"
	"fmt"
	"io"
	"maps"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secretgroup_certificate "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_certificate"
)

func dataSourceSecretGroupCertificate() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretGroupCertificateRead,
		Description: `
		Query a specific certificate for a secret-group in a given organization and environment.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Id assigned to this certificate",
			},
			"sg_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The secret-group id where the certificate instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the certificate's secret group is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the certificate's secret group is defined.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the certificate",
			},
			"expiration_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The expiration date of the certificate",
			},
			"certificate_file_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The file name of the certificate",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The specific type of the certificate",
			},
			"path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The path of the keystore",
			},
			"details": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Details of the certificate",
				Elem: &schema.Resource{
					Schema: SG_CERTIFICATE_DETAILS_SCHEMA,
				},
			},
		},
	}
}

func dataSourceSecretGroupCertificateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	id := d.Get("id").(string)
	authctx := getSgCertificateAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.sgcertificateclient.DefaultApi.GetSecretGroupCertificateDetails(authctx, orgid, envid, sgid, id).Execute()
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
			Summary:  "Unable to get certificate " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process result
	data := flattenSgCertificate(res)
	if err := setSgCertificateAttributesToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set certificate " + id + " details attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(*res.GetMeta().Id)

	return diags
}

func flattenSgCertificate(cert *secretgroup_certificate.Certificate) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := cert.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := cert.GetExpirationDateOk(); ok {
		item["expiration_date"] = *val
	}
	if val, ok := cert.GetCertificateFileNameOk(); ok {
		item["certificate_file_name"] = *val
	}
	if val, ok := cert.GetTypeOk(); ok {
		item["type"] = *val
	}
	if val, ok := cert.GetMetaOk(); ok {
		maps.Copy(item, flattenSgCertificateMeta(val))
	}
	if val, ok := cert.GetDetailsOk(); ok {
		item["details"] = []interface{}{flattenSgCertificateDetails(val)}
	}

	return item
}

func flattenSgCertificateDetails(cert *secretgroup_certificate.CertificateDetails) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := cert.GetIssuerOk(); ok {
		item["issuer"] = flattenSgCertificateIssuerSubject(val)
	}
	if val, ok := cert.GetSubjectOk(); ok {
		item["subject"] = flattenSgCertificateIssuerSubject(val)
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
		item["validity"] = flattenSgCertificateCertificateValidity(val)
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

func flattenSgCertificateIssuerSubject(is *secretgroup_certificate.IssuerSubject) map[string]interface{} {
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

func flattenSgCertificateCertificateValidity(validity *secretgroup_certificate.CertificateValidity) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := validity.GetNotBeforeOk(); ok {
		item["not_before"] = *val
	}
	if val, ok := validity.GetNotAfterOk(); ok {
		item["not_after"] = *val
	}
	return item
}

func flattenSgCertificateMeta(meta *secretgroup_certificate.Meta) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := meta.GetIdOk(); ok {
		item["id"] = *val
	}
	if val, ok := meta.GetPathOk(); ok {
		item["path"] = *val
	}
	return item
}

func setSgCertificateAttributesToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getSgCertificateAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set certificate attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getSgCertificateAttributes() []string {
	attributes := [...]string{
		"name", "expiration_date", "certificate_file_name",
		"type", "path", "details",
	}
	return attributes[:]
}
