package anypoint

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/private_space"
)

var PRIVATE_SPACE_SUMMARY_SCHEMA = map[string]*schema.Schema{
	"id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The unique identifier of the private space.",
	},
	"name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The name of the private space.",
	},
	"org_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ID of the organization to which the private space belongs.",
	},
	"root_org_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ID of the root organization to which the private space belongs.",
	},
	"status": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The current status of the private space.",
	},
}

func dataSourcePrivateSpaces() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePrivateSpacesRead,
		Description: `
		Reads all ` + "`" + `private spaces` + "`" + ` in your business group.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The master organization id where the private spaces are defined.",
			},
			"private_spaces": {
				Type:        schema.TypeList,
				Description: "List of private spaces for the given org",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: PRIVATE_SPACE_SUMMARY_SCHEMA,
				},
			},
		},
	}
}

func dataSourcePrivateSpacesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	//request
	res, httpr, err := pco.privatespaceclient.DefaultAPI.GetPrivateSpaces(authctx, orgid).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Get Private Spaces for org " + orgid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	private_spaces := flattenPrivateSpaceSummary(res)
	//save in data source schema
	if err := d.Set("private_spaces", private_spaces); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set Private Spaces for org " + orgid,
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenPrivateSpaceSummary(data *private_space.PrivateSpaceSummary) []map[string]any {
	if data == nil {
		return []map[string]any{}
	}
	content := data.GetContent()
	result := make([]map[string]any, len(content))
	for i, item := range content {
		result[i] = flattenPrivateSpaceSummaryContent(&item)
	}
	return result
}

func flattenPrivateSpaceSummaryContent(data *private_space.PrivateSpaceSummaryContentItem) map[string]any {
	result := make(map[string]any)
	if data == nil {
		return result
	}
	result["id"] = data.GetId()
	result["name"] = data.GetName()
	result["org_id"] = data.GetOrganizationId()
	result["root_org_id"] = data.GetRootOrganizationId()
	result["status"] = data.GetStatus()

	return result
}
