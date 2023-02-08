package anypoint

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	amq "github.com/mulesoft-consulting/anypoint-client-go/amq"
)

func dataSourceAMQ() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAMQRead,
		Description: `
		Reads all ` + "`" + `Anypoint MQs` + "`" + ` in your environment's region.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the Anypoint MQ is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the Anypoint MQ is defined.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The region id where the Anypoint MQ is defined. Refer to Anypoint Platform official documentation for the list of available regions",
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
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"inclusion": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "all",
							Description:      "Defines what properties to fetch",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"all", "minimal"}, false)),
						},
						"destination_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "all",
							Description:      "Defines what type to fetch",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"all", "queue", "exchange"}, false)),
						},
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
			"queues": {
				Type:        schema.TypeList,
				Description: "List of queues defined in the given region",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique id of this Anypoint MQ.",
						},
						"default_ttl": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The default TTL applied to messages in milliseconds.",
						},
						"default_lock_ttl": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The default time to live of the created locks in milliseconds.",
						},
						"default_delivery_delay": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The default delivery delay in seconds.",
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
						"dead_letter_queue_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The queue Id of the dead letter queue to bind to this queue. A FIFO DLQ only works with FIFO queues.",
						},
						"max_deliveries": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The maximum number of attempts after which the message will be routed to DLQ. This field can only be used when dead_letter_queue_id attribute is present.",
						},
						"fifo": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to make this queue a FIFO.",
						},
					},
				},
			},
		},
	}
}

func dataSourceAMQRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	req, errDiags := parseAMQSearchOpts(req, searchopts)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}
	//Executing Request
	res, httpr, err := req.Execute()
	defer httpr.Body.Close()
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
			Summary:  "Unable to Get AMQs",
			Detail:   details,
		})
		return diags
	}
	//process data
	amqinstance := flattenAMQsData(&res)
	//save in data source schema
	if err := d.Set("queues", amqinstance); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set AMQs",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

// Parses search parameters
func parseAMQSearchOpts(req amq.DefaultApiApiGetAMQListRequest, params *schema.Set) (amq.DefaultApiApiGetAMQListRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if params.Len() == 0 {
		return req, diags
	}

	opts := params.List()[0]

	for k, v := range opts.(map[string]interface{}) {

		if k == "inclusion" {
			req = req.Inclusion(v.(string))
			continue
		}
		if k == "destination_type" {
			req = req.DestinationType(v.(string))
			continue
		}
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
func setAMQAttributesToResourceData(d *schema.ResourceData, amqitem map[string]interface{}) error {
	attributes := getAMQCoreAttributes()
	if amqitem != nil {
		for _, attr := range attributes {
			if err := d.Set(attr, amqitem[attr]); err != nil {
				return fmt.Errorf("unable to set AMQ attribute %s\n details: %s", attr, err)
			}
		}
	}
	return nil
}

// flattens a list of Anypoint MQs
func flattenAMQsData(queues *[]amq.Queue) []interface{} {
	if queues != nil && len(*queues) > 0 {
		res := make([]interface{}, len(*queues))
		for i, q := range *queues {
			res[i] = flattenAMQData(&q)
		}
		return res
	}

	return make([]interface{}, 0)
}

// flattens and maps a given Anypoint MQ object
func flattenAMQData(queue *amq.Queue) map[string]interface{} {
	if queue != nil {
		item := make(map[string]interface{})

		item["queue_id"] = queue.GetQueueId()
		item["default_ttl"] = queue.GetDefaultTtl()
		item["default_lock_ttl"] = queue.GetDefaultLockTtl()
		item["default_delivery_delay"] = queue.GetDefaultDeliveryDelay()
		item["type"] = queue.GetType()
		item["encrypted"] = queue.GetEncrypted()
		if v, ok := queue.GetDeadLetterQueueIdOk(); ok {
			item["dead_letter_queue_id"] = v
		}
		if v, ok := queue.GetMaxDeliveriesOk(); ok {
			item["max_deliveries"] = v
		}
		item["fifo"] = queue.GetFifo()

		return item
	}

	return nil
}

func getAMQCoreAttributes() []string {
	attributes := [...]string{
		"queue_id", "default_ttl", "default_lock_ttl", "default_delivery_delay", "type",
		"encrypted", "dead_letter_queue_id", "max_deliveries", "fifo",
	}
	return attributes[:]
}
