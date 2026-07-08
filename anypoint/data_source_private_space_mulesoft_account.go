package anypoint

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePrivateSpaceMulesoftAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePrivateSpaceMulesoftAccountRead,
		Description: "Returns the MuleSoft AWS account id (region-bound to the private space). " +
			"Use as the principal of a `aws_ram_principal_association` so MuleSoft can attach a Transit Gateway via the shared RAM resource.",
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the private space lives.",
			},
			"private_space_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The private space id.",
			},
			"account_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The MuleSoft AWS account id (region-bound).",
			},
		},
	}
}

func dataSourcePrivateSpaceMulesoftAccountRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	psid := d.Get("private_space_id").(string)
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	res, httpr, err := pco.privatespaceclient.DefaultAPI.GetPrivateSpaceMulesoftAccount(authctx, orgid, psid).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Read MuleSoft AWS account for private space " + psid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.SetId(orgid + "/" + psid + "/mulesoft_account")
	d.Set("account_id", res)
	return diags
}
