package anypoint

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePrivateSpaceIamRoles() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePrivateSpaceIamRolesRead,
		Description: `
		Reads all ` + "`" + `private space iam roles` + "`" + ` in your private space.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"private_space_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"iam_roles": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourcePrivateSpaceIamRolesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	private_space_id := d.Get("private_space_id").(string)
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	//request
	res, httpr, err := pco.privatespaceclient.DefaultAPI.GetPrivateSpaceIamRoles(authctx, orgid, private_space_id).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error reading private space iam roles",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//set state
	if err := d.Set("iam_roles", res.GetRoles()); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error setting private space iam roles",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
