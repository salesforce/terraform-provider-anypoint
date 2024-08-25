package anypoint

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rtf "github.com/mulesoft-anypoint/anypoint-client-go/rtf"
)

func dataSourceFabricsHelmRepoProps() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFabricsHelmRepoPropsRead,
		Description: `
		Reads ` + "`" + `Runtime Fabrics'` + "`" + ` Helm repository properties.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Description: "The business group id",
				Required:    true,
			},
			"rtf_image_registry_endpoint": {
				Type:        schema.TypeString,
				Description: "The runtime fabrics image registry endpoint",
				Computed:    true,
			},
			"rtf_image_registry_user": {
				Type:        schema.TypeString,
				Description: "The user to authenticated to the image registry",
				Computed:    true,
			},
			"rtf_image_registry_password": {
				Type:        schema.TypeString,
				Description: "The password to authenticated to the image registry",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func dataSourceFabricsHelmRepoPropsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	authctx := getFabricsAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.rtfclient.DefaultApi.GetFabricsHelmRepoProps(authctx, orgid).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get fabrics helm repository props",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	data := flattenFabricsHelmRepoProps(res)
	//save in data source schema
	if err := setFabricsHelmRepoResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set fabrics helm repository props attributes",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenFabricsHelmRepoProps(props *rtf.FabricsHelmRepoProps) map[string]interface{} {
	data := make(map[string]interface{})
	if props == nil {
		return data
	}

	if val, ok := props.GetRTF_IMAGE_REGISTRY_ENDPOINTOk(); ok {
		data["rtf_image_registry_endpoint"] = *val
	}
	if val, ok := props.GetRTF_IMAGE_REGISTRY_USEROk(); ok {
		data["rtf_image_registry_user"] = *val
	}
	if val, ok := props.GetRTF_IMAGE_REGISTRY_PASSWORDOk(); ok {
		data["rtf_image_registry_password"] = *val
	}

	return data
}

func setFabricsHelmRepoResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getFabricsHelmRepoAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set fabrics helm repo attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getFabricsHelmRepoAttributes() []string {
	attributes := [...]string{
		"rtf_image_registry_endpoint", "rtf_image_registry_user", "rtf_image_registry_password",
	}
	return attributes[:]
}
