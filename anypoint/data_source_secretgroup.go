package anypoint

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSecretGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretGroupRead,
		Description: `
		Query a specific secret-group in a given organization and environment.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Id assigned to this secret group",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization id where the secret group instance is defined.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id where the secret group instance is defined.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the secret group",
			},
			"downloadable": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the secret group is downloadable or not",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time at which this secret group was created",
			},
			"modified_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time at which this secret group was last modified",
			},
			"modified_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Username of the anypoint platform user who modified this secret group",
			},
			"locked": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether this secret group is currently locked",
			},
			"locked_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Username of the anypoint platform user who currently holds the lock on this secret group. This is present only if 'locked' is set to true.",
			},
			"current_state": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `
				The state the secret group is currently in. If the state is anything but 'Clear', only the corresponding operation is allowed on the secret group, as follows.

					* If currentState="Finishing", only operation allowed - delete lock with action = finish.
					* If currentState="Cancelling", only operation allowed - delete lock with action = cancel.
					* If currentState="Deleting", only operation allowed - delete secret group.
				`,
			},
		},
	}
}

func dataSourceSecretGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//init vars
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	id := d.Get("id").(string)
	authctx := getSecretGroupAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.secretgroupclient.DefaultApi.GetSecretGroup(authctx, orgid, envid, id).Execute()
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
			Summary:  "Unable to get secret group " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process response data
	data := flattenSecretGroupResult(res)
	if err := setSecretGroupAttributesToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set secret group details attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(*res.GetMeta().Id)

	return diags
}

func setSecretGroupAttributesToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getSecretGroupAttributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set secret group attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getSecretGroupAttributes() []string {
	attributes := [...]string{
		"name", "downloadable", "created_at", "modified_at",
		"modified_by", "locked", "locked_by", "current_state",
	}
	return attributes[:]
}
