package anypoint

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rtf "github.com/mulesoft-anypoint/anypoint-client-go/rtf"
)

func dataSourceFabricsCollection() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAllFabricsRead,
		Description: `
		Reads all ` + "`" + `Runtime Fabrics'` + "`" + ` available in your org.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Description: "The business group id",
				Required:    true,
			},
			"list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique id of the fabrics instance in the platform.",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The organization id where the fabrics is hosted.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of this fabrics instance.",
						},
						"region": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region where fabrics instance is hosted.",
						},
						"vendor": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The vendor name of the kubernetes instance hosting fabrics.",
						},
						"vendor_metadata": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The vendor metadata",
						},
						"version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The version of fabrics.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the farbics instance.",
						},
						"desired_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The desired version of fabrics.",
						},
						"available_upgrade_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The available upgrade version of fabrics.",
						},
						"created_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The creation date of the fabrics instance.",
						},
						"upgrade": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The status of the fabrics. Only available when instance is created and not activated yet.",
							Elem:        FabricsUpgradeDefinition,
						},
						"nodes": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        NodeDefinition,
							Description: "The list of fabrics nodes.",
						},
						"activation_data": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The activation data to use during installation of fabrics on the kubernetes cluster. Only available when instance is created and not activated yet.",
						},
						"seconds_since_heartbeat": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of seconds since last heartbeat.",
						},
						"kubernetes_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The kubernetes version of the cluster.",
						},
						"namespace": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The namespace where runtime fabrics is installed.",
						},
						"license_expiry_date": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The expiry date of the license (timestamp).",
						},
						"is_managed": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this cluster is managed.",
						},
						"is_helm_managed": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this cluster is managed by helmet.",
						},
						"app_scoped_log_forwarding": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether app scoped log forwarding is active.",
						},
						"cluster_configuration_level": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The configuration level of the cluster (production or development).",
						},
						"features": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The features of this cluster.",
							Elem:        FabricsFeaturesDefinition,
						},
						"ingress": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The ingress configurations of this cluster.",
							Elem:        FabricsIngressDomainsDefinition,
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Description: "The total number of available results",
				Computed:    true,
			},
		},
	}
}

func dataSourceAllFabricsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	authctx := getFabricsAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.rtfclient.DefaultApi.GetAllFabrics(authctx, orgid).Execute()
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
			Summary:  "Unable to get fabrics list",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	list := flattenFabricsCollectionData(res)
	//save in data source schema
	if err := d.Set("list", list); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set fabrics list",
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("total", len(list)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number fabrics",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

/*
* Transforms a set of runtime fabrics to the dataSourceFabricsCollection schema
* @param fabricsCollection *[]rtf.Fabrics the list of fabrics
* @return list of generic items
 */
func flattenFabricsCollectionData(fabricsCollection []rtf.Fabrics) []interface{} {
	if len(fabricsCollection) == 0 {
		return []interface{}{}
	}

	data := make([]interface{}, len(fabricsCollection))
	for i, fabrics := range fabricsCollection {
		data[i] = flattenFabricsData(&fabrics)
	}
	return data
}
