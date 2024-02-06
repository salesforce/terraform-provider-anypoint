package anypoint

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mulesoft-anypoint/anypoint-client-go/ame"
	amq "github.com/mulesoft-anypoint/anypoint-client-go/amq"
)

func dataSourceAME() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAMERead,
		Description: `
		Reads all ` + "`" + `Anypoint MQs` + "`" + ` in your environment's region.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the Anypoint MQ Exchange is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the Anypoint MQ Exchange is defined.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The region id where the Anypoint MQ Exchange is defined. Refer to Anypoint Platform official documentation for the list of available regions",
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice(
						[]string{
							"us-east-1", "us-east-2", "us-west-2", "ca-central-1", "eu-west-1", "eu-west-2",
							"ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "eu-central-1",
						},
						false,
					),
				),
			},
			"params": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The search parameters. Should only provide one occurrence of the block.",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"offset": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Skip over a number of elements by specifying an offset value for the query.",
						},
						"limit": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     20,
							Description: "Limit the number of elements in the response.",
						},
						"starts_with": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Searchs the field from the left using the passed string.",
						},
						"destination_ids": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Includes only results with the given Ids.",
							Elem:        schema.TypeString,
						},
					},
				},
			},
			//Results
			"exchanges": {
				Type:        schema.TypeList,
				Description: "List of exchanges defined in the given region",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exchange_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique id of this Anypoint MQ Exchange.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type is always queue.",
						},
						"encrypted": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "To encrypt the queue.",
						},
					},
				},
			},
		},
	}
}

func dataSourceAMERead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchopts := d.Get("params").(*schema.Set)
	regionid := d.Get("region_id").(string)
	envid := d.Get("env_id").(string)
	orgid := d.Get("org_id").(string)
	authctx := getAMQAuthCtx(ctx, &pco)

	//Preparing request
	req := pco.amqclient.DefaultApi.GetAMQList(authctx, orgid, envid, regionid)
	req, errDiags := parseAMESearchOpts(req, searchopts)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}
	//Executing Request
	res, httpr, err := req.Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Get AMEs",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	amqinstance := flattenAMEsData(&res)
	//save in data source schema
	if err := d.Set("exchanges", amqinstance); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set AMEs",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

// Parses search parameters
func parseAMESearchOpts(req amq.DefaultApiApiGetAMQListRequest, params *schema.Set) (amq.DefaultApiApiGetAMQListRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	req = req.Inclusion("ALL")
	req = req.DestinationType("exchange")

	if params.Len() == 0 {
		return req, diags
	}

	opts := params.List()[0]

	for k, v := range opts.(map[string]interface{}) {
		if k == "offset" {
			req = req.Offset(v.(int32))
			continue
		}
		if k == "limit" {
			req = req.Limit(v.(int32))
			continue
		}
		if k == "starts_with" {
			req = req.StartsWith(v.(string))
			continue
		}
		if k == "destination_ids" {
			req = req.DestinationIds(v.([]string))
			continue
		}
	}

	return req, diags
}

// Copies the given amq instance into the given resource data
func setAMEAttributesToResourceData(d *schema.ResourceData, ameitem map[string]interface{}) error {
	attributes := getAMECoreAttributes()
	if ameitem != nil {
		for _, attr := range attributes {
			if err := d.Set(attr, ameitem[attr]); err != nil {
				return fmt.Errorf("unable to set AME attribute %s\n details: %s", attr, err)
			}
		}
	}
	return nil
}

// flattens a list of Anypoint MQs
func flattenAMEsData(queues *[]amq.Queue) []interface{} {
	if queues != nil && len(*queues) > 0 {
		res := make([]interface{}, len(*queues))
		for i, q := range *queues {
			res[i] = flattenAMQEData(&q)
		}
		return res
	}

	return make([]interface{}, 0)
}

// flattens and maps a given Anypoint MQ object to extract exchange data only
func flattenAMQEData(queue *amq.Queue) map[string]interface{} {
	if queue != nil {
		item := make(map[string]interface{})
		item["exchange_id"] = queue.GetExchangeId()
		item["type"] = queue.GetType()
		item["encrypted"] = queue.GetEncrypted()
		return item
	}

	return nil
}

// flattens and maps a given Anypoint MQ Exchange object
func flattenAMEData(exchange *ame.Exchange) map[string]interface{} {
	if exchange != nil {
		item := make(map[string]interface{})
		item["exchange_id"] = exchange.GetExchangeId()
		item["type"] = exchange.GetType()
		item["encrypted"] = exchange.GetEncrypted()
		return item
	}

	return nil
}

func getAMECoreAttributes() []string {
	attributes := [...]string{
		"exchange_id", "type", "encrypted",
	}
	return attributes[:]
}
