package anypoint

import (
	"context"
	"io"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/secretgroup"
)

func resourceSecretGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretGroupCreate,
		ReadContext:   resourceSecretGroupRead,
		UpdateContext: resourceSecretGroupUpdate,
		DeleteContext: resourceSecretGroupDelete,
		Description: `
		Create a secret group for a given organization and environment.
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
				Description: "The unique id of this secret group",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organization id where the secret group instance is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The environment id where the secret group instance is defined.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the secret group",
			},
			"downloadable": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
				Description: "Setting this to true indicates that the secrets from this secret group are allowed to be downloadable by end users, altough, through other applications.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time at which this secret group was created",
			},
			"modified_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time at which this secret group was last modified",
			},
			"modified_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Username of the anypoint platform user who modified this secret group",
			},
			"locked": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether this secret group is currently locked",
			},
			"locked_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Username of the anypoint platform user who currently holds the lock on this secret group. This is present only if 'locked' is set to true.",
			},
			"current_state": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `
				The state the secret group is currently in. If the state is anything but 'Clear', only the corresponding operation is allowed on the secret group, as follows.

					* If currentState="Finishing", only operation allowed - delete lock with action = finish.
					* If currentState="Cancelling", only operation allowed - delete lock with action = cancel.
					* If currentState="Deleting", only operation allowed - delete secret group.
				`,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSecretGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	authctx := getSecretGroupAuthCtx(ctx, &pco)
	body := newSecretGroupPostBody(d) // init request body
	// execute post request
	res, httpr, err := pco.secretgroupclient.DefaultApi.PostSecretGroup(authctx, orgid, envid).SecretGroupPostBody(*body).Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create secret group for org " + orgid + " and env " + envid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//update ids following the creation
	id := res.GetId()
	d.SetId(id)

	return resourceSecretGroupRead(ctx, d, m)
}

func resourceSecretGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	id := d.Get("id").(string)
	if isComposedResourceId(id) {
		orgid, envid, id = decomposeSecretGroupId(d)
	}
	authctx := getSecretGroupAuthCtx(ctx, &pco)
	res, httpr, err := pco.secretgroupclient.DefaultApi.GetSecretGroup(authctx, orgid, envid, id).Execute()
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
			Summary:  "Unable to get secret group " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process result
	data := flattenSecretGroupResult(res)
	if err := setSecretGroupAttributesToResourceData(d, data); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set secret group " + id,
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(id)
	d.Set("env_id", envid)
	d.Set("org_id", orgid)

	return diags
}

func resourceSecretGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	if d.HasChange("name") {
		orgid := d.Get("org_id").(string)
		envid := d.Get("env_id").(string)
		id := d.Get("id").(string)
		authctx := getSecretGroupAuthCtx(ctx, &pco)
		body := newSecretGroupPatchBody(d)
		_, httpr, err := pco.secretgroupclient.DefaultApi.PatchSecretGroup(authctx, orgid, envid, id).SecretGroupPatchBody(*body).Execute()
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
				Summary:  "Unable to update secret group " + id,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		d.Set("last_updated", time.Now().Format(time.RFC850))
		return resourceSecretGroupRead(ctx, d, m)
	}

	return diags
}

func resourceSecretGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	id := d.Get("id").(string)
	authctx := getSecretGroupAuthCtx(ctx, &pco)
	httpr, err := pco.secretgroupclient.DefaultApi.DeleteSecretGroup(authctx, orgid, envid, id).Execute()
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
			Summary:  "Unable to delete secret group " + id,
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

func newSecretGroupPostBody(d *schema.ResourceData) *secretgroup.SecretGroupPostBody {
	body := secretgroup.NewSecretGroupPostBody()
	if val, ok := d.GetOk("name"); ok {
		body.SetName(val.(string))
	}
	if val, ok := d.GetOk("downloadable"); ok {
		body.SetDownloadable(val.(bool))
	}
	return body
}

func newSecretGroupPatchBody(d *schema.ResourceData) *secretgroup.SecretGroupPatchBody {
	body := secretgroup.NewSecretGroupPatchBody()
	if val, ok := d.GetOk("name"); ok {
		body.SetName(val.(string))
	}
	return body
}

// returns the composed of the secret group
func decomposeSecretGroupId(d *schema.ResourceData) (string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2]
}

/*
 * Returns authentication context (includes authorization header)
 */
func getSecretGroupAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, secretgroup.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, secretgroup.ContextServerIndex, pco.server_index)
}
