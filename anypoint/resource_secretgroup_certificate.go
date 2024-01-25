package anypoint

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	secretgroup_certificate "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_certificate"
)

func resourceSecretGroupCertificate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretGroupCertificateCreate,
		ReadContext:   resourceSecretGroupCertificateRead,
		UpdateContext: resourceSecretGroupCertificateUpdate,
		DeleteContext: resourceSecretGroupCertificateDelete,
		Description: `
		Create and manage a certificate for a secret-group in a given organization and environment.
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
				Description: "Id assigned to this certificate.",
			},
			"sg_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The secret-group id where the certificate instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organization id where the certificate's secret group is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The environment id where the certificate's secret group is defined.",
			},
			"allow_expired_cert": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "With 'true' to allow uploading expired certificates",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the certificate",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The specific type of the certificate",
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice(
						[]string{"PEM"},
						false,
					),
				),
			},
			"certificate": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The path to The file containing the certificate in PEM format",
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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSecretGroupCertificateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	name := d.Get("name").(string)
	allow_expired_cert := d.Get("allow_expired_cert").(bool)
	authctx := getSgCertificateAuthCtx(ctx, &pco)
	//prepare request
	req := pco.sgcertificateclient.DefaultApi.PostSecretGroupCertificate(authctx, orgid, envid, sgid).AllowExpiredCert(allow_expired_cert)
	req, err := loadSgCertificatePostBody(req, d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create certificate " + name,
			Detail:   err.Error(),
		})
		return diags
	}
	//Execute request
	res, httpr, err := req.Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create certificate " + name,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	id := res.GetId()
	d.SetId(id)
	return resourceSecretGroupCertificateRead(ctx, d, m)
}

func resourceSecretGroupCertificateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	id := d.Get("id").(string)
	if isComposedResourceId(id) {
		orgid, envid, sgid, id = decomposeSgCertificateId(d)
	}
	authctx := getSgCertificateAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.sgcertificateclient.DefaultApi.GetSecretGroupCertificateDetails(authctx, orgid, envid, sgid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
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

	d.SetId(id)
	d.Set("sg_id", sgid)
	d.Set("env_id", envid)
	d.Set("org_id", orgid)

	return diags
}

func resourceSecretGroupCertificateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if d.HasChanges(getSgCertificateUpdatableAttributes()...) {
		pco := m.(ProviderConfOutput)
		orgid := d.Get("org_id").(string)
		envid := d.Get("env_id").(string)
		sgid := d.Get("sg_id").(string)
		id := d.Get("id").(string)
		allow_expired_cert := d.Get("allow_expired_cert").(bool)
		authctx := getSgCertificateAuthCtx(ctx, &pco)
		req := pco.sgcertificateclient.DefaultApi.PutSecretGroupCertificate(authctx, orgid, envid, sgid, id).AllowExpiredCert(allow_expired_cert)
		req, err := loadSgCertificatePutBody(req, d)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update certificate " + id,
				Detail:   err.Error(),
			})
			return diags
		}
		//Execute request
		_, httpr, err := req.Execute()
		if err != nil {
			var details string
			if httpr != nil && httpr.StatusCode >= 400 {
				b, _ := io.ReadAll(httpr.Body)
				details = string(b)
			} else {
				details = err.Error()
			}
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update certificate " + id,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		d.Set("last_updated", time.Now().Format(time.RFC850))
		return resourceSecretGroupCertificateRead(ctx, d, m)
	}

	return diags
}

func resourceSecretGroupCertificateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// NOTE: The delete action is not supported for this resource.
	// a certificate cannot be deleted, only secret-group (parent) can be deleted
	// Therefore we are only removing reference here
	d.SetId("")
	return diags
}

func loadSgCertificatePostBody(req secretgroup_certificate.DefaultApiPostSecretGroupCertificateRequest, d *schema.ResourceData) (secretgroup_certificate.DefaultApiPostSecretGroupCertificateRequest, error) {
	// if val, ok := d.GetOk("expiration_date"); ok {
	// 	req = req.ExpirationDate(val.(string))
	// }
	if val, ok := d.GetOk("name"); ok {
		req = req.Name(val.(string))
	}
	if val, ok := d.GetOk("type"); ok {
		req = req.Type_(val.(string))
	}
	if val, ok := d.GetOk("certificate"); ok {
		file, err := os.Open(val.(string))
		if err != nil {
			return req, err
		}
		req = req.CertStore(file)
	}
	return req, nil
}

func loadSgCertificatePutBody(req secretgroup_certificate.DefaultApiPutSecretGroupCertificateRequest, d *schema.ResourceData) (secretgroup_certificate.DefaultApiPutSecretGroupCertificateRequest, error) {
	// if val, ok := d.GetOk("expiration_date"); ok {
	// 	req = req.ExpirationDate(val.(string))
	// }
	if val, ok := d.GetOk("name"); ok {
		req = req.Name(val.(string))
	}
	if val, ok := d.GetOk("type"); ok {
		req = req.Type_(val.(string))
	}
	if val, ok := d.GetOk("certificate"); ok {
		file, err := os.Open(val.(string))
		if err != nil {
			return req, err
		}
		req = req.CertStore(file)
	}
	return req, nil
}

// returns the composed of the secret
func decomposeSgCertificateId(d *schema.ResourceData) (string, string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2], s[3]
}

func getSgCertificateUpdatableAttributes() []string {
	attributes := [...]string{
		"allow_expired_cert", "name", "type", "certificate",
	}
	return attributes[:]
}

/*
 * Returns authentication context (includes authorization header)
 */
func getSgCertificateAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, secretgroup_certificate.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, secretgroup_certificate.ContextServerIndex, pco.server_index)
}
