package anypoint

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	env "github.com/mulesoft-consulting/cloudhub-client-go/env"
)

func resourceENV() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceENVCreate,
		ReadContext:   resourceENVRead,
		UpdateContext: resourceENVUpdate,
		DeleteContext: resourceENVDelete,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"is_production": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceENVCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)

	authctx := getENVAuthCtx(ctx, &pco)

	body := newENVPostBody(d)

	//request env creation
	res, httpr, err := pco.envclient.DefaultApi.OrganizationsOrgIdEnvironmentsPost(authctx, orgid).EnvCore(*body).Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Create ENV",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(res.GetId())

	resourceENVRead(ctx, d, m)

	return diags
}

func resourceENVRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	envid := d.Id()
	orgid := d.Get("org_id").(string)

	authctx := getENVAuthCtx(ctx, &pco)

	res, httpr, err := pco.envclient.DefaultApi.OrganizationsOrgIdEnvironmentsEnvironmentIdGet(authctx, orgid, envid).Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Get ENV",
			Detail:   details,
		})
		return diags
	}

	//process data
	envinstance := flattenENVData(&res)
	//save in data source schema
	if err := setENVCoreAttributesToResourceData(d, envinstance); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set ENV",
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func resourceENVUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	envid := d.Id()
	orgid := d.Get("org_id").(string)

	authctx := getENVAuthCtx(ctx, &pco)

	if d.HasChanges(getENVCoreAttributes()...) {
		body := newENVPutBody(d)
		//request env creation
		_, httpr, err := pco.envclient.DefaultApi.OrganizationsOrgIdEnvironmentsEnvironmentIdPut(authctx, orgid, envid).EnvCore(*body).Execute()
		if err != nil {
			var details string
			if httpr != nil {
				b, _ := ioutil.ReadAll(httpr.Body)
				details = string(b)
			} else {
				details = err.Error()
			}
			diags := append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to Update ENV",
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceENVRead(ctx, d, m)
}

func resourceENVDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	envid := d.Id()
	orgid := d.Get("org_id").(string)

	authctx := getENVAuthCtx(ctx, &pco)

	httpr, err := pco.envclient.DefaultApi.OrganizationsOrgIdEnvironmentsEnvironmentIdDelete(authctx, orgid, envid).Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Delete ENV",
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

/*
 * Creates a new ENV Core Struct from the resource data schema
 */
func newENVPostBody(d *schema.ResourceData) *env.EnvCore {
	body := env.NewEnvCoreWithDefaults()

	body.SetName(d.Get("name").(string))
	body.SetType(d.Get("type").(string))
	return body
}

/*
 * Creates a new ENV Core Struct from the resource data schema
 */
func newENVPutBody(d *schema.ResourceData) *env.EnvCore {
	body := env.NewEnvCoreWithDefaults()

	body.SetName(d.Get("name").(string))
	return body
}

/*
 * Returns authentication context (includes authorization header)
 */
func getENVAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	return context.WithValue(ctx, env.ContextAccessToken, pco.access_token)
}
