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

func dataSourceSecretGroupTlsContextFG() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretGroupTlsContextFGRead,
		Description: `
		Query a specific tls-context of type "FlexGateway" for a secret-group in a given organization and environment.
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
			"cipher_suites": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of enabled cipher suites for Mule target.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"min_tls_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Minimum TLS version supported.",
			},
			"max_tls_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Maximum TLS version supported.",
			},
			"alpn_protocols": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "supported HTTP versions in the most-to-least preferred order. At least one version must be specified.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"inbound_settings": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Properties that are applicable only when the TLS context is used to secure inbound traffic.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_client_cert_validation": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether the client certificate validation must be enforced.",
						},
					},
				},
			},
			"outbound_settings": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Properties that are applicable only when the TLS context is used to secure outbound traffic.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"skip_server_cert_validation": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "flag that indicates whether the server certificate validation must be skipped.",
						},
					},
				},
			},
		},
	}
}

func dataSourceSecretGroupTlsContextFGRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	if !isSgTlsContextFlexGateway(res) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong target type for tls-context " + id,
			Detail:   "source is not of type FlexGateway",
		})
		return diags
	}
	data := flattenSgTlsContextFlexGateway(res)
	if err := setSgTlsContextFGAttributesToResourceData(d, data); err != nil {
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

func flattenSgTlsContextFlexGateway(fg *secretgroup_tlscontext.TlsContextDetails) map[string]interface{} {
	item := make(map[string]interface{})
	if meta, ok := fg.GetMetaOk(); ok {
		maps.Copy(item, flattenSgTlsContextMeta(meta))
	}
	if val, ok := fg.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := fg.GetExpirationDateOk(); ok {
		item["expiration_date"] = *val
	}
	if val, ok := fg.GetTargetOk(); ok {
		item["target"] = *val
	}
	if val, ok := fg.GetKeystoreOk(); ok {
		item["keystore_path"] = val.GetPath()
	}
	if val, ok := fg.GetTruststoreOk(); ok {
		item["truststore_path"] = val.GetPath()
	}
	if val, ok := fg.GetCipherSuitesOk(); ok {
		item["cipher_suites"] = val
	}
	if val, ok := fg.GetMinTlsVersionOk(); ok {
		item["min_tls_version"] = *val
	}
	if val, ok := fg.GetMaxTlsVersionOk(); ok {
		item["max_tls_version"] = *val
	}
	if val, ok := fg.GetAlpnProtocolsOk(); ok {
		item["alpn_protocols"] = val
	}
	if val, ok := fg.GetInboundSettingsOk(); ok {
		item["inbound_settings"] = []interface{}{flattenSgTlsContextFlexGatewayInboundSetting(val)}
	}
	if val, ok := fg.GetOutboundSettingsOk(); ok {
		item["outbound_settings"] = []interface{}{flattenSgTlsContextFlexGatewayOutboundSetting(val)}
	}
	return item
}

func flattenSgTlsContextFlexGatewayInboundSetting(inbound *secretgroup_tlscontext.TlsContextDetailsInboundSettings) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := inbound.GetEnableClientCertValidationOk(); ok {
		item["enable_client_cert_validation"] = *val
	}
	return item
}

func flattenSgTlsContextFlexGatewayOutboundSetting(inbound *secretgroup_tlscontext.TlsContextDetailsOutboundSettings) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := inbound.GetSkipServerCertValidationOk(); ok {
		item["skip_server_cert_validation"] = *val
	}
	return item
}

func setSgTlsContextFGAttributesToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getSgTlsContextFGAttributes()
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

func getSgTlsContextFGAttributes() []string {
	attributes := [...]string{
		"name", "expiration_date", "target", "path", "keystore_path",
		"truststore_path", "cipher_suites", "min_tls_version", "max_tls_version",
		"alpn_protocols", "inbound_settings", "outbound_settings",
	}
	return attributes[:]
}

// returns true if the target is of type FlexGateway
func isSgTlsContextFlexGateway(tls *secretgroup_tlscontext.TlsContextDetails) bool {
	return tls.GetTarget() == "FlexGateway"
}
