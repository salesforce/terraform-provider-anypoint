package anypoint

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFlexGatewayRegistrationToken() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFlexGatewayRegistrationTokenRead,
		Description: `
		Retrieve a flex gateway registration token used to register a new flex gateway instance.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the flex gateway targets are defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the flex gateway targets are defined.",
			},
			"registration_token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The registration token that can be used to register a new flex gateway",
			},
		},
	}
}

func dataSourceFlexGatewayRegistrationTokenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	authctx := getFlexGatewayAuthCtx(ctx, &pco)

	res, httpr, err := pco.flexgatewayclient.DefaultApi.GetFlexGatewayRegistrationToken(authctx, orgid, envid).Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get flex gateway registration token ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	if err := d.Set("registration_token", res.GetRegistrationToken()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set registration token of fex-gateway org " + orgid + " and env " + envid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
