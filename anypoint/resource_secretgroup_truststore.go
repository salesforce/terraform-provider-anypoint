package anypoint

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	secretgroup_truststore "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_truststore"
)

func resourceSecretGroupTruststore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretGroupTruststoreCreate,
		ReadContext:   resourceSecretGroupTruststoreRead,
		UpdateContext: resourceSecretGroupTruststoreUpdate,
		DeleteContext: resourceSecretGroupTruststoreDelete,
		Description: `
		Create and manage a truststore for a given secret-group, organization and environment.
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
			"allow_expired_cert": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "With 'true' to allow uploading expired certificates",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the truststore",
			},
			"expiration_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The expiration date of the truststore",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The specific type of the truststore",
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice(
						[]string{"PEM", "JKS", "JCEKS", "PKCS12"},
						false,
					),
				),
			},
			"truststore": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Path to the file containing one or more trusted certificate entries",
			},
			"store_passphrase": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The passphrase with which the trustStore file is protected. Required in case of JKS, JCEKS and PKCS12 types",
			},
			"algorithm": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Algorithm used to create the truststore manager factory which will make use of this truststore. Only present in the case of JKS, JCEKS and PKCS12 types",
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
		CustomizeDiff: func(ctx context.Context, rd *schema.ResourceDiff, i interface{}) error {
			return validateTruststoreInput(rd)
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSecretGroupTruststoreCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	name := d.Get("name").(string)
	allow_expired_cert := d.Get("allow_expired_cert").(bool)
	authctx := getSgTruststoreAuthCtx(ctx, &pco)
	//prepare request
	req := pco.sgtruststoreclient.DefaultApi.PostSecretGroupTruststore(authctx, orgid, envid, sgid).AllowExpiredCert(allow_expired_cert)
	req, err := loadSgTruststorePostBody(req, d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create truststore " + name,
			Detail:   err.Error(),
		})
		return diags
	}
	//Execute request
	res, httpr, err := req.Execute()
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
			Summary:  "Unable to create truststore " + name,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	id := res.GetId()
	d.SetId(id)

	return resourceSecretGroupTruststoreRead(ctx, d, m)
}

func resourceSecretGroupTruststoreRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	id := d.Get("id").(string)
	if isComposedResourceId(id) {
		orgid, envid, sgid, id = decomposeSgTruststoreId(d)
	}
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
			Summary:  "Unable to read truststore " + id,
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
	d.SetId(id)
	d.Set("sg_id", sgid)
	d.Set("env_id", envid)
	d.Set("org_id", orgid)

	return diags
}

func resourceSecretGroupTruststoreUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var attributes []string
	t := d.Get("type").(string)
	if isSgTruststorePEM(t) {
		attributes = getSgTruststorePEMUpdatableAttributes()
	} else {
		attributes = getSgTruststoreOthersUpdatableAttributes()
	}
	if d.HasChanges(attributes...) {
		pco := m.(ProviderConfOutput)
		allow_expired_cert := d.Get("allow_expired_cert").(bool)
		orgid := d.Get("org_id").(string)
		envid := d.Get("env_id").(string)
		sgid := d.Get("sg_id").(string)
		id := d.Get("id").(string)
		authctx := getSgTruststoreAuthCtx(ctx, &pco)
		req := pco.sgtruststoreclient.DefaultApi.PutSecretGroupTruststore(authctx, orgid, envid, sgid, id).AllowExpiredCert(allow_expired_cert)
		req, err := loadSgTruststorePutBody(req, d)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update truststore " + id,
				Detail:   err.Error(),
			})
			return diags
		}
		//Execute request
		_, httpr, err := req.Execute()
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
				Summary:  "Unable to update truststore " + id,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		d.Set("last_updated", time.Now().Format(time.RFC850))
		return resourceSecretGroupTruststoreRead(ctx, d, m)
	}

	return diags
}

func resourceSecretGroupTruststoreDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// NOTE: The delete action is not supported for this resource.
	// a truststore cannot be deleted, only secret-group (parent) can be deleted
	// Therefore we are only removing reference here
	d.SetId("")
	return diags
}

func loadSgTruststorePostBody(req secretgroup_truststore.DefaultApiPostSecretGroupTruststoreRequest, d *schema.ResourceData) (secretgroup_truststore.DefaultApiPostSecretGroupTruststoreRequest, error) {
	t := d.Get("type").(string)
	if isSgTruststorePEM(t) {
		return loadSgTruststorePEMPostBody(req, d)
	} else {
		return loadSgTruststoreOthersPostBody(req, d)
	}
}

func loadSgTruststorePEMPostBody(req secretgroup_truststore.DefaultApiPostSecretGroupTruststoreRequest, d *schema.ResourceData) (secretgroup_truststore.DefaultApiPostSecretGroupTruststoreRequest, error) {
	// if val, ok := d.GetOk("expiration_date"); ok {
	// 	req = req.ExpirationDate(val.(string))
	// }
	if val, ok := d.GetOk("truststore"); ok {
		file, err := os.Open(val.(string))
		if err != nil {
			return req, err
		}
		req = req.TrustStore(file)
	}
	if val, ok := d.GetOk("name"); ok {
		req = req.Name(val.(string))
	}
	if val, ok := d.GetOk("type"); ok {
		req = req.Type_(val.(string))
	}

	return req, nil
}

func loadSgTruststoreOthersPostBody(req secretgroup_truststore.DefaultApiPostSecretGroupTruststoreRequest, d *schema.ResourceData) (secretgroup_truststore.DefaultApiPostSecretGroupTruststoreRequest, error) {
	// if val, ok := d.GetOk("expiration_date"); ok {
	// 	req = req.ExpirationDate(val.(string))
	// }
	if val, ok := d.GetOk("name"); ok {
		req = req.Name(val.(string))
	}
	if val, ok := d.GetOk("type"); ok {
		req = req.Type_(val.(string))
	}
	if val, ok := d.GetOk("algorithm"); ok {
		req = req.Algorithm(val.(string))
	}
	if val, ok := d.GetOk("store_passphrase"); ok {
		req = req.StorePassphrase(val.(string))
	}
	if val, ok := d.GetOk("truststore"); ok {
		file, err := os.Open(val.(string))
		if err != nil {
			return req, err
		}
		req = req.TrustStore(file)
	}
	return req, nil
}

func loadSgTruststorePutBody(req secretgroup_truststore.DefaultApiPutSecretGroupTruststoreRequest, d *schema.ResourceData) (secretgroup_truststore.DefaultApiPutSecretGroupTruststoreRequest, error) {
	t := d.Get("type").(string)
	if isSgTruststorePEM(t) {
		return loadSgTruststorePEMPutBody(req, d)
	} else {
		return loadSgTruststoreOthersPutBody(req, d)
	}
}

func loadSgTruststorePEMPutBody(req secretgroup_truststore.DefaultApiPutSecretGroupTruststoreRequest, d *schema.ResourceData) (secretgroup_truststore.DefaultApiPutSecretGroupTruststoreRequest, error) {
	// if val, ok := d.GetOk("expiration_date"); ok {
	// 	req = req.ExpirationDate(val.(string))
	// }
	if val, ok := d.GetOk("truststore"); ok {
		file, err := os.Open(val.(string))
		if err != nil {
			return req, err
		}
		req = req.TrustStore(file)
	}
	if val, ok := d.GetOk("name"); ok {
		req = req.Name(val.(string))
	}
	if val, ok := d.GetOk("type"); ok {
		req = req.Type_(val.(string))
	}
	return req, nil
}

func loadSgTruststoreOthersPutBody(req secretgroup_truststore.DefaultApiPutSecretGroupTruststoreRequest, d *schema.ResourceData) (secretgroup_truststore.DefaultApiPutSecretGroupTruststoreRequest, error) {
	if val, ok := d.GetOk("algorithm"); ok {
		req = req.Algorithm(val.(string))
	}
	if val, ok := d.GetOk("store_passphrase"); ok {
		req = req.StorePassphrase(val.(string))
	}
	return loadSgTruststorePEMPutBody(req, d)
}

// depending on the type of the truststore, checks if required properties are checked
func validateTruststoreInput(d *schema.ResourceDiff) error {
	t := d.Get("type").(string)
	var required []string
	if isSgTruststorePEM(t) {
		required = []string{"truststore"}
	} else {
		required = []string{"truststore", "store_passphrase"}
	}
	for _, r := range required {
		if _, ok := d.GetOk(r); !ok {
			return fmt.Errorf("missing required attribute \"%s\" path for truststore type %s", r, t)
		}
	}
	return nil
}

// returns the composed of the secret
func decomposeSgTruststoreId(d *schema.ResourceData) (string, string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2], s[3]
}

func getSgTruststorePEMUpdatableAttributes() []string {
	attributes := [...]string{
		"allow_expired_cert", "name", "type", "truststore",
	}
	return attributes[:]
}

func getSgTruststoreOthersUpdatableAttributes() []string {
	attributes := [...]string{
		"allow_expired_cert", "name", "type", "truststore", "store_passphrase", "algorithm",
	}
	return attributes[:]
}

/*
 * Returns authentication context (includes authorization header)
 */
func getSgTruststoreAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, secretgroup_truststore.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, secretgroup_truststore.ContextServerIndex, pco.server_index)
}
