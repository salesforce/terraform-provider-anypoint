package anypoint

import (
	"context"
	"io"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_crl_distributor_configs"
)

func resourceSecretGroupCrlDistribCfgs() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretGroupCrlDistribCfgsCreate,
		ReadContext:   resourceSecretGroupCrlDistribCfgsRead,
		UpdateContext: resourceSecretGroupCrlDistribCfgsUpdate,
		DeleteContext: resourceSecretGroupCrlDistribCfgsDelete,
		Description: `
		Create and manage crl-distributor-configs for a secret-group in a given organization and environment.
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
				Required:    true,
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
				Required:    true,
				Description: "URL from where complete CRL file is retrieved",
			},
			"frequency": {
				Type:             schema.TypeInt,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(2, 1000)),
				Description:      "How frequently should the distributor site be checked for new crl files(in minutes). Value should be between 2 and 1000",
			},
			"distributor_certificate_path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Refers to secret of type certificate",
			},
			"delta_crl_issuer_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL from where the changes in CRL file can be retrieved",
			},
			"ca_certificate_path": {
				Type:     schema.TypeString,
				Optional: true,
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

func resourceSecretGroupCrlDistribCfgsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	name := d.Get("name").(string)
	authctx := getSgCrlDistribCfgsAuthCtx(ctx, &pco)
	// body request
	body := newSgCrlDistribCfgsReqBody(d)
	//perform request
	res, httpr, err := pco.sgcrldistribcfgsclient.DefaultApi.PostSecretGroupCrlDistribCfgs(authctx, orgid, envid, sgid).CrlDistribCfgsReqBody(*body).Execute()
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
			Summary:  "Unable to create crl-distributor-configs " + name,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.SetId(res.GetId())
	return resourceSecretGroupCrlDistribCfgsRead(ctx, d, m)
}

func resourceSecretGroupCrlDistribCfgsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	id := d.Get("id").(string)
	authctx := getSgCrlDistribCfgsAuthCtx(ctx, &pco)
	if isComposedResourceId(id) {
		orgid, envid, sgid, id = decomposeSgCrlDistribCfgsId(d)
	}
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
	d.SetId(id)
	d.Set("sg_id", sgid)
	d.Set("env_id", envid)
	d.Set("org_id", orgid)

	return diags
}

func resourceSecretGroupCrlDistribCfgsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if d.HasChanges(getSgCrlDistribCfgsUpdatableAttributes()...) {
		pco := m.(ProviderConfOutput)
		orgid := d.Get("org_id").(string)
		envid := d.Get("env_id").(string)
		sgid := d.Get("sg_id").(string)
		id := d.Get("id").(string)
		authctx := getSgTlsContextAuthCtx(ctx, &pco)
		//prepare body
		body := newSgCrlDistribCfgsReqBody(d)
		// perform request
		_, httpr, err := pco.sgcrldistribcfgsclient.DefaultApi.PutSecretGroupTlsContext(authctx, orgid, envid, sgid, id).CrlDistribCfgsReqBody(*body).Execute()
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
				Summary:  "Unable to update crl-distributor-configs " + id,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		d.Set("last_updated", time.Now().Format(time.RFC850))
		return resourceSecretGroupCrlDistribCfgsRead(ctx, d, m)
	}
	return diags
}

func resourceSecretGroupCrlDistribCfgsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// NOTE: The delete action is not supported for this resource.
	// a keystore cannot be deleted, only secret-group (parent) can be deleted
	// Therefore we are only removing reference here
	d.SetId("")
	return diags
}

func newSgCrlDistribCfgsReqBody(d *schema.ResourceData) *secretgroup_crl_distributor_configs.CrlDistribCfgsReqBody {
	body := secretgroup_crl_distributor_configs.NewCrlDistribCfgsReqBody()
	if val, ok := d.GetOk("name"); ok {
		body.SetName(val.(string))
	}
	if val, ok := d.GetOk("complete_crl_issuer_url"); ok {
		body.SetCompleteCrlIssuerUrl(val.(string))
	}
	if val, ok := d.GetOk("frequency"); ok {
		body.SetFrequency(int32(val.(int)))
	}
	if val, ok := d.GetOk("distributor_certificate_path"); ok {
		sp := secretgroup_crl_distributor_configs.NewSecretPath()
		sp.SetPath(val.(string))
		body.SetDistributorCertificate(*sp)
	}
	if val, ok := d.GetOk("delta_crl_issuer_url"); ok {
		body.SetDeltaCrlIssuerUrl(val.(string))
	}
	if val, ok := d.GetOk("ca_certificate_path"); ok {
		sp := secretgroup_crl_distributor_configs.NewSecretPath()
		sp.SetPath(val.(string))
		body.SetCaCertificate(*sp)
	}
	return body
}

func getSgCrlDistribCfgsUpdatableAttributes() []string {
	attributes := [...]string{
		"name", "complete_crl_issuer_url", "frequency",
		"distributor_certificate_path", "delta_crl_issuer_url",
		"ca_certificate_path",
	}
	return attributes[:]
}

/*
 * Returns authentication context (includes authorization header)
 */
func getSgCrlDistribCfgsAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, secretgroup_crl_distributor_configs.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, secretgroup_crl_distributor_configs.ContextServerIndex, pco.server_index)
}

// returns the composed of the secret
func decomposeSgCrlDistribCfgsId(d *schema.ResourceData) (string, string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2], s[3]
}
