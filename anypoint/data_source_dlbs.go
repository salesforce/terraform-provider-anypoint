package anypoint

import (
	"context"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/anypoint-client-go/dlb"
)

func dataSourceDLBs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDLBsRead,
		Description: `
		Reads all ` + "`" + `dedicated load balancer` + "`" + ` instances in a given VPC.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Description: "Business Group Id",
				Required:    true,
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Description: "Vitual Private Network Id",
				Required:    true,
			},
			"dlbs": {
				Type:        schema.TypeList,
				Description: "List of DLBs for the given vpc",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"deployment_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_config": {
							Type:     schema.TypeMap,
							Computed: true,
						},
						"ip_addresses": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ip_whitelist": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"http_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"default_ssl_endpoint": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ssl_endpoints": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"private_key_digest": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"public_key_label": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"public_key_digest": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"public_key_cn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"private_key_label": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"verify_client_mode": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"mappings": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"input_uri": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"app_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"app_uri": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"upstream_protocol": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"static_ips_disabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"workers": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"default_cipher_suite": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"keep_url_encoding": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"tlsv1": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"upstream_tlsv12": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"proxy_read_timeout": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ip_addresses_info": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"status": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"static_ip": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"double_static_ips": {
							Type:     schema.TypeBool,
							Computed: true,
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

func dataSourceDLBsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	vpcid := d.Get("vpc_id").(string)

	authctx := getDLBAuthCtx(ctx, &pco)

	//request dlb
	res, httpr, err := pco.dlbclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdLoadbalancersGet(authctx, orgid, vpcid).Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Get DLBs for org " + orgid + " and vpc " + vpcid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	dlbs := flattenDLBsData(res.GetData())

	//save in data source schema
	if err := d.Set("dlbs", dlbs); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set DLBs for org " + orgid + " and vpc " + vpcid,
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("total", res.GetTotal()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number DLBs for org " + orgid + " and vpc " + vpcid,
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

/*
 * Transforms a list of dlb.Dlb objects to the dataSourceDLBs schema
 */
func flattenDLBsData(dlbs []dlb.Dlb) []interface{} {
	result := make([]interface{}, len(dlbs))
	for i, dlb := range dlbs {
		item := flattenDLBData(&dlb)
		result[i] = item
	}
	return result
}
