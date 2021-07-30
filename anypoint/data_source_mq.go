package anypoint

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	amq "github.com/mulesoft-consulting/cloudhub-client-go/mq"
)

func dataSourceMQ() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMQRead,
		Schema: map[string]*schema.Schema{
			"queue_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"fifo": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"defaultttl": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"defaultlockttl": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"defaultdeliverydelay": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceMQRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	envid := d.Get("env_id").(string)
	orgid := d.Get("org_id").(string)
	regionid := d.Get("region_id").(string)
	queueid := d.Get("queue_id").(string)

	if envid == "" || orgid == "" || regionid == "" || queueid == "" {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "ENV id (env_id), Organization ID (org_id), Region id (region_id), Queue id (queue_id) are required",
			Detail:   "ENV id (env_id), Organization ID (org_id), Region id (region_id), Queue id (queue_id) must be provided",
		})
		return diags
	}

	authctx := getMQAuthCtx(ctx, &pco)

	//request mq
	res, httpr, err := pco.amqclient.DefaultApi.V1OrganizationsOrgIdEnvironmentsEnvIdRegionsRegionIdDestinationsQueuesQueueIdGet(authctx, orgid, envid, regionid, queueid).Execute()
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
			Summary:  "Unable to Get Queue in DS",
			Detail:   details,
		})
		return diags
	}
	//process data
	mqinstance := flattenMQData(&res)
	//save in data source schema
	if err := setMQAttributesToResourceData(d, mqinstance); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set MQ",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(queueid)

	return diags
}

/*
* Copies the given mq instance into the given resource data
* @param d *schema.ResourceData the resource data schema
* @param mqitem map[string]interface{} the mq instance
 */
func setMQAttributesToResourceData(d *schema.ResourceData, mqitem map[string]interface{}) error {
	attributes := getMQAttributes()
	if mqitem != nil {
		for _, attr := range attributes {
			if err := d.Set(attr, mqitem[attr]); err != nil {
				return fmt.Errorf("unable to set MQ attribute %s\n details: %s", attr, err)
			}
		}
	}
	return nil
}

/*
* Transforms a mq.Queue object to the dataSourceMQ schema
* @param mqitem *amq.Queue the mq struct
* @return the mq mapped struct
 */
func flattenMQData(mqitem *amq.Queue) map[string]interface{} {
	if mqitem != nil {
		item := make(map[string]interface{})

		item["queue_id"] = mqitem.GetQueueId()
		item["fifo"] = mqitem.GetFifo()
		item["defaultttl"] = mqitem.GetDefaultTtl()
		item["defaultlockttl"] = mqitem.GetDefaultLockTtl()
		item["defaultdeliverydelay"] = mqitem.GetDefaultDeliveryDelay()
		item["type"] = mqitem.GetType()
		item["encrypted"] = mqitem.GetEncrypted()
		return item
	}

	return nil
}

func getMQAttributes() []string {
	attributes := [...]string{
		"encrypted", "type", "queue_id", "defaultttl",
		"defaultlockttl",
	}
	return attributes[:]
}
