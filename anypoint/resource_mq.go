package anypoint

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	amq "github.com/mulesoft-consulting/cloudhub-client-go/mq"
)

func resourceMQ() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMQCreate,
		ReadContext:   resourceMQRead,
		UpdateContext: resourceMQUpdate,
		DeleteContext: resourceMQDelete,
		Schema: map[string]*schema.Schema{
			"defaultttl": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"defaultlockttl": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"region_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"queue_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"env_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceMQCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	envid := d.Get("env_id").(string)
	orgid := d.Get("org_id").(string)
	regionid := d.Get("region_id").(string)
	queueid := d.Get("queue_id").(string)

	authctx := getMQAuthCtx(ctx, &pco)

	body := newMQPutBody(d)

	//request mq creation
	_, httpr, err := pco.amqclient.DefaultApi.V1OrganizationsOrgIdEnvironmentsEnvIdRegionsRegionIdDestinationsQueuesQueueIdPut(authctx, orgid, envid, regionid, queueid).QueueOptional(*body).Execute()
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
			Summary:  "Unable to Create MQ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(queueid)

	resourceMQRead(ctx, d, m)

	return diags
}

func resourceMQRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	//queueid := d.Id()
	envid := d.Get("env_id").(string)
	orgid := d.Get("org_id").(string)
	regionid := d.Get("region_id").(string)
	queueid := d.Get("queue_id").(string)

	authctx := getMQAuthCtx(ctx, &pco)

	res, httpr, err := pco.amqclient.DefaultApi.V1OrganizationsOrgIdEnvironmentsEnvIdRegionsRegionIdDestinationsQueuesQueueIdGet(authctx, orgid, envid, regionid, queueid).Execute()
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
			Summary:  "Unable to Get Queue in RS qid-" + queueid + ",oid-" + orgid + ",rid-" + regionid + ",eid-" + envid,
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

	return diags
}

func resourceMQUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	envid := d.Get("env_id").(string)
	orgid := d.Get("org_id").(string)
	regionid := d.Get("region_id").(string)
	queueid := d.Get("queue_id").(string)

	authctx := getMQAuthCtx(ctx, &pco)

	if d.HasChanges(getMQAttributes()...) {
		body := newMQPatchBody(d)
		//request mq creation
		_, httpr, err := pco.amqclient.DefaultApi.V1OrganizationsOrgIdEnvironmentsEnvIdRegionsRegionIdDestinationsQueuesQueueIdPatch(authctx, orgid, envid, regionid, queueid).QueueOptional(*body).Execute()
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
				Summary:  "Unable to Update Queue",
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceMQRead(ctx, d, m)
}

func resourceMQDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	envid := d.Get("env_id").(string)
	orgid := d.Get("org_id").(string)
	regionid := d.Get("region_id").(string)
	queueid := d.Get("queue_id").(string)

	authctx := getMQAuthCtx(ctx, &pco)

	httpr, err := pco.amqclient.DefaultApi.V1OrganizationsOrgIdEnvironmentsEnvIdRegionsRegionIdDestinationsQueuesQueueIdDelete(authctx, orgid, envid, regionid, queueid).Execute()
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
			Summary:  "Unable to Delete Queue",
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

/*
 * Creates a new Queue struct from the resource data schema
 */
func newMQPutBody(d *schema.ResourceData) *amq.QueueOptional {
	body := amq.NewQueueOptionalWithDefaults()
	//body.SetDefaultTtl(d.Get("defaultttl").(int32))
	//body.SetDefaultLockTtl(d.Get("defaultlockttl").(int32))
	body.SetDefaultTtl(604800000)
	body.SetDefaultLockTtl(120000)
	body.SetType(d.Get("type").(string))
	body.SetEncrypted(d.Get("encrypted").(bool))
	return body
}

func newMQPatchBody(d *schema.ResourceData) *amq.QueueOptional {
	body := amq.NewQueueOptionalWithDefaults()
	//body.SetDefaultTtl(d.Get("defaultttl").(int32))
	//body.SetDefaultLockTtl(d.Get("defaultlockttl").(int32))
	body.SetDefaultTtl(604800000)
	body.SetDefaultLockTtl(120000)
	body.SetType(d.Get("type").(string))
	body.SetEncrypted(d.Get("encrypted").(bool))
	return body
}

/*
 * Returns authentication context (includes authorization header)
 */
func getMQAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	return context.WithValue(ctx, amq.ContextAccessToken, pco.access_token)
}
