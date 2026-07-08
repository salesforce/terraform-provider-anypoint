package anypoint

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	apim_tier "github.com/mulesoft-anypoint/anypoint-client-go/apim_tier"
)

func dataSourceApiInstanceSlaTier() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApiInstanceSlaTierRead,
		Description: "Reads a single SLA tier from an API Manager instance.",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The composite id of this resource: org_id/env_id/api_id/tier_id.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization id where the resource is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The environment id where the resource is defined.",
			},
			"api_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The API Manager instance id.",
			},
			"tier_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The numeric SLA tier id assigned by the platform.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The SLA tier name.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The SLA tier description.",
			},
			"auto_approve": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether contract requests for this tier are auto-approved.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The tier status.",
			},
			"api_version_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The API version id.",
			},
			"limits": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Rate limit definitions for this tier.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"visible": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this limit is visible to API consumers.",
						},
						"time_period_in_milliseconds": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The time window in milliseconds.",
						},
						"maximum_requests": {
							Type:        schema.TypeInt,
							Computed:    true,
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

func dataSourceApiInstanceSlaTierRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	id := d.Get("id").(string)
	orgid, envid, apiid, tierIdStr, err := splitApimTierId(id)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse SLA tier id " + id,
			Detail:   err.Error(),
		})
		return diags
	}
	tierId, err := strconv.Atoi(tierIdStr)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse tier_id from " + id,
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
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "SLA tier " + tierIdStr + " not found for api " + apiid,
			Detail:   "Tier not found in list response.",
		})
		return diags
	}

	d.SetId(id)
	d.Set("org_id", orgid)
	d.Set("env_id", envid)
	d.Set("api_id", apiid)

	return setApimTierAttrs(d, tier)
}

func setApimTierAttrs(d *schema.ResourceData, tier *apim_tier.SlaTier) diag.Diagnostics {
	var diags diag.Diagnostics

	if err := d.Set("tier_id", int(tier.GetId())); err != nil {
		diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set tier_id", Detail: err.Error()})
		return diags
	}
	if err := d.Set("name", tier.GetName()); err != nil {
		diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set name", Detail: err.Error()})
		return diags
	}
	if err := d.Set("description", tier.GetDescription()); err != nil {
		diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set description", Detail: err.Error()})
		return diags
	}
	if err := d.Set("auto_approve", tier.GetAutoApprove()); err != nil {
		diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set auto_approve", Detail: err.Error()})
		return diags
	}
	if err := d.Set("status", tier.GetStatus()); err != nil {
		diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set status", Detail: err.Error()})
		return diags
	}
	if err := d.Set("api_version_id", tier.GetApiVersionId()); err != nil {
		diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set api_version_id", Detail: err.Error()})
		return diags
	}
	if err := d.Set("application_count", int(tier.GetApplicationCount())); err != nil {
		diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set application_count", Detail: err.Error()})
		return diags
	}
	if err := d.Set("master_organization_id", tier.GetMasterOrganizationId()); err != nil {
		diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set master_organization_id", Detail: err.Error()})
		return diags
	}
	if err := d.Set("organization_id", tier.GetOrganizationId()); err != nil {
		diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set organization_id", Detail: err.Error()})
		return diags
	}

	if audit, ok := tier.GetAuditOk(); ok && audit != nil {
		if created, ok := audit.GetCreatedOk(); ok && created != nil {
			d.Set("created_at", created.GetDate())
		}
		if updated, ok := audit.GetUpdatedOk(); ok && updated != nil {
			d.Set("updated_at", updated.GetDate())
		}
	}

	limits := flattenApimTierLimits(tier.GetLimits())
	if err := d.Set("limits", limits); err != nil {
		diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: "Unable to set limits", Detail: err.Error()})
		return diags
	}

	return diags
}

func flattenApimTierLimits(limits []apim_tier.SlaLimit) []map[string]any {
	result := make([]map[string]any, len(limits))
	for i, l := range limits {
		result[i] = map[string]any{
			"visible":                     l.GetVisible(),
			"time_period_in_milliseconds": int(l.GetTimePeriodInMilliseconds()),
			"maximum_requests":            int(l.GetMaximumRequests()),
		}
	}
	return result
}
