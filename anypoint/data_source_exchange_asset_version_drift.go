package anypoint

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceExchangeAssetVersionDrift() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceExchangeAssetVersionDriftRead,
		Description: `
		Compares a declared Exchange asset version against the latest available version
		from Anypoint Exchange and classifies the drift as ` + "`none`" + `, ` + "`patch`" + `, ` + "`minor`" + `,
		` + "`major`" + `, or ` + "`unknown`" + `.

		Useful for surfacing outdated asset references (apim policy templates, custom
		policy templates, API specs, any Exchange asset) on dashboards or in CI drift
		reports without forcing a Terraform plan to mutate the resource.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Composite id of the drift query (asset_group_id/asset_id/declared_version).",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Anypoint organization id used as the authentication context for the Exchange lookup.",
			},
			"asset_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Exchange asset group id (also called organization id in some Exchange responses). For MuleSoft-published policies this is the MuleSoft global org id.",
			},
			"asset_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Exchange asset id.",
			},
			"declared_version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The version currently declared by the consumer (e.g. `anypoint_apim_policy_custom.x.asset_version`).",
			},
			"latest_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The latest version reported by Exchange for this asset. Empty when the asset is not found.",
			},
			"available_versions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "All available versions for this asset, sorted descending. Non-SemVer values appear at the tail, sorted lexicographically among themselves.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"drift_severity": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "One of `none`, `patch`, `minor`, `major`, `unknown`. `none` means the declared version is at or ahead of the latest; `unknown` means either side could not be parsed as SemVer or the asset was not found.",
			},
			"is_outdated": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True when the declared version is strictly behind the latest version (under SemVer comparison). False for `none` and `unknown`.",
			},
			"is_known": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True when Exchange returned a version for the requested asset. False when the asset does not exist or could not be read.",
			},
		},
	}
}

func dataSourceExchangeAssetVersionDriftRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	orgId := d.Get("org_id").(string)
	assetGroupId := d.Get("asset_group_id").(string)
	assetId := d.Get("asset_id").(string)
	declaredVersion := d.Get("declared_version").(string)

	authctx := getExchangeAssetAuthCtx(ctx, &pco)
	asset, httpr, err := pco.exchangeassetsclient.DefaultAPI.
		AssetsOrgIdAssetIdAssetGet(authctx, assetGroupId, assetId).Execute()

	d.SetId(ComposeResourceId([]string{assetGroupId, assetId, declaredVersion}))
	_ = orgId // present for auth context selection consistency with sibling data sources

	if err != nil {
		if httpr != nil && (httpr.StatusCode == http.StatusNotFound || httpr.StatusCode == http.StatusUnauthorized) {
			// Treat 404 / 401 as "asset unknown" rather than a hard error so consumers
			// can render a degraded drift report instead of failing the whole plan.
			defer httpr.Body.Close()
			result := ClassifyVersionDrift(declaredVersion, "", nil)
			return setExchangeAssetVersionDriftAttrs(d, result, false)
		}
		return exchangeAssetHTTPDiag(httpr, err, "Unable to read exchange asset "+assetGroupId+"/"+assetId+" for version drift")
	}
	defer httpr.Body.Close()

	latest := asset.GetVersion()
	available := append([]string{latest}, asset.GetVersions()...)
	result := ClassifyVersionDrift(declaredVersion, latest, available)

	if diags := setExchangeAssetVersionDriftAttrs(d, result, true); diags != nil {
		return diags
	}
	return diags
}

func setExchangeAssetVersionDriftAttrs(d *schema.ResourceData, r VersionDriftResult, isKnown bool) diag.Diagnostics {
	setters := []struct {
		key string
		val any
	}{
		{"latest_version", r.LatestVersion},
		{"available_versions", r.AvailableVersions},
		{"drift_severity", r.DriftSeverity},
		{"is_outdated", r.IsOutdated},
		{"is_known", isKnown},
	}
	for _, s := range setters {
		if err := d.Set(s.key, s.val); err != nil {
			return diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Unable to set version drift attribute " + s.key,
				Detail:   err.Error(),
			}}
		}
	}
	return nil
}
