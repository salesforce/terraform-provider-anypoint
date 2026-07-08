package anypoint

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	exchange_assets "github.com/mulesoft-anypoint/anypoint-client-go/exchange_assets"
)

func dataSourceExchangeAsset() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceExchangeAssetRead,
		Description: "Reads a single Exchange asset. If `version` is omitted, the latest published version is returned.",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The composite id of this asset: org_id/group_id/asset_id/version.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id that owns the asset.",
			},
			"asset_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The asset slug.",
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The asset version. Omit to read the latest.",
			},
			"group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The business group id the asset belongs to.",
			},
			"name":            {Type: schema.TypeString, Computed: true},
			"description":     {Type: schema.TypeString, Computed: true},
			"classifier":      {Type: schema.TypeString, Computed: true},
			"minor_version":   {Type: schema.TypeString, Computed: true},
			"version_group":   {Type: schema.TypeString, Computed: true},
			"type":            {Type: schema.TypeString, Computed: true},
			"status":          {Type: schema.TypeString, Computed: true},
			"is_public":       {Type: schema.TypeBool, Computed: true},
			"is_snapshot":     {Type: schema.TypeBool, Computed: true},
			"organization_id": {Type: schema.TypeString, Computed: true},
			"created_by_id":   {Type: schema.TypeString, Computed: true},
			"created_at":      {Type: schema.TypeString, Computed: true},
			"contact_name":    {Type: schema.TypeString, Computed: true},
			"contact_email":   {Type: schema.TypeString, Computed: true},
			"labels":          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"versions":        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"attributes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"key":   {Type: schema.TypeString, Computed: true},
					"value": {Type: schema.TypeString, Computed: true},
				}},
			},
			"categories": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"key":          {Type: schema.TypeString, Computed: true},
					"display_name": {Type: schema.TypeString, Computed: true},
					"value":        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				}},
			},
			"custom_fields": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"key":          {Type: schema.TypeString, Computed: true},
					"display_name": {Type: schema.TypeString, Computed: true},
					"data_type":    {Type: schema.TypeString, Computed: true},
					"text_value":   {Type: schema.TypeString, Computed: true},
					"date_value":   {Type: schema.TypeString, Computed: true},
				}},
			},
			"dependencies_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"organization_id": {Type: schema.TypeString, Computed: true},
					"group_id":        {Type: schema.TypeString, Computed: true},
					"asset_id":        {Type: schema.TypeString, Computed: true},
					"version":         {Type: schema.TypeString, Computed: true},
				}},
			},
		},
	}
}

func dataSourceExchangeAssetRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	orgId := d.Get("org_id").(string)
	assetId := d.Get("asset_id").(string)
	version := d.Get("version").(string)

	authctx := getExchangeAssetAuthCtx(ctx, &pco)

	var (
		asset *exchange_assets.Asset
		httpr *http.Response
		err   error
	)
	if version != "" {
		asset, httpr, err = pco.exchangeassetsclient.DefaultAPI.
			AssetsOrgIdAssetIdVersionAssetGet(authctx, orgId, assetId, version).Execute()
	} else {
		asset, httpr, err = pco.exchangeassetsclient.DefaultAPI.
			AssetsOrgIdAssetIdAssetGet(authctx, orgId, assetId).Execute()
	}
	if err != nil {
		return exchangeAssetHTTPDiag(httpr, err, "Unable to read exchange asset "+assetId)
	}
	defer httpr.Body.Close()

	effectiveVersion := asset.GetVersion()
	groupId := asset.GetGroupId()
	d.SetId(ComposeResourceId([]string{orgId, groupId, assetId, effectiveVersion}))

	if diags := setExchangeAssetAttrs(d, asset); diags != nil {
		return diags
	}
	d.Set("group_id", groupId)
	d.Set("version", effectiveVersion)
	d.Set("classifier", classifierFromAssetAttributes(asset.GetAttributes()))

	return diags
}

func setExchangeAssetAttrs(d *schema.ResourceData, asset *exchange_assets.Asset) diag.Diagnostics {
	attrs := asset.GetAttributes()
	setters := []struct {
		key string
		val any
	}{
		{"name", asset.GetName()},
		{"description", asset.GetDescription()},
		{"minor_version", asset.GetMinorVersion()},
		{"version_group", asset.GetVersionGroup()},
		{"type", asset.GetType()},
		{"status", asset.GetStatus()},
		{"is_public", asset.GetIsPublic()},
		{"is_snapshot", asset.GetIsSnapshot()},
		{"organization_id", asset.GetOrganizationId()},
		{"created_by_id", asset.GetCreatedById()},
		{"created_at", asset.GetCreatedAt()},
		{"contact_name", asset.GetContactName()},
		{"contact_email", asset.GetContactEmail()},
		{"labels", asset.GetLabels()},
		{"versions", asset.GetVersions()},
		{"attributes", flattenExchangeAttributes(attrs)},
		{"categories", flattenExchangeCategories(asset.GetCategories())},
		{"custom_fields", flattenExchangeCustomFields(asset.GetCustomFields())},
		{"dependencies_list", flattenExchangeDependencies(asset.GetDependencies())},
	}
	for _, s := range setters {
		if err := d.Set(s.key, s.val); err != nil {
			return diag.Diagnostics{{Severity: diag.Error, Summary: "Unable to set " + s.key, Detail: err.Error()}}
		}
	}
	return nil
}

// classifierFromAssetAttributes recovers the original `classifier` POST input from the
// API attribute set. Exchange surfaces it under the `original-format` key. Returns ""
// when the attribute is missing.
func classifierFromAssetAttributes(attrs []exchange_assets.Attribute) string {
	for _, a := range attrs {
		if a.GetKey() == "original-format" {
			return a.GetValue()
		}
	}
	return ""
}

func flattenExchangeAttributes(attrs []exchange_assets.Attribute) []map[string]any {
	out := make([]map[string]any, len(attrs))
	for i, a := range attrs {
		out[i] = map[string]any{"key": a.GetKey(), "value": a.GetValue()}
	}
	return out
}

func flattenExchangeCategories(cats []exchange_assets.Category) []map[string]any {
	out := make([]map[string]any, len(cats))
	for i, c := range cats {
		out[i] = map[string]any{
			"key":          c.GetKey(),
			"display_name": c.GetDisplayName(),
			"value":        c.GetValue(),
		}
	}
	return out
}

func flattenExchangeCustomFields(fields []exchange_assets.CustomField) []map[string]any {
	out := make([]map[string]any, len(fields))
	for i, f := range fields {
		out[i] = map[string]any{
			"key":          f.GetKey(),
			"display_name": f.GetDisplayName(),
			"data_type":    f.GetDataType(),
			"text_value":   f.GetTextValue(),
			"date_value":   f.GetDateValue(),
		}
	}
	return out
}

func flattenExchangeDependencies(deps []exchange_assets.Dependency) []map[string]any {
	out := make([]map[string]any, len(deps))
	for i, dep := range deps {
		out[i] = map[string]any{
			"organization_id": dep.GetOrganizationId(),
			"group_id":        dep.GetGroupId(),
			"asset_id":        dep.GetAssetId(),
			"version":         dep.GetVersion(),
		}
	}
	return out
}
