package anypoint

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mulesoft-anypoint/anypoint-client-go/private_space_tlscontext"
)

func preparePrivateSpaceTlsContextPemResourceSchema() map[string]*schema.Schema {
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
	pstc_schema["certificate"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The certificate.",
	}
	pstc_schema["key"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: "The private key.",
	}
	pstc_schema["capath"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The CA Path Certificate.",
	}
	pstc_schema["key_passphrase"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Sensitive:   true,
		Description: "The private key passphrase.",
	}
	pstc_schema["key_file_name"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The private key filename.",
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
					Optional:    true,
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

func resourcePrivateSpaceTlsContextPem() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePrivateSpaceTlsContextPemCreate,
		ReadContext:   resourcePrivateSpaceTlsContextPemRead,
		UpdateContext: resourcePrivateSpaceTlsContextPemUpdate,
		DeleteContext: resourcePrivateSpaceTlsContextPemDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: `
		Manages a ` + "`" + `private space tls context (of type PEM)` + "`" + ` in your private space.
		**NOTE**: This resource when applied on creation of the private space, becomes the default tls context until the private space is properly provisioned (about 30minutes). Meanwhile the resource cannot be deleted.
		`,
		Schema: preparePrivateSpaceTlsContextPemResourceSchema(),
	}
}

func resourcePrivateSpaceTlsContextPemCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	private_space_id := d.Get("private_space_id").(string)
	authctx := getPrivateSpaceTlsContextAuthCtx(ctx, &pco)
	body := createPrivateSpaceTlsContextPemBody(d)
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

	return resourcePrivateSpaceTlsContextPemRead(ctx, d, m)
}

func resourcePrivateSpaceTlsContextPemRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	private_space_id := d.Get("private_space_id").(string)
	id := d.Get("id").(string)
	if isComposedResourceId(id) {
		orgid, private_space_id, id, diags = decomposePrivateSpaceTlsContextId(d)
	}
	if diags.HasError() {
		return diags
	}
	authctx := getPrivateSpaceTlsContextAuthCtx(ctx, &pco)
	//request
	res, httpr, err := pco.privatespacetlscontextclient.DefaultApi.GetTlsContext(authctx, orgid, private_space_id, id).Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		details := extractAPIErrorDetail(err, httpr)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Get Private Space TLS Context for org " + orgid + " and private space " + private_space_id + " and id " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	private_space_tls_context := flattenPrivateSpaceTlsContextData(res)
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

func resourcePrivateSpaceTlsContextPemUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	private_space_id := d.Get("private_space_id").(string)
	id := d.Get("id").(string)
	authctx := getPrivateSpaceTlsContextAuthCtx(ctx, &pco)
	body := createPrivateSpaceTlsContextPemBody(d)

	if d.HasChanges(updatablePrivateSpaceTlsContextPemAttributes()...) {
		//request
		_, httpr, err := pco.privatespacetlscontextclient.DefaultApi.UpdateTlsContext(authctx, orgid, private_space_id, id).Body(*body).Execute()
		if err != nil {
			details := extractAPIErrorDetail(err, httpr)
			diags := append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to Update Private Space TLS Context for org " + orgid + " and private space " + private_space_id + " and id " + id,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()

		return resourcePrivateSpaceTlsContextPemRead(ctx, d, m)
	}
	return diags
}

func resourcePrivateSpaceTlsContextPemDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	private_space_id := d.Get("private_space_id").(string)
	id := d.Get("id").(string)
	authctx := getPrivateSpaceTlsContextAuthCtx(ctx, &pco)
	//request
	httpr, err := pco.privatespacetlscontextclient.DefaultApi.DeleteTlsContext(authctx, orgid, private_space_id, id).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Delete Private Space TLS Context for org " + orgid + " and id " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.SetId("")
	return diags
}

func createPrivateSpaceTlsContextPemBody(d *schema.ResourceData) *private_space_tlscontext.TlsContextPostBody {
	body := private_space_tlscontext.NewTlsContextPostBody()
	body.SetName(d.Get("name").(string))
	tlsconfig := private_space_tlscontext.NewTlsContextPostBodyTlsConfig()
	tlsconfig.SetKeyStore(private_space_tlscontext.TlsContextPostBodyTlsConfigKeyStore{
		TlsContextPostBodyKeyStorePEM: createPrivateSpaceTlsContextPemKeystore(d),
	})
	tlsconfig.SetTrustStore(createPrivateSpaceTlsContextTrustStore(d))
	body.SetCiphers(*createPrivateSpaceTlsContextCiphers(d))
	body.SetTlsConfig(*tlsconfig)
	return body
}

func createPrivateSpaceTlsContextPemKeystore(d *schema.ResourceData) *private_space_tlscontext.TlsContextPostBodyKeyStorePEM {
	keystore := private_space_tlscontext.NewTlsContextPostBodyKeyStorePEMWithDefaults()
	keystore.SetKey(d.Get("key").(string))
	keystore.SetCertificate(d.Get("certificate").(string))
	if val := d.Get("capath"); val != nil && val.(string) != "" {
		keystore.SetCapath(val.(string))
	}
	if val := d.Get("key_passphrase"); val != nil && val.(string) != "" {
		keystore.SetKeyPassphrase(val.(string))
	}
	if val := d.Get("key_file_name"); val != nil && val.(string) != "" {
		keystore.SetKeyFileName(val.(string))
	}
	if val := d.Get("certificate_file_name"); val != nil && val.(string) != "" {
		keystore.SetCertificateFileName(val.(string))
	}
	if val := d.Get("capath_file_name"); val != nil && val.(string) != "" {
		keystore.SetCapathFileName(val.(string))
	}
	return keystore
}

