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

func dataSourceSecretGroupTlsContextMule() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretGroupTlsContextMuleRead,
		Description: `
		Query a specific tls-context of type "Mule" for a secret-group in a given organization and environment.
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
			"cipher_suites": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of enabled cipher suites for Mule target.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"insecure": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Setting this flag to true indicates that certificate validation should not be enforced, i.e. the truststore, even though set, is ignored at runtime. Only available for \"Mule\" target",
			},
		},
	}
}

func dataSourceSecretGroupTlsContextMuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		if httpr != nil {
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
	d.SetId(*res.GetMeta().Id)

	return diags
}

func flattenSgTlsContextMule(mule *secretgroup_tlscontext.TlsContextDetails) map[string]interface{} {
	item := make(map[string]interface{})
	if meta, ok := mule.GetMetaOk(); ok {
		maps.Copy(item, flattenSgTlsContextMeta(meta))
	}
	if val, ok := mule.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := mule.GetExpirationDateOk(); ok {
		item["expiration_date"] = *val
	}
	if val, ok := mule.GetTargetOk(); ok {
		item["target"] = *val
	}
	if val, ok := mule.GetKeystoreOk(); ok {
		item["keystore_path"] = val.GetPath()
	}
	if val, ok := mule.GetTruststoreOk(); ok {
		item["truststore_path"] = val.GetPath()
	}
	if val, ok := mule.GetAcceptableTlsVersionsOk(); ok {
		item["acceptable_tls_versions"] = []interface{}{flattenSgTlsContextMuleAcceptableTlsVersions(val)}
	}
	if val, ok := mule.GetCipherSuitesOk(); ok {
		item["cipher_suites"] = val
	}
	if val, ok := mule.GetInsecureOk(); ok {
		item["insecure"] = *val
	}
	return item
}

func flattenSgTlsContextMuleAcceptableTlsVersions(atv *secretgroup_tlscontext.AcceptableTlsVersions) map[string]interface{} {
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

func setSgTlsContextMuleAttributesToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getSgTlsContextMuleAttributes()
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

func getSgTlsContextMuleAttributes() []string {
	attributes := [...]string{
		"name", "expiration_date", "target", "path", "keystore_path",
		"truststore_path", "acceptable_tls_versions", "cipher_suites", "insecure",
	}
	return attributes[:]
}

// returns true if target is of type Mule
func isSgTlsContextMule(tls *secretgroup_tlscontext.TlsContextDetails) bool {
	return tls.GetTarget() == "Mule"
}
