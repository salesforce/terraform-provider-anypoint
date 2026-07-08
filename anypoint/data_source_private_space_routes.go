package anypoint

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	private_space "github.com/mulesoft-anypoint/anypoint-client-go/private_space"
)

func dataSourcePrivateSpaceRoutes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePrivateSpaceRoutesRead,
		Description: "Reads the static route table of a private space network (destinations + targets).",
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"private_space_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"routes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination": {Type: schema.TypeString, Computed: true},
						"status":      {Type: schema.TypeString, Computed: true},
						"target":      {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourcePrivateSpaceRoutesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	psid := d.Get("private_space_id").(string)
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	list, httpr, err := pco.privatespaceclient.DefaultAPI.GetPrivateSpaceRoutes(authctx, orgid, psid).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Read routes for private space " + psid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	if err := d.Set("routes", flattenPrivateSpaceRoutes(list)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set private space routes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func flattenPrivateSpaceRoutes(list []private_space.PrivateSpaceRoute) []map[string]any {
	out := make([]map[string]any, 0, len(list))
	for i := range list {
		r := &list[i]
		out = append(out, map[string]any{
			"destination": r.GetDestination(),
			"status":      r.GetStatus(),
			"target":      r.GetTarget(),
		})
	}
	return out
}
