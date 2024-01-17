package anypoint

import (
	"context"
	"io"
	"maps"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secretgroup "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup"
)

func dataSourceSecretGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretGroupsRead,
		Description: `
		Query all or part of available secret groups for a given organization and environment.
		`,
		Schema: map[string]*schema.Schema{
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
			"params": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The search parameters. Should only provide one occurrence of the block.",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"downloadable": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Filter and fetch list of secret groups based on value of 'downloadable' flag",
						},
					},
				},
			},
			"secretgroups": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List secret groups result of the query",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Id assigned to this secret group",
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
				},
			},
		},
	}
}

func dataSourceSecretGroupsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//init vars
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	searchOpts := d.Get("params").(*schema.Set)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	authctx := getSecretGroupAuthCtx(ctx, &pco)
	req := pco.secretgroupclient.DefaultApi.GetEnvSecretGroups(authctx, orgid, envid)
	req, errDiags := parseSecretGroupsSearchOpts(req, searchOpts)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}
	//execut request
	res, httpr, err := req.Execute()
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
			Summary:  "Unable to get secret groups for org " + orgid + " and env " + envid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	data := flattenSecretGroupCollectionResult(res)
	if err := d.Set("secretgroups", data); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set result for secret groups query in org " + orgid + " and env " + envid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func flattenSecretGroupCollectionResult(secretgroups []secretgroup.SecretGroup) []interface{} {
	if len(secretgroups) > 0 {
		res := make([]interface{}, len(secretgroups))
		for i, sg := range secretgroups {
			res[i] = flattenSecretGroupResult(&sg)
		}
		return res
	}
	return make([]interface{}, 0)
}

func flattenSecretGroupResult(secretgroup *secretgroup.SecretGroup) map[string]interface{} {
	item := make(map[string]interface{})
	if secretgroup == nil {
		return item
	}
	if val, ok := secretgroup.GetDownloadableOk(); ok {
		item["downloadable"] = *val
	}
	if val, ok := secretgroup.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := secretgroup.GetMetaOk(); ok {
		maps.Copy(item, flattenSecretGroupMeta(val))
	}

	return item
}

func flattenSecretGroupMeta(meta *secretgroup.Meta) map[string]interface{} {
	item := make(map[string]interface{})
	if meta == nil {
		return item
	}
	if val, ok := meta.GetIdOk(); ok {
		item["id"] = *val
	}
	if val, ok := meta.GetCreatedAtOk(); ok {
		item["created_at"] = *val
	}
	if val, ok := meta.GetModifiedAtOk(); ok {
		item["modified_at"] = *val
	}
	if val, ok := meta.GetModifiedByOk(); ok {
		item["modified_by"] = *val
	}
	if val, ok := meta.GetLockedOk(); ok {
		item["locked"] = *val
	}
	if val, ok := meta.GetLockedByOk(); ok {
		item["locked_by"] = *val
	}
	if val, ok := meta.GetCurrentStateOk(); ok {
		item["current_state"] = *val
	}
	return item
}

/*
Parses the secret group search options in order to check if the required search parameters are set correctly.
Appends the parameters to the given request
*/
func parseSecretGroupsSearchOpts(req secretgroup.DefaultApiGetEnvSecretGroupsRequest, params *schema.Set) (secretgroup.DefaultApiGetEnvSecretGroupsRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if params.Len() == 0 {
		return req, diags
	}
	opts := params.List()[0]
	for k, v := range opts.(map[string]interface{}) {
		if k == "downloadable" {
			req = req.Downloadable(v.(bool))
			continue
		}
	}
	return req, diags
}
