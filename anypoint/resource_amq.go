package anypoint

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	amq "github.com/mulesoft-consulting/anypoint-client-go/amq"
)

func resourceAMQ() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAMQCreate,
		ReadContext:   resourceAMQRead,
		UpdateContext: resourceAMQUpdate,
		DeleteContext: resourceAMQDelete,
		Description: `
		Creates an ` + "`" + `Anypoint MQ` + "`" + ` in your ` + "`" + `region` + "`" + `.
		`,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The last time this resource has been updated locally.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique id of this Anypoint MQ generated by the provider composed of {orgId}_{envId}_{regionId}_{queueId}.",
			},
			"queue_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The unique id of this Anypoint MQ.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organization id where the Anypoint MQ is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The environment id where the Anypoint MQ is defined.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
			"default_ttl": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The default TTL applied to messages in milliseconds.",
			},
			"default_lock_ttl": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The default time to live of the created locks in milliseconds.",
			},
			"default_delivery_delay": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The default delivery delay in seconds.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type is always queue.",
			},
			"encrypted": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "To encrypt the queue.",
			},
			"dead_letter_queue_id": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"max_deliveries"},
				Description:  "The queue Id of the dead letter queue to bind to this queue. A FIFO DLQ only works with FIFO queues.",
			},
			"max_deliveries": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"dead_letter_queue_id"},
				Description:  "The maximum number of attempts after which the message will be routed to DLQ. This field can only be used when dead_letter_queue_id attribute is present.",
			},
			"fifo": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Whether to make this queue a FIFO.",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceAMQCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	regionid := d.Get("region_id").(string)
	queueid := d.Get("queue_id").(string)
	authctx := getAMQAuthCtx(ctx, &pco)
	body := newAMQCreateBody(d)

	//request resource creation
	_, httpr, err := pco.amqclient.DefaultApi.CreateAMQ(authctx, orgid, envid, regionid, queueid).QueueBody(*body).Execute()
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
			Summary:  "Unable to create AMQ ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(ComposeResourceId([]string{orgid, envid, regionid, queueid}))

	return resourceAMQRead(ctx, d, m)
}

func resourceAMQRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, envid, regionid, queueid := decomposeAMQId(d)
	authctx := getAMQAuthCtx(ctx, &pco)

	//request resource
	res, httpr, err := pco.amqclient.DefaultApi.GetAMQ(authctx, orgid, envid, regionid, queueid).Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get AMQ " + d.Id(),
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	//process data
	queue := flattenAMQData(&res)
	//save in data source schema
	if err := setAMQAttributesToResourceData(d, queue); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set AMQ " + d.Id(),
			Detail:   err.Error(),
		})
		return diags
	}
	// setting resource id components for import purposes
	d.Set("org_id", orgid)
	d.Set("env_id", envid)
	d.Set("region_id", regionid)
	d.Set("queue_id", queueid)

	return diags
}

func resourceAMQUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, envid, regionid, queueid := decomposeAMQId(d)
	authctx := getAMQAuthCtx(ctx, &pco)

	if d.HasChanges(getAMQPatchWatchAttributes()...) {
		body := newAMQCreateBody(d)
		//request user creation
		_, httpr, err := pco.amqclient.DefaultApi.UpdateAMQ(authctx, orgid, envid, regionid, queueid).QueueBody(*body).Execute()
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
				Summary:  "Unable to patch AMQ " + d.Id(),
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}
	return resourceAMQRead(ctx, d, m)
}

func resourceAMQDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, envid, regionid, queueid := decomposeAMQId(d)
	authctx := getAMQAuthCtx(ctx, &pco)

	httpr, err := pco.amqclient.DefaultApi.DeleteAMQ(authctx, orgid, envid, regionid, queueid).Execute()
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
			Summary:  "Unable to delete AMQ " + d.Id(),
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

// Creates AMQ body
func newAMQCreateBody(d *schema.ResourceData) *amq.QueueBody {
	body := new(amq.QueueBody)

	body.SetType("queue")
	if defaultTtl := d.Get("default_ttl"); defaultTtl != nil {
		body.SetDefaultTtl(int32(defaultTtl.(int)))
	}
	if defaultLockTtl := d.Get("default_lock_ttl"); defaultLockTtl != nil {
		body.SetDefaultLockTtl(int32(defaultLockTtl.(int)))
	}
	if encrypted := d.Get("encrypted"); encrypted != nil {
		body.SetEncrypted(encrypted.(bool))
	}
	if defaultDeliveryDelay := d.Get("default_delivery_delay"); defaultDeliveryDelay != nil {
		body.SetDefaultDeliveryDelay(int32(defaultDeliveryDelay.(int)))
	}
	if deadLetterQueueId := d.Get("dead_letter_queue_id"); deadLetterQueueId != nil {
		val := deadLetterQueueId.(string)
		if len(val) > 0 {
			body.SetDeadLetterQueueId(val)
		} else {
			body.UnsetDeadLetterQueueId()
		}
	}
	if maxDeliveries := d.Get("max_deliveries"); maxDeliveries != nil {
		val := int32(maxDeliveries.(int))
		if val > 0 {
			body.SetMaxDeliveries(val)
		} else {
			body.UnsetDeadLetterQueueId()
		}
	}
	if fifo := d.Get("fifo"); fifo != nil {
		body.SetFifo(fifo.(bool))
	}

	return body
}

func decomposeAMQId(d *schema.ResourceData) (string, string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2], s[3]
}

/*
List of attributes that requires patching the team
*/
func getAMQPatchWatchAttributes() []string {
	attributes := [...]string{
		"default_ttl", "default_lock_ttl", "encrypted", "default_delivery_delay",
		"dead_letter_queue_id", "max_deliveries",
	}
	return attributes[:]
}

/*
 * Returns authentication context (includes authorization header)
 */
func getAMQAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, amq.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, amq.ContextServerIndex, pco.server_index)
}
