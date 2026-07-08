package anypoint

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mulesoft-anypoint/anypoint-client-go/private_space_tlscontext"
)

func preparePrivateSpaceTlsContextJKSResourceSchema() map[string]*schema.Schema {
	pstc_schema := cloneSchema(PRIVATE_SPACE_TLSCONTEXT_SCHEMA)
	pstc_schema["org_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The master organization id where the private space is defined.",
	}
	pstc_schema["private_space_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The private space id.",
	}
	pstc_schema["name"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The name of the tls context.",
	}
	pstc_schema["keystore"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The keystore file content encoded in base64.",
	}
	pstc_schema["keystore_passphrase"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The keystore passphrase.",
	}
	pstc_schema["key_passphrase"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "",
		Description: "The private key passphrase.",
	}
	pstc_schema["alias"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "",
		Description: "The alias of the certificate.",
	}
	pstc_schema["keystore_file_name"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The keystore filename.",
	}
	pstc_schema["trust_store"] = &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "The trust store.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"source": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"PEM"}, false)),
					Description:      "The source of the trust store, can be 'PEM' only for now.",
					Default:          "PEM",
				},
				"content": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The content of the certificate in PEM format used as trust store.",
				},
			},
		},
	}
	pstc_schema["certificate_file_name"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The certificate filename.",
	}
	pstc_schema["capath_file_name"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The capath filename.",
	}
	pstc_schema["cipher_aes128_gcm_sha256"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_aes128_sha256"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_aes256_gcm_sha384"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_aes256_sha256"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_dhe_rsa_aes128_sha256"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_dhe_rsa_aes256_gcm_sha384"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_dhe_rsa_aes256_sha256"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_ecdhe_ecdsa_aes128_gcm_sha256"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_ecdhe_ecdsa_aes256_gcm_sha384"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_ecdhe_rsa_aes128_gcm_sha256"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_ecdhe_rsa_aes256_gcm_sha384"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_ecdhe_ecdsa_chacha20_poly1305"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_ecdhe_rsa_chacha20_poly1305"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_dhe_rsa_chacha20_poly1305"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_tls_aes256_gcm_sha384"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_tls_chacha20_poly1305_sha256"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "The ciphers used by the tls context.",
	}
	pstc_schema["cipher_tls_aes128_gcm_sha256"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "The ciphers used by the tls context.",
	}

	return pstc_schema
}

func resourcePrivateSpaceTlsContextJKS() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePrivateSpaceTlsContextJKSCreate,
		ReadContext:   resourcePrivateSpaceTlsContextJKSRead,
		UpdateContext: resourcePrivateSpaceTlsContextJKSUpdate,
		DeleteContext: resourcePrivateSpaceTlsContextJKSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: `
		Manages a ` + "`" + `private space tls context (of type JKS)` + "`" + ` in your private space.
		**NOTE**: This resource when applied on creation of the private space, becomes the default tls context until the private space is properly provisioned (about 30minutes). Meanwhile the resource cannot be deleted.
		`,
		Schema: preparePrivateSpaceTlsContextJKSResourceSchema(),
	}
}

func resourcePrivateSpaceTlsContextJKSCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	private_space_id := d.Get("private_space_id").(string)
	authctx := getPrivateSpaceTlsContextAuthCtx(ctx, &pco)
	body := createPrivateSpaceTlsContextJksBody(d)
	//request
	tlscontext, httpr, err := pco.privatespacetlscontextclient.DefaultApi.CreateTlsContext(authctx, orgid, private_space_id).TlsContextPostBody(*body).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Create Private Space TLS Context for org " + orgid + " and private space " + private_space_id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(tlscontext.GetId())

	return resourcePrivateSpaceTlsContextJKSRead(ctx, d, m)
}

func resourcePrivateSpaceTlsContextJKSRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	private_space_id := d.Get("private_space_id").(string)
	authctx := getPrivateSpaceTlsContextAuthCtx(ctx, &pco)
	id := d.Get("id").(string)
	if isComposedResourceId(id) {
		orgid, private_space_id, id, diags = decomposePrivateSpaceTlsContextId(d)
	}
	if diags.HasError() {
		return diags
	}
	//request
	tlscontext, httpr, err := pco.privatespacetlscontextclient.DefaultApi.GetTlsContext(authctx, orgid, private_space_id, id).Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		details := extractAPIErrorDetail(err, httpr)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Read Private Space TLS Context for org " + orgid + " and private space " + private_space_id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	private_space_tls_context := flattenPrivateSpaceTlsContextData(tlscontext)
	//save in data source schema
	if err := setPrivateSpaceTlsContextAttributesToResourceData(d, private_space_tls_context); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set Private Space TLS Context " + id + " in org " + orgid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(id)
	d.Set("org_id", orgid)
	d.Set("private_space_id", private_space_id)
	return diags
}

func resourcePrivateSpaceTlsContextJKSUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	private_space_id := d.Get("private_space_id").(string)
	authctx := getPrivateSpaceTlsContextAuthCtx(ctx, &pco)
	id := d.Get("id").(string)
	if d.HasChanges(updatablePrivateSpaceTlsContextJksAttributes()...) {
		body := createPrivateSpaceTlsContextJksBody(d)
		//request
		tlscontext, httpr, err := pco.privatespacetlscontextclient.DefaultApi.UpdateTlsContext(authctx, orgid, private_space_id, id).Body(*body).Execute()
		if err != nil {
			details := extractAPIErrorDetail(err, httpr)
			diags := append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to Update Private Space TLS Context for org " + orgid + " and private space " + private_space_id,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()

		d.SetId(tlscontext.GetId())

		return resourcePrivateSpaceTlsContextJKSRead(ctx, d, m)
	}
	return diags
}

func resourcePrivateSpaceTlsContextJKSDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	private_space_id := d.Get("private_space_id").(string)
	authctx := getPrivateSpaceTlsContextAuthCtx(ctx, &pco)
	id := d.Get("id").(string)
	//request
	httpr, err := pco.privatespacetlscontextclient.DefaultApi.DeleteTlsContext(authctx, orgid, private_space_id, id).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Delete Private Space TLS Context for org " + orgid + " and private space " + private_space_id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId("")
	return diags
}

func createPrivateSpaceTlsContextJksBody(d *schema.ResourceData) *private_space_tlscontext.TlsContextPostBody {
	body := private_space_tlscontext.NewTlsContextPostBody()
	body.SetName(d.Get("name").(string))
	tlsconfig := private_space_tlscontext.NewTlsContextPostBodyTlsConfig()
	tlsconfig.SetKeyStore(private_space_tlscontext.TlsContextPostBodyTlsConfigKeyStore{
		TlsContextPostBodyKeyStoreJKS: createPrivateSpaceTlsContextJksKeystore(d),
	})
	tlsconfig.SetTrustStore(createPrivateSpaceTlsContextTrustStore(d)) //defined in the tlscontext PEM resource file
	body.SetCiphers(*createPrivateSpaceTlsContextCiphers(d))           //defined in the tlscontext PEM resource file
	body.SetTlsConfig(*tlsconfig)
	return body
}

func createPrivateSpaceTlsContextJksKeystore(d *schema.ResourceData) *private_space_tlscontext.TlsContextPostBodyKeyStoreJKS {
	keystore := private_space_tlscontext.NewTlsContextPostBodyKeyStoreJKSWithDefaults()
	keystore.SetKeystoreBase64(d.Get("keystore").(string))
	keystore.SetStorePassphrase(d.Get("keystore_passphrase").(string))
	keystore.SetKeyPassphrase(d.Get("key_passphrase").(string))
	keystore.SetAlias(d.Get("alias").(string))
	if val := d.Get("keystore_file_name"); val != nil && val.(string) != "" {
		keystore.SetKeystoreFileName(val.(string))
	}

	return keystore
}

func updatablePrivateSpaceTlsContextJksAttributes() []string {
	return []string{
		"name",
		"keystore",
		"keystore_passphrase",
		"key_passphrase",
		"alias",
		"keystore_file_name",
		"trust_store",
		"cipher_aes128_gcm_sha256",
		"cipher_aes128_sha256",
		"cipher_aes256_gcm_sha384",
		"cipher_aes256_sha256",
		"cipher_dhe_rsa_aes128_gcm_sha256",
		"cipher_dhe_rsa_aes128_sha256",
		"cipher_dhe_rsa_aes256_gcm_sha384",
		"cipher_dhe_rsa_aes256_sha256",
		"cipher_ecdhe_ecdsa_aes128_gcm_sha256",
		"cipher_ecdhe_ecdsa_aes128_sha1",
		"cipher_ecdhe_ecdsa_aes256_gcm_sha384",
		"cipher_ecdhe_ecdsa_aes256_sha1",
		"cipher_ecdhe_rsa_aes128_gcm_sha256",
		"cipher_ecdhe_rsa_aes128_sha1",
		"cipher_ecdhe_rsa_aes256_gcm_sha384",
		"cipher_ecdhe_rsa_aes256_sha1",
		"cipher_ecdhe_ecdsa_chacha20_poly1305",
		"cipher_ecdhe_rsa_chacha20_poly1305",
		"cipher_dhe_rsa_chacha20_poly1305",
		"cipher_tls_aes256_gcm_sha384",
		"cipher_tls_chacha20_poly1305_sha256",
		"cipher_tls_aes128_gcm_sha256",
	}
}
