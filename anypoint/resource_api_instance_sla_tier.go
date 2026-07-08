package anypoint

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	apim_tier "github.com/mulesoft-anypoint/anypoint-client-go/apim_tier"
)

func resourceApiInstanceSlaTier() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApiInstanceSlaTierCreate,
		ReadContext:   resourceApiInstanceSlaTierRead,
		UpdateContext: resourceApiInstanceSlaTierUpdate,
		DeleteContext: resourceApiInstanceSlaTierDelete,
		Description:   "Creates and manages an SLA tier on an API Manager instance.",
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
				Description: "The unique composite id of this resource: org_id/env_id/api_id/tier_id.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organization id where the resource is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The environment id where the resource is defined.",
			},
			"api_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The API Manager instance id.",
			},
			"tier_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The numeric SLA tier id assigned by the platform.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The SLA tier name.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The SLA tier description.",
			},
			"auto_approve": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether contract requests for this tier are auto-approved.",
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ACTIVE",
				ValidateFunc: validation.StringInSlice([]string{
					"ACTIVE",
					"DEPRECATED",
				}, false),
				Description: "The tier status. One of: ACTIVE, DEPRECATED.",
			},
			"api_version_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The API version id. Defaults to api_id when not provided.",
			},
			"limits": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "Rate limit definitions for this tier. At least one must have visible=true.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"visible": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Whether this limit is visible to API consumers.",
						},
						"time_period_in_milliseconds": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The time window in milliseconds.",
						},
						"maximum_requests": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The maximum number of requests allowed in the time window.",
						},
					},
				},
			},
			"application_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of contracts bound to this tier.",
			},
			"master_organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The root organization id.",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization id returned by the platform.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date the tier was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date the tier was last updated.",
			},
		},
	}
}

func resourceApiInstanceSlaTierCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	apiid := d.Get("api_id").(string)

	authctx := getApimTierAuthCtx(ctx, &pco)
	body := newApimTierPostBody(d)

	res, httpr, err := pco.apimtierclient.DefaultAPI.CreateApiInstanceTier(authctx, orgid, envid, apiid).SlaTierPostBody(*body).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create SLA tier for api " + apiid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	tierId := res.GetId()
	d.SetId(ComposeResourceId([]string{orgid, envid, apiid, strconv.Itoa(int(tierId))}))
	d.Set("tier_id", int(tierId))

	return resourceApiInstanceSlaTierRead(ctx, d, m)
}

func resourceApiInstanceSlaTierRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	orgid, envid, apiid, tierIdStr, err := splitApimTierId(d.Id())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse SLA tier id " + d.Id(),
			Detail:   err.Error(),
		})
		return diags
	}
	tierId, err := strconv.Atoi(tierIdStr)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse tier_id from " + d.Id(),
			Detail:   err.Error(),
		})
		return diags
	}

	authctx := getApimTierAuthCtx(ctx, &pco)

	tier, diags := apimTierReadById(authctx, pco, orgid, envid, apiid, int32(tierId))
	if diags != nil {
		return diags
	}
	if tier == nil {
		d.SetId("")
		return nil
	}

	if err := setApimTierAttrs(d, tier); err != nil {
		return err
	}

	d.Set("org_id", orgid)
	d.Set("env_id", envid)
	d.Set("api_id", apiid)

	return diags
}

func resourceApiInstanceSlaTierUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	if !d.HasChanges("name", "description", "auto_approve", "status", "limits", "api_version_id") {
		return diags
	}

	orgid, envid, apiid, tierIdStr, err := splitApimTierId(d.Id())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse SLA tier id " + d.Id(),
			Detail:   err.Error(),
		})
		return diags
	}
	tierId, err := strconv.Atoi(tierIdStr)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse tier_id from " + d.Id(),
			Detail:   err.Error(),
		})
		return diags
	}

	authctx := getApimTierAuthCtx(ctx, &pco)
	body := newApimTierPutBody(d, int32(tierId))

	res, httpr, err := pco.apimtierclient.DefaultAPI.UpdateApiInstanceTier(authctx, orgid, envid, apiid, int32(tierId)).SlaTierPutBody(*body).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to update SLA tier " + tierIdStr,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	if err := setApimTierAttrs(d, res); err != nil {
		return err
	}
	d.Set("last_updated", time.Now().Format(time.RFC850))

	return diags
}

func resourceApiInstanceSlaTierDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	orgid, envid, apiid, tierIdStr, err := splitApimTierId(d.Id())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse SLA tier id " + d.Id(),
			Detail:   err.Error(),
		})
		return diags
	}
	tierId, err := strconv.Atoi(tierIdStr)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse tier_id from " + d.Id(),
			Detail:   err.Error(),
		})
		return diags
	}

	authctx := getApimTierAuthCtx(ctx, &pco)

	httpr, err := pco.apimtierclient.DefaultAPI.DeleteApiInstanceTier(authctx, orgid, envid, apiid, int32(tierId)).Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == 404 {
			d.SetId("")
			return diags
		}
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete SLA tier " + tierIdStr,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	return diags
}

// apimTierReadById lists all tiers and filters by id. No single-GET endpoint exists.
func apimTierReadById(
	authctx context.Context,
	pco ProviderConfOutput,
	orgid, envid, apiid string,
	tierId int32,
) (*apim_tier.SlaTier, diag.Diagnostics) {
	offset := int32(0)
	limit := int32(200)

	for {
		res, httpr, err := pco.apimtierclient.DefaultAPI.GetApiInstanceTiers(authctx, orgid, envid, apiid).
			Limit(limit).Offset(offset).Execute()
		if err != nil {
			details := extractAPIErrorDetail(err, httpr)
			return nil, diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Unable to read SLA tiers for api " + apiid,
				Detail:   details,
			}}
		}
		defer httpr.Body.Close()

		tiers := res.GetTiers()
		for i := range tiers {
			if tiers[i].GetId() == tierId {
				return &tiers[i], nil
			}
		}

		total := int32(res.GetTotal())
		offset += limit
		if offset >= total {
			break
		}
	}

	return nil, nil
}

func newApimTierPostBody(d *schema.ResourceData) *apim_tier.SlaTierPostBody {
	body := apim_tier.NewSlaTierPostBody()
	body.SetName(d.Get("name").(string))
	body.SetAutoApprove(d.Get("auto_approve").(bool))
	body.SetStatus(d.Get("status").(string))

	if v, ok := d.GetOk("description"); ok {
		body.SetDescription(v.(string))
	}

	apiVersionId := d.Get("api_version_id").(string)
	if apiVersionId == "" {
		apiVersionId = d.Get("api_id").(string)
	}
	body.SetApiVersionId(apiVersionId)

	body.SetLimits(expandApimTierLimits(d))
	return body
}

func newApimTierPutBody(d *schema.ResourceData, tierId int32) *apim_tier.SlaTierPutBody {
	body := apim_tier.NewSlaTierPutBody()
	body.SetId(tierId)
	body.SetName(d.Get("name").(string))
	body.SetAutoApprove(d.Get("auto_approve").(bool))
	body.SetStatus(d.Get("status").(string))

	if v, ok := d.GetOk("description"); ok {
		body.SetDescription(v.(string))
	}

	apiVersionId := d.Get("api_version_id").(string)
	if apiVersionId == "" {
		apiVersionId = d.Get("api_id").(string)
	}
	body.SetApiVersionId(apiVersionId)

	apiidStr := d.Get("api_id").(string)
	if apiid, err := strconv.Atoi(apiidStr); err == nil {
		body.SetApiId(int32(apiid))
	}

	body.SetLimits(expandApimTierLimits(d))
	return body
}

func expandApimTierLimits(d *schema.ResourceData) []apim_tier.SlaLimit {
	raw := d.Get("limits").([]any)
	limits := make([]apim_tier.SlaLimit, len(raw))
	for i, item := range raw {
		m := item.(map[string]any)
		l := apim_tier.NewSlaLimit()
		l.SetVisible(m["visible"].(bool))
		l.SetTimePeriodInMilliseconds(int64(m["time_period_in_milliseconds"].(int)))
		l.SetMaximumRequests(int64(m["maximum_requests"].(int)))
		limits[i] = *l
	}
	return limits
}

func splitApimTierId(id string) (orgid, envid, apiid, tierid string, err error) {
	parts := DecomposeResourceId(id)
	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("expected org_id/env_id/api_id/tier_id, got: %s", id)
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

func getApimTierAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, apim_tier.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, apim_tier.ContextServerIndex, pco.server_index)
}
