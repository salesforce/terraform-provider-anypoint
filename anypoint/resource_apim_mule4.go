package anypoint

import (
	"context"
	"fmt"
	"io"
	"maps"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iancoleman/strcase"
	apim "github.com/mulesoft-anypoint/anypoint-client-go/apim"
)

const APIM_MULE4_TECHNOLOGY = "mule4"

func resourceApimMule4() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApimMule4Create,
		ReadContext:   resourceApimMule4Read,
		UpdateContext: resourceApimMule4Update,
		DeleteContext: resourceApimMule4Delete,
		Description: `
		Create an API Manager Instance of type Mule4.
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
				Description: "The Instance's unique id",
			},
			"audit": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The instance's auditing data",
			},
			"master_organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The root business group id",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organization id where the api manager instance is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The environment id where the api manager instance is defined.",
			},
			"apim_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The api manager instance id.",
			},
			"instance_label": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "The instance's label.",
			},
			"asset_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The API specification's business group id",
			},
			"asset_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The API specification's asset id in exchange",
			},
			"asset_version": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The API specification's version number in exchange",
			},
			"product_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's asset major version number ",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the instance",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of tags",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"order": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The order of this instance in the API Manager instances list",
			},
			"provider_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "The client identity provider's id to use for this instance",
			},
			"deprecated": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "True if the instance is deprecated",
			},
			"last_active_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date of last activity for this instance",
			},
			"is_public": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "If this API is Public",
			},
			"technology": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of API Manager instance. Always equals to 'mule4'",
			},
			"endpoint_uri": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The endpoint URI of this instance API",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
			},
			"endpoint_audit": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The instance's endpoint auditing data",
			},
			"endpoint_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The instance's endpoint id",
			},
			"endpoint_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The endpoint's specification type",
			},
			"endpoint_proxy_uri": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Endpoint's Proxy URI",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
			},
			"endpoint_deployment_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "HY",
				Description: "Endpoint's deployment type",
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice(
						[]string{"CH", "HY", "RF", "SM", "CH2"},
						false,
					),
				),
			},
			// "routing": {
			// 	Type:        schema.TypeList,
			// 	Computed:    true,
			// 	Description: "The instance's routing mapping. Does not exist for mule4 instances",
			// },
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The API Instance status",
			},
			"autodiscovery_instance_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance's discovery name",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceApimMule4Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// init variables
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	authctx := getApimAuthCtx(ctx, &pco)
	// init request body
	body := newApimMule4PostBody(d)
	// execute post request
	res, httpr, err := pco.apimclient.DefaultApi.PostApimInstance(authctx, orgid, envid).ApimInstancePostBody(*body).Execute()
	defer httpr.Body.Close()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create api manager mule4 for org " + orgid + " and env " + envid,
			Detail:   details,
		})
		return diags
	}

	//update ids following the creation
	id := res.GetId()
	d.SetId(ComposeResourceId([]string{orgid, envid, strconv.Itoa(int(id))}))
	d.Set("apim_id", id)

	//perform read
	diags = append(diags, resourceApimMule4Read(ctx, d, m)...)
	return diags
}

// refresh the state of the flex gateway instance
func resourceApimMule4Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, envid, id := decomposeApimMule4Id(d)
	authctx := getApimAuthCtx(ctx, &pco)

	res, httpr, err := pco.apimclient.DefaultApi.GetApimInstanceDetails(authctx, orgid, envid, id).Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get API manager's mule4 instance " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	details := flattenApimInstanceDetails(&res)
	if err := setApimMule4AttributesToResourceData(d, details); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set API manager's mule4 instance " + id + " details attributes",
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

// updates the whole apim instance in case of changes
func resourceApimMule4Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, envid, apimid := decomposeApimMule4Id(d)

	if d.HasChanges(getApimMule4UpdatableAttributes()...) {
		body := newApimMule4PatchBody(d)
		authctx := getApimAuthCtx(ctx, &pco)
		_, httpr, err := pco.apimclient.DefaultApi.PatchApimInstance(authctx, orgid, envid, apimid).Body(body).Execute()
		if err != nil {
			var details error
			if httpr != nil {
				b, _ := io.ReadAll(httpr.Body)
				details = fmt.Errorf(string(b))
			} else {
				details = err
			}
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update api manager instance " + apimid,
				Detail:   details.Error(),
			})
		}
		defer httpr.Body.Close()
	}

	diags = append(diags, resourceApimMule4Read(ctx, d, m)...)
	return diags
}

// deletes the api manager instnace mule4
func resourceApimMule4Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, envid, id := decomposeApimMule4Id(d)
	authctx := getApimAuthCtx(ctx, &pco)

	httpr, err := pco.apimclient.DefaultApi.DeleteApimInstance(authctx, orgid, envid, id).Execute()
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
			Summary:  "Unable to Delete API Manager's Mule4 Instance",
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

func newApimMule4PostBody(d *schema.ResourceData) *apim.ApimInstancePostBody {
	body := apim.NewApimInstancePostBody()
	endpoint := newApimFlexGatewayEndpointPostBody(d)
	spec := newApimFlexGatewaySpecPostBody(d)

	if val, ok := d.GetOk("instance_label"); ok {
		body.SetInstanceLabel(val.(string))
	} else {
		body.SetInstanceLabelNil()
	}
	body.SetTechnology(FLEX_GATEWAY_TECHNOLOGY)
	body.SetEndpoint(*endpoint)
	body.SetDeploymentNil()
	body.SetSpec(*spec)

	return body
}

// creates patch body depending on the changes occured on the updatable attributes
func newApimMule4PatchBody(d *schema.ResourceData) map[string]interface{} {
	body := make(map[string]interface{})
	attributes := FilterStrList(getApimMule4UpdatableAttributes(), func(s string) bool {
		return !strings.HasPrefix(s, "endpoint")
	})
	for _, attr := range attributes {
		if d.HasChange(attr) {
			body[strcase.ToCamel(attr)] = d.Get(attr)
		}
	}
	maps.Copy(body, newPatchBodyMap4FlattenedAttr("endpoint", d))
	return body
}

func setApimMule4AttributesToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getApimMule4DetailsAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set api manager instance attribute %s\n details: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func decomposeApimMule4Id(d *schema.ResourceData) (string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2]
}

func getApimMule4DetailsAttributes() []string {
	attributes := [...]string{
		"org_id", "env_id", "audit", "master_organization_id", "instance_label", "asset_group_id",
		"asset_id", "asset_version", "product_version", "description", "tags", "order", "provider_id",
		"deprecated", "last_active_date", "endpoint_uri", "is_public", "technology", "endpoint_audit",
		"endpoint_id", "endpoint_type", "endpoint_proxy_uri", "endpoint_deployment_type",
		"status", "autodiscovery_instance_name",
	}
	return attributes[:]
}

func getApimMule4UpdatableAttributes() []string {
	attributes := [...]string{
		"instance_label", "description", "tags", "provider_id",
		"deprecated", "endpoint_uri", "endpoint_proxy_uri", "endpoint_deployment_type",
	}
	return attributes[:]
}
