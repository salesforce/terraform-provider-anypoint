package anypoint

import (
	"context"
	"fmt"
	"io"
	"maps"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iancoleman/strcase"
	"github.com/mulesoft-anypoint/anypoint-client-go/apim_policy"
)

func resourceApimInstancePolicyJwtValidation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApimInstancePolicyJwtValidationCreate,
		ReadContext:   resourceApimInstancePolicyJwtValidationRead,
		UpdateContext: resourceApimInstancePolicyJwtValidationUpdate,
		DeleteContext: resourceApimInstancePolicyJwtValidationDelete,
		Description: `
		Create and manage an API Policy of type jwt-validation.
		`,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The last time this resource has been updated locally.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy's unique id",
			},
			"apim_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The api manager instance id where the api instance is defined.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organization id where the api instance is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The environment id where api instance is defined.",
			},
			"audit": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The instance's auditing data",
			},
			"master_organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization id where the api instance is defined.",
			},
			"configuration_data": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "The policy configuration data",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"jwt_origin": {
							Type:     schema.TypeString,
							Required: true,
							Description: `
							Whether to use custom header or to use bearer authentication header.
							Values can be either "httpBearerAuthenticationHeader" or "customExpression".
							In the case of using "httpBearerAuthenticationHeader", you don't need to supply jwt_expression.
							`,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice(
									[]string{
										"httpBearerAuthenticationHeader",
										"customExpression",
									},
									false,
								),
							),
						},
						"jwt_expression": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#[attributes.headers['jwt']]",
							Description: "Mule Expression to be used to extract the JWT from API requests",
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringMatch(
									regexp.MustCompile(`^\s*#\[.+\]\s*$`),
									"field value should be a dataweave expression",
								),
							),
						},
						"signing_method": {
							Type:     schema.TypeString,
							Required: true,
							Description: `
							Specifies the method to be used by the policy to decode the JWT.
							Values can be either "rsa", "hmac", "es" and "none".
							`,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice(
									[]string{
										"rsa",
										"hmac",
										"es",
										"none",
									},
									false,
								),
							),
						},
						"signing_key_length": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  256,
							Description: `
							Specifies the length of the key to be in the signing method for HMAC, or the SHA algorithm used for RSA or ES.
							Ignore this field if the JWT Signing Method was set to None.
							`,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.IntInSlice(
									[]int{
										256,
										384,
										512,
									},
								),
							),
						},
						"jwt_key_origin": {
							Type:     schema.TypeString,
							Required: true,
							Description: `
							Origin of the JWT Key. The JWKS option is only supported if the JWT Signing Method was set to RSA or ES.
							Ignore this field if the JWT Signing Method was set to None.
							`,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice(
									[]string{
										"jwks",
										"text",
									},
									false,
								),
							),
						},
						"jwks_url": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "http://your-jwks-service.example:80/base/path",
							Description: `
							The Url to the JWKS server that contains the public keys for the signature validation.
							Ignore this field if the JWT Signing Method was set to None.
							`,
						},
						"jwks_service_time_to_live": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  60,
							Description: `
							The amount of time, in minutes, that the JWKS will be considered valid. Once the JWKS has expired, it will have to be retrieved again.
							Default value is 1 hour. Ignore this field if the JWT Signing Method was set to None.
							`,
						},
						"jwks_service_connection_timeout": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  10000,
							Description: `
							Timeout specification, in milliseconds, when reaching the JWKS service. Default value is 10 seconds.
							`,
						},
						"text_key": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "empty",
							// Sensitive: true,
							Description: `
							The shared secret in case the JWT Signing Method is set to HMAC.
							Include the public PEM key without -----BEGIN PUBLIC KEY----- and -----END PUBLIC KEY----- for RSA or ES signing.
							Ignore this field if the JWT Signing Method was set to None.
							`,
						},
						"skip_client_id_validation": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Skips client application's API contract validation.",
						},
						"client_id_expression": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#[vars.claimSet.client_id]",
							Description: "Expression to obtain the Client ID from the request in order to validate it.",
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringMatch(
									regexp.MustCompile(`^\s*#\[.+\]\s*$`),
									"field value should be a dataweave expression",
								),
							),
						},
						"validate_aud_claim": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "The JWT will be valid only if the aud claim contains at least one audiences value defined here.",
						},
						"mandatory_aud_claim": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether to make Audience claim mandatory. If a claim is marked as mandatory, and this claim is not present in the incoming JWT, the request will fail.",
						},
						"supported_audiences": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "aud.example.com",
							Description: "Comma separated list of supported audience values.",
						},
						"mandatory_exp_claim": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether to make Expiration claim mandatory. If a claim is marked as mandatory, and this claim is not present in the incoming JWT, the request will fail.",
						},
						"mandatory_nbf_claim": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether to make Not Before claim mandatory. If a claim is marked as mandatory, and this claim is not present in the incoming JWT, the request will fail.",
						},
						"validate_custom_claim": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "The JWT will be valid only if all DataWeave expressions defined in custom claims are.",
						},
						"mandatory_custom_claims": {
							Type:     schema.TypeList,
							Optional: true,
							Description: `
							Specify the Claim Name and the literal to validate the value of a claim E.g foo : fooValue If more complex validations must be made or the claim value is an array or an object, provide Claim Name and DataWeave expression to validate the value of a claim.
							E.g. foo : #[vars.claimSet.foo == 'fooValue'] If a claim is marked as mandatory and this claim is not present in the incoming jwt, the request will fail.
							`,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The claim name",
									},
									"value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The value to compare against in literal or dataweave expression",
									},
								},
							},
						},
						"non_mandatory_custom_claims": {
							Type:     schema.TypeList,
							Optional: true,
							Description: `
							Specify the Claim Name and the literal to validate the value of a claim E.g foo : fooValue If more complex validations must be made or the claim value is an array or an object, provide Claim Name and DataWeave expression to validate the value of a claim.
							E.g. foo : #[vars.claimSet.foo == 'fooValue'] If a claim is marked as non-mandatory and this claim is not present in the incoming jwt, the request will not fail.
							`,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The claim name",
									},
									"value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The value to compare against in literal or dataweave expression",
									},
								},
							},
						},
					},
				},
			},
			"policy_template_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy template id",
			},
			"order": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The policy order.",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the policy is disabled.",
			},
			"pointcut_data": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The Method & resource conditions",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method_regex": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "The list of HTTP methods",
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateDiagFunc: validation.ToDiagFunc(
									validation.StringInSlice(
										[]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD", "TRACE"},
										false,
									),
								),
							},
						},
						"uri_template_regex": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "URI template regex",
						},
					},
				},
			},
			"asset_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "68ef9520-24e9-4cf2-b2f5-620025690913",
				Description: "The policy template group id in anypoint exchange. Don't change unless mulesoft has renamed the policy group id.",
			},
			"asset_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "jwt-validation",
				Description: "The policy template id in anypoint exchange. Don't change unless mulesoft has renamed the policy asset id.",
			},
			"asset_version": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "1.3.2",
				Description: "the policy template version in anypoint exchange.",
			},
		},
		CustomizeDiff: func(ctx context.Context, rd *schema.ResourceDiff, i interface{}) error {
			return validateJwtValidationCfg(rd)
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceApimInstancePolicyJwtValidationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	//prepare body
	body := newApimPolicyJwtValidationBody(d)
	//perform request
	res, httpr, err := pco.apimpolicyclient.DefaultApi.PostApimPolicy(authctx, orgid, envid, apimid).ApimPolicyBody(*body).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create policy jwt-validation for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	id := res.GetId()
	d.SetId(strconv.Itoa(int(id)))
	diags = append(diags, resourceApimInstancePolicyJwtValidationRead(ctx, d, m)...)
	//in case disabled
	disabled := d.Get("disabled").(bool)
	if disabled {
		diags = append(diags, disableApimInstancePolicyJwtValidation(ctx, d, m)...)
		diags = append(diags, resourceApimInstancePolicyJwtValidationRead(ctx, d, m)...)
	}

	return diags
}

func resourceApimInstancePolicyJwtValidationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	if isComposedResourceId(id) {
		orgid, envid, apimid, id = decomposeApimPolicyJwtValidationId(d)
	}
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.apimpolicyclient.DefaultApi.GetApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read policy jwt-validation " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// process data
	data := flattenApimInstancePolicy(res)
	data["configuration_data"] = []interface{}{flattenApimPolicyJwtValidationCfg(d, res)}
	if err := setApimInstancePolicyAttributesToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set api policy jwt-validation details attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(id)
	d.Set("apim_id", apimid)
	d.Set("env_id", envid)
	d.Set("org_id", orgid)
	return diags
}

func resourceApimInstancePolicyJwtValidationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	//detect change
	if d.HasChanges("configuration_data", "pointcut_data") {
		pco := m.(ProviderConfOutput)
		orgid := d.Get("org_id").(string)
		envid := d.Get("env_id").(string)
		apimid := d.Get("apim_id").(string)
		id := d.Get("id").(string)
		authctx := getApimPolicyAuthCtx(ctx, &pco)
		//prepare body
		body := newApimPolicyJwtValidationPatchBody(d)
		//perform request
		_, httpr, err := pco.apimpolicyclient.DefaultApi.PatchApimPolicy(authctx, orgid, envid, apimid, id).Body(body).Execute()
		if err != nil {
			var details string
			if httpr != nil && httpr.StatusCode >= 400 {
				defer httpr.Body.Close()
				b, _ := io.ReadAll(httpr.Body)
				details = string(b)
			} else {
				details = err.Error()
			}
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update policy jwt-validation for api " + apimid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		diags = append(diags, resourceApimInstancePolicyJwtValidationRead(ctx, d, m)...)
	}
	if d.HasChange("disabled") {
		disabled := d.Get("disabled").(bool)
		if disabled {
			diags = append(diags, disableApimInstancePolicyJwtValidation(ctx, d, m)...)
		} else {
			diags = append(diags, enableApimInstancePolicyJwtValidation(ctx, d, m)...)
		}
		diags = append(diags, resourceApimInstancePolicyJwtValidationRead(ctx, d, m)...)
	}

	return diags
}

func resourceApimInstancePolicyJwtValidationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	httpr, err := pco.apimpolicyclient.DefaultApi.DeleteApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete policy jwt-validation " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return diags
}

func enableApimInstancePolicyJwtValidation(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	_, httpr, err := pco.apimpolicyclient.DefaultApi.EnableApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to enable policy jwt-validation " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	return diags
}

func disableApimInstancePolicyJwtValidation(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apimid := d.Get("apim_id").(string)
	id := d.Get("id").(string)
	authctx := getApimPolicyAuthCtx(ctx, &pco)
	_, httpr, err := pco.apimpolicyclient.DefaultApi.DisableApimPolicy(authctx, orgid, envid, apimid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to disable policy jwt-validation " + id + " for api " + apimid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	return diags
}

func flattenApimPolicyJwtValidationCfg(d *schema.ResourceData, policy *apim_policy.ApimPolicy) map[string]interface{} {
	data := make(map[string]interface{})
	cfg := policy.GetConfigurationData()
	for k, v := range cfg {
		k_snake := strcase.ToSnake(k)
		data[k_snake] = v
	}
	l := d.Get("configuration_data").([]interface{})
	dst := l[0].(map[string]interface{})
	maps.Copy(dst, data)
	return dst
}

func newApimPolicyJwtValidationBody(d *schema.ResourceData) *apim_policy.ApimPolicyBody {
	body := apim_policy.NewApimPolicyBody()
	if val, ok := d.GetOk("configuration_data"); ok {
		l := val.([]interface{})
		cfg := l[0].(map[string]interface{})
		data := newApimPolicyJwtValidationCfg(cfg)
		body.SetConfigurationData(data)
	}
	if val, ok := d.GetOk("pointcut_data"); ok {
		body.SetPointcutData(newApimPolicyJwtValidationPointcutDataBody(val.([]interface{})))
	}
	if val, ok := d.GetOk("asset_group_id"); ok {
		body.SetGroupId(val.(string))
	}
	if val, ok := d.GetOk("asset_id"); ok {
		body.SetAssetId(val.(string))
	}
	if val, ok := d.GetOk("asset_version"); ok {
		body.SetAssetVersion(val.(string))
	}
	return body
}

func newApimPolicyJwtValidationPatchBody(d *schema.ResourceData) map[string]interface{} {
	body := make(map[string]interface{})
	if val, ok := d.GetOk("configuration_data"); ok {
		l := val.([]interface{})
		cfg := l[0].(map[string]interface{})
		data := newApimPolicyJwtValidationCfg(cfg)
		body["configurationData"] = data
	}
	if val, ok := d.GetOk("pointcut_data"); ok {
		collection := newApimPolicyClientIdEnfPointcutDataBody(val.([]interface{}))
		slice := make([]map[string]interface{}, len(collection))
		for i, item := range collection {
			m, _ := item.ToMap()
			slice[i] = m
		}
		body["pointcutData"] = slice
	} else {
		body["pointcutData"] = nil
	}
	if val, ok := d.GetOk("asset_group_id"); ok {
		body["groupId"] = val
	}
	if val, ok := d.GetOk("asset_id"); ok {
		body["assetId"] = val
	}
	if val, ok := d.GetOk("asset_version"); ok {
		body["assetVersion"] = val
	}
	return body
}

func newApimPolicyJwtValidationCfg(input map[string]interface{}) map[string]interface{} {
	body := make(map[string]interface{})
	attributes := getApimPolicyJwtValidationCfgAttributes()
	for _, attr := range attributes {
		if val, ok := input[attr]; ok {
			body[strcase.ToLowerCamel(attr)] = val
		}
	}
	return body
}

func newApimPolicyJwtValidationPointcutDataBody(collection []interface{}) []apim_policy.PointcutDataItem {
	slice := make([]apim_policy.PointcutDataItem, len(collection))
	for i, item := range collection {
		data := item.(map[string]interface{})
		body := apim_policy.NewPointcutDataItem()
		if val, ok := data["method_regex"]; ok && val != nil {
			set := val.(*schema.Set)
			body.SetMethodRegex(JoinStringInterfaceSlice(set.List(), "|"))
		}
		if val, ok := data["uri_template_regex"]; ok {
			body.SetUriTemplateRegex(val.(string))
		}
		slice[i] = *body
	}
	return slice
}

func validateJwtValidationCfg(d *schema.ResourceDiff) error {
	c := d.Get("configuration_data")
	l := c.([]interface{})
	cfg := l[0].(map[string]interface{})
	jwt_origin := cfg["jwt_origin"].(string)
	if _, ok := cfg["jwt_expression"]; !ok && jwt_origin == "customExpression" {
		return fmt.Errorf("attribute jwt_expression is required in \"customExpression\" mode")
	}
	jwt_key_origin := cfg["jwt_key_origin"].(string)
	jwks_attributes := []string{"jwks_url", "jwks_service_time_to_live", "jwks_service_connection_timeout"}
	if jwt_key_origin == "jwks" {
		for _, attr := range jwks_attributes {
			if _, ok := cfg[attr]; !ok {
				return fmt.Errorf("attribute %s is required in \"jwks\" mode", attr)
			}
		}
	} else if jwt_key_origin == "text" {
		if _, ok := cfg["text_key"]; !ok {
			return fmt.Errorf("attribute text_key is required in \"text\" mode")
		}
	}
	skip_client_id_validation := cfg["skip_client_id_validation"].(bool)
	if _, ok := cfg["client_id_expression"]; !ok && !skip_client_id_validation {
		return fmt.Errorf("attribute client_id_expression is required when skip_client_id_validation is false")
	}
	validate_aud_claim := cfg["validate_aud_claim"].(bool)
	if _, ok := cfg["supported_audiences"]; !ok && validate_aud_claim {
		return fmt.Errorf("attribute supported_audiences is required when validate_aud_claim is true")
	}

	return nil
}

func getApimPolicyJwtValidationCfgAttributes() []string {
	return []string{
		"jwt_origin", "jwt_expression", "signing_method", "signing_key_length",
		"jwt_key_origin", "jwks_url", "jwks_service_time_to_live", "jwks_service_connection_timeout",
		"text_key", "skip_client_id_validation", "client_id_expression", "validate_aud_claim",
		"mandatory_aud_claim", "supported_audiences", "mandatory_exp_claim", "mandatory_nbf_claim",
		"validate_custom_claim", "mandatory_custom_claims", "non_mandatory_custom_claims",
	}
}

func decomposeApimPolicyJwtValidationId(d *schema.ResourceData) (string, string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2], s[3]
}
