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
	"github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_keystore"
)

func resourceSecretGroupKeystore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretGroupKeystoreCreate,
		ReadContext:   resourceSecretGroupKeystoreRead,
		UpdateContext: resourceSecretGroupKeystoreUpdate,
		DeleteContext: resourceSecretGroupKeystoreDelete,
		Description: `
		Create and manage a keystore for a secret-group in a given organization and environment.
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
				Description: "Id assigned to this keystore.",
			},
			"sg_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The secret-group id where the keystore instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organization id where the keystore's secret group is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The environment id where the keystore's secret group is defined.",
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
				Description: "The name of the keystore",
			},
			"expiration_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The expiration date of the keystore",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The specific type of the keystore",
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice(
						[]string{"PEM", "JKS", "JCEKS", "PKCS12"},
						false,
					),
				),
			},
			"key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The path to the encrypted private key. Required in case of PEM type.",
			},
			"key_passphrase": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Passphrase with which private key for a particular alias is protected.",
			},
			"certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The path to the public certificate. Required in the case of PEM type.",
			},
			"capath": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The path to the concatenated chain of CA certificates, except the leaf, leading up to the root CA. Can only be set in case of PEM type.",
			},
			"store_passphrase": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Passphrase with which keystore is protected. Required in case of JKS, JCEKS and PKCS12 types",
			},
			"keystore": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The path to the file containing one or more certificate entries. Required in case of JKS, JCEKS and PKCS12 types",
			},
			"algorithm": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Algorithm used to create the keystore manager factory which will make use of this keystore",
			},
			"alias": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The alias name of the entry that contains the certificate",
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
		CustomizeDiff: func(ctx context.Context, rd *schema.ResourceDiff, i interface{}) error {
			return validateKeystoreInput(rd)
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSecretGroupKeystoreCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	name := d.Get("name").(string)
	allow_expired_cert := d.Get("allow_expired_cert").(bool)
	authctx := getSgKeystoreAuthCtx(ctx, &pco)
	//prepare request
	req := pco.sgkeystoreclient.DefaultApi.PostSecretGroupKeystores(authctx, orgid, envid, sgid).AllowExpiredCert(allow_expired_cert)
	req, err := loadSgKeystorePostBody(req, d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create keystore " + name,
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
			Summary:  "Unable to create keystore " + name,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	id := res.GetId()
	d.SetId(id)

	return resourceSecretGroupKeystoreRead(ctx, d, m)
}

func resourceSecretGroupKeystoreRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	sgid := d.Get("sg_id").(string)
	id := d.Get("id").(string)
	if isComposedResourceId(id) {
		orgid, envid, sgid, id = decomposeSgKeystoreId(d)
	}
	authctx := getSgKeystoreAuthCtx(ctx, &pco)
	res, httpr, err := pco.sgkeystoreclient.DefaultApi.GetSecretGroupKeystoreDetails(authctx, orgid, envid, sgid, id).Execute()
	if err != nil && httpr.StatusCode >= 400 {
		var details string
		if httpr != nil {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read keystore " + id,
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
	d.SetId(id)
	d.Set("sg_id", sgid)
	d.Set("env_id", envid)
	d.Set("org_id", orgid)

	return diags
}

func resourceSecretGroupKeystoreUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var attributes []string
	t := d.Get("type").(string)
	if isSgKeystorePEM(t) {
		attributes = getSgKeystorePEMUpdatableAttributes()
	} else {
		attributes = getSgKeystoreOthersUpdatableAttributes()
	}
	if d.HasChanges(attributes...) {
		pco := m.(ProviderConfOutput)
		orgid := d.Get("org_id").(string)
		envid := d.Get("env_id").(string)
		sgid := d.Get("sg_id").(string)
		id := d.Get("id").(string)
		allow_expired_cert := d.Get("allow_expired_cert").(bool)
		authctx := getSgKeystoreAuthCtx(ctx, &pco)
		//prepare request
		req := pco.sgkeystoreclient.DefaultApi.PutSecretGroupKeystore(authctx, orgid, envid, sgid, id).AllowExpiredCert(allow_expired_cert)
		req, err := loadSgKeystorePutBody(req, d)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update keystore " + id,
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
				Summary:  "Unable to update keystore " + id,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		d.Set("last_updated", time.Now().Format(time.RFC850))
		return resourceSecretGroupKeystoreRead(ctx, d, m)
	}
	return diags
}

func resourceSecretGroupKeystoreDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// NOTE: The delete action is not supported for this resource.
	// a keystore cannot be deleted, only secret-group (parent) can be deleted
	// Therefore we are only removing reference here
	d.SetId("")
	return diags
}

func loadSgKeystorePostBody(req secretgroup_keystore.DefaultApiPostSecretGroupKeystoresRequest, d *schema.ResourceData) (secretgroup_keystore.DefaultApiPostSecretGroupKeystoresRequest, error) {
	t := d.Get("type").(string)
	if isSgKeystorePEM(t) {
		return loadSgKeystorePEMPostBody(req, d)
	} else {
		return loadSgKeystoreOthersPostBody(req, d)
	}
}

func loadSgKeystorePEMPostBody(req secretgroup_keystore.DefaultApiPostSecretGroupKeystoresRequest, d *schema.ResourceData) (secretgroup_keystore.DefaultApiPostSecretGroupKeystoresRequest, error) {
	// if val, ok := d.GetOk("expiration_date"); ok {
	// 	req = req.ExpirationDate(val.(string))
	// }
	if val, ok := d.GetOk("key"); ok {
		file, err := os.Open(val.(string))
		if err != nil {
			return req, err
		}
		req = req.Key(file)
	}
	if val, ok := d.GetOk("name"); ok {
		req = req.Name(val.(string))
	}
	if val, ok := d.GetOk("key_passphrase"); ok {
		req = req.KeyPassphrase(val.(string))
	}
	if val, ok := d.GetOk("certificate"); ok {
		file, err := os.Open(val.(string))
		if err != nil {
			return req, err
		}
		req = req.Certificate(file)
	}
	if val, ok := d.GetOk("type"); ok {
		req = req.Type_(val.(string))
	}
	if val, ok := d.GetOk("capath"); ok {
		file, err := os.Open(val.(string))
		if err != nil {
			return req, err
		}
		req = req.Capath(file)
	}

	return req, nil
}

func loadSgKeystoreOthersPostBody(req secretgroup_keystore.DefaultApiPostSecretGroupKeystoresRequest, d *schema.ResourceData) (secretgroup_keystore.DefaultApiPostSecretGroupKeystoresRequest, error) {
	// if val, ok := d.GetOk("expiration_date"); ok {
	// 	req = req.ExpirationDate(val.(string))
	// }
	if val, ok := d.GetOk("keystore"); ok {
		file, err := os.Open(val.(string))
		if err != nil {
			return req, err
		}
		req = req.KeyStore(file)
	}
	if val, ok := d.GetOk("name"); ok {
		req = req.Name(val.(string))
	}
	if val, ok := d.GetOk("key_passphrase"); ok {
		req = req.KeyPassphrase(val.(string))
	}
	if val, ok := d.GetOk("store_passphrase"); ok {
		req = req.StorePassphrase(val.(string))
	}
	if val, ok := d.GetOk("type"); ok {
		req = req.Type_(val.(string))
	}
	if val, ok := d.GetOk("algorithm"); ok {
		req = req.Algorithm(val.(string))
	}
	if val, ok := d.GetOk("alias"); ok {
		req = req.Alias(val.(string))
	}

	return req, nil
}

func loadSgKeystorePutBody(req secretgroup_keystore.DefaultApiPutSecretGroupKeystoreRequest, d *schema.ResourceData) (secretgroup_keystore.DefaultApiPutSecretGroupKeystoreRequest, error) {
	t := d.Get("type").(string)
	if isSgKeystorePEM(t) {
		return loadSgKeystorePEMPutBody(req, d)
	} else {
		return loadSgKeystoreOthersPutBody(req, d)
	}
}

func loadSgKeystorePEMPutBody(req secretgroup_keystore.DefaultApiPutSecretGroupKeystoreRequest, d *schema.ResourceData) (secretgroup_keystore.DefaultApiPutSecretGroupKeystoreRequest, error) {
	// if val, ok := d.GetOk("expiration_date"); ok {
	// 	req = req.ExpirationDate(val.(string))
	// }
	if val, ok := d.GetOk("key"); ok {
		file, err := os.Open(val.(string))
		if err != nil {
			return req, err
		}
		req = req.Key(file)
	}
	if val, ok := d.GetOk("name"); ok {
		req = req.Name(val.(string))
	}
	if val, ok := d.GetOk("key_passphrase"); ok {
		req = req.KeyPassphrase(val.(string))
	}
	if val, ok := d.GetOk("certificate"); ok {
		file, err := os.Open(val.(string))
		if err != nil {
			return req, err
		}
		req = req.Certificate(file)
	}
	if val, ok := d.GetOk("type"); ok {
		req = req.Type_(val.(string))
	}
	if val, ok := d.GetOk("capath"); ok {
		file, err := os.Open(val.(string))
		if err != nil {
			return req, err
		}
		req = req.Capath(file)
	}

	return req, nil
}

func loadSgKeystoreOthersPutBody(req secretgroup_keystore.DefaultApiPutSecretGroupKeystoreRequest, d *schema.ResourceData) (secretgroup_keystore.DefaultApiPutSecretGroupKeystoreRequest, error) {
	// if val, ok := d.GetOk("expiration_date"); ok {
	// 	req = req.ExpirationDate(val.(string))
	// }
	if val, ok := d.GetOk("keystore"); ok {
		file, err := os.Open(val.(string))
		if err != nil {
			return req, err
		}
		req = req.KeyStore(file)
	}
	if val, ok := d.GetOk("name"); ok {
		req = req.Name(val.(string))
	}
	if val, ok := d.GetOk("key_passphrase"); ok {
		req = req.KeyPassphrase(val.(string))
	}
	if val, ok := d.GetOk("store_passphrase"); ok {
		req = req.StorePassphrase(val.(string))
	}
	if val, ok := d.GetOk("type"); ok {
		req = req.Type_(val.(string))
	}
	if val, ok := d.GetOk("algorithm"); ok {
		req = req.Algorithm(val.(string))
	}
	if val, ok := d.GetOk("alias"); ok {
		req = req.Alias(val.(string))
	}

	return req, nil
}

// returns the composed of the secret
func decomposeSgKeystoreId(d *schema.ResourceData) (string, string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2], s[3]
}

// depending on the type of the keystore, checks if required properties are checked
func validateKeystoreInput(d *schema.ResourceDiff) error {
	t := d.Get("type").(string)
	var required []string
	if isSgKeystorePEM(t) {
		required = []string{"key", "certificate"}
	} else {
		required = []string{"keystore", "alias", "store_passphrase"}
	}
	for _, r := range required {
		if _, ok := d.GetOk(r); !ok {
			return fmt.Errorf("missing required attribute \"%s\" path for keystore type %s", r, t)
		}
	}
	return nil
}

func getSgKeystorePEMUpdatableAttributes() []string {
	attributes := [...]string{
		"allow_expired_cert", "name", "type", "key", "key_passphrase",
		"certificate", "capath",
	}
	return attributes[:]
}

func getSgKeystoreOthersUpdatableAttributes() []string {
	attributes := [...]string{
		"allow_expired_cert", "name", "type", "keystore", "key_passphrase",
		"store_passphrase", "algorithm", "alias",
	}
	return attributes[:]
}

/*
 * Returns authentication context (includes authorization header)
 */
func getSgKeystoreAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, secretgroup_keystore.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, secretgroup_keystore.ContextServerIndex, pco.server_index)
}
