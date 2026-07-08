package anypoint

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePrivateSpaceSupportedVpnConfigs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePrivateSpaceSupportedVpnConfigsRead,
		Description: "Returns the vendor / model / firmware-version compatibility matrix maintained by " +
			"Anypoint for private-space VPN customer endpoints. Read-only reference data — use it to " +
			"validate that your customer-side gear is supported before standing up a `anypoint_private_space_vpn` resource.",
		Schema: map[string]*schema.Schema{
			"org_id": {Type: schema.TypeString, Required: true, Description: "The organization id used to authenticate the request. The response itself is org-independent."},
			"vendors": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Supported vendors with their model and firmware compatibility matrix.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {Type: schema.TypeString, Computed: true, Description: "Vendor name."},
						"models": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {Type: schema.TypeString, Computed: true, Description: "Model family."},
									"firmware_versions": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "Supported firmware version strings.",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourcePrivateSpaceSupportedVpnConfigsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	res, httpr, err := pco.privatespaceclient.DefaultAPI.GetPrivateSpaceSupportedVpnConfigs(authctx, orgid).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Read supported VPN configs for org " + orgid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	matrix := *res
	vendorNames := make([]string, 0, len(matrix))
	for k := range matrix {
		vendorNames = append(vendorNames, k)
	}
	sort.Strings(vendorNames)
	vendors := make([]map[string]any, 0, len(vendorNames))
	for _, vname := range vendorNames {
		modelMap := matrix[vname]
		modelNames := make([]string, 0, len(modelMap))
		for k := range modelMap {
			modelNames = append(modelNames, k)
		}
		sort.Strings(modelNames)
		models := make([]map[string]any, 0, len(modelNames))
		for _, mname := range modelNames {
			models = append(models, map[string]any{
				"name":              mname,
				"firmware_versions": modelMap[mname],
			})
		}
		vendors = append(vendors, map[string]any{
			"name":   vname,
			"models": models,
		})
	}
	if err := d.Set("vendors", vendors); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set vendors attribute",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(ComposeResourceId([]string{orgid, "supportedVpnConfigs"}))
	return diags
}
