package anypoint

import (
	"context"
	"io"
	"maps"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secretgroup_certificate "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_certificate"
)

func dataSourceSecretGroupCertificates() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretGroupCertificatesRead,
		Description: `
		Query all or part of available certificates for a given secret-group, organization and environment.
		`,
		Schema: map[string]*schema.Schema{
			"sg_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The secret-group id where the keystore instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the keystore instance is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the keystore instance is defined.",
			},
			"certificates": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List certificates result of the query",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The specific type of the certificate",
						},
						"path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The path of the certificate",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The id of the certificate",
						},
					},
				},
			},
		},
	}
}

func dataSourceSecretGroupCertificatesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	authctx := getSgCertificateAuthCtx(ctx, &pco)
	// perform request
	res, httpr, err := pco.sgcertificateclient.DefaultApi.GetSecretGroupCertificates(authctx, orgid, envid, sgid).Execute()
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
			Summary:  "Unable to get certificates for secret-group " + sgid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// process response
	data := flattenSgCertificateSummaryCollection(res)
	if err := d.Set("certificates", data); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set certificates for secret-group " + sgid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenSgCertificateSummaryCollection(collection []secretgroup_certificate.CertificateSummary) []interface{} {
	length := len(collection)
	if length > 0 {
		res := make([]interface{}, length)
		for i, cert := range collection {
			res[i] = flattenSgCertificateSummary(&cert)
		}
		return res
	}
	return make([]interface{}, 0)
}

func flattenSgCertificateSummary(cert *secretgroup_certificate.CertificateSummary) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := cert.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := cert.GetExpirationDateOk(); ok {
		item["expiration_date"] = *val
	}
	if val, ok := cert.GetTypeOk(); ok {
		item["type"] = *val
	}
	if val, ok := cert.GetMetaOk(); ok {
		maps.Copy(item, flattenSgCertificateMeta(val))
	}
	return item
}
