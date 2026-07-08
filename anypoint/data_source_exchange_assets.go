package anypoint

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	exchange_assets "github.com/mulesoft-anypoint/anypoint-client-go/exchange_assets"
)

func dataSourceExchangeAssets() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceExchangeAssetsRead,
		Description: "Searches Exchange assets via the /assets/search endpoint.",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Synthetic id for this search result set.",
			},
			"types": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Asset type filter (required by the API).",
			},
			"search": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Substring match against the asset name.",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by organization domain.",
			},
			"master_organization_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by master organization id.",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by organization id.",
			},
			"offset": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Pagination offset.",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum number of records to return.",
			},
			"sort": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Field to sort the results by.",
			},
			"ascending": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Sort order: 'true' or 'false'.",
			},
			"shared_with_me": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Restrict to assets shared with the caller.",
			},
			"include_snapshots": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Include snapshot versions in the result set.",
			},
			"assets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"id":               {Type: schema.TypeString, Computed: true},
					"group_id":         {Type: schema.TypeString, Computed: true},
					"asset_id":         {Type: schema.TypeString, Computed: true},
					"version":          {Type: schema.TypeString, Computed: true},
					"minor_version":    {Type: schema.TypeString, Computed: true},
					"version_group":    {Type: schema.TypeString, Computed: true},
					"name":             {Type: schema.TypeString, Computed: true},
					"description":      {Type: schema.TypeString, Computed: true},
					"type":             {Type: schema.TypeString, Computed: true},
					"status":           {Type: schema.TypeString, Computed: true},
					"is_public":        {Type: schema.TypeBool, Computed: true},
					"is_snapshot":      {Type: schema.TypeBool, Computed: true},
					"contact_name":     {Type: schema.TypeString, Computed: true},
					"contact_email":    {Type: schema.TypeString, Computed: true},
					"created_at":       {Type: schema.TypeString, Computed: true},
					"created_date":     {Type: schema.TypeString, Computed: true},
					"updated_date":     {Type: schema.TypeString, Computed: true},
					"modified_at":      {Type: schema.TypeString, Computed: true},
					"icon":             {Type: schema.TypeString, Computed: true},
					"min_mule_version": {Type: schema.TypeString, Computed: true},
					"rating":           {Type: schema.TypeInt, Computed: true},
					"number_of_rates":  {Type: schema.TypeInt, Computed: true},
					"labels":           {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				}},
			},
		},
	}
}

func dataSourceExchangeAssetsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	authctx := getExchangeAssetAuthCtx(ctx, &pco)

	req := pco.exchangeassetsclient.DefaultAPI.AssetsSearchGet(authctx).
		Types(d.Get("types").(string))

	if v, ok := d.GetOk("search"); ok {
		req = req.Search(v.(string))
	}
	if v, ok := d.GetOk("domain"); ok {
		req = req.Domain(v.(string))
	}
	if v, ok := d.GetOk("master_organization_id"); ok {
		req = req.MasterOrganizationId(v.(string))
	}
	if v, ok := d.GetOk("organization_id"); ok {
		req = req.OrganizationId(v.(string))
	}
	if v, ok := d.GetOk("offset"); ok {
		req = req.Offset(int32(v.(int)))
	}
	if v, ok := d.GetOk("limit"); ok {
		req = req.Limit(int32(v.(int)))
	}
	if v, ok := d.GetOk("sort"); ok {
		req = req.Sort(v.(string))
	}
	if v, ok := d.GetOk("ascending"); ok {
		req = req.Ascending(v.(string))
	}
	if v, ok := d.GetOk("shared_with_me"); ok {
		req = req.SharedWithMe(v.(bool))
	}
	if v, ok := d.GetOk("include_snapshots"); ok {
		req = req.IncludeSnapshots(v.(bool))
	}

	res, httpr, err := req.Execute()
	if err != nil {
		return exchangeAssetHTTPDiag(httpr, err, "Unable to search exchange assets")
	}
	defer httpr.Body.Close()

	if err := d.Set("assets", flattenExchangeSearchResults(res)); err != nil {
		return diag.Diagnostics{{Severity: diag.Error, Summary: "Unable to set assets", Detail: err.Error()}}
	}

	d.SetId(time.Now().UTC().Format(time.RFC3339Nano))

	return diags
}

func flattenExchangeSearchResults(items []exchange_assets.AssetSearchResultItem) []map[string]any {
	out := make([]map[string]any, len(items))
	for i, it := range items {
		out[i] = map[string]any{
			"id":               it.GetId(),
			"group_id":         it.GetGroupId(),
			"asset_id":         it.GetAssetId(),
			"version":          it.GetVersion(),
			"minor_version":    it.GetMinorVersion(),
			"version_group":    it.GetVersionGroup(),
			"name":             it.GetName(),
			"description":      it.GetDescription(),
			"type":             it.GetType(),
			"status":           it.GetStatus(),
			"is_public":        it.GetIsPublic(),
			"is_snapshot":      it.GetIsSnapshot(),
			"contact_name":     it.GetContactName(),
			"contact_email":    it.GetContactEmail(),
			"created_at":       it.GetCreatedAt(),
			"created_date":     it.GetCreatedDate(),
			"updated_date":     it.GetUpdatedDate(),
			"modified_at":      it.GetModifiedAt(),
			"icon":             it.GetIcon(),
			"min_mule_version": it.GetMinMuleVersion(),
			"rating":           int(it.GetRating()),
			"number_of_rates":  int(it.GetNumberOfRates()),
			"labels":           it.GetLabels(),
		}
	}
	return out
}
