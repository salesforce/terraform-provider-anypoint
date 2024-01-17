package anypoint

import (
	"context"
	"fmt"
	"io"
	"maps"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secretgroup_crl_distributor_configs "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_crl_distributor_configs"
)

func dataSourceSecretGroupCrlDistribCfgs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretGroupCrlDistribCfgsRead,
		Description: `
		Query a specific crl-distributor-configs for a secret-group in a given organization and environment.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Id assigned to this crl-distributor-configs",
			},
			"sg_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The secret-group id where the crl-distributor-configs instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the crl-distributor-configs's secret group is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the crl-distributor-configs's secret group is defined.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the crl-distributor-configs",
			},
			"expiration_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The expiration date of the crl-distributor-configs",
			},
			"path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The path of the crl-distributor-configs",
			},
			"complete_crl_issuer_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL from where complete CRL file is retrieved",
			},
			"frequency": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "How frequently should the distributor site be checked for new crl files(in minutes)",
			},
			"distributor_certificate_path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Refers to secret of type certificate",
			},
			"delta_crl_issuer_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL from where the changes in CRL file can be retrieved",
			},
			"ca_certificate_path": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `
				Refers to a secret of type certificate. Select the CA certificate associated with the retrieved CRL file.
				If selected, the retrieved CRL file may contain revoked and/or held certificates issued by this CA.
				The CA subject name is obtained as part of the CRL file that is retrieved.
				However, the CRL distributor that issued and signed the CRL file may not be the issuing CA.
				If this CA certificate is encountered during chain-of-trust processing then a CRL file for this CA must have been successfully retrieved, validated and still in affect (not expired) or the chain-of trust processing fails depending on how the 'Require CRL for all CAs' flag setting configured as described below.
					* If the TLS Context secret has the 'Require CRL for all CAs' flag set to false, then the CA certificate should be selected. If not selected then prior to successful retrieval and processing of the CRL file there exists a window of time when a revoked CA certificate could be considered valid in chain-of-trust processing.
					* Else if its set to true, then its not necessary to select the CA certificate.
				`,
			},
		},
	}
}

func dataSourceSecretGroupCrlDistribCfgsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	id := d.Get("id").(string)
	authctx := getSgCrlDistribCfgsAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.sgcrldistribcfgsclient.DefaultApi.GetSecretGroupCrlDistribCfgsDetails(authctx, orgid, envid, sgid, id).Execute()
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
			Summary:  "Unable to get crl-distributor-configs " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process result
	data := flattenSgCrlDistribCfgsDetails(res)
	if err := setSgCrlDistribCfgsAttributesToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set crl-distributor-configs " + id + " details attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(*res.GetMeta().Id)
	return diags
}

func flattenSgCrlDistribCfgsDetails(cdc *secretgroup_crl_distributor_configs.CrlDistribCfgsDetails) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := cdc.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := cdc.GetExpirationDateOk(); ok {
		item["expiration_date"] = *val
	}
	if val, ok := cdc.GetMetaOk(); ok {
		maps.Copy(item, flattenSgCrlDistribCfgsMeta(val))
	}
	if val, ok := cdc.GetCompleteCrlIssuerUrlOk(); ok {
		item["complete_crl_issuer_url"] = *val
	}
	if val, ok := cdc.GetFrequencyOk(); ok {
		item["frequency"] = int(*val)
	}
	if val, ok := cdc.GetDistributorCertificateOk(); ok {
		item["distributor_certificate_path"] = val.GetPath()
	}
	if val, ok := cdc.GetDeltaCrlIssuerUrlOk(); ok {
		item["delta_crl_issuer_url"] = *val
	}
	if val, ok := cdc.GetCaCertificateOk(); ok {
		item["ca_certificate_path"] = val.GetPath()
	}
	return item
}

func setSgCrlDistribCfgsAttributesToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getSgCrlDistribCfgsAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set crl-distributor-configs attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getSgCrlDistribCfgsAttributes() []string {
	attributes := [...]string{
		"name", "expiration_date", "path", "complete_crl_issuer_url",
		"frequency", "distributor_certificate_path", "delta_crl_issuer_url",
		"ca_certificate_path",
	}
	return attributes[:]
}