func createPrivateSpaceTlsContextTrustStore(d *schema.ResourceData) []private_space_tlscontext.TlsContextPostBodyTrustStorePEM {
	result := []private_space_tlscontext.TlsContextPostBodyTrustStorePEM{}
	if data, ok := d.GetOk("trust_store"); ok {
		for _, item := range data.([]any) {
			itemMap := item.(map[string]any)
			truststore := private_space_tlscontext.NewTlsContextPostBodyTrustStorePEM()
			truststore.SetSource(itemMap["source"].(string))
			truststore.SetTrustStorePEM(itemMap["trust_store_pem"].(string))
			result = append(result, *truststore)
		}
	}
	return result
}

func createPrivateSpaceTlsContextCiphers(d *schema.ResourceData) *private_space_tlscontext.Ciphers {
	ciphers := private_space_tlscontext.NewCiphers()
	if val, ok := d.GetOk("cipher_aes128_gcm_sha256"); ok {
		ciphers.SetAes128GcmSha256(val.(bool))
	}
	if val, ok := d.GetOk("cipher_aes128_sha256"); ok {
		ciphers.SetAes128Sha256(val.(bool))
	}
	if val, ok := d.GetOk("cipher_aes256_gcm_sha384"); ok {
		ciphers.SetAes256GcmSha384(val.(bool))
	}
	if val, ok := d.GetOk("cipher_aes256_sha256"); ok {
		ciphers.SetAes256Sha256(val.(bool))
	}
	if val, ok := d.GetOk("cipher_dhe_rsa_aes128_sha256"); ok {
		ciphers.SetDheRsaAes128Sha256(val.(bool))
	}
	if val, ok := d.GetOk("cipher_dhe_rsa_aes256_gcm_sha384"); ok {
		ciphers.SetDheRsaAes256GcmSha384(val.(bool))
	}
	if val, ok := d.GetOk("cipher_dhe_rsa_aes256_sha256"); ok {
		ciphers.SetDheRsaAes256Sha256(val.(bool))
	}
	if val, ok := d.GetOk("cipher_ecdhe_ecdsa_aes128_gcm_sha256"); ok {
		ciphers.SetEcdheEcdsaAes128GcmSha256(val.(bool))
	}
	if val, ok := d.GetOk("cipher_ecdhe_ecdsa_aes256_gcm_sha384"); ok {
		ciphers.SetEcdheEcdsaAes256GcmSha384(val.(bool))
	}
	if val, ok := d.GetOk("cipher_ecdhe_rsa_aes128_gcm_sha256"); ok {
		ciphers.SetEcdheRsaAes128GcmSha256(val.(bool))
	}
	if val, ok := d.GetOk("cipher_ecdhe_rsa_aes256_gcm_sha384"); ok {
		ciphers.SetEcdheRsaAes256GcmSha384(val.(bool))
	}
	if val, ok := d.GetOk("cipher_ecdhe_ecdsa_chacha20_poly1305"); ok {
		ciphers.SetEcdheEcdsaChacha20Poly1305(val.(bool))
	}
	if val, ok := d.GetOk("cipher_ecdhe_rsa_chacha20_poly1305"); ok {
		ciphers.SetEcdheRsaChacha20Poly1305(val.(bool))
	}
	if val, ok := d.GetOk("cipher_dhe_rsa_chacha20_poly1305"); ok {
		ciphers.SetDheRsaChacha20Poly1305(val.(bool))
	}
	if val, ok := d.GetOk("cipher_tls_aes256_gcm_sha384"); ok {
		ciphers.SetTlsAes256GcmSha384(val.(bool))
	}
	if val, ok := d.GetOk("cipher_tls_chacha20_poly1305_sha256"); ok {
		ciphers.SetTlsChacha20Poly1305Sha256(val.(bool))
	}
	if val, ok := d.GetOk("cipher_tls_aes128_gcm_sha256"); ok {
		ciphers.SetTlsAes128GcmSha256(val.(bool))
	}
	return ciphers
}

func getPrivateSpaceTlsContextAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, private_space_tlscontext.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, private_space_tlscontext.ContextServerIndex, pco.server_index)
}

func decomposePrivateSpaceTlsContextId(d *schema.ResourceData) (string, string, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	s := DecomposeResourceId(d.Id())
	if len(s) != 3 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid Private Space TLS Context ID format",
			Detail:   fmt.Sprintf("Expected ORG_ID/PRIVATE_SPACE_ID/TLS_CONTEXT_ID, got %s", d.Id()),
		})
		return "", "", "", diags
	}
	return s[0], s[1], s[2], diags
}

func updatablePrivateSpaceTlsContextPemAttributes() []string {
	return []string{
		"name",
		"key",
		"certificate",
		"key_passphrase",
		"key_file_name",
		"certificate_file_name",
		"capath_file_name",
		"capath",
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
